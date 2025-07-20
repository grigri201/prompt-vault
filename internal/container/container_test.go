package container

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/grigri201/prompt-vault/internal/config"
)

func TestNewContainer(t *testing.T) {
	cont := NewContainer()

	if cont.PathManager == nil {
		t.Error("PathManager should not be nil")
	}
	if cont.CacheManager == nil {
		t.Error("CacheManager should not be nil")
	}
	if cont.ConfigManager == nil {
		t.Error("ConfigManager should not be nil")
	}
	if cont.AuthManager == nil {
		t.Error("AuthManager should not be nil")
	}
	if cont.GistClient != nil {
		t.Error("GistClient should be nil before initialization")
	}
}

func TestNewTestContainer(t *testing.T) {
	tempDir := t.TempDir()
	cont := NewTestContainer(tempDir)

	if cont.PathManager == nil {
		t.Error("PathManager should not be nil")
	}

	// Verify paths use the test directory
	cachePath := cont.PathManager.GetCachePath()
	if !filepath.HasPrefix(cachePath, tempDir) {
		t.Errorf("Cache path %q should be under test directory %q", cachePath, tempDir)
	}

	configPath := cont.PathManager.GetConfigPath()
	if !filepath.HasPrefix(configPath, tempDir) {
		t.Errorf("Config path %q should be under test directory %q", configPath, tempDir)
	}
}

func TestContainer_Initialize(t *testing.T) {
	tempDir := t.TempDir()
	cont := NewTestContainer(tempDir)
	ctx := context.Background()

	// Test initial state
	if cont.IsInitialized() {
		t.Error("Container should not be initialized before Initialize()")
	}

	// Test initialization
	if err := cont.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if !cont.IsInitialized() {
		t.Error("Container should be initialized after Initialize()")
	}

	// Verify directories were created
	cachePath := cont.PathManager.GetCachePath()
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}

	configDir := filepath.Dir(cont.PathManager.GetConfigPath())
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	// Config file is not created automatically by GetConfig, only when saved
	// So we don't check for file existence here

	// Test idempotency
	if err := cont.Initialize(ctx); err != nil {
		t.Errorf("Second Initialize() error = %v", err)
	}
}

func TestContainer_InitializeWithToken(t *testing.T) {
	tempDir := t.TempDir()
	cont := NewTestContainer(tempDir)
	ctx := context.Background()
	testToken := "test-github-token"

	// Test initialization with token
	if err := cont.InitializeWithToken(ctx, testToken); err != nil {
		t.Fatalf("InitializeWithToken() error = %v", err)
	}

	// Verify Gist client was created
	if cont.GistClient == nil {
		t.Error("GistClient should not be nil after InitializeWithToken")
	}

	// Verify token was saved to config
	cfg, err := cont.ConfigManager.GetConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Token != testToken {
		t.Errorf("Config token = %q, want %q", cfg.Token, testToken)
	}
}

func TestContainer_Initialize_WithExistingConfig(t *testing.T) {
	tempDir := t.TempDir()
	cont := NewTestContainer(tempDir)
	ctx := context.Background()

	// Create config with token
	testToken := "existing-token"
	cfg := &config.Config{Token: testToken}

	// Ensure config directory exists
	configDir := filepath.Dir(cont.PathManager.GetConfigPath())
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Save config
	if err := cont.ConfigManager.SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Initialize container
	if err := cont.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify Gist client was created with existing token
	if cont.GistClient == nil {
		t.Error("GistClient should not be nil when config has token")
	}
}

func TestContainer_RequireGistClient(t *testing.T) {
	tempDir := t.TempDir()
	cont := NewTestContainer(tempDir)
	ctx := context.Background()

	// Test before initialization
	_, err := cont.RequireGistClient()
	if err == nil {
		t.Error("RequireGistClient() should return error when GistClient is nil")
	}

	// Initialize without token
	if err := cont.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test without token
	_, err = cont.RequireGistClient()
	if err == nil {
		t.Error("RequireGistClient() should return error when GistClient is nil")
	}

	// Initialize with token
	if err := cont.InitializeWithToken(ctx, "test-token"); err != nil {
		t.Fatalf("InitializeWithToken() error = %v", err)
	}

	// Test with token
	client, err := cont.RequireGistClient()
	if err != nil {
		t.Errorf("RequireGistClient() error = %v", err)
	}
	if client == nil {
		t.Error("RequireGistClient() returned nil client")
	}
}

func TestContainer_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	cont := NewTestContainer(tempDir)
	ctx := context.Background()

	// Initialize first
	if err := cont.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test cleanup
	if err := cont.Cleanup(); err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}

	// Verify state after cleanup
	if cont.IsInitialized() {
		t.Error("Container should not be initialized after Cleanup()")
	}
}
