package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/models"
)

func TestGetCommand_GistURLDisplay(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "displays gist URL in single result",
			args: []string{"get", "unique"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "abc123unique",
							GistURL:     "https://gist.github.com/testuser/abc123unique",
							Name:        "Unique Prompt",
							Author:      "testuser",
							Category:    "special",
							Version:     "1.0",
							Description: "A unique prompt for testing",
							Tags:        []string{"unique", "test"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				// Save prompt with content
				for i := range index.Entries {
					savePromptToCache(t, cacheManager, index.Entries[i], "Test content")
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Found 1 prompt(s):",
				"[1] Unique Prompt by testuser",
				"Gist URL: https://gist.github.com/testuser/abc123unique",
			},
			wantErr: false,
		},
		{
			name: "displays gist URLs in multiple results",
			args: []string{"get", "test"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "test123",
							GistURL:     "https://gist.github.com/testuser/test123",
							Name:        "Test Prompt 1",
							Author:      "testuser",
							Category:    "testing",
							Version:     "1.0",
							Tags:        []string{"test"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "test456",
							GistURL:     "https://gist.github.com/testuser/test456",
							Name:        "Test Prompt 2",
							Author:      "testuser",
							Category:    "testing",
							Version:     "1.0",
							Tags:        []string{"test"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				// Save prompts with content
				for i := range index.Entries {
					savePromptToCache(t, cacheManager, index.Entries[i], "Test content")
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Found 2 prompt(s):",
				"[1] Test Prompt 1 by testuser",
				"Gist URL: https://gist.github.com/testuser/test123",
				"[2] Test Prompt 2 by testuser",
				"Gist URL: https://gist.github.com/testuser/test456",
			},
			wantErr: false,
		},
		{
			name: "handles empty gist URL gracefully",
			args: []string{"get", "empty"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "empty123",
							GistURL:     "", // Empty URL
							Name:        "Empty URL Prompt",
							Author:      "testuser",
							Category:    "testing",
							Version:     "1.0",
							Description: "Prompt without URL",
							Tags:        []string{"empty"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				// Save prompt with content
				for i := range index.Entries {
					savePromptToCache(t, cacheManager, index.Entries[i], "Test content")
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Found 1 prompt(s):",
				"[1] Empty URL Prompt by testuser",
				"Gist URL:", // Should show label even if URL is empty
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp cache
			tempDir := t.TempDir()
			cacheManager := cache.NewManagerWithPath(filepath.Join(tempDir, "cache", "prompts"))
			if err := cacheManager.InitializeCache(); err != nil {
				t.Fatal(err)
			}

			// Setup cache
			if tt.setupCache != nil {
				tt.setupCache(t, cacheManager)
			}

			// Override cache path for the command
			originalGetCachePath := getCachePathFunc
			getCachePathFunc = func() string {
				return filepath.Join(tempDir, "cache", "prompts")
			}
			defer func() {
				getCachePathFunc = originalGetCachePath
			}()

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
		})
	}
}

// Test for gist URL in clipboard success message
func TestGetCommand_GistURLInSuccessMessage(t *testing.T) {
	// This test would require mocking the clipboard and selector interaction
	// Since the test environment doesn't have TTY support, we'll need to
	// implement this test differently or skip it for now
	t.Skip("Clipboard success message test requires TTY support")
	
	// Expected behavior: After copying to clipboard, the success message should include:
	// ✓ Prompt copied to clipboard!
	// Gist URL: https://gist.github.com/testuser/abc123def456
}