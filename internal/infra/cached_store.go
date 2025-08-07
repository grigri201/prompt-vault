package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/utils"
)

// CachedStore is a decorator that adds caching capability to any Store implementation
// It implements the "remote-first, cache fallback" strategy as specified in requirements 2.1 and 5.2
type CachedStore struct {
	remote      Store         // The underlying remote store (typically GitHubStore)
	cache       *CacheManager // Local file cache manager
	config      config.Store  // Configuration store for settings
	forceRemote bool          // Flag to force remote operations (requirement 5.2)
}

// NewCachedStore creates a new CachedStore instance that wraps a remote Store with local caching
// The forceRemote parameter controls whether to bypass cache and always use remote operations
func NewCachedStore(remote Store, cacheManager *CacheManager, configStore config.Store, forceRemote bool) Store {
	return &CachedStore{
		remote:      remote,
		cache:       cacheManager,
		config:      configStore,
		forceRemote: forceRemote,
	}
}

// List retrieves all prompts using the remote-first strategy
// Per requirement 2.1: Try remote first, fallback to cache on failure
// Per requirement 5.2: If forceRemote is true, skip cache fallback
func (c *CachedStore) List() ([]model.Prompt, error) {
	// Always try remote first (remote-first strategy)
	prompts, err := c.remote.List()
	if err == nil {
		// Remote success: update cache and return remote data
		if cacheErr := c.updateCacheFromPrompts(prompts); cacheErr != nil {
			// Log cache update failure but don't fail the operation
			// The user still gets the correct data from remote
		}
		return prompts, nil
	}

	// Remote failed: check if we should fallback to cache
	if c.forceRemote {
		// forceRemote is true: don't use cache, return the remote error
		return nil, fmt.Errorf("remote operation failed and forceRemote is enabled: %w", err)
	}

	// Try to fallback to cache
	return c.listFromCache(err)
}

// Add creates a new prompt using the remote store and updates the cache
func (c *CachedStore) Add(prompt model.Prompt) error {
	// Add to remote store
	if err := c.remote.Add(prompt); err != nil {
		return err
	}

	// Update cache with the new prompt
	// Load current index, add the new prompt, and save
	index, err := c.cache.LoadIndex()
	if err != nil {
		// If cache doesn't exist, create a new one
		index = &model.Index{
			Prompts:     []model.IndexedPrompt{},
			LastUpdated: time.Now(),
		}
	}

	// Create indexed prompt entry
	indexedPrompt := model.IndexedPrompt{
		GistURL:     prompt.GistURL,
		FilePath:    fmt.Sprintf("%s.yaml", prompt.Name),
		Author:      prompt.Author,
		Name:        prompt.Name,
		LastUpdated: time.Now(),
	}

	// Add to index
	index.Prompts = append(index.Prompts, indexedPrompt)
	index.LastUpdated = time.Now()

	// Save updated index (ignore cache errors to not fail the operation)
	c.cache.SaveIndex(index)

	// Save prompt content to cache
	if prompt.ID != "" {
		c.cache.SaveContent(prompt.ID, prompt.Content)
	}

	return nil
}

// Delete removes a prompt using the remote store and updates the cache
func (c *CachedStore) Delete(keyword string) error {
	// Delete from remote store
	if err := c.remote.Delete(keyword); err != nil {
		return err
	}

	// Update cache by removing the deleted prompt
	index, err := c.cache.LoadIndex()
	if err != nil {
		// Cache doesn't exist or is corrupted, skip cache update
		return nil
	}

	// Remove matching prompts from cache index
	var updatedPrompts []model.IndexedPrompt
	for _, indexedPrompt := range index.Prompts {
		gistID := utils.ExtractGistIDFromURL(indexedPrompt.GistURL)
		// Keep prompts that don't match the deletion keyword
		if gistID != keyword && indexedPrompt.FilePath != keyword && indexedPrompt.Name != keyword {
			updatedPrompts = append(updatedPrompts, indexedPrompt)
		}
	}

	// Update and save index
	index.Prompts = updatedPrompts
	index.LastUpdated = time.Now()
	c.cache.SaveIndex(index)

	return nil
}

