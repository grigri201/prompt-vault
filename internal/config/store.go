package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/grigri/pv/internal/errors"
)

// fileStore implements the Store interface using file system storage
type fileStore struct {
	configPath string
}

// NewFileStore creates a new file-based configuration store
func NewFileStore() (Store, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	return &fileStore{
		configPath: configPath,
	}, nil
}

// SaveToken saves the GitHub authentication token
func (s *fileStore) SaveToken(token string) error {
	// Load existing config or create new one
	config := &Config{}
	data, err := os.ReadFile(s.configPath)
	if err == nil {
		// If file exists, unmarshal existing config
		if err := json.Unmarshal(data, config); err != nil {
			return errors.NewAppError(errors.ErrStorage, "failed to parse config file", err)
		}
	}

	// Obfuscate token before saving
	obfuscatedToken := obfuscate(token)
	config.GitHubToken = obfuscatedToken

	// Marshal config
	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to marshal config", err)
	}

	// Write to file with restricted permissions
	if err := writeFileWithPermissions(s.configPath, data); err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to save token", err)
	}

	return nil
}

// GetToken retrieves the stored GitHub authentication token
func (s *fileStore) GetToken() (string, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.ErrTokenNotFound
		}
		return "", errors.NewAppError(errors.ErrStorage, "failed to read config file", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return "", errors.NewAppError(errors.ErrStorage, "failed to parse config file", err)
	}

	if config.GitHubToken == "" {
		return "", errors.ErrTokenNotFound
	}

	// Deobfuscate token
	token := deobfuscate(config.GitHubToken)
	return token, nil
}

// DeleteToken removes the stored token
func (s *fileStore) DeleteToken() error {
	// Load existing config
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Nothing to delete
			return nil
		}
		return errors.NewAppError(errors.ErrStorage, "failed to read config file", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to parse config file", err)
	}

	// Clear token
	config.GitHubToken = ""

	// Save updated config
	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to marshal config", err)
	}

	if err := writeFileWithPermissions(s.configPath, data); err != nil {
		return errors.NewAppError(errors.ErrStorage, "failed to save config", err)
	}

	return nil
}

// GetConfigPath returns the configuration file path
func (s *fileStore) GetConfigPath() string {
	return s.configPath
}

// getConfigDir is a function variable that returns the configuration directory path based on the OS
// It's a variable so it can be overridden in tests
var getConfigDir = func() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\pv
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "pv")
	default:
		// Linux/macOS: ~/.config/pv (XDG Base Directory Specification)
		// First try XDG_CONFIG_HOME, then fallback to ~/.config
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			configHome = filepath.Join(homeDir, ".config")
		}
		configDir = filepath.Join(configHome, "pv")
	}

	return configDir, nil
}

// writeFileWithPermissions writes data to a file with appropriate permissions
func writeFileWithPermissions(path string, data []byte) error {
	// Write to a temporary file first
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return err
	}

	// Set appropriate permissions based on OS
	if runtime.GOOS != "windows" {
		// Unix-like systems: ensure 0600 permissions
		if err := os.Chmod(tempPath, 0600); err != nil {
			os.Remove(tempPath)
			return err
		}
	}
	// Note: Windows ACL handling would go here if needed

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return err
	}

	return nil
}

// WriteFileWithPermissions is a public wrapper around writeFileWithPermissions
// It writes data to a file with appropriate secure permissions (0600)
func WriteFileWithPermissions(path string, data []byte) error {
	return writeFileWithPermissions(path, data)
}
