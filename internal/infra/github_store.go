package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/utils"
	"golang.org/x/oauth2"
)

const (
	// IndexGistDescription is the description for the index gist
	IndexGistDescription = "pv-prompts-index"
	// IndexFileName is the filename for the index data in the gist
	IndexFileName = "index.json"
)

// Custom error types for better user experience
var (
	ErrNoIndex    = fmt.Errorf("no prompt index found - this appears to be your first time using pv")
	ErrEmptyIndex = fmt.Errorf("no prompts found in your collection")
)

// GitHubStore implements the Store interface using GitHub Gists
type GitHubStore struct {
	client      *github.Client
	configStore config.Store
	indexGistID string
}

// NewGitHubStore creates a new GitHubStore instance
func NewGitHubStore(configStore config.Store) Store {
	return &GitHubStore{
		configStore: configStore,
	}
}

// ensureInitialized ensures the GitHub client is initialized
func (g *GitHubStore) ensureInitialized() error {
	if g.client != nil {
		return nil
	}

	token, err := g.configStore.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get GitHub token: %w", err)
	}

	if token == "" {
		return fmt.Errorf("GitHub token is not configured")
	}

	// Create GitHub client with authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	g.client = github.NewClient(tc)

	// Initialize index gist
	if err := g.initializeIndex(); err != nil {
		return fmt.Errorf("failed to initialize index: %w", err)
	}

	return nil
}

// initializeIndex finds or creates the index gist
func (g *GitHubStore) initializeIndex() error {
	ctx := context.Background()

	// First, try to find existing index gist
	gists, _, err := g.client.Gists.List(ctx, "", &github.GistListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list gists: %w", err)
	}

	for _, gist := range gists {
		if gist.GetDescription() == IndexGistDescription {
			g.indexGistID = gist.GetID()
			return nil
		}
	}

	// Create new index gist if not found
	emptyIndex := model.Index{
		Prompts:     []model.IndexedPrompt{},
		Exports:     []model.IndexedPrompt{},
		LastUpdated: time.Now(),
	}

	indexContent, err := json.MarshalIndent(emptyIndex, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal empty index: %w", err)
	}

	gist := &github.Gist{
		Description: github.Ptr(IndexGistDescription),
		Public:      github.Ptr(false),
		Files: map[github.GistFilename]github.GistFile{
			IndexFileName: {
				Content: github.Ptr(string(indexContent)),
			},
		},
	}

	createdGist, _, err := g.client.Gists.Create(ctx, gist)
	if err != nil {
		return fmt.Errorf("failed to create index gist: %w", err)
	}

	g.indexGistID = createdGist.GetID()
	return nil
}

