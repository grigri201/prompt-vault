package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/gist"
)

func TestManager_SaveToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		username string
		setup    func(t *testing.T, configPath string)
		wantErr  bool
		verify   func(t *testing.T, configPath string)
	}{
		{
			name:     "saves token successfully",
			token:    "ghp_testtoken123",
			username: "testuser",
			wantErr:  false,
			verify: func(t *testing.T, configPath string) {
				// Verify config file exists
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Error("Config file was not created")
				}

				// Verify file permissions (should be 0600)
				info, err := os.Stat(configPath)
				if err != nil {
					t.Fatalf("Failed to stat config file: %v", err)
				}
				if perm := info.Mode().Perm(); perm != 0600 {
					t.Errorf("Config file permissions = %v, want 0600", perm)
				}

				// Verify content
				cfg, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Token != "ghp_testtoken123" {
					t.Errorf("Token = %v, want ghp_testtoken123", cfg.Token)
				}
				if cfg.Username != "testuser" {
					t.Errorf("Username = %v, want testuser", cfg.Username)
				}
			},
		},
		{
			name:     "updates existing token",
			token:    "ghp_newtoken456",
			username: "newuser",
			setup: func(t *testing.T, configPath string) {
				// Create existing config
				cfg := &config.Config{
					Token:    "ghp_oldtoken123",
					Username: "olduser",
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, configPath string) {
				cfg, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Token != "ghp_newtoken456" {
					t.Errorf("Token = %v, want ghp_newtoken456", cfg.Token)
				}
				if cfg.Username != "newuser" {
					t.Errorf("Username = %v, want newuser", cfg.Username)
				}
			},
		},
		{
			name:     "rejects empty token",
			token:    "",
			username: "testuser",
			wantErr:  true,
		},
		{
			name:     "rejects empty username",
			token:    "ghp_testtoken123",
			username: "",
			wantErr:  true,
		},
		{
			name:     "handles config directory creation",
			token:    "ghp_testtoken123",
			username: "testuser",
			setup: func(t *testing.T, configPath string) {
				// Remove config directory
				configDir := filepath.Dir(configPath)
				if err := os.RemoveAll(configDir); err != nil {
					t.Fatalf("Failed to remove config dir: %v", err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, configPath string) {
				// Verify directory was created
				configDir := filepath.Dir(configPath)
				if _, err := os.Stat(configDir); os.IsNotExist(err) {
					t.Error("Config directory was not created")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "prompt-vault", "config.yaml")

			// Create manager
			m := &Manager{
				configPath: configPath,
			}

			if tt.setup != nil {
				tt.setup(t, configPath)
			}

			err := m.SaveTokenWithUsername(tt.token, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, configPath)
			}
		})
	}
}

func TestManager_GetToken(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T, configPath string)
		wantToken    string
		wantUsername string
		wantErr      bool
	}{
		{
			name: "retrieves token successfully",
			setup: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Token:    "ghp_testtoken123",
					Username: "testuser",
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			wantToken:    "ghp_testtoken123",
			wantUsername: "testuser",
			wantErr:      false,
		},
		{
			name:         "returns error when config doesn't exist",
			wantToken:    "",
			wantUsername: "",
			wantErr:      true,
		},
		{
			name: "returns error when token is empty",
			setup: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Token:    "",
					Username: "testuser",
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			wantToken:    "",
			wantUsername: "",
			wantErr:      true,
		},
		{
			name: "handles corrupted config file",
			setup: func(t *testing.T, configPath string) {
				// Create directory
				configDir := filepath.Dir(configPath)
				if err := os.MkdirAll(configDir, 0700); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				// Write invalid YAML
				if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0600); err != nil {
					t.Fatalf("Failed to write corrupted config: %v", err)
				}
			},
			wantToken:    "",
			wantUsername: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "prompt-vault", "config.yaml")

			// Create manager
			m := &Manager{
				configPath: configPath,
			}

			if tt.setup != nil {
				tt.setup(t, configPath)
			}

			token, username, err := m.GetTokenWithUsername()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if token != tt.wantToken {
				t.Errorf("GetToken() token = %v, want %v", token, tt.wantToken)
			}

			if username != tt.wantUsername {
				t.Errorf("GetToken() username = %v, want %v", username, tt.wantUsername)
			}
		})
	}
}

