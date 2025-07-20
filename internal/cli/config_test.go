package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grigri201/prompt-vault/internal/config"
)

func TestConfigCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func(t *testing.T, configPath string)
		expectedOutput []string
		wantErr        bool
	}{
		{
			name: "display config with auth",
			setupConfig: func(t *testing.T, configPath string) {
				cfg := &config.Config{
					Username: "testuser",
					Token:    "test-token",
				}
				if err := cfg.SaveToFile(configPath); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			expectedOutput: []string{
				"Configuration Information:",
				"==========================",
				"Config Directory:",
				"Config File:",
				"Cache Directory:",
				"Current Settings:",
				"GitHub Username: testuser",
				"GitHub Token: ****** (set)",
				"Last Sync: Never",
			},
			wantErr: false,
		},
		{
			name: "display config without auth",
			setupConfig: func(t *testing.T, configPath string) {
				// No config file created - should use defaults
			},
			expectedOutput: []string{
				"Configuration Information:",
				"==========================",
				"Config Directory:",
				"Config File:",
				"Cache Directory:",
				"Current Settings:",
				"GitHub Username:",
				"GitHub Token: (not set)",
				"Last Sync: Never",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()
			configDir := filepath.Join(tempDir, ".config", "prompt-vault")
			configPath := filepath.Join(configDir, "config.yaml")

			// Set HOME to temp directory
			oldHome := os.Getenv("HOME")
			if err := os.Setenv("HOME", tempDir); err != nil {
				t.Fatalf("Failed to set HOME: %v", err)
			}
			defer func() {
				_ = os.Setenv("HOME", oldHome)
			}()

			// Create config directory
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatalf("Failed to create config dir: %v", err)
			}

			// Setup config if needed
			if tt.setupConfig != nil {
				tt.setupConfig(t, configPath)
			}

			// Execute command
			cmd := newConfigCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)

			err := cmd.Execute()

			// Check error
			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output
			output := out.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestConfigCommand_Integration(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "prompt-vault")
	configPath := filepath.Join(configDir, "config.yaml")

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", oldHome)
	}()

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create a config with all fields
	cfg := &config.Config{
		Username: "johndoe",
		Token:    "ghp_1234567890abcdef",
	}
	if err := cfg.SaveToFile(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Execute command
	cmd := newConfigCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Check output
	output := out.String()
	
	// Verify all expected information is present
	expectedStrings := []string{
		"Configuration Information:",
		configDir,
		configPath,
		"GitHub Username: johndoe",
		"GitHub Token: ****** (set)",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but got:\n%s", expected, output)
		}
	}
	
	// Verify token is masked
	if strings.Contains(output, "ghp_1234567890abcdef") {
		t.Errorf("Token should be masked but was found in output")
	}
}

func TestConfigCommand_OutputFormat(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", oldHome)
	}()

	// Execute command
	cmd := newConfigCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Check output format
	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Verify header format
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines of output, got %d", len(lines))
	}
	
	if lines[0] != "Configuration Information:" {
		t.Errorf("Expected first line to be 'Configuration Information:', got %q", lines[0])
	}
	
	if lines[1] != "==========================" {
		t.Errorf("Expected second line to be '==========================', got %q", lines[1])
	}
	
	// Verify sections exist
	expectedSections := []string{
		"Config Directory:",
		"Config File:",
		"Cache Directory:",
		"Current Settings:",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected output to contain section %q", section)
		}
	}
}