// loadIndex loads the current index from the index gist
func (g *GitHubStore) loadIndex() (*model.Index, error) {
	ctx := context.Background()

	gist, _, err := g.client.Gists.Get(ctx, g.indexGistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index gist: %w", err)
	}

	indexFile, exists := gist.Files[IndexFileName]
	if !exists {
		return nil, fmt.Errorf("index file not found in gist")
	}

	var index model.Index
	if err := json.Unmarshal([]byte(indexFile.GetContent()), &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	return &index, nil
}

// GetRawIndexContent retrieves the raw index.json content from GitHub
func (g *GitHubStore) GetRawIndexContent() (string, error) {
	if err := g.ensureInitialized(); err != nil {
		return "", err
	}

	ctx := context.Background()

	gist, _, err := g.client.Gists.Get(ctx, g.indexGistID)
	if err != nil {
		return "", fmt.Errorf("failed to get index gist: %w", err)
	}

	indexFile, exists := gist.Files[IndexFileName]
	if !exists {
		return "", fmt.Errorf("index file not found in gist")
	}

	return indexFile.GetContent(), nil
}

// saveIndex saves the index to the index gist
func (g *GitHubStore) saveIndex(index *model.Index) error {
	ctx := context.Background()

	index.LastUpdated = time.Now()

	indexContent, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	gist := &github.Gist{
		Files: map[github.GistFilename]github.GistFile{
			IndexFileName: {
				Content: github.Ptr(string(indexContent)),
			},
		},
	}

	_, _, err = g.client.Gists.Edit(ctx, g.indexGistID, gist)
	if err != nil {
		return fmt.Errorf("failed to update index gist: %w", err)
	}

	return nil
}

// List returns all prompts from the index
func (g *GitHubStore) List() ([]model.Prompt, error) {
	if err := g.ensureInitialized(); err != nil {
		return nil, err
	}

	index, err := g.loadIndex()
	if err != nil {
		// Check if this is because no index gist exists (404 error specifically)
		if strings.Contains(err.Error(), "failed to get index gist") && strings.Contains(err.Error(), "404") {
			return nil, ErrNoIndex
		}
		return nil, err
	}

	// Check if index is empty
	if len(index.Prompts) == 0 {
		return nil, ErrEmptyIndex
	}

	var prompts []model.Prompt
	ctx := context.Background()

	for _, indexedPrompt := range index.Prompts {
		// Extract gist ID from URL
		gistID := utils.ExtractGistIDFromURL(indexedPrompt.GistURL)
		if gistID == "" {
			continue
		}

		// Get gist details
		gist, _, err := g.client.Gists.Get(ctx, gistID)
		if err != nil {
			continue // Skip if gist is not accessible
		}

		// Find the prompt file
		_, exists := gist.Files[github.GistFilename(indexedPrompt.FilePath)]
		if !exists {
			continue
		}

		prompt := model.Prompt{
			ID:      gistID,
			Name:    indexedPrompt.Name,   // 直接使用索引中的 name
			Author:  indexedPrompt.Author, // 直接使用索引中的 author
			GistURL: indexedPrompt.GistURL,
		}

		prompts = append(prompts, prompt)
	}

	// If we processed some indexed prompts but none were accessible, treat as empty
	if len(prompts) == 0 {
		return nil, ErrEmptyIndex
	}

	return prompts, nil
}

// findExistingPrompt searches for an existing prompt with the same name and author
func (g *GitHubStore) findExistingPrompt(name, author string) (*model.Prompt, error) {
	if err := g.ensureInitialized(); err != nil {
		return nil, err
	}

	allPrompts, err := g.List()
	if err != nil {
		// If we get ErrNoIndex or ErrEmptyIndex, that's fine - no existing prompts
		if err == ErrNoIndex || err == ErrEmptyIndex {
			return nil, nil
		}
		return nil, err
	}
	fmt.Println("allPrompts:", allPrompts)

	for _, prompt := range allPrompts {
		fmt.Println("name equals:", prompt.Name, name, prompt.Name == name)
		fmt.Println("author equals:", prompt.Author, author, prompt.Author == author)
		if prompt.Name == name && prompt.Author == author {
			return &prompt, nil
		}
	}

	return nil, nil
}

// FindExistingPromptByURL 根据 gist URL 查找已存在的提示词
func (g *GitHubStore) FindExistingPromptByURL(gistURL string) (*model.Prompt, error) {
	if err := g.ensureInitialized(); err != nil {
		return nil, err
	}

	allPrompts, err := g.List()
	if err != nil {
		// If we get ErrNoIndex or ErrEmptyIndex, that's fine - no existing prompts
		if err == ErrNoIndex || err == ErrEmptyIndex {
			return nil, nil
		}
		return nil, err
	}

	for _, prompt := range allPrompts {
		if prompt.GistURL == gistURL {
			return &prompt, nil
		}
	}

	return nil, nil
}

// Add creates a new prompt gist and updates the index

func (g *GitHubStore) Add(prompt model.Prompt) error {
	if err := g.ensureInitialized(); err != nil {
		return err
	}

	// Check if a prompt with the same name and author already exists
	existingPrompt, err := g.findExistingPrompt(prompt.Name, prompt.Author)
	if err != nil {
		return fmt.Errorf("failed to check for existing prompt: %w", err)
	}

	fmt.Println(
		"existingPrompt",
		existingPrompt,
	)

	// If an existing prompt is found, update it instead of creating a new one
	if existingPrompt != nil {
		// Copy the new content and other updatable fields to the existing prompt
		existingPrompt.Content = prompt.Content
		existingPrompt.Description = prompt.Description
		existingPrompt.Tags = prompt.Tags
		existingPrompt.Version = prompt.Version

		return g.Update(*existingPrompt)
	}

	ctx := context.Background()

	// Create filename using YAML format from prompt name
	fileName := prompt.Name + ".yaml"

	// Build gist description with prompt metadata
	description := fmt.Sprintf("Prompt: %s", prompt.Name)
	if prompt.Description != "" {
		description = fmt.Sprintf("Prompt: %s - %s", prompt.Name, prompt.Description)
	}

	// Create new gist with actual prompt content
	gist := &github.Gist{
		Description: github.Ptr(description),
		Public:      github.Ptr(false),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(fileName): {
				Content: github.Ptr(prompt.Content),
			},
		},
	}

	createdGist, _, err := g.client.Gists.Create(ctx, gist)
	if err != nil {
		return fmt.Errorf("failed to create gist: %w", err)
	}

	// Update index
	index, err := g.loadIndex()
	if err != nil {
		return err
	}

	indexedPrompt := model.IndexedPrompt{
		GistURL:     createdGist.GetHTMLURL(),
		FilePath:    fileName,
		Author:      prompt.Author, // 使用 YAML 中的 author
		Name:        prompt.Name,   // 使用 YAML 中的 name
		LastUpdated: time.Now(),
	}

	index.Prompts = append(index.Prompts, indexedPrompt)

	return g.saveIndex(index)
}

