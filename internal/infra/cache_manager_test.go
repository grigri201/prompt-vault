package infra

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	appErrors "github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
)

func TestNewCacheManager(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, error)
		expectError bool
	}{
		{
			name: "successful creation",
			setupFunc: func() (string, error) {
				tempDir := t.TempDir()
				return filepath.Join(tempDir, "pv"), nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test cache directory
			cacheDir, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Create cache manager directly for testing
			manager := &CacheManager{cacheDir: cacheDir}

			if tt.expectError {
				// Test with invalid cache directory setup
				manager.cacheDir = "/invalid/path/that/cannot/be/created"
				err := manager.EnsureCacheDir()
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if manager == nil {
					t.Errorf("Expected CacheManager instance but got nil")
				}
				if manager.cacheDir != cacheDir {
					t.Errorf("Expected cache dir %s, got %s", cacheDir, manager.cacheDir)
				}
			}
		})
	}
}

func TestCacheManager_EnsureCacheDir(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) error
		expectError bool
		checkFunc   func(string) error
	}{
		{
			name:        "create new cache directory",
			setupFunc:   nil, // No setup needed
			expectError: false,
			checkFunc: func(cacheDir string) error {
				// Check main directory exists
				if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
					return fmt.Errorf("cache directory not created")
				}

				// Check prompts subdirectory exists
				promptsDir := filepath.Join(cacheDir, "prompts")
				if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
					return fmt.Errorf("prompts directory not created")
				}

				// Check permissions on non-Windows systems
				if runtime.GOOS != "windows" {
					stat, err := os.Stat(cacheDir)
					if err != nil {
						return err
					}
					if stat.Mode().Perm() != 0700 {
						return fmt.Errorf("incorrect cache directory permissions: got %v, want %v", stat.Mode().Perm(), 0700)
					}

					promptsStat, err := os.Stat(promptsDir)
					if err != nil {
						return err
					}
					if promptsStat.Mode().Perm() != 0700 {
						return fmt.Errorf("incorrect prompts directory permissions: got %v, want %v", promptsStat.Mode().Perm(), 0700)
					}
				}

				return nil
			},
		},
		{
			name: "existing directory with wrong permissions",
			setupFunc: func(cacheDir string) error {
				// Create directory with wrong permissions
				if err := os.MkdirAll(cacheDir, 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(cacheDir, "prompts"), 0755)
			},
			expectError: false,
			checkFunc: func(cacheDir string) error {
				// Should fix permissions
				if runtime.GOOS != "windows" {
					stat, err := os.Stat(cacheDir)
					if err != nil {
						return err
					}
					if stat.Mode().Perm() != 0700 {
						return fmt.Errorf("permissions not fixed: got %v, want %v", stat.Mode().Perm(), 0700)
					}
				}
				return nil
			},
		},
		{
			name: "cannot create directory (permission denied)",
			setupFunc: func(cacheDir string) error {
				// Create a file where directory should be (to simulate permission issue)
				parentDir := filepath.Dir(cacheDir)
				if err := os.MkdirAll(parentDir, 0755); err != nil {
					return err
				}
				// Create a file with the same name as the cache directory
				return os.WriteFile(cacheDir, []byte("block"), 0644)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, "test-cache")

			// Setup test conditions
			if tt.setupFunc != nil {
				if err := tt.setupFunc(cacheDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Create cache manager with test directory
			manager := &CacheManager{cacheDir: cacheDir}

			// Test EnsureCacheDir
			err := manager.EnsureCacheDir()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Run additional checks
				if tt.checkFunc != nil {
					if checkErr := tt.checkFunc(cacheDir); checkErr != nil {
						t.Errorf("Check failed: %v", checkErr)
					}
				}
			}
		})
	}
}

