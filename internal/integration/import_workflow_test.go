package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/imports"
	"github.com/grigri201/prompt-vault/internal/models"
)

// TestImportWorkflow_EndToEnd tests the complete import workflow
func TestImportWorkflow_EndToEnd(t *testing.T) {
	ctx := context.Background()

	// Setup mock gist client
	mockClient := &MockImportGistClient{
		gists: make(map[string]*github.Gist),
	}

	// Add a public gist to import
	publicGist := &github.Gist{
		ID:          github.String("public123"),
		HTMLURL:     github.String("https://gist.github.com/otheruser/public123"),
		Description: github.String("Awesome Prompt"),
		Public:      github.Bool(true),
		Files: map[github.GistFilename]github.GistFile{
			"awesome-prompt.yaml": {
				Content: github.String(`---
name: Awesome Prompt
author: otheruser
category: productivity
tags: [awesome, productivity, ai]
version: "1.0.0"
description: An awesome prompt for productivity
---

You are an AI assistant focused on productivity.

Task: {task}
Context: {context}

Please provide a detailed response.`),
			},
		},
	}
	mockClient.gists["public123"] = publicGist

	// Setup mock UI
	mockUI := &MockUI{
		confirmResponses: []bool{true}, // Will confirm update when asked
	}

	// Create import manager
	manager := imports.NewManager(mockClient, mockUI)

	// Create an index
	index := &models.Index{
		Username:        "testuser",
		Entries:         []models.IndexEntry{},
		ImportedEntries: []models.IndexEntry{},
		UpdatedAt:       time.Now(),
	}

	// Test 1: Import a new prompt
	t.Run("first_import", func(t *testing.T) {
		result, err := manager.ImportPrompt(ctx, "https://gist.github.com/otheruser/public123", index)
		if err != nil {
			t.Fatalf("ImportPrompt() error = %v", err)
		}

		if result.IsUpdate {
			t.Error("Expected IsUpdate = false for first import")
		}

		if result.GistID != "public123" {
			t.Errorf("Expected GistID = public123, got %s", result.GistID)
		}

		// Verify index was updated
		if len(index.ImportedEntries) != 1 {
			t.Fatalf("Expected 1 imported entry, got %d", len(index.ImportedEntries))
		}

		entry := index.ImportedEntries[0]
		if entry.Name != "Awesome Prompt" {
			t.Errorf("Expected Name = Awesome Prompt, got %s", entry.Name)
		}
		if entry.Author != "otheruser" {
			t.Errorf("Expected Author = otheruser, got %s", entry.Author)
		}
		if entry.Version != "1.0.0" {
			t.Errorf("Expected Version = 1.0.0, got %s", entry.Version)
		}
	})

	// Test 2: Update the gist and import again
	t.Run("update_import", func(t *testing.T) {
		// Update the gist
		publicGist.Files = map[github.GistFilename]github.GistFile{
			"awesome-prompt.yaml": {
				Content: github.String(`---
name: Awesome Prompt
author: otheruser
category: productivity
tags: [awesome, productivity, ai, enhanced]
version: "2.0.0"
description: An enhanced awesome prompt for productivity
---

You are an advanced AI assistant focused on productivity.

Task: {task}
Context: {context}
Priority: {priority}

Please provide a detailed and actionable response.`),
			},
		}

		// Import again
		result, err := manager.ImportPrompt(ctx, "https://gist.github.com/otheruser/public123", index)
		if err != nil {
			t.Fatalf("ImportPrompt() error = %v", err)
		}

		if !result.IsUpdate {
			t.Error("Expected IsUpdate = true for update")
		}

		if result.OldVersion != "1.0.0" {
			t.Errorf("Expected OldVersion = 1.0.0, got %s", result.OldVersion)
		}

		if result.NewVersion != "2.0.0" {
			t.Errorf("Expected NewVersion = 2.0.0, got %s", result.NewVersion)
		}

		// Verify UI was asked for confirmation
		if len(mockUI.confirmCalls) == 0 {
			t.Error("Expected confirmation to be requested")
		}

		// Verify index was updated
		if len(index.ImportedEntries) != 1 {
			t.Fatalf("Expected still 1 imported entry, got %d", len(index.ImportedEntries))
		}

		entry := index.ImportedEntries[0]
		if entry.Version != "2.0.0" {
			t.Errorf("Expected Version = 2.0.0, got %s", entry.Version)
		}
		if len(entry.Tags) != 4 {
			t.Errorf("Expected 4 tags, got %d", len(entry.Tags))
		}
	})

	// Test 3: Try to import a private gist
	t.Run("import_private_error", func(t *testing.T) {
		// Add a private gist
		privateGist := &github.Gist{
			ID:     github.String("private456"),
			Public: github.Bool(false),
		}
		mockClient.gists["private456"] = privateGist

		_, err := manager.ImportPrompt(ctx, "https://gist.github.com/testuser/private456", index)
		if err == nil {
			t.Fatal("Expected error when importing private gist")
		}

		if !contains(err.Error(), "private") {
			t.Errorf("Expected error about private gist, got: %v", err)
		}
	})
}

