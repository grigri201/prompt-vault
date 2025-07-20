package paths

import (
	"os"
	"path/filepath"
)

// PathManager handles all application paths
type PathManager struct {
	homeDir string
}

// NewPathManager creates a new path manager
func NewPathManager() *PathManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return &PathManager{homeDir: homeDir}
}

// NewPathManagerWithHome creates a path manager with custom home
func NewPathManagerWithHome(homeDir string) *PathManager {
	return &PathManager{homeDir: homeDir}
}

// GetCachePath returns the cache directory path
func (pm *PathManager) GetCachePath() string {
	return filepath.Join(pm.homeDir, ".cache", "prompt-vault", "prompts")
}

// GetConfigPath returns the config file path
func (pm *PathManager) GetConfigPath() string {
	return filepath.Join(pm.homeDir, ".config", "prompt-vault", "config.yaml")
}

// GetIndexPath returns the index file path
func (pm *PathManager) GetIndexPath() string {
	return filepath.Join(pm.GetCachePath(), "index.json")
}

// GetPromptPath returns the path for a specific prompt file
func (pm *PathManager) GetPromptPath(id string) string {
	return filepath.Join(pm.GetCachePath(), id+".md")
}

// EnsureDir creates directory if it doesn't exist
func (pm *PathManager) EnsureDir(path string) error {
	return os.MkdirAll(path, 0700)
}

// EnsureCacheDir ensures the cache directory exists
func (pm *PathManager) EnsureCacheDir() error {
	return pm.EnsureDir(pm.GetCachePath())
}

// EnsureConfigDir ensures the config directory exists
func (pm *PathManager) EnsureConfigDir() error {
	configDir := filepath.Dir(pm.GetConfigPath())
	return pm.EnsureDir(configDir)
}

// AtomicWrite performs atomic file write
func (pm *PathManager) AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := pm.EnsureDir(dir); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()

	// Cleanup function
	cleanup := func() {
		tempFile.Close()
		os.Remove(tempPath)
	}

	// Write data
	if _, err := tempFile.Write(data); err != nil {
		cleanup()
		return err
	}

	// Close the file
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}

	// Set permissions
	if err := os.Chmod(tempPath, perm); err != nil {
		os.Remove(tempPath)
		return err
	}

	// Atomic rename
	return os.Rename(tempPath, path)
}

// FileExists checks if a file exists
func (pm *PathManager) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads a file and returns its contents
func (pm *PathManager) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// RemoveFile removes a file
func (pm *PathManager) RemoveFile(path string) error {
	return os.Remove(path)
}

// ListFiles lists all files in a directory matching a pattern
func (pm *PathManager) ListFiles(dir, pattern string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, pattern))
}
