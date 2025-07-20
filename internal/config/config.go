package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/grigri201/prompt-vault/internal/managers"
	"github.com/grigri201/prompt-vault/internal/paths"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Token    string    `yaml:"token"`
	Username string    `yaml:"username"`
	LastSync time.Time `yaml:"last_sync"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}

// SaveToFile saves the configuration to a file
func (c *Config) SaveToFile(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveToFileAtomic saves the configuration to a file atomically using PathManager
func (c *Config) SaveToFileAtomic(pm *paths.PathManager, path string) error {
	// Marshal config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file atomically with secure permissions
	if err := pm.AtomicWrite(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadFromFile loads the configuration from a file
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		// Return the original error so os.IsNotExist can work
		return err
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		Token:    "",
		Username: "",
		LastSync: time.Time{},
	}
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	pm := paths.NewPathManager()
	return pm.GetConfigPath()
}

// Manager handles configuration operations
type Manager struct {
	managers.BaseManager
	pathManager *paths.PathManager
	config      *Config
	mu          sync.RWMutex
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	pm := paths.NewPathManager()
	return &Manager{
		pathManager: pm,
	}
}

// NewManagerWithPath creates a new configuration manager with a custom path
func NewManagerWithPath(path string) *Manager {
	// For backward compatibility, extract home directory from path
	homeDir := extractHomeDir(path)
	pm := paths.NewPathManagerWithHome(homeDir)
	return &Manager{
		pathManager: pm,
	}
}

// NewManagerWithPathManager creates a new configuration manager with a path manager
func NewManagerWithPathManager(pm *paths.PathManager) *Manager {
	return &Manager{
		pathManager: pm,
	}
}

// extractHomeDir extracts the home directory from a config path
func extractHomeDir(configPath string) string {
	// If path contains .config/prompt-vault/config.yaml, extract the base
	const configSubPath = ".config/prompt-vault/config.yaml"
	dir := filepath.Dir(configPath)
	if filepath.Base(configPath) == "config.yaml" && filepath.Base(dir) == "prompt-vault" {
		// Go up to find the home directory
		return filepath.Dir(filepath.Dir(dir))
	}
	// Otherwise, use the parent of the config path
	return filepath.Dir(configPath)
}

// Initialize implements managers.Manager interface
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.pathManager.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	m.SetInitialized(true)
	return nil
}

// Cleanup implements managers.Manager interface
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = nil
	m.SetInitialized(false)
	return nil
}

// GetConfig returns the current configuration, loading it if necessary
func (m *Manager) GetConfig() (*Config, error) {
	m.mu.RLock()
	if m.config != nil {
		m.mu.RUnlock()
		return m.config, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if m.config != nil {
		return m.config, nil
	}

	config := DefaultConfig()
	configPath := m.pathManager.GetConfigPath()

	// Try to load existing config
	if err := config.LoadFromFile(configPath); err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			m.config = config
			return config, nil
		}
		// For other errors, return the error
		return nil, err
	}

	m.config = config
	return config, nil
}

// SaveConfig saves the configuration to file
func (m *Manager) SaveConfig(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	configPath := m.pathManager.GetConfigPath()
	if err := config.SaveToFileAtomic(m.pathManager, configPath); err != nil {
		return err
	}
	m.config = config
	return nil
}

// UpdateLastSync updates the last sync time and saves the config
func (m *Manager) UpdateLastSync() error {
	config, err := m.GetConfig()
	if err != nil {
		return err
	}

	config.LastSync = time.Now()
	return m.SaveConfig(config)
}

// Save is a convenience method that saves the config to the specified path
func (c *Config) Save(path string) error {
	return c.SaveToFile(path)
}

// Load is a convenience function that loads a config from the specified path
func Load(path string) (*Config, error) {
	config := &Config{}
	if err := config.LoadFromFile(path); err != nil {
		return nil, err
	}
	return config, nil
}
