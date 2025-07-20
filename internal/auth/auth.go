package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
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
		return errors.NewValidationErrorMsg("SaveToken", "token is required")
	}
	if username == "" {
		return errors.NewValidationErrorMsg("SaveToken", "username is required")
	}

	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return errors.WrapWithMessage(err, "failed to create config directory")
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
		return errors.WrapWithMessage(err, "failed to save config")
	}

	return nil
}

// GetToken retrieves the stored token and username
func (m *Manager) GetToken() (string, string, error) {
	// Load config
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return "", "", errors.WrapWithMessage(err, "no token found")
	}

	// Validate token exists
	if cfg.Token == "" {
		return "", "", errors.NewAuthErrorMsg("GetToken", "no token found in config")
	}

	return cfg.Token, cfg.Username, nil
}

// RemoveToken removes the stored token and username
func (m *Manager) RemoveToken() error {
	// Load config
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to load config")
	}

	// Clear token and username
	cfg.Token = ""
	cfg.Username = ""

	// Save config
	if err := cfg.Save(m.configPath); err != nil {
		return errors.WrapWithMessage(err, "failed to save config")
	}

	return nil
}

// GetConfigPath returns the path to the config file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// ValidateToken validates the stored token with GitHub API
func (m *Manager) ValidateToken(ctx context.Context) (string, error) {
	// Get stored token
	token, username, err := m.GetToken()
	if err != nil {
		return "", err
	}

	// Create GitHub client
	client, err := gist.NewClient(token)
	if err != nil {
		return "", errors.WrapWithMessage(err, "failed to create GitHub client")
	}

	// Validate token and get username
	validatedUsername, err := client.ValidateToken(ctx)
	if err != nil {
		return "", err
	}

	// Update stored username if different
	if username != validatedUsername {
		if err := m.SaveToken(token, validatedUsername); err != nil {
			// Log error but don't fail validation
			// In real implementation, we might want to log this
		}
	}

	return validatedUsername, nil
}

// GetCurrentUser returns the current authenticated user's username
// It first checks the cached username, and if not found, fetches from API
func (m *Manager) GetCurrentUser(ctx context.Context) (string, error) {
	// Get stored token and username
	token, username, err := m.GetToken()
	if err != nil {
		return "", err
	}

	// If username is already cached, return it
	if username != "" {
		return username, nil
	}

	// Username not cached, fetch from API
	client, err := gist.NewClient(token)
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Validate token and get username
	validatedUsername, err := client.ValidateToken(ctx)
	if err != nil {
		return "", err
	}

	// Cache the username
	if err := m.SaveToken(token, validatedUsername); err != nil {
		// Log error but don't fail the operation
		// In real implementation, we might want to log this
	}

	return validatedUsername, nil
}
