package infra

import (
	"fmt"
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
	// Build new index from prompts
	var indexedPrompts []model.IndexedPrompt
	for _, prompt := range prompts {
		indexedPrompt := model.IndexedPrompt{
			GistURL:     prompt.GistURL,
			FilePath:    fmt.Sprintf("%s.yaml", prompt.Name),
			Author:      prompt.Author,
			Name:        prompt.Name,
			LastUpdated: time.Now(),
		}
		indexedPrompts = append(indexedPrompts, indexedPrompt)
	}

	// Create and save new index
	index := &model.Index{
		Prompts:     indexedPrompts,
		LastUpdated: time.Now(),
	}

	return c.cache.SaveIndex(index)
}

// contains performs case-insensitive substring matching
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
