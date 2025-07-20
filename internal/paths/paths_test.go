package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPathManager(t *testing.T) {
	pm := NewPathManager()
	if pm.homeDir == "" {
		t.Error("NewPathManager() should set homeDir")
	}
}

func TestNewPathManagerWithHome(t *testing.T) {
	customHome := "/custom/home"
	pm := NewPathManagerWithHome(customHome)

	if pm.homeDir != customHome {
		t.Errorf("NewPathManagerWithHome() homeDir = %v, want %v", pm.homeDir, customHome)
	}
}

func TestPathManager_GetPaths(t *testing.T) {
	pm := NewPathManagerWithHome("/home/test")

	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{
			name: "GetCachePath",
			fn:   pm.GetCachePath,
			want: "/home/test/.cache/prompt-vault/prompts",
		},
		{
			name: "GetConfigPath",
			fn:   pm.GetConfigPath,
			want: "/home/test/.config/prompt-vault/config.yaml",
		},
		{
			name: "GetIndexPath",
			fn:   pm.GetIndexPath,
			want: "/home/test/.cache/prompt-vault/prompts/index.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPathManager_GetPromptPath(t *testing.T) {
	pm := NewPathManagerWithHome("/home/test")

	got := pm.GetPromptPath("test-id")
	want := "/home/test/.cache/prompt-vault/prompts/test-id.md"

	if got != want {
		t.Errorf("GetPromptPath() = %v, want %v", got, want)
	}
}

func TestPathManager_EnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	testDir := filepath.Join(tempDir, "test", "nested", "dir")

	if err := pm.EnsureDir(testDir); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("EnsureDir() did not create directory")
	}

	// Test idempotency
	if err := pm.EnsureDir(testDir); err != nil {
		t.Errorf("EnsureDir() on existing dir error = %v", err)
	}
}

func TestPathManager_EnsureCacheDir(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	if err := pm.EnsureCacheDir(); err != nil {
		t.Fatalf("EnsureCacheDir() error = %v", err)
	}

	cacheDir := pm.GetCachePath()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("EnsureCacheDir() did not create cache directory")
	}
}

func TestPathManager_EnsureConfigDir(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	if err := pm.EnsureConfigDir(); err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	configDir := filepath.Dir(pm.GetConfigPath())
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("EnsureConfigDir() did not create config directory")
	}
}

func TestPathManager_AtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("test content")
	testPerm := os.FileMode(0600)

	// Test atomic write
	if err := pm.AtomicWrite(testFile, testData, testPerm); err != nil {
		t.Fatalf("AtomicWrite() error = %v", err)
	}

	// Verify file exists
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat written file: %v", err)
	}

	// Verify permissions
	if info.Mode() != testPerm {
		t.Errorf("File permissions = %v, want %v", info.Mode(), testPerm)
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("File content = %v, want %v", string(content), string(testData))
	}

	// Test atomic write to non-existent directory
	nestedFile := filepath.Join(tempDir, "nested", "dir", "test.txt")
	if err := pm.AtomicWrite(nestedFile, testData, testPerm); err != nil {
		t.Fatalf("AtomicWrite() to nested dir error = %v", err)
	}

	if _, err := os.Stat(nestedFile); err != nil {
		t.Error("AtomicWrite() did not create nested file")
	}
}

func TestPathManager_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	// Test non-existent file
	if pm.FileExists(filepath.Join(tempDir, "nonexistent.txt")) {
		t.Error("FileExists() returned true for non-existent file")
	}

	// Create a file
	testFile := filepath.Join(tempDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !pm.FileExists(testFile) {
		t.Error("FileExists() returned false for existing file")
	}
}

func TestPathManager_ReadFile(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	testFile := filepath.Join(tempDir, "read.txt")
	testData := []byte("read test content")

	// Write test file
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test read
	content, err := pm.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("ReadFile() content = %v, want %v", string(content), string(testData))
	}

	// Test read non-existent file
	_, err = pm.ReadFile(filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("ReadFile() should return error for non-existent file")
	}
}

func TestPathManager_RemoveFile(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	testFile := filepath.Join(tempDir, "remove.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("remove"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test remove
	if err := pm.RemoveFile(testFile); err != nil {
		t.Fatalf("RemoveFile() error = %v", err)
	}

	// Verify file is removed
	if pm.FileExists(testFile) {
		t.Error("RemoveFile() did not remove the file")
	}

	// Test remove non-existent file
	err := pm.RemoveFile(testFile)
	if err == nil {
		t.Error("RemoveFile() should return error for non-existent file")
	}
}

func TestPathManager_ListFiles(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPathManagerWithHome(tempDir)

	// Create test files
	testFiles := []string{"test1.md", "test2.md", "test.txt", "other.md"}
	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		if err := os.WriteFile(path, []byte("content"), 0600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test list with pattern
	files, err := pm.ListFiles(tempDir, "*.md")
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 3 {
		t.Errorf("ListFiles() returned %d files, want 3", len(files))
	}

	// Test list with non-matching pattern
	files, err = pm.ListFiles(tempDir, "*.xyz")
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("ListFiles() returned %d files for non-matching pattern, want 0", len(files))
	}
}
