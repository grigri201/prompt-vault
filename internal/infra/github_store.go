package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/model"
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
		gistID := extractGistIDFromURL(indexedPrompt.GistURL)
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
			Name:    strings.TrimSuffix(indexedPrompt.FilePath, ".md"),
			Author:  gist.GetOwner().GetLogin(),
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

// Add creates a new prompt gist and updates the index
func (g *GitHubStore) Add(prompt model.Prompt) error {
	if err := g.ensureInitialized(); err != nil {
		return err
	}

	ctx := context.Background()

	// Create filename from prompt name
	fileName := prompt.Name + ".md"

	// Create new gist
	gist := &github.Gist{
		Description: github.Ptr(fmt.Sprintf("Prompt: %s", prompt.Name)),
		Public:      github.Ptr(false),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(fileName): {
				Content: github.Ptr("# " + prompt.Name + "\n\n<!-- Add your prompt content here -->"),
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
		gistID := extractGistIDFromURL(indexedPrompt.GistURL)
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
	gist, _, err := g.client.Gists.Get(ctx, prompt.ID)
	if err != nil {
		return fmt.Errorf("failed to get gist: %w", err)
	}

	// Update gist description
	gist.Description = github.Ptr(fmt.Sprintf("Prompt: %s", prompt.Name))

	// Update the gist
	_, _, err = g.client.Gists.Edit(ctx, prompt.ID, gist)
	if err != nil {
		return fmt.Errorf("failed to update gist: %w", err)
	}

	// Update index timestamp
	index, err := g.loadIndex()
	if err != nil {
		return err
	}

	for i, indexedPrompt := range index.Prompts {
		if extractGistIDFromURL(indexedPrompt.GistURL) == prompt.ID {
			index.Prompts[i].LastUpdated = time.Now()
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

// extractGistIDFromURL extracts the gist ID from a GitHub gist URL
func extractGistIDFromURL(url string) string {
	// GitHub gist URLs are in the format: https://gist.github.com/{user}/{gist_id}
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}
