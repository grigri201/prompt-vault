package auth

import (
	"context"
	"os"
	"path/filepath"

	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
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

// GetToken retrieves the stored token (interface requirement)
func (m *Manager) GetToken() (string, error) {
	token, _, err := m.GetTokenWithUsername()
	return token, err
}

// GetTokenWithUsername retrieves the stored token and username
func (m *Manager) GetTokenWithUsername() (string, string, error) {
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
	token, username, err := m.GetTokenWithUsername()
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
		if err := m.SaveTokenWithUsername(token, validatedUsername); err != nil {
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
	token, username, err := m.GetTokenWithUsername()
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
		return "", errors.WrapError("AuthenticateWithToken", err)
	}

	// Validate token and get username
	validatedUsername, err := client.ValidateToken(ctx)
	if err != nil {
		return "", err
	}

	// Cache the username
	if err := m.SaveTokenWithUsername(token, validatedUsername); err != nil {
		// Log error but don't fail the operation
		// In real implementation, we might want to log this
	}

	return validatedUsername, nil
}

// AuthenticateWithToken validates a token and stores authentication data
func (m *Manager) AuthenticateWithToken(ctx context.Context, token string) (string, error) {
	// Validate token with GitHub API
	client, err := gist.NewClient(token)
	if err != nil {
		return "", errors.WrapError("AuthenticateWithToken", err)
	}

	// Get username from API
	username, err := client.ValidateToken(ctx)
	if err != nil {
		return "", err
	}

	// Save token and username
	if err := m.SaveTokenWithUsername(token, username); err != nil {
		return "", errors.WrapError("AuthenticateWithToken", err)
	}

	return username, nil
}

// SaveToken saves just the token (interface requirement)
func (m *Manager) SaveToken(token string) error {
	// Get current username to preserve it
	_, username, err := m.GetTokenWithUsername()
	if err != nil {
		// If no existing config, we'll need to get username from token
		client, err := gist.NewClient(token)
		if err != nil {
			return errors.WrapError("SaveToken", err)
		}

		ctx := context.Background()
		username, err = client.ValidateToken(ctx)
		if err != nil {
			return errors.WrapError("SaveToken", err)
		}
	}

	return m.SaveTokenWithUsername(token, username)
}

// SaveTokenWithUsername saves both token and username (renamed from original SaveToken)
func (m *Manager) SaveTokenWithUsername(token, username string) error {
	// Validate inputs
	if token == "" {
		return errors.NewValidationErrorMsg("SaveTokenWithUsername", "token is required")
	}
	if username == "" {
		return errors.NewValidationErrorMsg("SaveTokenWithUsername", "username is required")
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

// SaveUsername saves the username
func (m *Manager) SaveUsername(username string) error {
	token, _, err := m.GetTokenWithUsername()
	if err != nil {
		return errors.WrapError("SaveUsername", err)
	}

	return m.SaveTokenWithUsername(token, username)
}

// GetUsername retrieves the username
func (m *Manager) GetUsername() (string, error) {
	_, username, err := m.GetTokenWithUsername()
	if err != nil {
		return "", err
	}
	return username, nil
}

// IsAuthenticated checks if the user is authenticated
func (m *Manager) IsAuthenticated() bool {
	token, _, err := m.GetTokenWithUsername()
	return err == nil && token != ""
}

// ClearAuthentication removes stored authentication data
func (m *Manager) ClearAuthentication() error {
	return m.RemoveToken()
}

// Initialize prepares the auth manager for use
func (m *Manager) Initialize(ctx context.Context) error {
	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return errors.WrapError("Initialize", err)
	}
	return nil
}

// Cleanup performs any necessary cleanup
func (m *Manager) Cleanup() error {
	// No cleanup needed for auth manager
	return nil
}

// Ensure Manager implements the interfaces
var (
	_ interfaces.AuthManager = (*Manager)(nil)
	_ interfaces.AuthReader  = (*Manager)(nil)
	_ interfaces.AuthWriter  = (*Manager)(nil)
)
