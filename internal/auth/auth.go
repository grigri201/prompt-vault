package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grigri201/prompt-vault/internal/config"
)

// Manager handles authentication token storage and retrieval
type Manager struct {
	configPath string
}

// NewManager creates a new authentication manager with default config path
func NewManager() *Manager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configPath := filepath.Join(homeDir, ".config", "prompt-vault", "config.yaml")
	return &Manager{
		configPath: configPath,
	}
}

// NewManagerWithPath creates a new authentication manager with custom config path
func NewManagerWithPath(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// SaveToken saves the authentication token and username to the config file
func (m *Manager) SaveToken(token, username string) error {
	// Validate inputs
	if token == "" {
		return errors.New("token is required")
	}
	if username == "" {
		return errors.New("username is required")
	}

	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create new one
	cfg, err := config.Load(m.configPath)
	if err != nil {
		// If config doesn't exist, create new one
		cfg = &config.Config{}
	}

	// Update token and username
	cfg.Token = token
	cfg.Username = username

	// Save config
	if err := cfg.Save(m.configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetToken retrieves the stored token and username
func (m *Manager) GetToken() (string, string, error) {
	// Load config
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return "", "", fmt.Errorf("no token found: %w", err)
	}

	// Validate token exists
	if cfg.Token == "" {
		return "", "", errors.New("no token found in config")
	}

	return cfg.Token, cfg.Username, nil
}

// RemoveToken removes the stored token and username
func (m *Manager) RemoveToken() error {
	// Load config
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Clear token and username
	cfg.Token = ""
	cfg.Username = ""

	// Save config
	if err := cfg.Save(m.configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the config file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}