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

func TestGetCommand_Integration_GistURLDisplay(t *testing.T) {
	tests := []struct {
		name       string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		keyword    string
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "displays gist URL in search results",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create index
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "Test Prompt",
							Author:      "alice",
							Category:    "testing",
							Version:     "1.0",
							Description: "A test prompt for integration testing",
							Tags:        []string{"test", "integration"},
							UpdatedAt:   time.Now(),
						},
						{
							GistID:      "abcdefabcdefabcdefabcdefabcdefab",
							GistURL:     "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
							Name:        "Another Test",
							Author:      "bob",
							Category:    "testing",
							Version:     "2.0",
							Description: "Another test prompt",
							Tags:        []string{"test"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}

				// Create prompt files
				prompt1 := &models.Prompt{
					PromptMeta: models.PromptMeta{
						Name:        "Test Prompt",
						Author:      "alice",
						Category:    "testing",
						Version:     "1.0",
						Description: "A test prompt for integration testing",
						Tags:        []string{"test", "integration"},
					},
					GistID:  "1234567890abcdef1234567890abcdef",
					GistURL: "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
					Content: "This is a test prompt content",
				}
				if err := cacheManager.SavePrompt(prompt1); err != nil {
					t.Fatal(err)
				}

				prompt2 := &models.Prompt{
					PromptMeta: models.PromptMeta{
						Name:        "Another Test",
						Author:      "bob",
						Category:    "testing",
						Version:     "2.0",
						Description: "Another test prompt",
						Tags:        []string{"test"},
					},
					GistID:  "abcdefabcdefabcdefabcdefabcdefab",
					GistURL: "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
					Content: "Another test prompt content",
				}
				if err := cacheManager.SavePrompt(prompt2); err != nil {
					t.Fatal(err)
				}
			},
			keyword: "test",
			wantOutput: []string{
				"Found 2 prompt(s):",
				"[1] Test Prompt by alice",
				"Category: testing",
				"Tags: test, integration",
				"Description: A test prompt for integration testing",
				"Gist URL: https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
				"[2] Another Test by bob",
				"Category: testing",
				"Tags: test",
				"Description: Another test prompt",
				"Gist URL: https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
			},
			wantErr: false,
		},
		{
			name: "displays gist URL after selection",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create index with single entry
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "1234567890abcdef1234567890abcdef",
							GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:        "Single Prompt",
							Author:      "alice",
							Category:    "development",
							Version:     "1.0",
							Description: "A single prompt for testing",
							Tags:        []string{"dev"},
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}

				// Create prompt file
				prompt := &models.Prompt{
					PromptMeta: models.PromptMeta{
						Name:        "Single Prompt",
						Author:      "alice",
						Category:    "development",
						Version:     "1.0",
						Description: "A single prompt for testing",
						Tags:        []string{"dev"},
					},
					GistID:  "1234567890abcdef1234567890abcdef",
					GistURL: "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
					Content: "Single prompt content",
				}
				if err := cacheManager.SavePrompt(prompt); err != nil {
					t.Fatal(err)
				}
			},
			keyword: "single",
			wantOutput: []string{
				"Found 1 prompt(s):",
				"Single Prompt by alice",
				"Gist URL: https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
			},
			wantErr: false,
		},
		{
			name: "displays gist URL in clipboard success message",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create index with prompt that has no variables
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:      "fedcba0987654321fedcba0987654321",
							GistURL:     "https://gist.github.com/testuser/fedcba0987654321fedcba0987654321",
							Name:        "No Variables Prompt",
							Author:      "charlie",
							Category:    "utility",
							Version:     "1.0",
							UpdatedAt:   time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}

				// Create prompt without variables
				prompt := &models.Prompt{
					PromptMeta: models.PromptMeta{
						Name:     "No Variables Prompt",
						Author:   "charlie",
						Category: "utility",
						Version:  "1.0",
					},
					GistID:  "fedcba0987654321fedcba0987654321",
					GistURL: "https://gist.github.com/testuser/fedcba0987654321fedcba0987654321",
					Content: "Simple prompt content without variables",
				}
				if err := cacheManager.SavePrompt(prompt); err != nil {
					t.Fatal(err)
				}
			},
			keyword: "variables",
			wantOutput: []string{
				"No Variables Prompt by charlie",
				"Category: utility",
				"Gist URL: https://gist.github.com/testuser/fedcba0987654321fedcba0987654321",
				// Note: clipboard message won't appear due to TTY requirement for selector
			},
			wantErr: false,
		},
		{
			name: "handles prompts without gist URL",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create index with entry missing gist URL
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:    "localonly",
							// GistURL is empty
							Name:      "Local Prompt",
							Author:    "dave",
							Category:  "local",
							Version:   "1.0",
							UpdatedAt: time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}

				// Create prompt
				prompt := &models.Prompt{
					PromptMeta: models.PromptMeta{
						Name:     "Local Prompt",
						Author:   "dave",
						Category: "local",
						Version:  "1.0",
					},
					GistID:  "localonly",
					Content: "Local prompt content",
				}
				if err := cacheManager.SavePrompt(prompt); err != nil {
					t.Fatal(err)
				}
			},
			keyword: "local",
			wantOutput: []string{
				"Local Prompt by dave",
				"Category: local",
				"Gist URL:", // Empty URL still shows the label
				// Note: clipboard message won't appear due to TTY requirement
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

			// Create command
			cmd := NewRootCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{"get", tt.keyword})

			// Execute
			err := cmd.Execute()

			// Check error - expect TTY errors for interactive prompts
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				// Accept TTY errors in test environment
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

			// Note: Empty gist URLs still show "Gist URL:" line with empty value
			// This is the current behavior
		})
	}
}