func TestCacheManager_LoadIndex(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) error
		expectError bool
		errorType   appErrors.ErrorType
		checkFunc   func(*model.Index) error
	}{
		{
			name: "load valid index",
			setupFunc: func(cacheDir string) error {
				// Create cache directory
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					return err
				}

				// Create valid index file
				index := &model.Index{
					Prompts: []model.IndexedPrompt{
						{
							GistURL: "https://gist.github.com/test/test-id",
							Name:    "Test Prompt",
							Author:  "test-author",
						},
					},
					LastUpdated: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				}

				data, err := json.MarshalIndent(index, "", "  ")
				if err != nil {
					return err
				}

				indexPath := filepath.Join(cacheDir, "index.json")
				return os.WriteFile(indexPath, data, 0600)
			},
			expectError: false,
			checkFunc: func(index *model.Index) error {
				if len(index.Prompts) != 1 {
					return fmt.Errorf("expected 1 prompt, got %d", len(index.Prompts))
				}
				if index.Prompts[0].Name != "Test Prompt" {
					return fmt.Errorf("unexpected prompt name: %s", index.Prompts[0].Name)
				}
				return nil
			},
		},
		{
			name:        "index file not found",
			setupFunc:   nil, // No setup - file won't exist
			expectError: true,
			errorType:   appErrors.ErrStorage,
		},
		{
			name: "corrupted index file (invalid JSON)",
			setupFunc: func(cacheDir string) error {
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					return err
				}
				indexPath := filepath.Join(cacheDir, "index.json")
				return os.WriteFile(indexPath, []byte("invalid json {"), 0600)
			},
			expectError: true,
			errorType:   appErrors.ErrStorage,
		},
		{
			name: "empty index file",
			setupFunc: func(cacheDir string) error {
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					return err
				}
				indexPath := filepath.Join(cacheDir, "index.json")
				return os.WriteFile(indexPath, []byte(""), 0600)
			},
			expectError: true,
			errorType:   appErrors.ErrStorage,
		},
		{
			name: "permission denied reading index",
			setupFunc: func(cacheDir string) error {
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					return err
				}
				indexPath := filepath.Join(cacheDir, "index.json")
				if err := os.WriteFile(indexPath, []byte("{}"), 0600); err != nil {
					return err
				}
				// Remove read permissions (only on Unix-like systems)
				if runtime.GOOS != "windows" {
					return os.Chmod(indexPath, 0000)
				}
				return nil
			},
			expectError: true,
			errorType:   appErrors.ErrStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, "test-cache")

			// Setup test conditions
			if tt.setupFunc != nil {
				if err := tt.setupFunc(cacheDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Create cache manager
			manager := &CacheManager{cacheDir: cacheDir}

			// Test LoadIndex
			index, err := manager.LoadIndex()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else {
					// Check error type
					var appErr appErrors.AppError
					if errors.As(err, &appErr) && appErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if index == nil {
					t.Errorf("Expected index but got nil")
				}

				// Run additional checks
				if tt.checkFunc != nil {
					if checkErr := tt.checkFunc(index); checkErr != nil {
						t.Errorf("Check failed: %v", checkErr)
					}
				}
			}
		})
	}
}

