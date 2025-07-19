package cache

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/models"
)

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
	cachePath := filepath.Join(tempDir, "cache", "prompts")

	manager := &Manager{
		cachePath: cachePath,
	}

	// Test initialization
	err := manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v", err)
	}

	// Check if directory exists
	info, err := os.Stat(cachePath)
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
	cachePath := filepath.Join(tempDir, "cache", "prompts")

	// Pre-create the directory
	err := os.MkdirAll(cachePath, 0700)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	manager := &Manager{
		cachePath: cachePath,
	}

	// Initialize should succeed even if directory exists
	err = manager.InitializeCache()
	if err != nil {
		t.Fatalf("InitializeCache() error = %v, want nil for existing directory", err)
	}
}

func TestManager_GetCacheDir(t *testing.T) {
	cachePath := "/test/cache/path"
	manager := &Manager{
		cachePath: cachePath,
	}

	result := manager.GetCacheDir()
	if result != cachePath {
		t.Errorf("GetCacheDir() = %s, want %s", result, cachePath)
	}
}

func TestManager_GetIndexPath(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache", "prompts")
	
	manager := &Manager{
		cachePath: cachePath,
	}

	expected := filepath.Join(filepath.Dir(cachePath), "index.json")
	actual := manager.GetIndexPath()

	if actual != expected {
		t.Errorf("GetIndexPath() = %s, want %s", actual, expected)
	}
}

func TestManager_GetMetadataPath(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache", "prompts")
	
	manager := &Manager{
		cachePath: cachePath,
	}

	expected := filepath.Join(filepath.Dir(cachePath), "metadata.json")
	actual := manager.GetMetadataPath()

	if actual != expected {
		t.Errorf("GetMetadataPath() = %s, want %s", actual, expected)
	}
}

func TestManager_GetPromptPath(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache", "prompts")
	
	manager := &Manager{
		cachePath: cachePath,
	}

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
	if manager.cachePath != expectedPath {
		t.Errorf("NewManager() cachePath = %s, want %s", manager.cachePath, expectedPath)
	}
}

func TestNewManagerWithPath(t *testing.T) {
	customPath := "/custom/cache/path"
	manager := NewManagerWithPath(customPath)
	
	if manager == nil {
		t.Fatal("NewManagerWithPath() returned nil")
	}

	if manager.cachePath != customPath {
		t.Errorf("NewManagerWithPath() cachePath = %s, want %s", manager.cachePath, customPath)
	}
}

func TestManager_Clean(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	cachePath := filepath.Join(cacheDir, "prompts")

	// Create cache structure with some files
	err := os.MkdirAll(cachePath, 0700)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test files
	testFiles := []string{
		filepath.Join(cachePath, "test1.yaml"),
		filepath.Join(cachePath, "test2.yaml"),
		filepath.Join(cacheDir, "index.json"),
		filepath.Join(cacheDir, "metadata.json"),
	}

	for _, file := range testFiles {
		err := os.WriteFile(file, []byte("test content"), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	manager := &Manager{
		cachePath: cachePath,
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
	manager := &Manager{
		cachePath: "/non/existent/path",
	}

	// Clean should not error on non-existent directory
	err := manager.Clean()
	if err != nil {
		t.Errorf("Clean() error = %v, want nil for non-existent directory", err)
	}
}

func TestManager_SaveAndGetPrompt(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache", "prompts")

	manager := &Manager{
		cachePath: cachePath,
	}

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
	cachePath := filepath.Join(tempDir, "cache", "prompts")

	manager := &Manager{
		cachePath: cachePath,
	}

	// Try to get non-existent prompt
	_, err := manager.GetPrompt("nonexistent")
	if err == nil {
		t.Error("GetPrompt() should return error for non-existent prompt")
	}
}

func TestManager_SavePrompt_InvalidContent(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache", "prompts")

	manager := &Manager{
		cachePath: cachePath,
	}

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
	cachePath := filepath.Join(tempDir, "cache", "prompts")

	manager := &Manager{
		cachePath: cachePath,
	}

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
	cachePath := filepath.Join(tempDir, "cache", "prompts")

	manager := &Manager{
		cachePath: cachePath,
	}

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