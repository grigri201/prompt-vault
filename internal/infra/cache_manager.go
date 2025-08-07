package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
)

// CacheManager handles local cache file operations for prompts and index data
type CacheManager struct {
	cacheDir string
}

// NewCacheManager creates a new CacheManager instance
func NewCacheManager() (*CacheManager, error) {
	cacheDir, err := getCacheDir()

	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	return &CacheManager{
		cacheDir: cacheDir,
	}, nil
}

// EnsureCacheDir creates the cache directory structure with proper permissions
func (c *CacheManager) EnsureCacheDir() error {
	// Create main cache directory
	if err := os.MkdirAll(c.cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create prompts subdirectory
	promptsDir := filepath.Join(c.cacheDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0700); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Set proper permissions on Unix-like systems
	if runtime.GOOS != "windows" {
		// Ensure cache directory has 0700 permissions
		if err := os.Chmod(c.cacheDir, 0700); err != nil {
			return fmt.Errorf("failed to set cache directory permissions: %w", err)
		}

		// Ensure prompts subdirectory has 0700 permissions
		if err := os.Chmod(promptsDir, 0700); err != nil {
			return fmt.Errorf("failed to set prompts directory permissions: %w", err)
		}
	}

	return nil
}

// LoadIndex reads the cache index from index.json file
// Returns a pointer to model.Index or error if the file doesn't exist or is corrupted
func (c *CacheManager) LoadIndex() (*model.Index, error) {
	indexPath := filepath.Join(c.cacheDir, "index.json")
	
	// Read the index file
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewAppError(errors.ErrStorage, "cache index not found", err)
		}
		return nil, errors.NewAppError(errors.ErrStorage, "failed to read cache index", err)
	}
	
	// Parse JSON data
	var index model.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, errors.NewAppError(errors.ErrStorage, "failed to parse cache index", err)
	}
	
	return &index, nil
}

// SaveIndex saves the model.Index to index.json using JSON serialization
// Reuses existing permission settings from config package
func (c *CacheManager) SaveIndex(index *model.Index) error {
	// Ensure cache directory exists
	if err := c.EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to ensure cache directory: %w", err)
	}
	
	// Marshal index to JSON with proper formatting
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to marshal cache index", err)
	}
	
	// Build path to index file
	indexPath := filepath.Join(c.cacheDir, "index.json")
	
	// Write to file with proper permissions using config package utility
	if err := config.WriteFileWithPermissions(indexPath, data); err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to save cache index", err)
	}
	
	return nil
}

// LoadContent reads raw YAML content from prompts/{gist_id}.yaml file
// Returns the original YAML content exactly as stored in GitHub Gist
func (c *CacheManager) LoadContent(gistID string) (string, error) {
	// Build path to the cached content file
	filename := fmt.Sprintf("%s.yaml", gistID)
	contentPath := filepath.Join(c.cacheDir, "prompts", filename)
	
	// Read the raw content file
	data, err := os.ReadFile(contentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.NewAppError(errors.ErrStorage, "cached content not found", err)
		}
		return "", errors.NewAppError(errors.ErrStorage, "failed to read cached content", err)
	}
	
	// Return raw content as string (identical to GitHub Gist)
	return string(data), nil
}

// SaveContent saves raw YAML content to prompts/{gist_id}.yaml file
// The content is saved exactly as received from GitHub Gist without any modification
func (c *CacheManager) SaveContent(gistID, content string) error {
	// Ensure cache directory exists
	if err := c.EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to ensure cache directory: %w", err)
	}
	
	// Build path to the content file
	filename := fmt.Sprintf("%s.yaml", gistID)
	contentPath := filepath.Join(c.cacheDir, "prompts", filename)
	
	// Save raw content using config package's secure file writing pattern
	// This ensures proper permissions (0600) and atomic writes
	if err := config.WriteFileWithPermissions(contentPath, []byte(content)); err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to save cached content", err)
	}
	
	return nil
}

// GetCacheInfo returns statistical information about the cache directory
// including last update time, total prompts count, and total cache size in bytes
func (c *CacheManager) GetCacheInfo() (*model.CacheInfo, error) {
	info := &model.CacheInfo{
		LastUpdated:  time.Time{}, // Zero time as default
		TotalPrompts: 0,
		CacheSize:    0,
	}
	
	// Get last updated time from index file if it exists
	index, err := c.LoadIndex()
	if err == nil {
		info.LastUpdated = index.LastUpdated
		info.TotalPrompts = len(index.Prompts)
	}
	// If LoadIndex fails, we continue with zero values - this is not a fatal error
	
	// Calculate total cache directory size
	cacheSize, err := c.calculateDirectorySize(c.cacheDir)
	if err != nil {
		// Log the error but don't fail - provide cache info with size 0
		// This handles cases where cache directory doesn't exist or is inaccessible
		info.CacheSize = 0
	} else {
		info.CacheSize = cacheSize
	}
	
	return info, nil
}

// GetCacheDir returns the cache directory path
func (c *CacheManager) GetCacheDir() string {
	return c.cacheDir
}

// calculateDirectorySize recursively calculates the total size of a directory
// Returns total size in bytes or error if directory traversal fails
func (c *CacheManager) calculateDirectorySize(dirPath string) (int64, error) {
	var totalSize int64
	
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		// Handle errors during directory walk gracefully
		if err != nil {
			// Skip files/directories that can't be accessed (permissions, etc.)
			// Don't fail the entire operation for individual file errors
			return nil
		}
		
		// Only count regular files, skip directories and special files
		if info.Mode().IsRegular() {
			totalSize += info.Size()
		}
		
		return nil
	})
	
	if err != nil {
		return 0, fmt.Errorf("failed to calculate directory size: %w", err)
	}
	
	return totalSize, nil
}

// getCacheDir wraps the config.GetCacheDir() function to get cache directory
// This function reuses the existing cache directory logic from internal/config/cache.go
func getCacheDir() (string, error) {
	// Use the GetCacheDir function from config package
	// This leverages the existing implementation that handles:
	// - PV_CACHE_DIR environment variable override
	// - Windows LOCALAPPDATA path
	// - Unix XDG_CACHE_HOME and ~/.cache fallback
	return config.GetCacheDir()
}