// Delete removes a prompt by deleting its gist and updating the index
func (g *GitHubStore) Delete(keyword string) error {
	if err := g.ensureInitialized(); err != nil {
		return err
	}

	ctx := context.Background()

	index, err := g.loadIndex()
	if err != nil {
		return err
	}

	// Find matching prompt in index
	var toDelete []int
	for i, indexedPrompt := range index.Prompts {
		gistID := utils.ExtractGistIDFromURL(indexedPrompt.GistURL)
		if gistID == keyword || strings.Contains(indexedPrompt.FilePath, keyword) {
			// Delete the gist
			_, err := g.client.Gists.Delete(ctx, gistID)
			if err != nil {
				return fmt.Errorf("failed to delete gist %s: %w", gistID, err)
			}
			toDelete = append(toDelete, i)
		}
	}

	// Remove from index (iterate backwards to maintain indices)
	for i := len(toDelete) - 1; i >= 0; i-- {
		idx := toDelete[i]
		index.Prompts = append(index.Prompts[:idx], index.Prompts[idx+1:]...)
	}

	return g.saveIndex(index)
}

// Update modifies an existing prompt
func (g *GitHubStore) Update(prompt model.Prompt) error {
	if err := g.ensureInitialized(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get existing gist
	existingGist, _, err := g.client.Gists.Get(ctx, prompt.ID)
	if err != nil {
		return fmt.Errorf("failed to get gist: %w", err)
	}

	// Find the existing file name (should be prompt.Name + ".yaml")
	fileName := prompt.Name + ".yaml"
	var existingFileName string
	for filename := range existingGist.Files {
		// Try to find the existing file - it might have a different name if the prompt name changed
		if strings.HasSuffix(string(filename), ".yaml") {
			existingFileName = string(filename)
			break
		}
	}

	// If no existing file found, use the new filename
	if existingFileName == "" {
		existingFileName = fileName
	}

	// Build gist description with prompt metadata
	description := fmt.Sprintf("Prompt: %s", prompt.Name)
	if prompt.Description != "" {
		description = fmt.Sprintf("Prompt: %s - %s", prompt.Name, prompt.Description)
	}

	// Prepare the updated gist
	updatedGist := &github.Gist{
		Description: github.Ptr(description),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(fileName): {
				Content: github.Ptr(prompt.Content),
			},
		},
	}

	// If the filename changed, we need to delete the old file
	if existingFileName != fileName {
		updatedGist.Files[github.GistFilename(existingFileName)] = github.GistFile{
			Content: nil, // Setting content to nil deletes the file
		}
	}

	// Update the gist
	_, _, err = g.client.Gists.Edit(ctx, prompt.ID, updatedGist)
	if err != nil {
		return fmt.Errorf("failed to update gist: %w", err)
	}

	// Update index timestamp and filename if changed
	index, err := g.loadIndex()
	if err != nil {
		return err
	}

	for i, indexedPrompt := range index.Prompts {
		if utils.ExtractGistIDFromURL(indexedPrompt.GistURL) == prompt.ID {
			index.Prompts[i].LastUpdated = time.Now()
			index.Prompts[i].Author = prompt.Author // 更新 author
			index.Prompts[i].Name = prompt.Name     // 更新 name
			// Update filename in index if it changed
			if existingFileName != fileName {
				index.Prompts[i].FilePath = fileName
			}
			break
		}
	}

	return g.saveIndex(index)
}

