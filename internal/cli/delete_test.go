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

func TestDeleteCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupCache     func(t *testing.T, cacheManager *cache.Manager)
		input          string
		wantOutput     []string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:    "requires prompt name argument",
			args:    []string{"delete"},
			wantErr: true,
		},
		{
			name: "prompt not found",
			args: []string{"delete", "nonexistent"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries:  []models.IndexEntry{},
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Error: Prompt 'nonexistent' not found",
			},
			wantErr:        true,
			wantErrMessage: "prompt not found",
		},
		{
			name: "cannot delete other user's prompt",
			args: []string{"delete", "other-prompt"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "other-prompt",
							Author:    "otheruser",
							GistID:    "gist123",
							Category:  "test",
							UpdatedAt: time.Now(),
						},
					},
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Error: You can only delete your own prompts",
				"This prompt belongs to: otheruser",
			},
			wantErr:        true,
			wantErrMessage: "permission denied",
		},
		{
			name: "cancels deletion on 'n'",
			args: []string{"delete", "my-prompt"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "my-prompt",
							Author:    "testuser",
							GistID:    "gist123",
							Category:  "test",
							UpdatedAt: time.Now(),
						},
					},
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			input: "n\n",
			wantOutput: []string{
				"Are you sure you want to delete 'my-prompt'? (y/N):",
				"Deletion cancelled",
			},
			wantErr: false,
		},
		{
			name: "successfully deletes prompt with confirmation",
			args: []string{"delete", "my-prompt"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "my-prompt",
							Author:    "testuser",
							GistID:    "gist123",
							Category:  "test",
							UpdatedAt: time.Now(),
						},
					},
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}

				// Save the prompt to cache
				prompt := &models.Prompt{
					PromptMeta: models.PromptMeta{
						Name:     "my-prompt",
						Author:   "testuser",
						Category: "test",
						Tags:     []string{"test"},
					},
					Content: "Test content",
					GistID:  "gist123",
				}
				if err := cacheManager.SavePrompt(prompt); err != nil {
					t.Fatal(err)
				}
			},
			input: "y\n",
			wantOutput: []string{
				"Are you sure you want to delete 'my-prompt'? (y/N):",
				"Successfully deleted prompt 'my-prompt'",
				"Note: This is a simplified version",
			},
			wantErr: false,
		},
		{
			name: "successfully deletes with force flag",
			args: []string{"delete", "my-prompt", "--force"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "my-prompt",
							Author:    "testuser",
							GistID:    "gist123",
							Category:  "test",
							UpdatedAt: time.Now(),
						},
					},
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Successfully deleted prompt 'my-prompt'",
				"Note: This is a simplified version",
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

			// Set input if provided
			if tt.input != "" {
				cmd.SetIn(strings.NewReader(tt.input))
			}

			// Execute
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErrMessage != "" && !strings.Contains(err.Error(), tt.wantErrMessage) {
				t.Errorf("Error message = %v, want containing %v", err.Error(), tt.wantErrMessage)
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

func TestDeleteCommand_Help(t *testing.T) {
	cmd := NewRootCmd()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"delete", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := out.String()
	expectedStrings := []string{
		"Delete a prompt template",
		"requires confirmation",
		"your own templates",
		"--force",
		"Skip confirmation prompt",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help text missing %q", expected)
		}
	}
}