// Update modifies an existing prompt using the remote store and updates the cache
func (c *CachedStore) Update(prompt model.Prompt) error {
	// Update in remote store
	if err := c.remote.Update(prompt); err != nil {
		return err
	}

	// Update cache
	index, err := c.cache.LoadIndex()
	if err != nil {
		// Cache doesn't exist, skip cache update
		return nil
	}

	// Update the prompt in cache index
	for i, indexedPrompt := range index.Prompts {
		gistID := utils.ExtractGistIDFromURL(indexedPrompt.GistURL)
		if gistID == prompt.ID {
			index.Prompts[i].Author = prompt.Author
			index.Prompts[i].Name = prompt.Name
			index.Prompts[i].FilePath = fmt.Sprintf("%s.yaml", prompt.Name)
			index.Prompts[i].LastUpdated = time.Now()
			break
		}
	}

	// Save updated index and content
	index.LastUpdated = time.Now()
	c.cache.SaveIndex(index)
	c.cache.SaveContent(prompt.ID, prompt.Content)

	return nil
}

// Get searches for prompts using the remote-first strategy
func (c *CachedStore) Get(keyword string) ([]model.Prompt, error) {
	// Try remote first
	prompts, err := c.remote.Get(keyword)
	if err == nil {
		// Remote success: update cache and return results
		if cacheErr := c.updateCacheFromPrompts(prompts); cacheErr != nil {
			// Log cache update failure but don't fail the operation
		}
		return prompts, nil
	}

	// Remote failed: check if we should fallback to cache
	if c.forceRemote {
		return nil, fmt.Errorf("remote operation failed and forceRemote is enabled: %w", err)
	}

	// Fallback to cache search
	return c.searchCache(keyword, err)
}

// GetContent retrieves prompt content using the remote-first strategy
// This is the core method for content retrieval with caching
func (c *CachedStore) GetContent(gistID string) (string, error) {
	// Try remote first (remote-first strategy)
	content, err := c.remote.GetContent(gistID)
	if err == nil {
		// Remote success: cache the content and return it
		if cacheErr := c.cache.SaveContent(gistID, content); cacheErr != nil {
			// Log cache save failure but don't fail the operation
			// User still gets the correct content from remote
		}
		return content, nil
	}

	// Remote failed: check if we should fallback to cache
	if c.forceRemote {
		return "", fmt.Errorf("remote operation failed and forceRemote is enabled: %w", err)
	}

	// Try to load from cache
	cachedContent, cacheErr := c.cache.LoadContent(gistID)
	if cacheErr != nil {
		// Both remote and cache failed
		return "", fmt.Errorf("remote failed and no cache available: remote error: %w, cache error: %v", err, cacheErr)
	}

	// Return cached content
	return cachedContent, nil
}

// Helper methods

