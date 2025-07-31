package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/grigri201/prompt-vault/internal/errors"
)

// AppConfig represents the application-wide configuration
type AppConfig struct {
	IndexGistID string `json:"index_gist_id"`
}

// GetAppConfigPath returns the default app config file path
func GetAppConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".config", "prompt-vault", "app_config.json")
}

// LoadAppConfig loads the app configuration from file
func LoadAppConfig() (*AppConfig, error) {
	path := GetAppConfigPath()
	return LoadAppConfigFromPath(path)
}

// LoadAppConfigFromPath loads the app configuration from a specific path
func LoadAppConfigFromPath(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &AppConfig{}, nil
		}
		return nil, errors.NewFileSystemError("LoadAppConfigFromPath", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.NewParsingError("LoadAppConfigFromPath", err)
	}

	return &config, nil
}

// SaveAppConfig saves the app configuration to file
func SaveAppConfig(config *AppConfig) error {
	path := GetAppConfigPath()
	return SaveAppConfigToPath(config, path)
}

// SaveAppConfigToPath saves the app configuration to a specific path
func SaveAppConfigToPath(config *AppConfig, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.NewFileSystemError("SaveAppConfigToPath", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.NewParsingError("SaveAppConfigToPath", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return errors.NewFileSystemError("SaveAppConfigToPath", err)
	}

	return nil
}
