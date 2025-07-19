package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not available
		homeDir = "."
	}
	return filepath.Join(homeDir, ".config", "prompt-vault", "config.yaml")
}

// Manager handles configuration operations
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		configPath: GetConfigPath(),
	}
}

// NewManagerWithPath creates a new configuration manager with a custom path
func NewManagerWithPath(path string) *Manager {
	return &Manager{
		configPath: path,
	}
}

// GetConfig returns the current configuration, loading it if necessary
func (m *Manager) GetConfig() (*Config, error) {
	if m.config != nil {
		return m.config, nil
	}

	config := DefaultConfig()
	
	// Try to load existing config
	if err := config.LoadFromFile(m.configPath); err != nil {
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
	if err := config.SaveToFile(m.configPath); err != nil {
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