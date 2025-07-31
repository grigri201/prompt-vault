package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/managers"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
	"github.com/grigri201/prompt-vault/internal/paths"
	"gopkg.in/yaml.v3"
)

// Manager handles cache operations for prompt templates
type Manager struct {
	managers.BaseManager
	pathManager *paths.PathManager
	mu          sync.RWMutex // Mutex for concurrent access
}

// NewManager creates a new cache manager with default path
func NewManager() *Manager {
	pm := paths.NewPathManager()
	return &Manager{
		pathManager: pm,
	}
}

// NewManagerWithPath creates a new cache manager with custom path
func NewManagerWithPath(path string) *Manager {
	// Extract home directory from the path to create appropriate PathManager
	// This maintains backward compatibility
	homeDir := extractHomeDir(path)
	pm := paths.NewPathManagerWithHome(homeDir)
	return &Manager{
		pathManager: pm,
	}
}

// NewManagerWithPathManager creates a new cache manager with a path manager
func NewManagerWithPathManager(pm *paths.PathManager) *Manager {
	return &Manager{
		pathManager: pm,
	}
}

// extractHomeDir extracts the home directory from a cache path
// For backward compatibility with existing NewManagerWithPath usage
func extractHomeDir(cachePath string) string {
	// If path contains .cache/prompt-vault/prompts, extract the base
	const cacheSubPath = ".cache/prompt-vault/prompts"
	if idx := strings.Index(cachePath, cacheSubPath); idx > 0 {
		return cachePath[:idx-1] // -1 to remove the trailing slash
	}
	// Otherwise, assume it's a custom path and use parent directory
	return filepath.Dir(filepath.Dir(filepath.Dir(cachePath)))
}

// GetCachePath returns the default cache directory path
func GetCachePath() string {
	pm := paths.NewPathManager()
	return pm.GetCachePath()
}

// Initialize implements managers.Manager interface
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.pathManager.EnsureCacheDir(); err != nil {
		return errors.WrapError("Initialize", err)
	}

	m.SetInitialized(true)
	return nil
}

// Cleanup implements managers.Manager interface
func (m *Manager) Cleanup() error {
	// Currently no cleanup needed for cache manager
	m.SetInitialized(false)
	return nil
}

// InitializeCache creates the cache directory structure if it doesn't exist
// Deprecated: Use Initialize(ctx) instead
func (m *Manager) InitializeCache() error {
	return m.Initialize(context.Background())
}

// GetCacheDir returns the cache directory path
func (m *Manager) GetCacheDir() string {
	return m.pathManager.GetCachePath()
}

// GetIndexPath returns the path to the index.json file
func (m *Manager) GetIndexPath() string {
	return m.pathManager.GetIndexPath()
}

// GetMetadataPath returns the path to the metadata.json file
func (m *Manager) GetMetadataPath() string {
	// Metadata is stored in parent directory of prompts
	return filepath.Join(filepath.Dir(m.pathManager.GetCachePath()), "metadata.json")
}

// GetPromptPath returns the path for a cached prompt file
func (m *Manager) GetPromptPath(gistID string) string {
	// pathManager.GetPromptPath expects ID without extension and adds .md
	// but cache uses .yaml extension
	return filepath.Join(m.pathManager.GetCachePath(), gistID+".yaml")
}

// Clean removes all cached files but keeps the directory structure
func (m *Manager) Clean() error {
	// Remove index.json if it exists
	indexPath := m.GetIndexPath()
	if err := m.pathManager.RemoveFile(indexPath); err != nil && !os.IsNotExist(err) {
		return errors.WrapError("clearCache", err)
	}

	// Remove metadata.json if it exists
	metadataPath := m.GetMetadataPath()
	if err := m.pathManager.RemoveFile(metadataPath); err != nil && !os.IsNotExist(err) {
		return errors.WrapError("clearCache", err)
	}

	// Remove all files in prompts directory
	cacheDir := m.pathManager.GetCachePath()
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, nothing to clean
			return nil
		}
		return errors.WrapError("clearCache", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(cacheDir, entry.Name())
			if err := m.pathManager.RemoveFile(path); err != nil {
				return errors.WrapError("clearCache", err)
			}
		}
	}

	return nil
}

// ClearCache removes all cached files (alias for Clean to implement interface)
func (m *Manager) ClearCache() error {
	return m.Clean()
}

