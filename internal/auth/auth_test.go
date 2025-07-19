package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grigri201/prompt-vault/internal/config"
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
			
			err := m.SaveToken(tt.token, tt.username)
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
			
			token, username, err := m.GetToken()
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