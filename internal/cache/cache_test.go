package cache

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/paths"
)

// createTestManager creates a Manager for testing with a specific cache path
func createTestManager(tempDir string, subPath string) *Manager {
	// Since PathManager always appends .cache/prompt-vault/prompts,
	// we need to calculate the home directory that would result in our desired path
	pm := paths.NewPathManagerWithHome(tempDir)
	return NewManagerWithPathManager(pm)
}

// createTestManagerWithExactPath creates a Manager that uses exact paths for testing
func createTestManagerWithExactPath(cachePath string) *Manager {
	// Calculate the home directory that would result in the exact cache path
	// by removing the .cache/prompt-vault/prompts suffix that PathManager adds
	const suffix = ".cache/prompt-vault/prompts"
	homeDir := cachePath

	// If the path ends with the standard suffix, extract the home directory
	if strings.HasSuffix(cachePath, suffix) {
		homeDir = strings.TrimSuffix(cachePath, "/"+suffix)
		homeDir = strings.TrimSuffix(homeDir, suffix) // Handle case without leading slash
	} else {
		// For custom paths, we need to adjust so that when PathManager adds the suffix,
		// we get close to the desired path. This is a limitation of the current design.
		// The best we can do is use the path as-is, knowing it will be modified.
		homeDir = filepath.Dir(filepath.Dir(filepath.Dir(cachePath)))
	}

	pm := paths.NewPathManagerWithHome(homeDir)
	return NewManagerWithPathManager(pm)
}

func TestGetCachePath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set test HOME
	testHome := t.TempDir()
	os.Setenv("HOME", testHome)

	expectedPath := filepath.Join(testHome, ".cache", "prompt-vault", "prompts")
	actualPath := GetCachePath()

	if actualPath != expectedPath {
		t.Errorf("GetCachePath() = %s, want %s", actualPath, expectedPath)
	}
}

func TestManager_InitializeCache(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create manager for testing
	manager := createTestManager(tempDir, "")

	// Test initialization
	err := manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v", err)
	}

	// Check if directory exists
	// Get the actual cache path from the manager
	actualCachePath := manager.GetCacheDir()
	info, err := os.Stat(actualCachePath)
	if err != nil {
		t.Fatalf("Cache directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Cache path is not a directory")
	}

	// Check permissions (should be 0700)
	if perm := info.Mode().Perm(); perm != 0700 {
		t.Errorf("Cache directory permissions = %v, want 0700", perm)
	}
}

func TestManager_InitializeCache_AlreadyExists(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create manager and get its cache path
	manager := createTestManager(tempDir, "")
	cachePath := manager.GetCacheDir()

	// Pre-create the directory
	err := os.MkdirAll(cachePath, 0700)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Initialize should succeed even if directory exists
	err = manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v, want nil for existing directory", err)
	}
}

func TestManager_GetCacheDir(t *testing.T) {
	tempDir := t.TempDir()
	manager := createTestManager(tempDir, "")
	expectedPath := filepath.Join(tempDir, ".cache", "prompt-vault", "prompts")

	result := manager.GetCacheDir()
	if result != expectedPath {
		t.Errorf("GetCacheDir() = %s, want %s", result, expectedPath)
	}
}

func TestManager_GetIndexPath(t *testing.T) {
	tempDir := t.TempDir()
	manager := createTestManager(tempDir, "")
	cachePath := manager.GetCacheDir()

	expected := filepath.Join(cachePath, "index.json")
	actual := manager.GetIndexPath()

	if actual != expected {
		t.Errorf("GetIndexPath() = %s, want %s", actual, expected)
	}
}

func TestManager_GetMetadataPath(t *testing.T) {
	tempDir := t.TempDir()
	manager := createTestManager(tempDir, "")
	cachePath := manager.GetCacheDir()

	expected := filepath.Join(filepath.Dir(cachePath), "metadata.json")
	actual := manager.GetMetadataPath()

	if actual != expected {
		t.Errorf("GetMetadataPath() = %s, want %s", actual, expected)
	}
}