// TestImportWorkflow_ErrorHandling tests error scenarios
func TestImportWorkflow_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		setupMock     func(*MockImportGistClient, *MockUI)
		gistURL       string
		expectError   bool
		errorContains string
	}{
		{
			name: "invalid_url",
			setupMock: func(client *MockImportGistClient, ui *MockUI) {
				// No setup needed
			},
			gistURL:       "not-a-url",
			expectError:   true,
			errorContains: "invalid URL",
		},
		{
			name: "non_github_url",
			setupMock: func(client *MockImportGistClient, ui *MockUI) {
				// No setup needed
			},
			gistURL:       "https://example.com/gist/123",
			expectError:   true,
			errorContains: "not a GitHub gist URL",
		},
		{
			name: "gist_not_found",
			setupMock: func(client *MockImportGistClient, ui *MockUI) {
				// No gists
			},
			gistURL:       "https://gist.github.com/user/nonexistent",
			expectError:   true,
			errorContains: "not found",
		},
		{
			name: "invalid_prompt_format",
			setupMock: func(client *MockImportGistClient, ui *MockUI) {
				client.gists["invalid123"] = &github.Gist{
					ID:     github.String("invalid123"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"not-a-prompt.txt": {
							Content: github.String("This is not a valid prompt file"),
						},
					},
				}
			},
			gistURL:       "https://gist.github.com/user/invalid123",
			expectError:   true,
			errorContains: "valid prompt",
		},
		{
			name: "missing_required_fields",
			setupMock: func(client *MockImportGistClient, ui *MockUI) {
				client.gists["incomplete123"] = &github.Gist{
					ID:     github.String("incomplete123"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"incomplete.yaml": {
							Content: github.String(`---
name: Incomplete Prompt
author: testuser
---
Content without required fields`),
						},
					},
				}
			},
			gistURL:       "https://gist.github.com/user/incomplete123",
			expectError:   true,
			errorContains: "required",
		},
		{
			name: "user_cancels_update",
			setupMock: func(client *MockImportGistClient, ui *MockUI) {
				client.gists["update123"] = &github.Gist{
					ID:     github.String("update123"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"prompt.yaml": {
							Content: github.String(`---
name: Update Test
author: otheruser
category: test
tags: [test]
version: "2.0.0"
---
Updated content`),
						},
					},
				}
				// User will cancel
				ui.confirmResponses = []bool{false}
			},
			gistURL:       "https://gist.github.com/user/update123",
			expectError:   true,
			errorContains: "cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockClient := &MockImportGistClient{
				gists: make(map[string]*github.Gist),
			}
			mockUI := &MockUI{}

			tt.setupMock(mockClient, mockUI)

			// Create manager
			manager := imports.NewManager(mockClient, mockUI)

			// Create index with existing entry for update test
			index := &models.Index{
				Username: "testuser",
				ImportedEntries: []models.IndexEntry{
					{
						GistID:  "update123",
						Name:    "Update Test",
						Version: "1.0.0",
					},
				},
			}

			// Execute
			_, err := manager.ImportPrompt(ctx, tt.gistURL, index)

			// Verify
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestImportWorkflow_IndexPersistence tests that index changes persist correctly
func TestImportWorkflow_IndexPersistence(t *testing.T) {
	ctx := context.Background()

	// Setup
	mockClient := &MockImportGistClient{
		gists: make(map[string]*github.Gist),
	}

	// Add multiple gists
	for i := 1; i <= 3; i++ {
		gistID := fmt.Sprintf("gist%d", i)
		mockClient.gists[gistID] = &github.Gist{
			ID:     github.String(gistID),
			Public: github.Bool(true),
			Files: map[github.GistFilename]github.GistFile{
				"prompt.yaml": {
					Content: github.String(fmt.Sprintf(`---
name: Prompt %d
author: user%d
category: test
tags: [test]
version: "1.0.0"
---
Content %d`, i, i, i)),
				},
			},
		}
	}

	mockUI := &MockUI{}
	manager := imports.NewManager(mockClient, mockUI)

	// Start with empty index
	index := &models.Index{
		Username:        "testuser",
		Entries:         []models.IndexEntry{},
		ImportedEntries: []models.IndexEntry{},
	}

	// Import multiple prompts
	for i := 1; i <= 3; i++ {
		gistURL := fmt.Sprintf("https://gist.github.com/user/gist%d", i)
		_, err := manager.ImportPrompt(ctx, gistURL, index)
		if err != nil {
			t.Fatalf("Failed to import gist%d: %v", i, err)
		}
	}

	// Verify all imports are in the index
	if len(index.ImportedEntries) != 3 {
		t.Fatalf("Expected 3 imported entries, got %d", len(index.ImportedEntries))
	}

	// Verify each entry
	for i := 1; i <= 3; i++ {
		found := false
		expectedName := fmt.Sprintf("Prompt %d", i)
		for _, entry := range index.ImportedEntries {
			if entry.Name == expectedName {
				found = true
				if entry.GistID != fmt.Sprintf("gist%d", i) {
					t.Errorf("Entry %s has wrong GistID", expectedName)
				}
				break
			}
		}
		if !found {
			t.Errorf("Entry %s not found in index", expectedName)
		}
	}
}

// MockImportGistClient extends MockGistClient with import-specific methods
type MockImportGistClient struct {
	gists        map[string]*github.Gist
	getGistError error
}

func (m *MockImportGistClient) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	if m.getGistError != nil {
		return nil, m.getGistError
	}
	gist, exists := m.gists[gistID]
	if !exists {
		return nil, fmt.Errorf("gist not found")
	}
	return gist, nil
}

func (m *MockImportGistClient) GetGistByURL(ctx context.Context, gistURL string) (*github.Gist, error) {
	// Extract gist ID from URL
	gistID, err := m.ExtractGistID(gistURL)
	if err != nil {
		return nil, err
	}
	return m.GetGist(ctx, gistID)
}

func (m *MockImportGistClient) ExtractGistID(gistURL string) (string, error) {
	// Simple extraction for testing
	if !contains(gistURL, "gist.github.com") {
		return "", fmt.Errorf("not a GitHub gist URL")
	}

	if gistURL == "not-a-url" {
		return "", fmt.Errorf("invalid URL")
	}

	// Extract last part of URL as gist ID
	parts := strings.Split(gistURL, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid gist URL format")
	}

	return parts[len(parts)-1], nil
}
