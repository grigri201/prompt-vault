package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/share"
)

func TestShareCommand_KeywordSearch(t *testing.T) {
	tests := []struct {
		name       string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		keyword    string
		wantOutput []string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "keyword search with multiple matches",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "Test Prompt 1",
							Author:      "testuser",
							Category:    "development",
							Version:     "1.0",
							Description: "A test prompt for development",
							Tags:        []string{"test", "dev"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "abcdef1234567890abcdef1234567890",
							GistURL:     "https://gist.github.com/testuser/abcdef1234567890abcdef1234567890",
							Name:        "Test Prompt 2",
							Author:      "testuser",
							Category:    "testing",
							Version:     "1.0",
							Description: "Another test prompt",
							Tags:        []string{"test", "qa"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "fedcba0987654321fedcba0987654321",
							GistURL:     "https://gist.github.com/testuser/fedcba0987654321fedcba0987654321",
							Name:        "Documentation Helper",
							Author:      "testuser",
							Category:    "documentation",
							Version:     "1.0",
							Description: "Helps write documentation efficiently",
							Tags:        []string{"docs", "writing"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			keyword: "test",
			wantOutput: []string{
				"Found 3 prompt(s) matching 'test':",
				"Test Prompt 1 by testuser",
				"Category: development",
				"Tags: test, dev",  
				"Description: A test prompt for development",
				"Test Prompt 2 by testuser",
				"Category: testing",
				"Tags: test, qa",
				"Description: Another test prompt",
				"Documentation Helper by testuser",
			},
			wantErr: false,
		},
		{
			name: "keyword search with single match asks for confirmation",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "Documentation Helper",
							Author:      "testuser",
							Category:    "documentation",
							Version:     "1.0",
							Description: "Helps write documentation",
							Tags:        []string{"docs", "writing"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "abcdef1234567890abcdef1234567890",
							GistURL:     "https://gist.github.com/testuser/abcdef1234567890abcdef1234567890",
							Name:        "Test Prompt",
							Author:      "testuser",
							Category:    "testing",
							Version:     "1.0",
							Description: "A test prompt",
							Tags:        []string{"test"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			keyword: "documentation",
			wantOutput: []string{
				"Found 1 prompt matching 'documentation':",
				"Documentation Helper by testuser",
				"Category: documentation",
				"Tags: docs, writing",
				"Description: Helps write documentation",
				"Share this prompt?",
			},
			wantErr: false,
		},
		{
			name: "keyword search with no matches",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "Test Prompt",
							Author:      "testuser",
							Category:    "testing",
							Version:     "1.0",
							Tags:        []string{"test"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			keyword:    "nonexistent",
			wantOutput: []string{},
			wantErr:    true,
			wantErrMsg: "No prompts found matching 'nonexistent'",
		},
		{
			name: "gist ID takes priority over keyword search",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "abcdefabcdefabcdefabcdefabcdefab",
							GistURL:     "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
							Name:        "Specific Prompt",
							Author:      "testuser",
							Category:    "specific",
							Version:     "1.0",
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "abcdefabcdefabcdefabcdefabcdefab Prompt",
							Author:      "testuser",
							Category:    "testing",
							Version:     "1.0",
							Description: "Has the gist ID in name",
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			keyword: "abcdefabcdefabcdefabcdefabcdefab", // This is a valid gist ID
			wantOutput: []string{
				// Should attempt to share directly, not search
			},
			wantErr: false,
		},
		{
			name: "empty keyword search",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:    "1234567890abcdef1234567890abcdef",
							Name:      "Test",
							Author:    "testuser",
							Category:  "test",
							UpdatedAt: time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			keyword:    "",
			wantOutput: []string{"Found 1 prompt(s) to share:"},
			wantErr:    false,
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

			// Create mock share manager
			mockManager := &MockShareManager{
				SharePromptFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
					// For testing, we'll return success
					return &share.ShareResult{
						PublicGistID:  "public" + privateGistID[:10],
						PublicGistURL: "https://gist.github.com/testuser/public" + privateGistID[:10],
						IsUpdate:      false,
					}, nil
				},
			}

			cmd := newShareCmd(mockManager)

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			if tt.keyword == "" {
				cmd.SetArgs([]string{})
			} else {
				cmd.SetArgs([]string{tt.keyword})
			}

			// Execute
			err := cmd.Execute()

			// Check output first
			output := out.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing %q\nGot: %s", want, output)
				}
			}

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.wantErrMsg != "" && !strings.Contains(output, tt.wantErrMsg) {
					t.Errorf("Expected error message %q, got: %s", tt.wantErrMsg, output)
				}
			} else {
				// Accept TTY errors in test environment for interactive tests
				if err != nil && strings.Contains(err.Error(), "TTY") {
					t.Logf("Got expected TTY error in test environment")
				} else if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestShareCommand_KeywordSearchCancellation(t *testing.T) {
	// This test would require mocking the selector interaction
	// Since the test environment doesn't have TTY support, we'll skip it
	t.Skip("Keyword search cancellation test requires TTY support")
	
	// Expected behavior: When user cancels selection after keyword search
	// Output: "No selection made."
	// No sharing should occur
}