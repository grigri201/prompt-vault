package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid config with token and username",
			config: Config{
				Token:    "ghp_testtoken123",
				Username: "john",
			},
			wantError: false,
		},
		{
			name: "missing token",
			config: Config{
				Username: "john",
			},
			wantError: true,
		},
		{
			name: "missing username",
			config: Config{
				Token: "ghp_testtoken123",
			},
			wantError: true,
		},
		{
			name:      "empty config",
			config:    Config{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set test HOME
	testHome := t.TempDir()
	os.Setenv("HOME", testHome)

	expectedPath := filepath.Join(testHome, ".config", "prompt-vault", "config.yaml")
	actualPath := GetConfigPath()

	if actualPath != expectedPath {
		t.Errorf("GetConfigPath() = %s, want %s", actualPath, expectedPath)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create test config
	original := Config{
		Token:    "ghp_testtoken123",
		Username: "john",
		LastSync: time.Date(2024, 1, 19, 10, 30, 0, 0, time.UTC),
	}

	// Test Save
	err := original.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Check file permissions (should be 0600)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("Config file permissions = %v, want 0600", perm)
	}

	// Test Load
	loaded := &Config{}
	err = loaded.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Compare loaded config with original
	if loaded.Token != original.Token {
		t.Errorf("Token = %s, want %s", loaded.Token, original.Token)
	}
	if loaded.Username != original.Username {
		t.Errorf("Username = %s, want %s", loaded.Username, original.Username)
	}
	if !loaded.LastSync.Equal(original.LastSync) {
		t.Errorf("LastSync = %v, want %v", loaded.LastSync, original.LastSync)
	}
}

func TestConfig_LoadFromFile_NotExist(t *testing.T) {
	config := &Config{}
	err := config.LoadFromFile("/nonexistent/path/config.yaml")

	if err == nil {
		t.Error("LoadFromFile() should return error for non-existent file")
	}

	if !os.IsNotExist(err) {
		t.Errorf("LoadFromFile() should return os.IsNotExist error, got %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Token != "" {
		t.Error("DefaultConfig() should have empty token")
	}
	if config.Username != "" {
		t.Error("DefaultConfig() should have empty username")
	}
	if !config.LastSync.IsZero() {
		t.Error("DefaultConfig() should have zero LastSync time")
	}
}

func TestConfig_Save_CreateDirectory(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "config.yaml")

	config := Config{
		Token:    "ghp_testtoken123",
		Username: "john",
	}

	// Save should create the directory
	err := config.SaveToFile(configPath)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestManager_GetConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".config", "prompt-vault", "config.yaml")

	// Create manager with test path
	manager := NewManagerWithPath(configPath)

	// First call should create default config
	config, err := manager.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config.Token != "" || config.Username != "" {
		t.Error("First GetConfig() should return default config")
	}

	// Save a config
	testConfig := Config{
		Token:    "ghp_testtoken123",
		Username: "john",
	}
	// Use the manager to save the config properly
	err = manager.SaveConfig(&testConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Create new manager to test loading
	manager2 := NewManagerWithPath(configPath)

	// Should load the saved config
	config2, err := manager2.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config2.Token != testConfig.Token {
		t.Errorf("GetConfig() Token = %s, want %s", config2.Token, testConfig.Token)
	}
}

func TestManager_SaveConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".config", "prompt-vault", "config.yaml")

	manager := NewManagerWithPath(configPath)

	config := &Config{
		Token:    "ghp_testtoken123",
		Username: "john",
	}

	err := manager.SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify the config was saved by creating a new manager
	manager2 := NewManagerWithPath(configPath)
	loaded, err := manager2.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if loaded.Token != config.Token {
		t.Errorf("Saved Token = %s, want %s", loaded.Token, config.Token)
	}
}

func TestManager_UpdateLastSync(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".config", "prompt-vault", "config.yaml")

	// Create manager
	manager := NewManagerWithPath(configPath)

	// Create and save initial config
	initialConfig := &Config{
		Token:    "ghp_testtoken123",
		Username: "john",
		LastSync: time.Time{}, // Zero time
	}

	// Save initial config through manager
	err := manager.SaveConfig(initialConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Record time before update
	beforeUpdate := time.Now()

	// Update last sync
	err = manager.UpdateLastSync()
	if err != nil {
		t.Fatalf("UpdateLastSync() error = %v", err)
	}

	// Get updated config through manager
	updatedConfig, err := manager.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	// Check that LastSync was updated
	if updatedConfig.LastSync.IsZero() {
		t.Error("LastSync should not be zero after update")
	}

	if updatedConfig.LastSync.Before(beforeUpdate) {
		t.Error("LastSync should be after the update call")
	}

	// Check that other fields remain unchanged
	if updatedConfig.Token != initialConfig.Token {
		t.Errorf("Token changed: got %s, want %s", updatedConfig.Token, initialConfig.Token)
	}
	if updatedConfig.Username != initialConfig.Username {
		t.Errorf("Username changed: got %s, want %s", updatedConfig.Username, initialConfig.Username)
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	// Since Manager no longer exposes configPath directly,
	// we test by trying to save and load a config
	testConfig := &Config{
		Token:    "test_token",
		Username: "test_user",
	}

	// Save config
	err := manager.SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Create new manager and load config
	manager2 := NewManager()
	loadedConfig, err := manager2.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	// Verify the config was saved to the default path
	if loadedConfig.Token != testConfig.Token {
		t.Errorf("Loaded config Token = %s, want %s", loadedConfig.Token, testConfig.Token)
	}
}

func TestNewManagerWithPath(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, ".config", "prompt-vault", "config.yaml")
	manager := NewManagerWithPath(customPath)

	// Test by saving and loading config
	testConfig := &Config{
		Token:    "test_token",
		Username: "test_user",
	}

	// Save config
	err := manager.SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify the config was saved by using another manager with the same path
	manager2 := NewManagerWithPath(customPath)
	loadedConfig, err := manager2.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if loadedConfig.Token != testConfig.Token {
		t.Errorf("Loaded config Token = %s, want %s", loadedConfig.Token, testConfig.Token)
	}
}
