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

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "shows empty list message when no prompts",
			args: []string{"get"},
			wantOutput: []string{
				"No prompts found",
				"Use 'pv sync' to download prompts",
			},
			wantErr: false,
		},
		{
			name: "searches by name",
			args: []string{"get", "coding"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:        "Coding Assistant",
							Author:      "testuser",
							Category:    "development",
							Version:     "1.0",
							Description: "Help with programming tasks",
							Tags:        []string{"programming", "assistant"},
							UpdatedAt:   time.Now(),
						},
						{
							Name:        "Email Writer",
							Author:      "testuser",
							Category:    "writing",
							Version:     "1.0",
							Description: "Professional email templates",
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
				"Found 1 prompt(s):",
				"[1] Coding Assistant by testuser",
				"Category: development",
				"Tags: programming, assistant",
				"Description: Help with programming tasks",
			},
			wantErr: false,
		},
		{
			name: "searches by category",
			args: []string{"get", "writing"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:        "Email Writer",
							Author:      "testuser",
							Category:    "writing",
							Version:     "1.0",
							Description: "Professional email templates",
							UpdatedAt:   time.Now(),
						},
						{
							Name:        "Code Review",
							Author:      "testuser",
							Category:    "development",
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
			wantOutput: []string{
				"Found 1 prompt(s):",
				"[1] Email Writer by testuser",
				"Category: writing",
			},
			wantErr: false,
		},
		{
			name: "searches by tag",
			args: []string{"get", "assistant"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Coding Assistant",
							Author:    "testuser",
							Category:  "development",
							Version:   "1.0",
							Tags:      []string{"programming", "assistant"},
							UpdatedAt: time.Now(),
						},
						{
							Name:      "Writing Helper",
							Author:    "testuser",
							Category:  "writing",
							Version:   "1.0",
							Tags:      []string{"writing", "assistant", "creative"},
							UpdatedAt: time.Now(),
						},
						{
							Name:      "Data Analyzer",
							Author:    "testuser",
							Category:  "analysis",
							Version:   "1.0",
							Tags:      []string{"data", "analysis"},
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
				"Found 2 prompt(s):",
				"[1] Coding Assistant by testuser",
				"[2] Writing Helper by testuser",
			},
			wantErr: false,
		},
		{
			name: "searches by author",
			args: []string{"get", "john"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Template 1",
							Author:    "john_doe",
							Category:  "general",
							Version:   "1.0",
							UpdatedAt: time.Now(),
						},
						{
							Name:      "Template 2",
							Author:    "jane_smith",
							Category:  "general",
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
			wantOutput: []string{
				"Found 1 prompt(s):",
				"[1] Template 1 by john_doe",
			},
			wantErr: false,
		},
		{
			name: "searches by description",
			args: []string{"get", "professional"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:        "Email Writer",
							Author:      "testuser",
							Category:    "writing",
							Version:     "1.0",
							Description: "Professional email templates for business",
							UpdatedAt:   time.Now(),
						},
						{
							Name:        "Code Helper",
							Author:      "testuser",
							Category:    "development",
							Version:     "1.0",
							Description: "Casual coding assistant",
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
				"Found 1 prompt(s):",
				"[1] Email Writer by testuser",
				"Description: Professional email templates for business",
			},
			wantErr: false,
		},
		{
			name: "shows all prompts when no keyword provided",
			args: []string{"get"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Template 1",
							Author:    "testuser",
							Category:  "general",
							Version:   "1.0",
							UpdatedAt: time.Now(),
						},
						{
							Name:      "Template 2",
							Author:    "testuser",
							Category:  "general",
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
			wantOutput: []string{
				"Found 2 prompt(s):",
				"[1] Template 1 by testuser",
				"[2] Template 2 by testuser",
			},
			wantErr: false,
		},
		{
			name: "no matches found",
			args: []string{"get", "nonexistent"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Template 1",
							Author:    "testuser",
							Category:  "general",
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
			wantOutput: []string{
				"No prompts found matching 'nonexistent'",
			},
			wantErr: false,
		},
		{
			name: "case insensitive search",
			args: []string{"get", "CODING"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Coding Assistant",
							Author:    "testuser",
							Category:  "development",
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
			wantOutput: []string{
				"Found 1 prompt(s):",
				"[1] Coding Assistant by testuser",
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

func TestGetCommand_Help(t *testing.T) {
	cmd := NewRootCmd()
	
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "--help"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	output := out.String()
	expectedStrings := []string{
		"Search for prompt templates by keyword across names, categories",
		"tags, authors, and descriptions",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help text missing %q", expected)
		}
	}
}