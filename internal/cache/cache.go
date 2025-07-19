package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
	"gopkg.in/yaml.v3"
)

// Manager handles cache operations for prompt templates
type Manager struct {
	cachePath string
	mu        sync.RWMutex // Mutex for concurrent access
}

// NewManager creates a new cache manager with default path
func NewManager() *Manager {
	return &Manager{
		cachePath: GetCachePath(),
	}
}

// NewManagerWithPath creates a new cache manager with custom path
func NewManagerWithPath(path string) *Manager {
	return &Manager{
		cachePath: path,
	}
}

// GetCachePath returns the default cache directory path
func GetCachePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not available
		homeDir = "."
	}
	return filepath.Join(homeDir, ".cache", "prompt-vault", "prompts")
}

// InitializeCache creates the cache directory structure if it doesn't exist
func (m *Manager) InitializeCache() error {
	// Create cache directory with secure permissions
	if err := os.MkdirAll(m.cachePath, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	return nil
}

// GetCacheDir returns the cache directory path
func (m *Manager) GetCacheDir() string {
	return m.cachePath
}

// GetIndexPath returns the path to the index.json file
func (m *Manager) GetIndexPath() string {
	// Index is stored in parent directory of prompts
	return filepath.Join(filepath.Dir(m.cachePath), "index.json")
}

// GetMetadataPath returns the path to the metadata.json file
func (m *Manager) GetMetadataPath() string {
	// Metadata is stored in parent directory of prompts
	return filepath.Join(filepath.Dir(m.cachePath), "metadata.json")
}

// GetPromptPath returns the path for a cached prompt file
func (m *Manager) GetPromptPath(gistID string) string {
	return filepath.Join(m.cachePath, gistID+".yaml")
}

// Clean removes all cached files but keeps the directory structure
func (m *Manager) Clean() error {
	// Remove index.json if it exists
	indexPath := m.GetIndexPath()
	if err := os.Remove(indexPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove index file: %w", err)
	}

	// Remove metadata.json if it exists
	metadataPath := m.GetMetadataPath()
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata file: %w", err)
	}

	// Remove all files in prompts directory
	entries, err := os.ReadDir(m.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, nothing to clean
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(m.cachePath, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove cached file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// SavePrompt saves a prompt to the cache
func (m *Manager) SavePrompt(prompt *models.Prompt) error {
	if prompt == nil {
		return fmt.Errorf("prompt is nil")
	}
	
	if prompt.GistID == "" {
		return fmt.Errorf("prompt GistID is empty")
	}

	// Lock for writing
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create metadata map for YAML marshaling
	meta := map[string]interface{}{
		"name":        prompt.Name,
		"author":      prompt.Author,
		"category":    prompt.Category,
		"tags":        prompt.Tags,
		"version":     prompt.Version,
		"description": prompt.Description,
	}

	// Marshal metadata to YAML
	metaYAML, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create complete content with front matter
	frontMatter := fmt.Sprintf("---\n%s---\n%s", string(metaYAML), prompt.Content)

	// Get prompt file path
	promptPath := m.GetPromptPath(prompt.GistID)

	// Write to file with secure permissions
	if err := os.WriteFile(promptPath, []byte(frontMatter), 0600); err != nil {
		return fmt.Errorf("failed to save prompt: %w", err)
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
	content, err := os.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("prompt not found in cache")
		}
		return nil, fmt.Errorf("failed to read prompt: %w", err)
	}

	// Parse the prompt file
	prompt, err := parser.ParsePromptFile(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse cached prompt: %w", err)
	}

	// Set the GistID (not stored in the file content)
	prompt.GistID = gistID

	return prompt, nil
}

// SaveIndex saves the index to the cache
func (m *Manager) SaveIndex(index *models.Index) error {
	if index == nil {
		return fmt.Errorf("index is nil")
	}
	
	if index.Username == "" {
		return fmt.Errorf("index username is empty")
	}

	// Lock for writing
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get index file path
	indexPath := m.GetIndexPath()

	// Marshal index to JSON with indentation
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	// Write to file atomically to prevent corruption
	// First write to temp file
	tempPath := indexPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp index file: %w", err)
	}

	// Rename temp file to final path (atomic on most filesystems)
	if err := os.Rename(tempPath, indexPath); err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to save index file: %w", err)
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
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Index doesn't exist yet, return nil
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	// Handle empty file
	if len(data) == 0 {
		return nil, fmt.Errorf("index file is empty")
	}

	// Unmarshal JSON
	var index models.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
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
	promptPath := filepath.Join(m.cachePath, safeFilename+".yaml")

	// Remove the file
	if err := os.Remove(promptPath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, not an error
			return nil
		}
		return err
	}

	return nil
}