func TestManager_GetPromptPath(t *testing.T) {
	tempDir := t.TempDir()
	manager := createTestManager(tempDir, "")
	cachePath := manager.GetCacheDir()

	gistID := "abc123def456"
	expected := filepath.Join(cachePath, gistID+".yaml")
	actual := manager.GetPromptPath(gistID)

	if actual != expected {
		t.Errorf("GetPromptPath() = %s, want %s", actual, expected)
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	expectedPath := GetCachePath()
	if manager.GetCacheDir() != expectedPath {
		t.Errorf("NewManager() cache dir = %s, want %s", manager.GetCacheDir(), expectedPath)
	}
}

func TestNewManagerWithPath(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, ".cache", "prompt-vault", "prompts")
	// NewManagerWithPath expects to extract home from the path
	manager := NewManagerWithPath(customPath)

	if manager == nil {
		t.Fatal("NewManagerWithPath() returned nil")
	}

	// The function should extract tempDir as home and recreate the path
	expectedPath := filepath.Join(tempDir, ".cache", "prompt-vault", "prompts")
	if manager.GetCacheDir() != expectedPath {
		t.Errorf("NewManagerWithPath() cache dir = %s, want %s", manager.GetCacheDir(), expectedPath)
	}
}

func TestManager_Clean(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create a PathManager with the test directory as home
	pm := paths.NewPathManagerWithHome(tempDir)
	manager := NewManagerWithPathManager(pm)

	// Get the actual paths from the manager
	cachePath := manager.GetCacheDir()

	// Initialize to create the directory structure
	err := manager.InitializeCache()
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Create test files
	testFiles := []string{
		filepath.Join(cachePath, "test1.yaml"),
		filepath.Join(cachePath, "test2.yaml"),
		manager.GetIndexPath(),
		manager.GetMetadataPath(),
	}

	for _, file := range testFiles {
		err := os.WriteFile(file, []byte("test content"), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Clean cache
	err = manager.Clean()
	if err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	// Check that all files are removed
	for _, file := range testFiles {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("File %s still exists after Clean()", file)
		}
	}

	// Check that cache directory still exists but is empty
	entries, err := os.ReadDir(cachePath)
	if err != nil {
		t.Fatalf("Failed to read cache directory: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("Cache directory is not empty after Clean(), found %d entries", len(entries))
	}
}

func TestManager_Clean_NonExistentDirectory(t *testing.T) {
	// Use a manager with non-existent path
	pm := paths.NewPathManagerWithHome("/non/existent/home")
	manager := NewManagerWithPathManager(pm)

	// Clean should not error on non-existent directory
	err := manager.Clean()
	if err != nil {
		t.Errorf("Clean() error = %v, want nil for non-existent directory", err)
	}
}

func TestManager_SaveAndGetPrompt(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Use createTestManager to handle path correctly
	manager := createTestManager(tempDir, "")

	// Initialize cache
	err := manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v", err)
	}

	// Create test prompt
	prompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:        "Test Prompt",
			Author:      "john",
			Category:    "test",
			Tags:        []string{"test", "example"},
			Version:     "1.0",
			Description: "A test prompt",
		},
		GistID:    "abc123def456",
		GistURL:   "https://gist.github.com/john/abc123def456",
		UpdatedAt: time.Date(2024, 1, 19, 10, 30, 0, 0, time.UTC),
		Content:   "Hello {name}! This is a test prompt.",
	}

	// Test SavePrompt
	err = manager.SavePrompt(prompt)
	if err != nil {
		t.Fatalf("SavePrompt() error = %v", err)
	}

	// Check file exists
	promptPath := manager.GetPromptPath(prompt.GistID)
	if _, err := os.Stat(promptPath); err != nil {
		t.Fatalf("Prompt file was not created: %v", err)
	}

	// Test GetPrompt
	loaded, err := manager.GetPrompt(prompt.GistID)
	if err != nil {
		t.Fatalf("GetPrompt() error = %v", err)
	}

	// Compare loaded prompt with original
	if loaded.Name != prompt.Name {
		t.Errorf("Name = %s, want %s", loaded.Name, prompt.Name)
	}
	if loaded.Content != prompt.Content {
		t.Errorf("Content = %s, want %s", loaded.Content, prompt.Content)
	}
	if loaded.GistID != prompt.GistID {
		t.Errorf("GistID = %s, want %s", loaded.GistID, prompt.GistID)
	}
}

func TestManager_GetPrompt_NotExist(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Use createTestManager to handle path correctly
	manager := createTestManager(tempDir, "")

	// Try to get non-existent prompt
	_, err := manager.GetPrompt("nonexistent")
	if err == nil {
		t.Error("GetPrompt() should return error for non-existent prompt")
	}
}