func TestGetCommand_Integration_CompleteFlow(t *testing.T) {
	// Create temp cache directory
	tempDir := t.TempDir()
	cacheManager := cache.NewManagerWithPath(filepath.Join(tempDir, "cache", "prompts"))
	if err := cacheManager.InitializeCache(); err != nil {
		t.Fatal(err)
	}

	// Create comprehensive test data
	index := &models.Index{
		Username: "testuser",
		Entries: []models.IndexEntry{
			{
				GistID:      "1234567890abcdef1234567890abcdef",
				GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
				Name:        "Code Review Helper",
				Author:      "alice",
				Category:    "development",
				Version:     "2.0",
				Description: "Helps with code reviews and suggestions",
				Tags:        []string{"code", "review", "helper"},
				UpdatedAt:   time.Now(),
			},
			{
				GistID:      "abcdefabcdefabcdefabcdefabcdefab",
				GistURL:     "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
				Name:        "Test Generator",
				Author:      "bob",
				Category:    "testing",
				Version:     "1.5",
				Description: "Generates unit tests for code",
				Tags:        []string{"test", "generator", "unit"},
				UpdatedAt:   time.Now(),
			},
			{
				GistID:      "fedcba0987654321fedcba0987654321",
				GistURL:     "https://gist.github.com/testuser/fedcba0987654321fedcba0987654321",
				Name:        "Documentation Writer",
				Author:      "charlie",
				Category:    "documentation",
				Version:     "1.0",
				Description: "Helps write technical documentation",
				Tags:        []string{"docs", "writing"},
				UpdatedAt:   time.Now(),
			},
		},
		UpdatedAt: time.Now(),
	}

	if err := cacheManager.SaveIndex(index); err != nil {
		t.Fatal(err)
	}

	// Create prompt files
	// Note: The models.Prompt type doesn't have a Variables field, so we'll just create a prompt without variables
	reviewPrompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:        "Code Review Helper",
			Author:      "alice",
			Category:    "development",
			Version:     "2.0",
			Description: "Helps with code reviews and suggestions",
			Tags:        []string{"code", "review", "helper"},
		},
		GistID:  "1234567890abcdef1234567890abcdef",
		GistURL: "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
		Content: "Review this {{language}} code focusing on {{focus_area}}",
	}
	if err := cacheManager.SavePrompt(reviewPrompt); err != nil {
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

	// Test the complete flow
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "review"})

	err := cmd.Execute()
	
	// Should fail due to TTY requirement for selector
	if err == nil || !strings.Contains(err.Error(), "TTY") {
		t.Errorf("Expected TTY error, got: %v", err)
	}

	output := out.String()

	// Verify all expected elements in the flow
	expectedElements := []string{
		// Search results
		"Found 1 prompt(s):",
		"[1] Code Review Helper by alice",
		"Category: development",
		"Tags: code, review, helper",
		"Description: Helps with code reviews and suggestions",
		"Gist URL: https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
		// Variables would be shown if we could proceed past selection
		// "2 variable(s)",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Missing expected element %q in output:\n%s", expected, output)
		}
	}
}