func TestCacheManager_SaveIndex(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) error
		index       *model.Index
		expectError bool
		errorType   appErrors.ErrorType
		checkFunc   func(string) error
	}{
		{
			name: "save valid index",
			index: &model.Index{
				Prompts: []model.IndexedPrompt{
					{
						GistURL: "https://gist.github.com/test/test-id",
						Name:    "Test Prompt",
						Author:  "test-author",
					},
				},
				LastUpdated: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expectError: false,
			checkFunc: func(cacheDir string) error {
				// Verify file was created
				indexPath := filepath.Join(cacheDir, "index.json")
				if _, err := os.Stat(indexPath); os.IsNotExist(err) {
					return fmt.Errorf("index file not created")
				}

				// Verify content
				data, err := os.ReadFile(indexPath)
				if err != nil {
					return fmt.Errorf("failed to read saved index: %v", err)
				}

				var savedIndex model.Index
				if err := json.Unmarshal(data, &savedIndex); err != nil {
					return fmt.Errorf("saved index is not valid JSON: %v", err)
				}

				if len(savedIndex.Prompts) != 1 {
					return fmt.Errorf("expected 1 prompt, got %d", len(savedIndex.Prompts))
				}

				// Check permissions on Unix-like systems
				if runtime.GOOS != "windows" {
					stat, err := os.Stat(indexPath)
					if err != nil {
						return err
					}
					if stat.Mode().Perm() != 0600 {
						return fmt.Errorf("incorrect file permissions: got %v, want %v", stat.Mode().Perm(), 0600)
					}
				}

				return nil
			},
		},
		{
			name: "save empty index",
			index: &model.Index{
				Prompts:     []model.IndexedPrompt{},
				LastUpdated: time.Now(),
			},
			expectError: false,
		},
		{
			name: "disk space issue simulation",
			setupFunc: func(cacheDir string) error {
				// Create the cache directory structure
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					return err
				}
				// On Unix-like systems, make directory read-only to simulate disk issues
				if runtime.GOOS != "windows" {
					return os.Chmod(cacheDir, 0500) // read + execute only
				}
				return nil
			},
			index: &model.Index{
				Prompts:     []model.IndexedPrompt{},
				LastUpdated: time.Now(),
			},
			expectError: true,
			errorType:   appErrors.ErrStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, "test-cache")

			// Setup test conditions
			if tt.setupFunc != nil {
				if err := tt.setupFunc(cacheDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Create cache manager
			manager := &CacheManager{cacheDir: cacheDir}

			// Test SaveIndex
			err := manager.SaveIndex(tt.index)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else {
					// Check error type
					var appErr appErrors.AppError
					if errors.As(err, &appErr) && appErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Run additional checks
				if tt.checkFunc != nil {
					if checkErr := tt.checkFunc(cacheDir); checkErr != nil {
						t.Errorf("Check failed: %v", checkErr)
					}
				}
			}
		})
	}
}