func TestManager_SavePrompt_InvalidContent(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Use createTestManager to handle path correctly
	manager := createTestManager(tempDir, "")

	// Initialize cache
	err := manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v", err)
	}

	// Test with nil prompt
	err = manager.SavePrompt(nil)
	if err == nil {
		t.Error("SavePrompt() should return error for nil prompt")
	}

	// Test with empty GistID
	prompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name: "Test",
		},
		GistID: "",
	}
	err = manager.SavePrompt(prompt)
	if err == nil {
		t.Error("SavePrompt() should return error for empty GistID")
	}
}

func TestManager_SavePrompt_ConcurrentAccess(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Use createTestManager to handle path correctly
	manager := createTestManager(tempDir, "")

	// Initialize cache
	err := manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v", err)
	}

	// Create multiple prompts
	prompts := []*models.Prompt{
		{
			PromptMeta: models.PromptMeta{
				Name:     "Prompt 1",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
			},
			GistID:  "gist1",
			Content: "Content 1",
		},
		{
			PromptMeta: models.PromptMeta{
				Name:     "Prompt 2",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
			},
			GistID:  "gist2",
			Content: "Content 2",
		},
		{
			PromptMeta: models.PromptMeta{
				Name:     "Prompt 3",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
			},
			GistID:  "gist3",
			Content: "Content 3",
		},
	}

	// Save prompts concurrently
	var wg sync.WaitGroup
	errors := make(chan error, len(prompts))

	for _, p := range prompts {
		wg.Add(1)
		go func(prompt *models.Prompt) {
			defer wg.Done()
			if err := manager.SavePrompt(prompt); err != nil {
				errors <- err
			}
		}(p)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent SavePrompt() error = %v", err)
	}

	// Verify all prompts were saved
	for _, p := range prompts {
		loaded, err := manager.GetPrompt(p.GistID)
		if err != nil {
			t.Errorf("GetPrompt(%s) error = %v", p.GistID, err)
			continue
		}
		if loaded.Name != p.Name {
			t.Errorf("Prompt %s: Name = %s, want %s", p.GistID, loaded.Name, p.Name)
		}
	}
}

