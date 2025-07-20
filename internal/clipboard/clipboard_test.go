package clipboard

import (
	"os"
	"runtime"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty text returns error",
			text:    "",
			wantErr: true,
			errMsg:  "cannot copy empty text",
		},
		{
			name:    "copies simple text",
			text:    "Hello, World!",
			wantErr: false,
		},
		{
			name:    "copies multiline text",
			text:    "Line 1\nLine 2\nLine 3",
			wantErr: false,
		},
		{
			name:    "copies text with special characters",
			text:    "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip clipboard tests in CI environment
			if os.Getenv("CI") != "" {
				t.Skip("Skipping clipboard test in CI environment")
			}

			// Check if clipboard is available on this platform
			if !IsAvailable() && !tt.wantErr {
				t.Skip("Clipboard not available on this platform")
			}

			err := Copy(tt.text)

			if (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, want containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestIsAvailable(t *testing.T) {
	// This test verifies that IsAvailable returns a boolean without error
	available := IsAvailable()

	// Log the result for debugging
	t.Logf("Clipboard available on %s: %v", runtime.GOOS, available)

	// On CI or test environments, we just check that it doesn't panic
	// The actual availability depends on the system configuration
}

func TestPlatformSpecific(t *testing.T) {
	// Skip in CI
	if os.Getenv("CI") != "" {
		t.Skip("Skipping platform-specific tests in CI")
	}

	switch runtime.GOOS {
	case "darwin":
		t.Run("macOS pbcopy", func(t *testing.T) {
			if !isCommandAvailable("pbcopy") {
				t.Skip("pbcopy not available")
			}

			err := copyDarwin("test text")
			if err != nil {
				t.Errorf("copyDarwin() error = %v", err)
			}
		})

	case "linux":
		t.Run("Linux clipboard utilities", func(t *testing.T) {
			// Test available clipboard utilities
			utilities := []struct {
				name string
				fn   func(string) error
			}{
				{"xclip", func(text string) error {
					if !isCommandAvailable("xclip") {
						t.Skip("xclip not available")
					}
					return copyLinux(text)
				}},
				{"xsel", func(text string) error {
					if !isCommandAvailable("xsel") {
						t.Skip("xsel not available")
					}
					return copyLinux(text)
				}},
				{"wl-copy", func(text string) error {
					if !isCommandAvailable("wl-copy") {
						t.Skip("wl-copy not available")
					}
					return copyLinux(text)
				}},
			}

			hasAny := false
			for _, util := range utilities {
				if isCommandAvailable(util.name) {
					hasAny = true
					t.Run(util.name, func(t *testing.T) {
						err := util.fn("test text")
						if err != nil {
							t.Errorf("%s error = %v", util.name, err)
						}
					})
				}
			}

			if !hasAny {
				err := copyLinux("test text")
				if err == nil || !contains(err.Error(), "no clipboard utility found") {
					t.Errorf("Expected error about missing clipboard utility, got: %v", err)
				}
			}
		})

	case "windows":
		t.Run("Windows clipboard", func(t *testing.T) {
			err := copyWindows("test text")
			if err != nil {
				t.Errorf("copyWindows() error = %v", err)
			}
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "common command exists",
			command:  "echo",
			expected: true,
		},
		{
			name:     "non-existent command",
			command:  "definitely-not-a-real-command-xyz123",
			expected: false,
		},
	}

	// Skip on Windows as 'which' may not be available
	if runtime.GOOS == "windows" {
		t.Skip("Skipping 'which' command test on Windows")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommandAvailable(tt.command)
			if result != tt.expected {
				t.Errorf("isCommandAvailable(%q) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestUnsupportedPlatform(t *testing.T) {
	// Save original GOOS
	originalGOOS := runtime.GOOS
	defer func() {
		// Note: We can't actually change runtime.GOOS at runtime,
		// but we can test the default case by calling the function directly
		_ = originalGOOS
	}()

	// Test that we handle unknown platforms gracefully
	// This would need build tags or mocking to test properly
	t.Skip("Cannot test unsupported platform without build tags")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && len(substr) > 0 &&
			(s[:len(substr)] == substr || contains(s[1:], substr)))
}

// MockCommand is used for testing command execution
type MockCommand struct {
	name string
	args []string
	err  error
}

func (m *MockCommand) Run() error {
	return m.err
}

// Test that the clipboard package exports the expected functions
func TestExportedFunctions(t *testing.T) {
	// This test ensures that the public API is maintained
	var _ func(string) error = Copy
	var _ func() bool = IsAvailable

	// Verify that the functions can be called
	_ = IsAvailable()

	// Copy with empty string should return error
	err := Copy("")
	if err == nil {
		t.Error("Copy(\"\") should return an error")
	}
}