// Get searches for prompts by keyword
func (g *GitHubStore) Get(keyword string) ([]model.Prompt, error) {
	if err := g.ensureInitialized(); err != nil {
		return nil, err
	}

	allPrompts, err := g.List()
	if err != nil {
		return nil, err
	}

	var matchingPrompts []model.Prompt
	keyword = strings.ToLower(keyword)

	for _, prompt := range allPrompts {
		if strings.Contains(strings.ToLower(prompt.Name), keyword) ||
			strings.Contains(strings.ToLower(prompt.Author), keyword) ||
			strings.Contains(strings.ToLower(prompt.ID), keyword) {
			matchingPrompts = append(matchingPrompts, prompt)
		}
	}

	return matchingPrompts, nil
}

// GetContent retrieves the actual content of a prompt from its GitHub Gist
func (g *GitHubStore) GetContent(gistID string) (string, error) {
	if err := g.ensureInitialized(); err != nil {
		return "", err
	}

	ctx := context.Background()

	// Get the gist
	gist, _, err := g.client.Gists.Get(ctx, gistID)
	if err != nil {
		return "", fmt.Errorf("failed to get gist %s: %w", gistID, err)
	}

	// Find the prompt file (should be a .yaml file)
	for filename, file := range gist.Files {
		if strings.HasSuffix(string(filename), ".yaml") {
			return file.GetContent(), nil
		}
	}

	return "", fmt.Errorf("no .yaml file found in gist %s", gistID)
}

// CreatePublicGist 创建新的公开 gist
func (g *GitHubStore) CreatePublicGist(prompt model.Prompt) (string, error) {
	// 构建 gist 文件内容
	filename := fmt.Sprintf("%s.yaml", prompt.Name)
	content := g.buildYAMLContent(prompt)

	public := true
	gist := &github.Gist{
		Description: &prompt.Description,
		Public:      &public,
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(filename): {
				Content: &content,
			},
		},
	}

	createdGist, _, err := g.client.Gists.Create(context.Background(), gist)
	if err != nil {
		return "", errors.NewShareError("创建公开 gist", "", err)
	}

	return createdGist.GetHTMLURL(), nil
}

// UpdateGist 更新现有 gist 的内容
func (g *GitHubStore) UpdateGist(gistURL string, prompt model.Prompt) error {
	gistID := g.extractGistID(gistURL)

	// 获取现有 gist
	existingGist, _, err := g.client.Gists.Get(context.Background(), gistID)
	if err != nil {
		return errors.NewShareError("获取现有 gist", gistURL, err)
	}

	// 构建新内容
	filename := fmt.Sprintf("%s.yaml", prompt.Name)
	content := g.buildYAMLContent(prompt)

	// 更新 gist
	gist := &github.Gist{
		Description: &prompt.Description,
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(filename): {
				Content: &content,
			},
		},
	}

	// 如果文件名改变了，删除旧文件
	for oldFilename := range existingGist.Files {
		if string(oldFilename) != filename {
			gist.Files[oldFilename] = github.GistFile{
				Content: nil, // 设置为 nil 删除文件
			}
		}
	}

	_, _, err = g.client.Gists.Edit(context.Background(), gistID, gist)
	if err != nil {
		return errors.NewShareError("更新 gist", gistURL, err)
	}

	return nil
}

