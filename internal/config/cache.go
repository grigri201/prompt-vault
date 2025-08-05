package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// getCacheDir is a function variable that returns the cache directory path based on the OS
// It's a variable so it can be overridden in tests
var getCacheDir = func() (string, error) {
	// Check for environment variable override first
	if cacheDir := os.Getenv("PV_CACHE_DIR"); cacheDir != "" {
		return cacheDir, nil
	}

	var cacheDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %LOCALAPPDATA%\pv
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		cacheDir = filepath.Join(localAppData, "pv")
	default:
		// Linux/macOS: ~/.cache/pv (XDG Base Directory Specification)
		// First try XDG_CACHE_HOME, then fallback to ~/.cache
		cacheHome := os.Getenv("XDG_CACHE_HOME")
		if cacheHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			cacheHome = filepath.Join(homeDir, ".cache")
		}
		cacheDir = filepath.Join(cacheHome, "pv")
	}

	return cacheDir, nil
}

// GetCacheDir returns the cache directory path for the application
// This is a public wrapper around the private getCacheDir function
func GetCacheDir() (string, error) {
	return getCacheDir()
}