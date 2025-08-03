package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/grigri/pv/internal/errors"
)

func TestFileStore_SaveAndGetToken(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the config directory for testing
	originalGetConfigDir := getConfigDir
	getConfigDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		getConfigDir = originalGetConfigDir
	}()

	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Test saving a token
	testToken := "ghp_testtoken123456789"
	err = store.SaveToken(testToken)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Test retrieving the token
	retrievedToken, err := store.GetToken()
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if retrievedToken != testToken {
		t.Errorf("Retrieved token doesn't match: got %q, want %q", retrievedToken, testToken)
	}

	// Test file permissions on Unix-like systems
	if runtime.GOOS != "windows" {
		configPath := store.GetConfigPath()
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("Failed to stat config file: %v", err)
		}

		// Check that file has 0600 permissions
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("Config file has wrong permissions: got %o, want %o", perm, 0600)
		}
	}
}

func TestFileStore_DeleteToken(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the config directory for testing
	originalGetConfigDir := getConfigDir
	getConfigDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		getConfigDir = originalGetConfigDir
	}()

	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Save a token first
	testToken := "ghp_testtoken123456789"
	err = store.SaveToken(testToken)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Delete the token
	err = store.DeleteToken()
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Try to get the token - should return ErrTokenNotFound
	_, err = store.GetToken()
	if err != errors.ErrTokenNotFound {
		t.Errorf("Expected ErrTokenNotFound, got: %v", err)
	}
}

func TestFileStore_GetTokenWhenNoConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the config directory for testing
	originalGetConfigDir := getConfigDir
	getConfigDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		getConfigDir = originalGetConfigDir
	}()

	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Try to get token when no config exists
	_, err = store.GetToken()
	if err != errors.ErrTokenNotFound {
		t.Errorf("Expected ErrTokenNotFound, got: %v", err)
	}
}

func TestObfuscation(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple token", "ghp_123456789"},
		{"token with special chars", "ghp_abc!@#$%^&*()_+"},
		{"long token", "ghp_verylongtokenwithalotofcharactersfortesting123456789"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Obfuscate and then deobfuscate
			obfuscated := obfuscate(tc.input)
			deobfuscated := deobfuscate(obfuscated)

			if deobfuscated != tc.input {
				t.Errorf("Obfuscation round trip failed: got %q, want %q", deobfuscated, tc.input)
			}

			// Ensure obfuscated value is different from input (unless empty)
			if tc.input != "" && obfuscated == tc.input {
				t.Errorf("Obfuscated value should be different from input")
			}
		})
	}
}

func TestConfigDirCreation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "subdir", "config")

	// Override the config directory for testing
	originalGetConfigDir := getConfigDir
	getConfigDir = func() (string, error) {
		return nonExistentDir, nil
	}
	defer func() {
		getConfigDir = originalGetConfigDir
	}()

	// Create store - should create the directory
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}

	// Check that directory was created
	if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created")
	}

	// Directory should have 0700 permissions on Unix-like systems
	if runtime.GOOS != "windows" {
		info, err := os.Stat(nonExistentDir)
		if err != nil {
			t.Fatalf("Failed to stat config directory: %v", err)
		}

		perm := info.Mode().Perm()
		if perm != 0700 {
			t.Errorf("Config directory has wrong permissions: got %o, want %o", perm, 0700)
		}
	}

	// Test that we can save a token in the newly created directory
	err = store.SaveToken("test_token")
	if err != nil {
		t.Errorf("Failed to save token in newly created directory: %v", err)
	}
}
