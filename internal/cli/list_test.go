package cli

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/models"
)

func TestListCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "shows empty list message",
			args: []string{"list"},
			wantOutput: []string{
				"No prompts found",
			},
			wantErr: false,
		},
		{
			name: "displays prompts in table format",
			args: []string{"list"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Save test index
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "Test Prompt",
							Author:    "testuser",
							Category:  "testing",
							Version:   "1.0",
							UpdatedAt: time.Now(),
						},
						{
							Name:      "Another Prompt",
							Author:    "testuser",
							Category:  "example",
							Version:   "2.0",
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
				"Name",
				"Author",
				"Category",
				"Version",
				"Test Prompt",
				"testuser",
				"testing",
				"1.0",
				"Another Prompt",
				"example",
				"2.0",
			},
			wantErr: false,
		},
		{
			name: "shows pagination info",
			args: []string{"list"},
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				// Create many entries to trigger pagination
				entries := make([]models.IndexEntry, 25)
				for i := 0; i < 25; i++ {
					entries[i] = models.IndexEntry{
						Name:      fmt.Sprintf("Prompt %d", i+1),
						Author:    "testuser",
						Category:  "test",
						Version:   "1.0",
						UpdatedAt: time.Now(),
					}
				}
				index := &models.Index{
					Username:  "testuser",
					Entries:   entries,
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			wantOutput: []string{
				"Page 1 of 2",
				"Showing 1-20 of 25",
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

func TestListCommand_Help(t *testing.T) {
	cmd := NewRootCmd()
	
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--help"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	output := out.String()
	expectedStrings := []string{
		"List all available prompt templates",
		"paginated table",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help text missing %q", expected)
		}
	}
}