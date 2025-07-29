package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/models"
)

func TestSyncCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "shows empty index message",
			args: []string{"sync"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create empty index
				index := &models.Index{
					Username:  "testuser",
					Entries:   []models.IndexEntry{},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Starting synchronization",
				"No prompts found in index",
				"Upload prompts using 'pv upload' to get started",
				"Sync completed successfully!",
				"Sync cache files to",
			},
			wantErr: false,
		},
		{
			name: "syncs with existing prompts",
			args: []string{"sync"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create index with prompts
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Prompt 1",
							Author:    "testuser",
							Category:  "test",
							GistID:    "gist1",
							UpdatedAt: time.Now(),
						},
						{
							Name:      "Prompt 2",
							Author:    "testuser",
							Category:  "example",
							GistID:    "gist2",
							UpdatedAt: time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Starting synchronization",
				"Sync completed successfully",
				"Downloaded: 2 prompts",
				"Total prompts: 2",
				"Sync cache files to",
			},
			wantErr: false,
		},
		{
			name: "syncs with verbose output",
			args: []string{"sync", "--verbose"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create index with prompts
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Test Prompt",
							Author:    "testuser",
							Category:  "test",
							GistID:    "gist123",
							UpdatedAt: time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Starting synchronization",
				"Checking for updates",
				"Downloading index from GitHub",
				"Found 1 prompts in index",
				"Downloading: Test Prompt",
				"Sync completed successfully",
				"Last sync:",
			},
			wantErr: false,
		},
		{
			name: "handles missing index gracefully",
			args: []string{"sync"},
			wantOutput: []string{
				"Starting synchronization",
				"No prompts found in index",
				"Sync completed successfully!",
				"Sync cache files to",
			},
			wantErr: false,
		},
		{
			name: "creates cache directory if missing",
			args: []string{"sync"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Don't initialize cache, to simulate missing directory
			},
			wantOutput: []string{
				"Starting synchronization",
				"Cache directory created at",
				"No prompts found in index",
				"Sync completed successfully!",
				"Sync cache files to",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp cache
			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, ".cache", "prompt-vault", "prompts")
			cacheManager := cache.NewManagerWithPath(cacheDir)
			
			// Only initialize cache if not testing missing directory
			if tt.name != "creates cache directory if missing" {
				if err := cacheManager.InitializeCache(); err != nil {
					t.Fatal(err)
				}
			}

			// Setup cache if needed
			if tt.setupCache != nil {
				tt.setupCache(t, cacheManager)
			}

			// Override cache path
			originalGetCachePath := getCachePathFunc
			getCachePathFunc = func() string {
				return cacheDir
			}
			defer func() {
				getCachePathFunc = originalGetCachePath
			}()

			// Create command
			cmd := NewRootCmd()

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tt.args)

			// Execute
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check output
			output := out.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing %q\nGot: %s", want, output)
				}
			}

			// If sync was successful, verify index was updated
			if !tt.wantErr && err == nil {
				updatedIndex, _ := cacheManager.GetIndex()
				if updatedIndex != nil && updatedIndex.UpdatedAt.IsZero() {
					t.Error("Index UpdatedAt was not set after sync")
				}
			}

			// Verify cache directory was created for missing directory test
			if tt.name == "creates cache directory if missing" && err == nil {
				if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
					t.Error("Cache directory was not created")
				}
			}
		})
	}
}

func TestSyncCommand_Help(t *testing.T) {
	cmd := NewRootCmd()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"sync", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := out.String()
	expectedStrings := []string{
		"Synchronize your local prompt cache with GitHub Gists",
		"downloads all prompts",
		"--verbose",
		"Show detailed progress",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help text missing %q", expected)
		}
	}
}