// listFromCache attempts to load prompts from local cache when remote fails
func (c *CachedStore) listFromCache(remoteErr error) ([]model.Prompt, error) {
	index, err := c.cache.LoadIndex()
	if err != nil {
		// Both remote and cache failed
		return nil, fmt.Errorf("remote failed and no cache available: remote error: %w, cache error: %v", remoteErr, err)
	}

	// Convert cached index to prompts
	var prompts []model.Prompt
	for _, indexedPrompt := range index.Prompts {
		gistID := utils.ExtractGistIDFromURL(indexedPrompt.GistURL)
		prompt := model.Prompt{
			ID:      gistID,
			Name:    indexedPrompt.Name,
			Author:  indexedPrompt.Author,
			GistURL: indexedPrompt.GistURL,
		}
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// searchCache searches for prompts in local cache by keyword
func (c *CachedStore) searchCache(keyword string, remoteErr error) ([]model.Prompt, error) {
	// First get all cached prompts
	allPrompts, err := c.listFromCache(remoteErr)
	if err != nil {
		return nil, err
	}

	// Filter by keyword (similar to GitHubStore.Get logic)
	var matchingPrompts []model.Prompt
	keywordLower := keyword
	for _, prompt := range allPrompts {
		if contains(prompt.Name, keywordLower) ||
			contains(prompt.Author, keywordLower) ||
			contains(prompt.ID, keywordLower) {
			matchingPrompts = append(matchingPrompts, prompt)
		}
	}

	return matchingPrompts, nil
}

// updateCacheFromPrompts updates the cache index with a list of prompts
func (c *CachedStore) updateCacheFromPrompts(prompts []model.Prompt) error {
	// Load existing cache to preserve timestamps where possible
	existingIndex, err := c.cache.LoadIndex()
	existingTimestamps := make(map[string]time.Time)
	if err == nil {
		// Build a map of existing timestamps by GistURL
		for _, existing := range existingIndex.Prompts {
			existingTimestamps[existing.GistURL] = existing.LastUpdated
		}
	}

	// Build new index from prompts, preserving existing timestamps where possible
	var indexedPrompts []model.IndexedPrompt
	for _, prompt := range prompts {
		// Check if we have an existing timestamp for this prompt
		var lastUpdated time.Time
		if existingTime, exists := existingTimestamps[prompt.GistURL]; exists {
			// Use existing timestamp to maintain consistency
			lastUpdated = existingTime
		} else {
			// New prompt, use current time
			lastUpdated = time.Now()
		}

		indexedPrompt := model.IndexedPrompt{
			GistURL:     prompt.GistURL,
			FilePath:    fmt.Sprintf("%s.yaml", prompt.Name),
			Author:      prompt.Author,
			Name:        prompt.Name,
			LastUpdated: lastUpdated,
		}
		indexedPrompts = append(indexedPrompts, indexedPrompt)
	}

	// Create new index, preserving existing overall LastUpdated if no new prompts
	var indexLastUpdated time.Time
	if err == nil && len(indexedPrompts) == len(existingIndex.Prompts) {
		// Same number of prompts, check if any are actually new
		hasNewPrompts := false
		for _, prompt := range prompts {
			if _, exists := existingTimestamps[prompt.GistURL]; !exists {
				hasNewPrompts = true
				break
			}
		}
		if !hasNewPrompts {
			// No new prompts, preserve existing index timestamp
			indexLastUpdated = existingIndex.LastUpdated
		} else {
			// Has new prompts, update index timestamp
			indexLastUpdated = time.Now()
		}
	} else {
		// Different number of prompts or no existing index, update timestamp
		indexLastUpdated = time.Now()
	}

	index := &model.Index{
		Prompts:     indexedPrompts,
		LastUpdated: indexLastUpdated,
	}

	// Preserve exports if they exist
	if err == nil && existingIndex.Exports != nil {
		index.Exports = existingIndex.Exports
	}

	return c.cache.SaveIndex(index)
}

// SyncRawIndex downloads the raw index.json content from GitHub and saves it to cache
func (c *CachedStore) SyncRawIndex() error {
	// Cast remote store to GitHubStore to access GetRawIndexContent
	githubStore, ok := c.remote.(*GitHubStore)
	if !ok {
		return fmt.Errorf("remote store is not a GitHubStore")
	}

	// Get raw index content from GitHub
	rawIndexContent, err := githubStore.GetRawIndexContent()
	if err != nil {
		return fmt.Errorf("failed to get raw index content: %w", err)
	}

	// Parse the raw content to validate it's valid JSON
	var index model.Index
	if err := json.Unmarshal([]byte(rawIndexContent), &index); err != nil {
		return fmt.Errorf("invalid index JSON from GitHub: %w", err)
	}
	
	// Debug output: show downloaded index content
	fmt.Printf("📥 从 GitHub 下载的原始 index 内容:\n%s\n", rawIndexContent)
	fmt.Printf("📋 解析后的 Index 结构:\n")
	fmt.Printf("  - Prompts 数量: %d\n", len(index.Prompts))
	fmt.Printf("  - Exports 数量: %d\n", len(index.Exports))
	fmt.Printf("  - 最后更新: %s\n", index.LastUpdated.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Ensure cache directory exists before writing files
	if err := c.cache.EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to ensure cache directory: %w", err)
	}

	// Save the raw content directly to cache as index.json
	indexPath := filepath.Join(c.cache.GetCacheDir(), "index.json")
	if err := config.WriteFileWithPermissions(indexPath, []byte(rawIndexContent)); err != nil {
		return fmt.Errorf("failed to save raw index to cache: %w", err)
	}
	
	// Debug output: show cached index.json content
	cachedContent, err := os.ReadFile(indexPath)
	if err != nil {
		fmt.Printf("⚠️  无法读取缓存的 index.json: %v\n", err)
	} else {
		fmt.Printf("💾 写入缓存的 index.json 内容:\n%s\n", string(cachedContent))
	}

	// Note: We don't call c.cache.SaveIndex() here to avoid overwriting the raw index.json
	// The raw content has already been saved directly to maintain exact timestamp formatting.

	return nil
}

// CreatePublicGist creates a new public gist and delegates to remote store
func (c *CachedStore) CreatePublicGist(prompt model.Prompt) (string, error) {
	return c.remote.CreatePublicGist(prompt)
}

// UpdateGist updates an existing gist and delegates to remote store
func (c *CachedStore) UpdateGist(gistURL string, prompt model.Prompt) error {
	return c.remote.UpdateGist(gistURL, prompt)
}

// GetGistInfo retrieves gist information and delegates to remote store
func (c *CachedStore) GetGistInfo(gistURL string) (*GistInfo, error) {
	return c.remote.GetGistInfo(gistURL)
}

// AddExport adds a prompt to the export index and updates the cache
func (c *CachedStore) AddExport(prompt model.IndexedPrompt) error {
	// Add to remote store
	if err := c.remote.AddExport(prompt); err != nil {
		return err
	}

	// Update cache with the new export
	// Load current index, add the new export, and save
	index, err := c.cache.LoadIndex()
	if err != nil {
		// If cache doesn't exist, create a new one
		index = &model.Index{
			Prompts:     []model.IndexedPrompt{},
			Exports:     []model.IndexedPrompt{},
			LastUpdated: time.Now(),
		}
	}

	// Ensure Exports is initialized (for compatibility with existing indexes)
	if index.Exports == nil {
		index.Exports = []model.IndexedPrompt{}
	}

	// Add to exports index
	index.Exports = append(index.Exports, prompt)
	index.LastUpdated = time.Now()

	// Save updated index (ignore cache errors to not fail the operation)
	c.cache.SaveIndex(index)

	return nil
}

// UpdateExport updates a prompt in the export index and updates the cache
func (c *CachedStore) UpdateExport(prompt model.IndexedPrompt) error {
	// Update in remote store
	if err := c.remote.UpdateExport(prompt); err != nil {
		return err
	}

	// Update cache
	index, err := c.cache.LoadIndex()
	if err != nil {
		// Cache doesn't exist, skip cache update
		return nil
	}

	// Ensure Exports is initialized (for compatibility with existing indexes)
	if index.Exports == nil {
		index.Exports = []model.IndexedPrompt{}
	}

	// Update the export in cache index
	for i, export := range index.Exports {
		if export.GistURL == prompt.GistURL {
			index.Exports[i] = prompt
			index.LastUpdated = time.Now()
			c.cache.SaveIndex(index)
			return nil
		}
	}

	// If export not found in cache, add it (similar to GitHubStore.UpdateExport behavior)
	index.Exports = append(index.Exports, prompt)
	index.LastUpdated = time.Now()
	c.cache.SaveIndex(index)

	return nil
}

// GetExports retrieves all exported prompts and delegates to remote store
func (c *CachedStore) GetExports() ([]model.IndexedPrompt, error) {
	return c.remote.GetExports()
}

// FindExistingPromptByURL 根据 gist URL 查找已存在的提示词
func (c *CachedStore) FindExistingPromptByURL(gistURL string) (*model.Prompt, error) {
	return c.remote.FindExistingPromptByURL(gistURL)
}

// contains performs case-insensitive substring matching
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