// GetGistInfo 获取 gist 的基本信息
func (g *GitHubStore) GetGistInfo(gistURL string) (*GistInfo, error) {
	if err := g.ensureInitialized(); err != nil {
		return nil, err
	}

	gistID := g.extractGistID(gistURL)

	gist, resp, err := g.client.Gists.Get(context.Background(), gistID)
	if err != nil {
		// 处理权限和访问错误
		if resp != nil && resp.StatusCode == 404 {
			return &GistInfo{
				ID:        gistID,
				URL:       gistURL,
				IsPublic:  false,
				HasAccess: false,
			}, nil
		}
		return nil, errors.NewShareError("获取 gist 信息", gistURL, err)
	}

	// 安全获取 owner 信息
	var owner string
	if gistOwner := gist.GetOwner(); gistOwner != nil {
		owner = gistOwner.GetLogin()
	}

	return &GistInfo{
		ID:          gistID,
		URL:         gistURL,
		IsPublic:    gist.GetPublic(),
		HasAccess:   true,
		Description: gist.GetDescription(),
		Owner:       owner,
	}, nil
}

// AddExport 添加新的导出记录
func (g *GitHubStore) AddExport(prompt model.IndexedPrompt) error {
	err := g.ensureInitialized()
	if err != nil {
		return err
	}

	index, err := g.loadIndex()
	if err != nil {
		return err
	}

	index.Exports = append(index.Exports, prompt)
	return g.saveIndex(index)
}

// UpdateExport 更新现有的导出记录
func (g *GitHubStore) UpdateExport(prompt model.IndexedPrompt) error {
	err := g.ensureInitialized()
	if err != nil {
		return err
	}

	index, err := g.loadIndex()
	if err != nil {
		return err
	}

	for i, export := range index.Exports {
		if export.GistURL == prompt.GistURL {
			index.Exports[i] = prompt
			return g.saveIndex(index)
		}
	}

	// 如果没找到，则添加
	index.Exports = append(index.Exports, prompt)
	return g.saveIndex(index)
}

// GetExports 获取所有导出记录
func (g *GitHubStore) GetExports() ([]model.IndexedPrompt, error) {
	err := g.ensureInitialized()
	if err != nil {
		return nil, err
	}

	index, err := g.loadIndex()
	if err != nil {
		return nil, err
	}

	return index.Exports, nil
}

// buildYAMLContent 构建 YAML 内容字符串
func (g *GitHubStore) buildYAMLContent(prompt model.Prompt) string {
	// 如果 prompt.Content 包含完整的原始 YAML 内容，则直接使用
	// 这确保了 share 命令生成的 gist 内容与原始文件完全一致
	if strings.TrimSpace(prompt.Content) != "" {
		return prompt.Content
	}

	// 如果 Content 为空，则回退到构建 YAML 内容（兼容性处理）
	return fmt.Sprintf(`name: %s
author: %s
description: %s
tags: %v
version: %s
content: |
%s`,
		prompt.Name,
		prompt.Author,
		prompt.Description,
		prompt.Tags,
		prompt.Version,
		g.indentContent(prompt.Content, "  "))
}

// indentContent 为内容添加缩进
func (g *GitHubStore) indentContent(content string, indent string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// extractGistID 从 URL 中提取 gist ID
func (g *GitHubStore) extractGistID(gistURL string) string {
	parts := strings.Split(gistURL, "/")
	return parts[len(parts)-1]
}