func TestManager_GetPrompt_CorruptedFile(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Use createTestManager to handle path correctly
	manager := createTestManager(tempDir, "")

	// Initialize cache
	err := manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v", err)
	}

	// Create corrupted YAML file
	gistID := "corrupted"
	promptPath := manager.GetPromptPath(gistID)
	corruptedContent := `---
name: "Test"
invalid yaml content {{{
---
Content`

	err = os.WriteFile(promptPath, []byte(corruptedContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Try to get corrupted prompt
	_, err = manager.GetPrompt(gistID)
	if err == nil {
		t.Error("GetPrompt() should return error for corrupted file")
	}
}

func TestManager_SaveIndex(t *testing.T) {
	validIndex := &models.Index{
		Username: "testuser",
		Entries: []models.IndexEntry{
			{
				GistID:      "gist1",
				Name:        "Prompt 1",
				Author:      "testuser",
				Category:    "testing",
				Tags:        []string{"test"},
				Description: "Test prompt 1",
				UpdatedAt:   time.Now(),
			},
			{
				GistID:      "gist2",
				Name:        "Prompt 2",
				Author:      "testuser",
				Category:    "development",
				Tags:        []string{"dev", "code"},
				Description: "Test prompt 2",
				UpdatedAt:   time.Now(),
			},
		},
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name    string
		index   *models.Index
		wantErr bool
		setup   func(t *testing.T, m *Manager)
	}{
		{
			name:    "saves valid index",
			index:   validIndex,
			wantErr: false,
		},
		{
			name:    "rejects nil index",
			index:   nil,
			wantErr: true,
		},
		{
			name: "rejects index with empty username",
			index: &models.Index{
				Username: "",
				Entries:  []models.IndexEntry{},
			},
			wantErr: true,
		},
		{
			name: "handles file write failure",
			index: &models.Index{
				Username: "testuser",
				Entries:  []models.IndexEntry{},
			},
			wantErr: true,
			setup: func(t *testing.T, m *Manager) {
				// Create a directory where the index file should be
				indexFile := m.GetIndexPath()
				if err := os.Mkdir(indexFile, 0755); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			// Use createTestManager to handle path correctly
			m := createTestManager(tempDir, "")

			if err := m.InitializeCache(); err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				tt.setup(t, m)
			}

			err := m.SaveIndex(tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveIndex() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.index != nil {
				// Verify file was created
				indexFile := m.GetIndexPath()
				if _, err := os.Stat(indexFile); os.IsNotExist(err) {
					t.Errorf("SaveIndex() did not create index file")
				}
			}
		})
	}
}

func TestManager_GetIndex(t *testing.T) {
	validIndex := &models.Index{
		Username: "testuser",
		Entries: []models.IndexEntry{
			{
				GistID:      "gist1",
				Name:        "Prompt 1",
				Author:      "testuser",
				Category:    "testing",
				Tags:        []string{"test"},
				Description: "Test prompt 1",
				UpdatedAt:   time.Now().UTC().Truncate(time.Second),
			},
		},
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *Manager)
		want    *models.Index
		wantErr bool
	}{
		{
			name: "retrieves existing index",
			setup: func(t *testing.T, m *Manager) {
				if err := m.SaveIndex(validIndex); err != nil {
					t.Fatal(err)
				}
			},
			want:    validIndex,
			wantErr: false,
		},
		{
			name:    "returns nil for non-existent index",
			want:    nil,
			wantErr: false,
		},
		{
			name: "handles corrupted index file",
			setup: func(t *testing.T, m *Manager) {
				indexFile := m.GetIndexPath()
				if err := os.WriteFile(indexFile, []byte("invalid json content"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
		{
			name: "handles empty index file",
			setup: func(t *testing.T, m *Manager) {
				indexFile := m.GetIndexPath()
				if err := os.WriteFile(indexFile, []byte(""), 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			// Use createTestManager to handle path correctly
			m := createTestManager(tempDir, "")

			if err := m.InitializeCache(); err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				tt.setup(t, m)
			}

			got, err := m.GetIndex()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIndex() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != nil && tt.want != nil {
				// Compare the indexes
				if got.Username != tt.want.Username {
					t.Errorf("GetIndex() Username = %v, want %v", got.Username, tt.want.Username)
				}
				if len(got.Entries) != len(tt.want.Entries) {
					t.Errorf("GetIndex() Entries length = %v, want %v", len(got.Entries), len(tt.want.Entries))
				}
				// Compare timestamps with truncation for consistency
				if !got.UpdatedAt.Truncate(time.Second).Equal(tt.want.UpdatedAt.Truncate(time.Second)) {
					t.Errorf("GetIndex() UpdatedAt = %v, want %v", got.UpdatedAt, tt.want.UpdatedAt)
				}
			}
		})
	}
}

func TestManager_IndexAtomicUpdate(t *testing.T) {
	tempDir := t.TempDir()
	// Use createTestManager to handle path correctly
	m := createTestManager(tempDir, "")

	if err := m.InitializeCache(); err != nil {
		t.Fatal(err)
	}

	// Create initial index
	initialIndex := &models.Index{
		Username: "testuser",
		Entries: []models.IndexEntry{
			{
				GistID:    "gist1",
				Name:      "Initial",
				Author:    "testuser",
				Category:  "test",
				UpdatedAt: time.Now(),
			},
		},
		UpdatedAt: time.Now(),
	}

	if err := m.SaveIndex(initialIndex); err != nil {
		t.Fatal(err)
	}

	// Simulate concurrent updates
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			index := &models.Index{
				Username: "testuser",
				Entries: []models.IndexEntry{
					{
						GistID:    "gist" + string(rune('0'+id)),
						Name:      "Updated " + string(rune('0'+id)),
						Author:    "testuser",
						Category:  "test",
						UpdatedAt: time.Now(),
					},
				},
				UpdatedAt: time.Now(),
			}

			err := m.SaveIndex(index)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all updates
	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Atomic update error: %v", err)
	}

	// Verify final state is valid
	finalIndex, err := m.GetIndex()
	if err != nil {
		t.Fatalf("Failed to get final index: %v", err)
	}

	if finalIndex == nil {
		t.Fatal("Final index is nil")
	}

	if finalIndex.Username != "testuser" {
		t.Errorf("Final index has incorrect username: %v", finalIndex.Username)
	}
}