func TestManager_RemoveToken(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, configPath string)
		wantErr bool
		verify  func(t *testing.T, configPath string)
	}{
		{
			name: "removes token successfully",
			setup: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Token:    "ghp_testtoken123",
					Username: "testuser",
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, configPath string) {
				cfg, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Token != "" {
					t.Errorf("Token = %v, want empty", cfg.Token)
				}
				if cfg.Username != "" {
					t.Errorf("Username = %v, want empty", cfg.Username)
				}
			},
		},
		{
			name:    "handles non-existent config",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "prompt-vault", "config.yaml")

			// Create manager
			m := &Manager{
				configPath: configPath,
			}

			if tt.setup != nil {
				tt.setup(t, configPath)
			}

			err := m.RemoveToken()
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, configPath)
			}
		})
	}
}

func TestNewManager(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set test HOME
	testHome := t.TempDir()
	os.Setenv("HOME", testHome)

	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}

	expectedPath := filepath.Join(testHome, ".config", "prompt-vault", "config.yaml")
	if m.configPath != expectedPath {
		t.Errorf("configPath = %v, want %v", m.configPath, expectedPath)
	}
}

func TestNewManagerWithPath(t *testing.T) {
	customPath := "/custom/path/config.yaml"
	m := NewManagerWithPath(customPath)

	if m == nil {
		t.Fatal("NewManagerWithPath() returned nil")
	}

	if m.configPath != customPath {
		t.Errorf("configPath = %v, want %v", m.configPath, customPath)
	}
}

func TestManager_ValidateToken(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func(t *testing.T, configPath string)
		setupGist      func(t *testing.T) *gist.Client
		wantUsername   string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: "validates token successfully",
			setupConfig: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Token:    "ghp_validtoken123",
					Username: "testuser",
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			setupGist: func(t *testing.T) *gist.Client {
				// This would be mocked in real implementation
				return nil
			},
			wantUsername: "testuser",
			wantErr:      false,
		},
		{
			name:           "returns error when no token stored",
			wantErr:        true,
			wantErrMessage: "no token found",
		},
		{
			name: "handles invalid token",
			setupConfig: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Token:    "invalid_token",
					Username: "",
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			wantErr:        true,
			wantErrMessage: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip actual API tests for now - would need mock
			if tt.name == "validates token successfully" {
				t.Skip("Skipping API test - requires mock implementation")
			}
			if tt.name == "handles invalid token" {
				t.Skip("Skipping API test - requires mock implementation")
			}

			// Create temp config directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "prompt-vault", "config.yaml")

			// Create manager
			m := &Manager{
				configPath: configPath,
			}

			if tt.setupConfig != nil {
				tt.setupConfig(t, configPath)
			}

			username, err := m.ValidateToken(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("ValidateToken() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}

			if !tt.wantErr && username != tt.wantUsername {
				t.Errorf("ValidateToken() username = %v, want %v", username, tt.wantUsername)
			}
		})
	}
}

func TestManager_GetCurrentUser(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func(t *testing.T, configPath string)
		wantUsername   string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: "retrieves cached username",
			setupConfig: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Token:    "ghp_validtoken123",
					Username: "cacheduser",
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			wantUsername: "cacheduser",
			wantErr:      false,
		},
		{
			name:           "returns error when no token stored",
			wantErr:        true,
			wantErrMessage: "no token found",
		},
		{
			name: "fetches username when not cached",
			setupConfig: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Token:    "ghp_validtoken123",
					Username: "", // Empty username
				}
				if err := cfg.Save(configPath); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
			},
			wantUsername: "",    // Would be fetched from API
			wantErr:      false, // In real test, this would depend on API
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip API test for now
			if tt.name == "fetches username when not cached" {
				t.Skip("Skipping API test - requires mock implementation")
			}

			// Create temp config directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "prompt-vault", "config.yaml")

			// Create manager
			m := &Manager{
				configPath: configPath,
			}

			if tt.setupConfig != nil {
				tt.setupConfig(t, configPath)
			}

			username, err := m.GetCurrentUser(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("GetCurrentUser() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}

			if !tt.wantErr && username != tt.wantUsername {
				t.Errorf("GetCurrentUser() username = %v, want %v", username, tt.wantUsername)
			}
		})
	}
}

// Helper function for string contains
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
