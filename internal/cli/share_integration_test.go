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

func TestShareCommand_Integration_CompleteFlow(t *testing.T) {
	tests := []struct {
		name       string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		args       []string
		shareFunc  func(ctx context.Context, privateGistID string) (*share.ShareResult, error)
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "no arguments shows all prompts for selection",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "First Prompt",
							Author:      "alice",
							Category:    "development",
							Version:     "1.0",
							Description: "Development helper prompt",
							Tags:        []string{"dev", "helper"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "abcdefabcdefabcdefabcdefabcdefab",
							GistURL:     "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
							Name:        "Second Prompt",
							Author:      "bob",
							Category:    "testing",
							Version:     "2.0",
							Description: "Testing automation prompt",
							Tags:        []string{"test", "automation"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{}, // No arguments
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "public" + privateGistID[:10],
					PublicGistURL: "https://gist.github.com/testuser/public" + privateGistID[:10],
					IsUpdate:      false,
				}, nil
			},
			wantOutput: []string{
				"Found 2 prompt(s) to share:",
				"[1] First Prompt by alice",
				"Category: development",
				"Tags: dev, helper",
				"Description: Development helper prompt",
				"[2] Second Prompt by bob",
				"Category: testing",
				"Tags: test, automation",
				"Description: Testing automation prompt",
			},
			wantErr: false,
		},
		{
			name: "keyword search with multiple matches",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "Code Helper",
							Author:      "alice",
							Category:    "development",
							Version:     "1.0",
							Description: "Helps with code development",
							Tags:        []string{"code", "helper"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "abcdefabcdefabcdefabcdefabcdefab",
							GistURL:     "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
							Name:        "Test Helper",
							Author:      "bob",
							Category:    "testing",
							Version:     "1.0",
							Description: "Helps with test creation",
							Tags:        []string{"test", "helper"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "fedcba0987654321fedcba0987654321",
							GistURL:     "https://gist.github.com/testuser/fedcba0987654321fedcba0987654321",
							Name:        "Documentation Writer",
							Author:      "charlie",
							Category:    "docs",
							Version:     "1.0",
							Tags:        []string{"docs"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{"helper"}, // Keyword search
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "public" + privateGistID[:10],
					PublicGistURL: "https://gist.github.com/testuser/public" + privateGistID[:10],
					IsUpdate:      false,
				}, nil
			},
			wantOutput: []string{
				"Found 2 prompt(s) matching 'helper':",
				"[1] Code Helper by alice",
				"Category: development",
				"Tags: code, helper",
				"Description: Helps with code development",
				"[2] Test Helper by bob",
				"Category: testing",
				"Tags: test, helper",
				"Description: Helps with test creation",
			},
			wantErr: false,
		},
		{
			name: "keyword search with single match shows confirmation",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "a01cde1234567890abcdef1234567890",
							GistURL:     "https://gist.github.com/testuser/a01cde1234567890abcdef1234567890",
							Name:        "Unique Prompt",
							Author:      "dave",
							Category:    "special",
							Version:     "1.0",
							Description: "A unique special prompt",
							Tags:        []string{"unique", "special"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{"unique"}, // Keyword that matches only one prompt
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "publicunique",
					PublicGistURL: "https://gist.github.com/testuser/publicunique",
					IsUpdate:      false,
				}, nil
			},
			wantOutput: []string{
				"Found 1 prompt matching 'unique':",
				"Unique Prompt by dave",
				"Category: special",
				"Tags: unique, special",
				"Description: A unique special prompt",
				"Share this prompt?",
			},
			wantErr: false,
		},
		{
			name: "direct gist ID bypasses search",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "d1234567890abcdef1234567890abcde",
							GistURL:     "https://gist.github.com/testuser/d1234567890abcdef1234567890abcde",
							Name:        "Direct Prompt",
							Author:      "eve",
							Category:    "direct",
							Version:     "1.0",
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{"d1234567890abcdef1234567890abcde"}, // Direct gist ID
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				if privateGistID != "d1234567890abcdef1234567890abcde" {
					t.Errorf("Expected gist ID 'd1234567890abcdef1234567890abcde', got '%s'", privateGistID)
				}
				return &share.ShareResult{
					PublicGistID:  "publicdirect",
					PublicGistURL: "https://gist.github.com/testuser/publicdirect",
					IsUpdate:      false,
				}, nil
			},
			wantOutput: []string{
				"Successfully created public gist: https://gist.github.com/testuser/publicdirect",
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
							GistID:    "someid123",
							Name:      "Some Prompt",
							Author:    "frank",
							Category:  "misc",
							Version:   "1.0",
							UpdatedAt: time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{"nonexistent"}, // Keyword that doesn't match
			wantOutput: []string{
				"No prompts found matching 'nonexistent'",
			},
			wantErr: true,
		},
		{
			name: "successful update of existing public gist",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Empty cache is fine, we're testing direct gist ID
			},
			args: []string{"abcdef1234567890abcdef1234567890"}, // Direct gist ID
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "publicupdate",
					PublicGistURL: "https://gist.github.com/testuser/publicupdate",
					IsUpdate:      true, // This is an update
				}, nil
			},
			wantOutput: []string{
				"Successfully updated public gist: https://gist.github.com/testuser/publicupdate",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp cache directory
			tempDir := t.TempDir()
			cacheManager := cache.NewManagerWithPath(filepath.Join(tempDir, "cache", "prompts"))
			if err := cacheManager.InitializeCache(); err != nil {
				t.Fatal(err)
			}

			// Setup cache
			if tt.setupCache != nil {
				tt.setupCache(t, cacheManager)
			}

			// Override cache path
			originalGetCachePath := getCachePathFunc
			getCachePathFunc = func() string {
				return filepath.Join(tempDir, "cache", "prompts")
			}
			defer func() {
				getCachePathFunc = originalGetCachePath
			}()

			// Create mock share manager
			mockManager := &MockShareManager{
				SharePromptFunc: tt.shareFunc,
			}

			// Create command with mock
			cmd := newShareCmd(mockManager)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tt.args)

			// Execute
			err := cmd.Execute()

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				// Accept TTY errors for interactive tests
				if err != nil && !strings.Contains(err.Error(), "TTY") {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Check output
			output := out.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestShareCommand_Integration_AllModes(t *testing.T) {
	// Setup comprehensive test data
	tempDir := t.TempDir()
	cacheManager := cache.NewManagerWithPath(filepath.Join(tempDir, "cache", "prompts"))
	if err := cacheManager.InitializeCache(); err != nil {
		t.Fatal(err)
	}

	// Create test index with various prompts
	index := &models.Index{
		Username: "testuser",
		Entries: []models.IndexEntry{
			{
				GistID:      "1234567890abcdef1234567890abcdef",
				GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
				Name:        "Development Assistant",
				Author:      "alice",
				Category:    "development",
				Version:     "3.0",
				Description: "Assists with software development tasks",
				Tags:        []string{"dev", "code", "assistant"},
				UpdatedAt:   time.Now(),
			},
			{
				GistID:      "1e51167890abcdef1234567890abcdef",
				GistURL:     "https://gist.github.com/testuser/1e51167890abcdef1234567890abcdef",
				Name:        "Testing Tool",
				Author:      "bob",
				Category:    "testing",
				Version:     "1.5",
				Description: "Automated testing helper",
				Tags:        []string{"test", "automation", "qa"},
				UpdatedAt:   time.Now(),
			},
			{
				GistID:      "d0c1234567890abcdef1234567890abc",
				GistURL:     "https://gist.github.com/testuser/d0c1234567890abcdef1234567890abc",
				Name:        "Documentation Writer",
				Author:      "charlie",
				Category:    "documentation",
				Version:     "2.1",
				Description: "Helps write technical documentation",
				Tags:        []string{"docs", "writing", "technical"},
				UpdatedAt:   time.Now(),
			},
		},
		UpdatedAt: time.Now(),
	}

	if err := cacheManager.SaveIndex(index); err != nil {
		t.Fatal(err)
	}

	// Override cache path
	originalGetCachePath := getCachePathFunc
	getCachePathFunc = func() string {
		return filepath.Join(tempDir, "cache", "prompts")
	}
	defer func() {
		getCachePathFunc = originalGetCachePath
	}()

	// Test mode 1: No arguments (should show all prompts)
	t.Run("mode_no_args", func(t *testing.T) {
		mockManager := &MockShareManager{
			SharePromptFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "public_" + privateGistID[:10],
					PublicGistURL: "https://gist.github.com/testuser/public_" + privateGistID[:10],
					IsUpdate:      false,
				}, nil
			},
		}

		cmd := newShareCmd(mockManager)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		// Expect TTY error for selector
		if err == nil || !strings.Contains(err.Error(), "TTY") {
			t.Errorf("Expected TTY error, got: %v", err)
		}

		output := out.String()
		expectedElements := []string{
			"Found 3 prompt(s) to share:",
			"Development Assistant",
			"Testing Tool",
			"Documentation Writer",
		}
		for _, expected := range expectedElements {
			if !strings.Contains(output, expected) {
				t.Errorf("Missing expected element %q in output", expected)
			}
		}
	})

	// Test mode 2: Keyword search
	t.Run("mode_keyword", func(t *testing.T) {
		mockManager := &MockShareManager{
			SharePromptFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "public_doc",
					PublicGistURL: "https://gist.github.com/testuser/public_doc",
					IsUpdate:      false,
				}, nil
			},
		}

		cmd := newShareCmd(mockManager)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"documentation"})

		err := cmd.Execute()
		// Expect TTY error for confirmation
		if err == nil || !strings.Contains(err.Error(), "TTY") {
			t.Errorf("Expected TTY error, got: %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "Found 1 prompt matching 'documentation':") {
			t.Error("Should find exactly one match for 'documentation'")
		}
		if !strings.Contains(output, "Documentation Writer") {
			t.Error("Should show the Documentation Writer prompt")
		}
		if !strings.Contains(output, "Share this prompt?") {
			t.Error("Should ask for confirmation with single match")
		}
	})

	// Test mode 3: Direct gist ID
	t.Run("mode_gist_id", func(t *testing.T) {
		mockManager := &MockShareManager{
			SharePromptFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				if privateGistID != "1234567890abcdef1234567890abcdef" {
					t.Errorf("Expected specific gist ID, got: %s", privateGistID)
				}
				return &share.ShareResult{
					PublicGistID:  "public123",
					PublicGistURL: "https://gist.github.com/testuser/public123",
					IsUpdate:      false,
				}, nil
			},
		}

		cmd := newShareCmd(mockManager)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"1234567890abcdef1234567890abcdef"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "Successfully created public gist") {
			t.Error("Should show success message")
		}
		if !strings.Contains(output, "https://gist.github.com/testuser/public123") {
			t.Error("Should show the public gist URL")
		}

		// Verify the mock was called with correct gist ID
		if len(mockManager.SharePromptCalls) != 1 {
			t.Errorf("Expected 1 call to SharePrompt, got %d", len(mockManager.SharePromptCalls))
		}
		if mockManager.SharePromptCalls[0] != "1234567890abcdef1234567890abcdef" {
			t.Errorf("SharePrompt called with wrong ID: %s", mockManager.SharePromptCalls[0])
		}
	})
}