// SavePrompt saves a prompt to the cache
func (m *Manager) SavePrompt(prompt *models.Prompt) error {
	if prompt == nil {
		return errors.NewValidationErrorMsg("SavePrompt", "prompt is nil")
	}

	if prompt.GistID == "" {
		return errors.NewValidationErrorMsg("SavePrompt", "prompt GistID is empty")
	}

	// Lock for writing
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create metadata map for YAML marshaling
	meta := map[string]interface{}{
		"name":        prompt.Name,
		"author":      prompt.Author,
		"tags":        prompt.Tags,
		"version":     prompt.Version,
		"description": prompt.Description,
	}

	// Marshal metadata to YAML
	metaYAML, err := yaml.Marshal(meta)
	if err != nil {
		return errors.WrapError("SavePrompt", err)
	}

	// Create complete content with front matter
	frontMatter := fmt.Sprintf("---\n%s---\n%s", string(metaYAML), prompt.Content)

	// Get prompt file path
	promptPath := m.GetPromptPath(prompt.GistID)

	// Write to file atomically with secure permissions
	if err := m.pathManager.AtomicWrite(promptPath, []byte(frontMatter), 0600); err != nil {
		return errors.WrapError("SavePrompt", err)
	}

	return nil
}

// GetPrompt retrieves a prompt from the cache
func (m *Manager) GetPrompt(gistID string) (*models.Prompt, error) {
	// Lock for reading
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get prompt file path
	promptPath := m.GetPromptPath(gistID)

	// Read file content
	content, err := m.pathManager.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewFileSystemErrorMsg("GetPrompt", "prompt not found in cache")
		}
		return nil, errors.WrapError("GetPrompt", err)
	}

	// Parse the prompt file
	prompt, err := parser.ParsePromptFile(string(content))
	if err != nil {
		return nil, errors.WrapError("GetPrompt", err)
	}

	// Set the GistID (not stored in the file content)
	prompt.GistID = gistID

	return prompt, nil
}

// SaveIndex saves the index to the cache
func (m *Manager) SaveIndex(index *models.Index) error {
	if index == nil {
		return errors.NewValidationErrorMsg("SaveIndex", "index is nil")
	}

	if index.Username == "" {
		return errors.NewValidationErrorMsg("SaveIndex", "index username is empty")
	}

	// Lock for writing
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get index file path
	indexPath := m.GetIndexPath()

	// Marshal index to JSON with indentation
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return errors.WrapError("SaveIndex", err)
	}

	// Write to file atomically using pathManager
	if err := m.pathManager.AtomicWrite(indexPath, data, 0600); err != nil {
		return errors.WrapError("SaveIndex", err)
	}

	return nil
}

// GetIndex retrieves the index from the cache
func (m *Manager) GetIndex() (*models.Index, error) {
	// Lock for reading
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get index file path
	indexPath := m.GetIndexPath()

	// Read file content
	data, err := m.pathManager.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Index doesn't exist yet, return nil
			return nil, nil
		}
		return nil, errors.WrapError("GetIndex", err)
	}

	// Handle empty file
	if len(data) == 0 {
		return nil, errors.NewFileSystemErrorMsg("GetIndex", "index file is empty")
	}

	// Unmarshal JSON
	var index models.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, errors.WrapError("GetIndex", err)
	}

	return &index, nil
}

// generateSafeFilename creates a safe filename from a prompt name
func generateSafeFilename(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove or replace unsafe characters
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	name = re.ReplaceAllString(name, "")

	// Convert to lowercase
	name = strings.ToLower(name)

	// Limit length
	if len(name) > 100 {
		name = name[:100]
	}

	return name
}

// DeletePrompt removes a prompt from the cache
func (m *Manager) DeletePrompt(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate safe filename
	safeFilename := generateSafeFilename(name)
	promptPath := filepath.Join(m.pathManager.GetCachePath(), safeFilename+".yaml")

	// Remove the file
	if err := m.pathManager.RemoveFile(promptPath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, not an error
			return nil
		}
		return err
	}

	return nil
}

// Ensure Manager implements the interfaces
var (
	_ interfaces.CacheManager = (*Manager)(nil)
	_ interfaces.CacheReader  = (*Manager)(nil)
	_ interfaces.CacheWriter  = (*Manager)(nil)
)
