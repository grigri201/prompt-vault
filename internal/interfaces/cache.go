package interfaces

import (
	"context"

	"github.com/grigri201/prompt-vault/internal/models"
)

// CacheReader defines read operations for cache
type CacheReader interface {
	// GetPrompt retrieves a cached prompt by its gist ID
	GetPrompt(gistID string) (*models.Prompt, error)

	// GetIndex retrieves the cached index
	GetIndex() (*models.Index, error)

	// GetCacheDir returns the cache directory path
	GetCacheDir() string
}

// CacheWriter defines write operations for cache
type CacheWriter interface {
	// SavePrompt saves a prompt to the cache
	SavePrompt(prompt *models.Prompt) error

	// SaveIndex saves the index to the cache
	SaveIndex(index *models.Index) error

	// DeletePrompt removes a prompt from the cache
	DeletePrompt(name string) error

	// ClearCache removes all cached files
	ClearCache() error
}

// CacheManager combines read and write operations with lifecycle management
type CacheManager interface {
	CacheReader
	CacheWriter

	// Initialize prepares the cache manager for use
	Initialize(ctx context.Context) error

	// Cleanup performs any necessary cleanup
	Cleanup() error
}
