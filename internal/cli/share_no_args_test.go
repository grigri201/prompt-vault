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

func TestShareCommand_NoArguments(t *testing.T) {
	tests := []struct {
		name       string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "shows empty list message when no prompts",
			setupCache: nil, // No cache setup, empty index
			wantOutput: []string{
				"No prompts found",
				"Use 'pv sync' to download prompts",
			},
			wantErr: false,
		},
		{
			name: "displays all prompts for selection",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "private123",
							GistURL:     "https://gist.github.com/testuser/private123",
							Name:        "Private Prompt 1",
							Author:      "testuser",
							Category:    "development",
							Version:     "1.0",
							Description: "A private prompt for testing",
							Tags:        []string{"test", "private"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "private456",
							GistURL:     "https://gist.github.com/testuser/private456",
							Name:        "Private Prompt 2",
							Author:      "testuser",
							Category:    "writing",
							Version:     "1.0",
							Description: "Another private prompt",
							Tags:        []string{"test", "private"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Found 2 prompt(s) to share:",
				"[1] Private Prompt 1 by testuser",
				"Category: development",
				"Tags: test, private",
				"Description: A private prompt for testing",
				"[2] Private Prompt 2 by testuser",
				"Category: writing",
				"Description: Another private prompt",
			},
			wantErr: false,
		},
		{
			name: "displays single prompt for sharing",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "single123",
							GistURL:     "https://gist.github.com/testuser/single123",
							Name:        "Single Prompt",
							Author:      "testuser",
							Category:    "general",
							Version:     "1.0",
							Tags:        []string{"single"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Found 1 prompt(s) to share:",
				"[1] Single Prompt by testuser",
				"Category: general",
				"Tags: single",
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

			// Setup cache if needed
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
			cmd.SetArgs([]string{"share"}) // No arguments

			// Execute
			err := cmd.Execute()

			// Check output first, regardless of error
			output := out.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing %q\nGot: %s", want, output)
				}
			}

			// Check error - expecting TTY error for selector tests
			if err != nil && strings.Contains(err.Error(), "TTY") {
				// This is expected in test environment, check if we got the right output
				t.Logf("Got expected TTY error in test environment")
			} else if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShareCommand_SelectionCancelled(t *testing.T) {
	// This test would require mocking the selector interaction
	// Since the test environment doesn't have TTY support, we'll skip it
	t.Skip("Selection cancellation test requires TTY support")
	
	// Expected behavior: When user cancels selection
	// Output: "No selection made."
	// No sharing should occur
}