func TestCacheManager_LoadContent(t *testing.T) {
	testContent := `name: Test Prompt
author: test-author
description: A test prompt
content: |
  This is test content
variables:
  - name: test_var
    description: Test variable`

	tests := []struct {
		name        string
		gistID      string
		setupFunc   func(string) error
		expectError bool
		errorType   appErrors.ErrorType
		checkFunc   func(string) error
	}{
		{
			name:   "load existing content",
			gistID: "test-gist-123",
			setupFunc: func(cacheDir string) error {
				promptsDir := filepath.Join(cacheDir, "prompts")
				if err := os.MkdirAll(promptsDir, 0700); err != nil {
					return err
				}
				contentPath := filepath.Join(promptsDir, "test-gist-123.yaml")
				return os.WriteFile(contentPath, []byte(testContent), 0600)
			},
			expectError: false,
			checkFunc: func(content string) error {
				if content != testContent {
					return fmt.Errorf("content mismatch: got %q, want %q", content, testContent)
				}
				return nil
			},
		},
		{
			name:        "content file not found",
			gistID:      "non-existent-gist",
			setupFunc:   nil,
			expectError: true,
			errorType:   appErrors.ErrStorage,
		},
		{
			name:   "empty content file",
			gistID: "empty-gist",
			setupFunc: func(cacheDir string) error {
				promptsDir := filepath.Join(cacheDir, "prompts")
				if err := os.MkdirAll(promptsDir, 0700); err != nil {
					return err
				}
				contentPath := filepath.Join(promptsDir, "empty-gist.yaml")
				return os.WriteFile(contentPath, []byte(""), 0600)
			},
			expectError: false,
			checkFunc: func(content string) error {
				if content != "" {
					return fmt.Errorf("expected empty content, got %q", content)
				}
				return nil
			},
		},
		{
			name:   "permission denied reading content",
			gistID: "protected-gist",
			setupFunc: func(cacheDir string) error {
				promptsDir := filepath.Join(cacheDir, "prompts")
				if err := os.MkdirAll(promptsDir, 0700); err != nil {
					return err
				}
				contentPath := filepath.Join(promptsDir, "protected-gist.yaml")
				if err := os.WriteFile(contentPath, []byte(testContent), 0600); err != nil {
					return err
				}
				// Remove read permissions (only on Unix-like systems)
				if runtime.GOOS != "windows" {
					return os.Chmod(contentPath, 0000)
				}
				return nil
			},
			expectError: true,
			errorType:   appErrors.ErrStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, "test-cache")

			// Setup test conditions
			if tt.setupFunc != nil {
				if err := tt.setupFunc(cacheDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Create cache manager
			manager := &CacheManager{cacheDir: cacheDir}

			// Test LoadContent
			content, err := manager.LoadContent(tt.gistID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else {
					// Check error type
					var appErr appErrors.AppError
					if errors.As(err, &appErr) && appErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Run additional checks
				if tt.checkFunc != nil {
					if checkErr := tt.checkFunc(content); checkErr != nil {
						t.Errorf("Check failed: %v", checkErr)
					}
				}
			}
		})
	}
}

func TestCacheManager_SaveContent(t *testing.T) {
	testContent := `name: Test Prompt
author: test-author
description: A test prompt
content: |
  This is test content`

	tests := []struct {
		name        string
		gistID      string
		content     string
		setupFunc   func(string) error
		expectError bool
		errorType   appErrors.ErrorType
		checkFunc   func(string, string) error
	}{
		{
			name:        "save new content",
			gistID:      "new-gist-123",
			content:     testContent,
			expectError: false,
			checkFunc: func(cacheDir, gistID string) error {
				// Verify file was created
				contentPath := filepath.Join(cacheDir, "prompts", gistID+".yaml")
				if _, err := os.Stat(contentPath); os.IsNotExist(err) {
					return fmt.Errorf("content file not created")
				}

				// Verify content
				data, err := os.ReadFile(contentPath)
				if err != nil {
					return fmt.Errorf("failed to read saved content: %v", err)
				}

				if string(data) != testContent {
					return fmt.Errorf("content mismatch: got %q, want %q", string(data), testContent)
				}

				// Check permissions on Unix-like systems
				if runtime.GOOS != "windows" {
					stat, err := os.Stat(contentPath)
					if err != nil {
						return err
					}
					if stat.Mode().Perm() != 0600 {
						return fmt.Errorf("incorrect file permissions: got %v, want %v", stat.Mode().Perm(), 0600)
					}
				}

				return nil
			},
		},
		{
			name:        "save empty content",
			gistID:      "empty-gist",
			content:     "",
			expectError: false,
			checkFunc: func(cacheDir, gistID string) error {
				contentPath := filepath.Join(cacheDir, "prompts", gistID+".yaml")
				data, err := os.ReadFile(contentPath)
				if err != nil {
					return fmt.Errorf("failed to read saved content: %v", err)
				}
				if string(data) != "" {
					return fmt.Errorf("expected empty content, got %q", string(data))
				}
				return nil
			},
		},
		{
			name:    "overwrite existing content",
			gistID:  "existing-gist",
			content: "new content",
			setupFunc: func(cacheDir string) error {
				promptsDir := filepath.Join(cacheDir, "prompts")
				if err := os.MkdirAll(promptsDir, 0700); err != nil {
					return err
				}
				contentPath := filepath.Join(promptsDir, "existing-gist.yaml")
				return os.WriteFile(contentPath, []byte("old content"), 0600)
			},
			expectError: false,
			checkFunc: func(cacheDir, gistID string) error {
				contentPath := filepath.Join(cacheDir, "prompts", gistID+".yaml")
				data, err := os.ReadFile(contentPath)
				if err != nil {
					return fmt.Errorf("failed to read saved content: %v", err)
				}
				if string(data) != "new content" {
					return fmt.Errorf("content not overwritten: got %q", string(data))
				}
				return nil
			},
		},
		{
			name:    "permission denied saving content",
			gistID:  "protected-gist",
			content: testContent,
			setupFunc: func(cacheDir string) error {
				promptsDir := filepath.Join(cacheDir, "prompts")
				if err := os.MkdirAll(promptsDir, 0700); err != nil {
					return err
				}
				// Create a file with the same name to block the write
				targetPath := filepath.Join(promptsDir, "protected-gist.yaml")
				if err := os.WriteFile(targetPath, []byte("existing"), 0600); err != nil {
					return err
				}
				// Make the existing file read-only on Unix-like systems
				if runtime.GOOS != "windows" {
					return os.Chmod(targetPath, 0400) // read only
				}
				return nil
			},
			expectError: false, // The config.WriteFileWithPermissions should overwrite the file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, "test-cache")

			// Setup test conditions
			if tt.setupFunc != nil {
				if err := tt.setupFunc(cacheDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Create cache manager
			manager := &CacheManager{cacheDir: cacheDir}

			// Test SaveContent
			err := manager.SaveContent(tt.gistID, tt.content)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else {
					// Check error type
					var appErr appErrors.AppError
					if errors.As(err, &appErr) && appErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Run additional checks
				if tt.checkFunc != nil {
					if checkErr := tt.checkFunc(cacheDir, tt.gistID); checkErr != nil {
						t.Errorf("Check failed: %v", checkErr)
					}
				}
			}
		})
	}
}

func TestCacheManager_GetCacheInfo(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string) error
		checkFunc func(*model.CacheInfo, string) error
	}{
		{
			name: "cache with index and content",
			setupFunc: func(cacheDir string) error {
				// Create cache structure
				promptsDir := filepath.Join(cacheDir, "prompts")
				if err := os.MkdirAll(promptsDir, 0700); err != nil {
					return err
				}

				// Create index file
				index := &model.Index{
					Prompts: []model.IndexedPrompt{
						{GistURL: "https://gist.github.com/test/gist1", Name: "Prompt 1"},
						{GistURL: "https://gist.github.com/test/gist2", Name: "Prompt 2"},
					},
					LastUpdated: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				}
				data, err := json.MarshalIndent(index, "", "  ")
				if err != nil {
					return err
				}
				indexPath := filepath.Join(cacheDir, "index.json")
				if err := os.WriteFile(indexPath, data, 0600); err != nil {
					return err
				}

				// Create content files
				content1Path := filepath.Join(promptsDir, "gist1.yaml")
				if err := os.WriteFile(content1Path, []byte("content1"), 0600); err != nil {
					return err
				}
				content2Path := filepath.Join(promptsDir, "gist2.yaml")
				return os.WriteFile(content2Path, []byte("content2"), 0600)
			},
			checkFunc: func(info *model.CacheInfo, cacheDir string) error {
				if info.TotalPrompts != 2 {
					return fmt.Errorf("expected 2 prompts, got %d", info.TotalPrompts)
				}
				expectedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
				if !info.LastUpdated.Equal(expectedTime) {
					return fmt.Errorf("expected last updated %v, got %v", expectedTime, info.LastUpdated)
				}
				if info.CacheSize <= 0 {
					return fmt.Errorf("expected positive cache size, got %d", info.CacheSize)
				}
				return nil
			},
		},
		{
			name:      "empty cache directory",
			setupFunc: nil,
			checkFunc: func(info *model.CacheInfo, cacheDir string) error {
				if info.TotalPrompts != 0 {
					return fmt.Errorf("expected 0 prompts, got %d", info.TotalPrompts)
				}
				if !info.LastUpdated.IsZero() {
					return fmt.Errorf("expected zero time, got %v", info.LastUpdated)
				}
				if info.CacheSize != 0 {
					return fmt.Errorf("expected 0 cache size, got %d", info.CacheSize)
				}
				return nil
			},
		},
		{
			name: "corrupted index file",
			setupFunc: func(cacheDir string) error {
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					return err
				}
				// Create corrupted index
				indexPath := filepath.Join(cacheDir, "index.json")
				return os.WriteFile(indexPath, []byte("invalid json"), 0600)
			},
			checkFunc: func(info *model.CacheInfo, cacheDir string) error {
				// Should still return info with zero values for prompts
				if info.TotalPrompts != 0 {
					return fmt.Errorf("expected 0 prompts for corrupted index, got %d", info.TotalPrompts)
				}
				if !info.LastUpdated.IsZero() {
					return fmt.Errorf("expected zero time for corrupted index, got %v", info.LastUpdated)
				}
				// CacheSize should still be calculated from the corrupted file
				if info.CacheSize <= 0 {
					return fmt.Errorf("expected positive cache size even with corrupted index, got %d", info.CacheSize)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, "test-cache")

			// Setup test conditions
			if tt.setupFunc != nil {
				if err := tt.setupFunc(cacheDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Create cache manager
			manager := &CacheManager{cacheDir: cacheDir}

			// Test GetCacheInfo
			info, err := manager.GetCacheInfo()

			// GetCacheInfo should never return an error
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if info == nil {
				t.Errorf("Expected CacheInfo but got nil")
			}

			// Run additional checks
			if tt.checkFunc != nil {
				if checkErr := tt.checkFunc(info, cacheDir); checkErr != nil {
					t.Errorf("Check failed: %v", checkErr)
				}
			}
		})
	}
}

func TestCacheManager_calculateDirectorySize(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string) error
		checkFunc func(int64, string) error
	}{
		{
			name: "directory with files",
			setupFunc: func(testDir string) error {
				// Create some files with known sizes
				if err := os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("hello"), 0600); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("world"), 0600); err != nil {
					return err
				}
				// Create subdirectory with file
				subDir := filepath.Join(testDir, "subdir")
				if err := os.MkdirAll(subDir, 0700); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("test"), 0600)
			},
			checkFunc: func(size int64, testDir string) error {
				// Expected: "hello" (5) + "world" (5) + "test" (4) = 14 bytes
				if size != 14 {
					return fmt.Errorf("expected size 14, got %d", size)
				}
				return nil
			},
		},
		{
			name:      "empty directory",
			setupFunc: nil,
			checkFunc: func(size int64, testDir string) error {
				if size != 0 {
					return fmt.Errorf("expected size 0 for empty directory, got %d", size)
				}
				return nil
			},
		},
		{
			name: "directory with permission issues",
			setupFunc: func(testDir string) error {
				// Create a file
				if err := os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("content"), 0600); err != nil {
					return err
				}
				// Create a subdirectory we can't access
				subDir := filepath.Join(testDir, "restricted")
				if err := os.MkdirAll(subDir, 0700); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(subDir, "secret.txt"), []byte("secret"), 0600); err != nil {
					return err
				}
				// Remove permissions on Unix-like systems
				if runtime.GOOS != "windows" {
					return os.Chmod(subDir, 0000)
				}
				return nil
			},
			checkFunc: func(size int64, testDir string) error {
				// Should count accessible files and gracefully skip inaccessible ones
				// Expected at least "content" (7 bytes)
				if size < 7 {
					return fmt.Errorf("expected at least 7 bytes, got %d", size)
				}
				// Restore permissions for cleanup
				if runtime.GOOS != "windows" {
					subDir := filepath.Join(testDir, "restricted")
					os.Chmod(subDir, 0700)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for testing
			testDir := t.TempDir()

			// Setup test conditions
			if tt.setupFunc != nil {
				if err := tt.setupFunc(testDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Create cache manager
			manager := &CacheManager{cacheDir: testDir}

			// Test calculateDirectorySize
			size, err := manager.calculateDirectorySize(testDir)

			// Should never return error due to graceful error handling
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Run additional checks
			if tt.checkFunc != nil {
				if checkErr := tt.checkFunc(size, testDir); checkErr != nil {
					t.Errorf("Check failed: %v", checkErr)
				}
			}
		})
	}
}

