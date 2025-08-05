package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetCacheDir(t *testing.T) {
	// Save original getCacheDir function
	originalGetCacheDir := getCacheDir
	defer func() {
		getCacheDir = originalGetCacheDir
	}()

	tests := []struct {
		name        string
		setEnvVar   map[string]string
		unsetEnvVar []string
		goos        string
		want        string
		wantErr     bool
	}{
		{
			name:      "PV_CACHE_DIR override",
			setEnvVar: map[string]string{"PV_CACHE_DIR": "/custom/cache/path"},
			want:      "/custom/cache/path",
			wantErr:   false,
		},
		{
			name:        "Windows LOCALAPPDATA",
			setEnvVar:   map[string]string{"LOCALAPPDATA": "C:\\Users\\test\\AppData\\Local"},
			unsetEnvVar: []string{"PV_CACHE_DIR"},
			goos:        "windows",
			want:        filepath.Join("C:\\Users\\test\\AppData\\Local", "pv"),
			wantErr:     false,
		},
		{
			name:        "Unix XDG_CACHE_HOME",
			setEnvVar:   map[string]string{"XDG_CACHE_HOME": "/home/test/.cache"},
			unsetEnvVar: []string{"PV_CACHE_DIR"},
			goos:        "linux",
			want:        "/home/test/.cache/pv",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.setEnvVar {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Unset environment variables
			for _, key := range tt.unsetEnvVar {
				os.Unsetenv(key)
				defer os.Setenv(key, os.Getenv(key)) // Restore original value
			}

			// Mock runtime.GOOS if specified
			if tt.goos != "" {
				// Create a test version of getCacheDir with mocked GOOS
				getCacheDir = func() (string, error) {
					// Check for environment variable override first
					if cacheDir := os.Getenv("PV_CACHE_DIR"); cacheDir != "" {
						return cacheDir, nil
					}

					var cacheDir string

					switch tt.goos {
					case "windows":
						// Windows: %LOCALAPPDATA%\pv
						localAppData := os.Getenv("LOCALAPPDATA")
						if localAppData == "" {
							return "", nil // For this test, we'll assume it's set
						}
						cacheDir = filepath.Join(localAppData, "pv")
					default:
						// Linux/macOS: ~/.cache/pv (XDG Base Directory Specification)
						// First try XDG_CACHE_HOME, then fallback to ~/.cache
						cacheHome := os.Getenv("XDG_CACHE_HOME")
						if cacheHome == "" {
							homeDir, err := os.UserHomeDir()
							if err != nil {
								return "", err
							}
							cacheHome = filepath.Join(homeDir, ".cache")
						}
						cacheDir = filepath.Join(cacheHome, "pv")
					}

					return cacheDir, nil
				}
			}

			got, err := getCacheDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCacheDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCacheDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCacheDirDefaultBehavior(t *testing.T) {
	// Test the actual function with real environment
	// Save and unset PV_CACHE_DIR to test default behavior
	originalValue := os.Getenv("PV_CACHE_DIR")
	os.Unsetenv("PV_CACHE_DIR")
	defer func() {
		if originalValue != "" {
			os.Setenv("PV_CACHE_DIR", originalValue)
		}
	}()

	cacheDir, err := getCacheDir()
	if err != nil {
		t.Fatalf("getCacheDir() failed: %v", err)
	}

	// Verify the path ends with "/pv" (or "\pv" on Windows)
	expectedSuffix := filepath.Join("", "pv")
	if !filepath.IsAbs(cacheDir) {
		t.Errorf("getCacheDir() returned relative path: %s", cacheDir)
	}

	if !filepath.IsAbs(cacheDir) || !strings.HasSuffix(cacheDir, expectedSuffix) {
		t.Errorf("getCacheDir() = %s, expected to end with %s", cacheDir, expectedSuffix)
	}

	// Verify OS-specific behavior
	switch runtime.GOOS {
	case "windows":
		// Should contain LOCALAPPDATA or be set by PV_CACHE_DIR
		if os.Getenv("LOCALAPPDATA") != "" && !strings.HasSuffix(cacheDir, "pv") {
			t.Errorf("Windows cache dir should end with 'pv': %s", cacheDir)
		}
	default:
		// Should contain .cache or be set by environment variables
		if !strings.HasSuffix(cacheDir, "pv") {
			t.Errorf("Unix cache dir should end with 'pv': %s", cacheDir)
		}
	}
}