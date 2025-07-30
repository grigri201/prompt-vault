package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/share"
)

// TestShareWorkflow_EndToEnd tests the complete share workflow
func TestShareWorkflow_EndToEnd(t *testing.T) {
	ctx := context.Background()

	// Setup mock gist client
	mockClient := &MockGistClient{
		gists:     make(map[string]*github.Gist),
		userGists: []*github.Gist{},
	}

	// Add a private gist
	privateGist := &github.Gist{
		ID:          github.String("private123"),
		HTMLURL:     github.String("https://gist.github.com/testuser/private123"),
		Description: github.String("Test Prompt"),
		Public:      github.Bool(false),
		Files: map[github.GistFilename]github.GistFile{
			"test-prompt.yaml": {
				Content: github.String(`---
name: Test Prompt
author: testuser
category: testing
tags: [test, example]
version: "1.0.0"
description: A test prompt for integration testing
---

You are a helpful assistant. 

User: {user_query}

Please help the user with their request.`),
			},
		},
	}
	mockClient.gists["private123"] = privateGist

	// Setup mock UI
	mockUI := &MockUI{
		confirmResponses: []bool{true}, // Will confirm update when asked
	}

	// Create share manager
	manager := share.NewManager(mockClient, mockUI, "testuser")

	// Test 1: Share a private gist for the first time
	t.Run("first_share", func(t *testing.T) {
		result, err := manager.SharePrompt(ctx, "private123")
		if err != nil {
			t.Fatalf("SharePrompt() error = %v", err)
		}

		if result.IsUpdate {
			t.Error("Expected IsUpdate = false for first share")
		}

		if result.PublicGistID == "" {
			t.Error("Expected PublicGistID to be set")
		}

		// Verify public gist was created
		publicGist, exists := mockClient.gists[result.PublicGistID]
		if !exists {
			t.Fatal("Public gist was not created")
		}

		if !*publicGist.Public {
			t.Error("Expected gist to be public")
		}

		// Verify parent field was added
		content := *publicGist.Files["test-prompt.yaml"].Content
		if !contains(content, "parent: private123") {
			t.Error("Expected parent field to be added")
		}
	})

	// Test 2: Update the private gist and share again
	t.Run("update_share", func(t *testing.T) {
		// Update private gist
		privateGist.Files = map[github.GistFilename]github.GistFile{
			"test-prompt.yaml": {
				Content: github.String(`---
name: Test Prompt
author: testuser
category: testing
tags: [test, example, updated]
version: "2.0.0"
description: An updated test prompt
---

You are a very helpful assistant. 

User: {user_query}

Please help the user with their request in detail.`),
			},
		}

		// Share again
		result, err := manager.SharePrompt(ctx, "private123")
		if err != nil {
			t.Fatalf("SharePrompt() error = %v", err)
		}

		if !result.IsUpdate {
			t.Error("Expected IsUpdate = true for update")
		}

		// Verify UI was asked for confirmation
		if len(mockUI.confirmCalls) == 0 {
			t.Error("Expected confirmation to be requested")
		}

		// Verify public gist was updated
		publicGist, _ := mockClient.gists[result.PublicGistID]
		content := *publicGist.Files["test-prompt.yaml"].Content
		if !contains(content, "version: 2.0.0") && !contains(content, "version: \"2.0.0\"") {
			t.Error("Expected version to be updated")
		}
		if !contains(content, "updated") {
			t.Error("Expected tags to be updated")
		}
	})

	// Test 3: Try to share a public gist
	t.Run("share_public_error", func(t *testing.T) {
		// Add a public gist
		publicGist := &github.Gist{
			ID:     github.String("public456"),
			Public: github.Bool(true),
		}
		mockClient.gists["public456"] = publicGist

		_, err := manager.SharePrompt(ctx, "public456")
		if err == nil {
			t.Fatal("Expected error when sharing public gist")
		}

		if !contains(err.Error(), "already public") {
			t.Errorf("Expected error about gist being public, got: %v", err)
		}
	})
}

// TestShareWorkflow_ErrorHandling tests error scenarios
func TestShareWorkflow_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		setupMock     func(*MockGistClient, *MockUI)
		gistID        string
		expectError   bool
		errorContains string
	}{
		{
			name: "gist_not_found",
			setupMock: func(client *MockGistClient, ui *MockUI) {
				// No gists
			},
			gistID:        "nonexistent",
			expectError:   true,
			errorContains: "not found",
		},
		{
			name: "network_error",
			setupMock: func(client *MockGistClient, ui *MockUI) {
				client.getGistError = fmt.Errorf("network error")
			},
			gistID:        "any",
			expectError:   true,
			errorContains: "network error",
		},
		{
			name: "user_cancels_update",
			setupMock: func(client *MockGistClient, ui *MockUI) {
				// Add private gist
				client.gists["private123"] = &github.Gist{
					ID:     github.String("private123"),
					Public: github.Bool(false),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
name: Test
author: testuser
category: test
tags: [test]
---
Content`),
						},
					},
				}
				// Add existing public version
				client.userGists = []*github.Gist{
					{
						ID:     github.String("public456"),
						Public: github.Bool(true),
						Files: map[github.GistFilename]github.GistFile{
							"test.yaml": {
								Content: github.String(`---
name: Test
parent: private123
---
Content`),
							},
						},
					},
				}
				// User will cancel
				ui.confirmResponses = []bool{false}
			},
			gistID:        "private123",
			expectError:   true,
			errorContains: "cancelled",
		},
		{
			name: "invalid_prompt_format",
			setupMock: func(client *MockGistClient, ui *MockUI) {
				client.gists["invalid123"] = &github.Gist{
					ID:     github.String("invalid123"),
					Public: github.Bool(false),
					Files: map[github.GistFilename]github.GistFile{
						"test.txt": {
							Content: github.String("Not a valid prompt"),
						},
					},
				}
			},
			gistID:        "invalid123",
			expectError:   true,
			errorContains: "YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockClient := &MockGistClient{
				gists:     make(map[string]*github.Gist),
				userGists: []*github.Gist{},
			}
			mockUI := &MockUI{}

			tt.setupMock(mockClient, mockUI)

			// Create manager
			manager := share.NewManager(mockClient, mockUI, "testuser")

			// Execute
			_, err := manager.SharePrompt(ctx, tt.gistID)

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

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