// TestCacheManager_EdgeCases tests various edge cases and boundary conditions
func TestCacheManager_EdgeCases(t *testing.T) {
	t.Run("very long gist ID", func(t *testing.T) {
		tempDir := t.TempDir()
		cacheDir := filepath.Join(tempDir, "test-cache")
		manager := &CacheManager{cacheDir: cacheDir}

		// Test with very long gist ID (but within filesystem limits)
		// Most filesystems support up to 255 bytes for filename, so let's use something reasonable
		longGistID := strings.Repeat("a", 200)
		content := "test content"

		err := manager.SaveContent(longGistID, content)
		if err != nil {
			// If the filesystem doesn't support long filenames, that's expected
			t.Logf("Long gist ID not supported on this filesystem: %v", err)
			return
		}

		// Try to load it back
		loadedContent, err := manager.LoadContent(longGistID)
		if err != nil {
			t.Errorf("Failed to load content with long gist ID: %v", err)
		}
		if loadedContent != content {
			t.Errorf("Content mismatch: got %q, want %q", loadedContent, content)
		}
	})

	t.Run("special characters in gist ID", func(t *testing.T) {
		tempDir := t.TempDir()
		cacheDir := filepath.Join(tempDir, "test-cache")
		manager := &CacheManager{cacheDir: cacheDir}

		// Test with gist ID containing special characters (but valid for filename)
		specialGistID := "gist-123_abc.def"
		content := "special content"

		err := manager.SaveContent(specialGistID, content)
		if err != nil {
			t.Errorf("Unexpected error with special gist ID: %v", err)
		}

		loadedContent, err := manager.LoadContent(specialGistID)
		if err != nil {
			t.Errorf("Failed to load content with special gist ID: %v", err)
		}
		if loadedContent != content {
			t.Errorf("Content mismatch: got %q, want %q", loadedContent, content)
		}
	})

	t.Run("concurrent access simulation", func(t *testing.T) {
		tempDir := t.TempDir()
		cacheDir := filepath.Join(tempDir, "test-cache")
		manager := &CacheManager{cacheDir: cacheDir}

		// Simulate concurrent index updates
		index1 := &model.Index{
			Prompts:     []model.IndexedPrompt{{GistURL: "https://gist.github.com/test/1"}},
			LastUpdated: time.Now(),
		}
		index2 := &model.Index{
			Prompts:     []model.IndexedPrompt{{GistURL: "https://gist.github.com/test/2"}},
			LastUpdated: time.Now(),
		}

		// Save first index
		if err := manager.SaveIndex(index1); err != nil {
			t.Errorf("Failed to save first index: %v", err)
		}

		// Save second index (overwrites first)
		if err := manager.SaveIndex(index2); err != nil {
			t.Errorf("Failed to save second index: %v", err)
		}

		// Load and verify we got the second index
		loadedIndex, err := manager.LoadIndex()
		if err != nil {
			t.Errorf("Failed to load index: %v", err)
		}
		if len(loadedIndex.Prompts) != 1 || loadedIndex.Prompts[0].GistURL != "https://gist.github.com/test/2" {
			t.Errorf("Expected second index, got %+v", loadedIndex)
		}
	})
}

