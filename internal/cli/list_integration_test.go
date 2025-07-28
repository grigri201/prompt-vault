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

func TestListCommand_Integration_GistURLDisplay(t *testing.T) {
	tests := []struct {
		name       string
		setupCache func(t *testing.T, cacheManager *cache.Manager)
		args       []string
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "displays gist URL in table header",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:    "1234567890abcdef1234567890abcdef",
							GistURL:   "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:      "Test Prompt",
							Author:    "testuser",
							Category:  "testing",
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
			args: []string{"list"},
			wantOutput: []string{
				"Name",
				"Author",
				"Category",
				"Version",
				"Updated",
				"Gist URL",
				"Test Prompt",
				"testuser",
				"testing",
				"1.0",
				"testuser/1234567890abcdef1234567890abcdef", // Truncated URL shows end part
			},
			wantErr: false,
		},
		{
			name: "handles empty gist URL gracefully",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							Name:      "No URL Prompt",
							Author:    "testuser",
							Category:  "testing",
							Version:   "1.0",
							UpdatedAt: time.Now(),
							// GistURL is empty
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{"list"},
			wantOutput: []string{
				"Gist URL",
				"No URL Prompt",
				"testuser",
			},
			wantErr: false,
		},
		{
			name: "displays multiple prompts with URLs",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:    "1234567890abcdef1234567890abcdef",
							GistURL:   "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
							Name:      "First Prompt",
							Author:    "alice",
							Category:  "development",
							Version:   "1.0",
							UpdatedAt: time.Now().Add(-24 * time.Hour),
						},
						{
							GistID:    "abcdefabcdefabcdefabcdefabcdefab",
							GistURL:   "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
							Name:      "Second Prompt",
							Author:    "bob",
							Category:  "testing",
							Version:   "2.0",
							UpdatedAt: time.Now(),
						},
						{
							GistID:    "fedcba0987654321fedcba0987654321",
							GistURL:   "https://gist.github.com/testuser/fedcba0987654321fedcba0987654321",
							Name:      "Third Prompt",
							Author:    "charlie",
							Category:  "documentation",
							Version:   "1.5",
							UpdatedAt: time.Now().Add(-48 * time.Hour),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{"list"},
			wantOutput: []string{
				"Gist URL",
				"First Prompt",
				"alice",
				"testuser/1234567890abcdef1234567890abcdef",
				"Second Prompt",
				"bob",
				"testuser/abcdefabcdefabcdefabcdefabcdefab",
				"Third Prompt",
				"charlie",
				"testuser/fedcba0987654321fedcba0987654321",
			},
			wantErr: false,
		},
		{
			name: "respects pagination with URLs",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				entries := make([]models.IndexEntry, 25)
				for i := 0; i < 25; i++ {
					gistID := fmt.Sprintf("%032d", i) // 32-char numeric string
					entries[i] = models.IndexEntry{
						GistID:    gistID,
						GistURL:   fmt.Sprintf("https://gist.github.com/testuser/%s", gistID),
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
			args: []string{"list", "--page", "2"},
			wantOutput: []string{
				"Gist URL",
				"Prompt 21", // Should show prompts 21-25 on page 2
				"Prompt 25",
				"Page 2 of 2",
			},
			wantErr: false,
		},
		{
			name: "handles very long URLs with truncation",
			setupCache: func(t *testing.T, cacheManager *cache.Manager) {
				index := &models.Index{
					Username: "testuser",
					Entries: []models.IndexEntry{
						{
							GistID:  "verylonggistidthatexceedsnormalsize123456789",
							GistURL: "https://gist.github.com/veryverylongusername/verylonggistidthatexceedsnormalsize123456789",
							Name:    "Long URL Prompt",
							Author:  "testuser",
							Category: "testing",
							Version: "1.0",
							UpdatedAt: time.Now(),
						},
					},
					UpdatedAt: time.Now(),
				}
				if err := cacheManager.SaveIndex(index); err != nil {
					t.Fatal(err)
				}
			},
			args: []string{"list"},
			wantOutput: []string{
				"Long URL Prompt",
				"verylonggistidthatexceedsnormalsize123456789", // Should show the gist ID part
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

			// Create command
			cmd := NewRootCmd()

			// Capture output
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
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

func TestListCommand_Integration_CompleteFlow(t *testing.T) {
	// Create temp cache directory
	tempDir := t.TempDir()
	cacheManager := cache.NewManagerWithPath(filepath.Join(tempDir, "cache", "prompts"))
	if err := cacheManager.InitializeCache(); err != nil {
		t.Fatal(err)
	}

	// Create a comprehensive test index
	index := &models.Index{
		Username: "testuser",
		Entries: []models.IndexEntry{
			{
				GistID:      "1234567890abcdef1234567890abcdef",
				GistURL:     "https://gist.github.com/testuser/1234567890abcdef1234567890abcdef",
				Name:        "Code Review Assistant",
				Author:      "alice",
				Category:    "development",
				Version:     "2.1",
				Description: "Helps with code reviews",
				Tags:        []string{"code", "review", "development"},
				UpdatedAt:   time.Now().Add(-2 * time.Hour),
			},
			{
				GistID:      "abcdefabcdefabcdefabcdefabcdefab",
				GistURL:     "https://gist.github.com/testuser/abcdefabcdefabcdefabcdefabcdefab",
				Name:        "Test Writer",
				Author:      "bob",
				Category:    "testing",
				Version:     "1.0",
				Description: "Generates unit tests",
				Tags:        []string{"test", "tdd", "unit"},
				UpdatedAt:   time.Now().Add(-24 * time.Hour),
			},
			{
				Name:        "Legacy Prompt",
				Author:      "charlie",
				Category:    "misc",
				Version:     "0.5",
				Description: "Old prompt without URL",
				UpdatedAt:   time.Now().Add(-72 * time.Hour),
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

	// Test list command
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute list command: %v", err)
	}

	output := out.String()

	// Verify complete table structure
	expectedElements := []string{
		// Headers
		"Name", "Author", "Category", "Version", "Updated", "Gist URL",
		// Separator line
		"----",
		// First entry
		"Code Review Assistant", "alice", "development", "2.1",
		"testuser/1234567890abcdef1234567890abcdef", // Truncated URL
		// Second entry
		"Test Writer", "bob", "testing", "1.0",
		"testuser/abcdefabcdefabcdefabcdefabcdefab", // Truncated URL
		// Third entry (no URL)
		"Legacy Prompt", "charlie", "misc", "0.5",
		// Footer
		"Last synced:", // Shows sync time instead of page
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Missing expected element %q in output:\n%s", expected, output)
		}
	}

	// Verify table alignment
	lines := strings.Split(output, "\n")
	headerLine := -1
	for i, line := range lines {
		if strings.Contains(line, "Name") && strings.Contains(line, "Gist URL") {
			headerLine = i
			break
		}
	}

	if headerLine == -1 {
		t.Error("Could not find header line in output")
	} else {
		// Check that data lines have consistent spacing
		if headerLine+2 < len(lines) {
			dataLine := lines[headerLine+2]
			if !strings.Contains(dataLine, "    ") { // Should have spaces between columns
				t.Error("Table columns not properly spaced")
			}
		}
	}
}