// TestCacheManager_DiskSpaceHandling tests handling of disk space issues
func TestCacheManager_DiskSpaceHandling(t *testing.T) {
	// This test is platform-specific and may not work on all systems
	if runtime.GOOS == "windows" {
		t.Skip("Disk space simulation not supported on Windows")
	}

	t.Run("no space left on device simulation", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a scenario where we can't create the cache directory
		// by having a file with the same name as the cache directory
		blockedCacheDir := filepath.Join(tempDir, "blocked-cache")
		if err := os.WriteFile(blockedCacheDir, []byte("blocker"), 0644); err != nil {
			t.Fatalf("Failed to create blocker file: %v", err)
		}

		invalidManager := &CacheManager{cacheDir: blockedCacheDir}

		// Try to save content - should fail gracefully because EnsureCacheDir can't create the directory
		err := invalidManager.SaveContent("test-gist", "some content")
		if err == nil {
			t.Errorf("Expected error when cache directory setup fails")
		}

		// Should contain the error about failing to ensure cache directory
		if !strings.Contains(err.Error(), "failed to ensure cache directory") {
			t.Errorf("Expected error about cache directory setup, got: %v", err)
		}
	})
}

// TestCacheManager_RealWorldScenarios tests real-world usage scenarios
func TestCacheManager_RealWorldScenarios(t *testing.T) {
	t.Run("large cache directory with many files", func(t *testing.T) {
		tempDir := t.TempDir()
		cacheDir := filepath.Join(tempDir, "test-cache")
		manager := &CacheManager{cacheDir: cacheDir}

		// Create a realistic cache scenario with many files
		numPrompts := 100
		prompts := make([]model.IndexedPrompt, numPrompts)
		for i := 0; i < numPrompts; i++ {
			gistID := fmt.Sprintf("gist-%03d", i)
			prompts[i] = model.IndexedPrompt{
				GistURL: fmt.Sprintf("https://gist.github.com/test/%s", gistID),
				Name:    fmt.Sprintf("Prompt %d", i),
				Author:  "test-author",
			}

			// Save content for each prompt
			content := fmt.Sprintf("Content for prompt %d", i)
			if err := manager.SaveContent(gistID, content); err != nil {
				t.Fatalf("Failed to save content %d: %v", i, err)
			}
		}

		// Save index
		index := &model.Index{
			Prompts:     prompts,
			LastUpdated: time.Now(),
		}
		if err := manager.SaveIndex(index); err != nil {
			t.Fatalf("Failed to save index: %v", err)
		}

		// Verify cache info
		info, err := manager.GetCacheInfo()
		if err != nil {
			t.Fatalf("Failed to get cache info: %v", err)
		}

		if info.TotalPrompts != numPrompts {
			t.Errorf("Expected %d prompts, got %d", numPrompts, info.TotalPrompts)
		}

		if info.CacheSize <= 0 {
			t.Errorf("Expected positive cache size, got %d", info.CacheSize)
		}

		// Test loading random content
		randomGistID := fmt.Sprintf("gist-%03d", 42)
		content, err := manager.LoadContent(randomGistID)
		if err != nil {
			t.Errorf("Failed to load random content: %v", err)
		}
		expectedContent := "Content for prompt 42"
		if content != expectedContent {
			t.Errorf("Content mismatch: got %q, want %q", content, expectedContent)
		}
	})

	t.Run("cache corruption recovery", func(t *testing.T) {
		tempDir := t.TempDir()
		cacheDir := filepath.Join(tempDir, "test-cache")
		manager := &CacheManager{cacheDir: cacheDir}

		// Save valid content first
		validContent := "valid content"
		if err := manager.SaveContent("valid-gist", validContent); err != nil {
			t.Fatalf("Failed to save valid content: %v", err)
		}

		// Corrupt the index file
		if err := os.MkdirAll(cacheDir, 0700); err != nil {
			t.Fatalf("Failed to create cache dir: %v", err)
		}
		indexPath := filepath.Join(cacheDir, "index.json")
		if err := os.WriteFile(indexPath, []byte("corrupted{json"), 0600); err != nil {
			t.Fatalf("Failed to create corrupted index: %v", err)
		}

		// GetCacheInfo should handle corruption gracefully
		info, err := manager.GetCacheInfo()
		if err != nil {
			t.Errorf("GetCacheInfo should handle corruption gracefully: %v", err)
		}

		// Should report 0 prompts due to corrupted index
		if info.TotalPrompts != 0 {
			t.Errorf("Expected 0 prompts with corrupted index, got %d", info.TotalPrompts)
		}

		// But cache size should still be calculated from existing files
		if info.CacheSize <= 0 {
			t.Errorf("Expected positive cache size, got %d", info.CacheSize)
		}

		// Should still be able to load valid content files
		content, err := manager.LoadContent("valid-gist")
		if err != nil {
			t.Errorf("Failed to load valid content after index corruption: %v", err)
		}
		if content != validContent {
			t.Errorf("Content mismatch after corruption: got %q, want %q", content, validContent)
		}

		// Should be able to save new index to recover
		newIndex := &model.Index{
			Prompts: []model.IndexedPrompt{
				{GistURL: "https://gist.github.com/test/valid-gist", Name: "Valid Prompt"},
			},
			LastUpdated: time.Now(),
		}
		if err := manager.SaveIndex(newIndex); err != nil {
			t.Errorf("Failed to save recovery index: %v", err)
		}

		// Verify recovery
		recoveredInfo, err := manager.GetCacheInfo()
		if err != nil {
			t.Errorf("Failed to get cache info after recovery: %v", err)
		}
		if recoveredInfo.TotalPrompts != 1 {
			t.Errorf("Expected 1 prompt after recovery, got %d", recoveredInfo.TotalPrompts)
		}
	})
}