package share

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/models"
)

// MockGistClient is a mock implementation of the gist client interface
type MockGistClient struct {
	// GetGist mock
	GetGistFunc  func(ctx context.Context, gistID string) (*github.Gist, error)
	GetGistCalls []string

	// CreatePublicGist mock
	CreatePublicGistFunc  func(ctx context.Context, gistName, description, content string) (string, string, error)
	CreatePublicGistCalls []struct {
		GistName    string
		Description string
		Content     string
	}

	// UpdateGist mock
	UpdateGistFunc  func(ctx context.Context, gistID, gistName, description, content string) (string, error)
	UpdateGistCalls []struct {
		GistID      string
		GistName    string
		Description string
		Content     string
	}

	// ListUserGists mock
	ListUserGistsFunc  func(ctx context.Context, username string) ([]*github.Gist, error)
	ListUserGistsCalls []string
}

func (m *MockGistClient) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	m.GetGistCalls = append(m.GetGistCalls, gistID)
	if m.GetGistFunc != nil {
		return m.GetGistFunc(ctx, gistID)
	}
	return nil, fmt.Errorf("GetGist not implemented")
}

func (m *MockGistClient) CreatePublicGist(ctx context.Context, gistName, description, content string) (string, string, error) {
	m.CreatePublicGistCalls = append(m.CreatePublicGistCalls, struct {
		GistName    string
		Description string
		Content     string
	}{gistName, description, content})
	if m.CreatePublicGistFunc != nil {
		return m.CreatePublicGistFunc(ctx, gistName, description, content)
	}
	return "", "", fmt.Errorf("CreatePublicGist not implemented")
}

func (m *MockGistClient) UpdateGist(ctx context.Context, gistID, gistName, description, content string) (string, error) {
	m.UpdateGistCalls = append(m.UpdateGistCalls, struct {
		GistID      string
		GistName    string
		Description string
		Content     string
	}{gistID, gistName, description, content})
	if m.UpdateGistFunc != nil {
		return m.UpdateGistFunc(ctx, gistID, gistName, description, content)
	}
	return "", fmt.Errorf("UpdateGist not implemented")
}

func (m *MockGistClient) ListUserGists(ctx context.Context, username string) ([]*github.Gist, error) {
	m.ListUserGistsCalls = append(m.ListUserGistsCalls, username)
	if m.ListUserGistsFunc != nil {
		return m.ListUserGistsFunc(ctx, username)
	}
	return nil, fmt.Errorf("ListUserGists not implemented")
}

// MockUI is a mock implementation of the UI interface
type MockUI struct {
	ConfirmFunc  func(message string) (bool, error)
	ConfirmCalls []string
}

func (m *MockUI) Confirm(message string) (bool, error) {
	m.ConfirmCalls = append(m.ConfirmCalls, message)
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(message)
	}
	return false, nil
}

func TestShareManager_SharePrompt_NewPublicGist(t *testing.T) {
	ctx := context.Background()

	// Setup mock gist client
	mockClient := &MockGistClient{
		GetGistFunc: func(ctx context.Context, gistID string) (*github.Gist, error) {
			if gistID == "private123" {
				content := `---
name: Test Prompt
author: testuser
category: test
tags: [test]
version: "1.0.0"
---

This is a test prompt.`
				return &github.Gist{
					ID:          github.String("private123"),
					HTMLURL:     github.String("https://gist.github.com/testuser/private123"),
					Description: github.String("Test prompt"),
					Public:      github.Bool(false), // Private gist
					Files: map[github.GistFilename]github.GistFile{
						"test-prompt.yaml": {
							Content: github.String(content),
						},
					},
				}, nil
			}
			return nil, fmt.Errorf("gist not found")
		},
		ListUserGistsFunc: func(ctx context.Context, username string) ([]*github.Gist, error) {
			// Return empty list - no existing public version
			return []*github.Gist{}, nil
		},
		CreatePublicGistFunc: func(ctx context.Context, gistName, description, content string) (string, string, error) {
			// Verify parent field is added
			if !contains(content, "parent: private123") {
				t.Error("Expected content to contain parent field")
			}
			return "public456", "https://gist.github.com/testuser/public456", nil
		},
	}

	mockUI := &MockUI{
		ConfirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		ui:         mockUI,
		username:   "testuser",
	}

	// Test sharing a private gist
	result, err := manager.SharePrompt(ctx, "private123")
	if err != nil {
		t.Fatalf("SharePrompt() error = %v", err)
	}

	// Verify result
	if result.PublicGistID != "public456" {
		t.Errorf("Expected PublicGistID = public456, got %s", result.PublicGistID)
	}
	if result.PublicGistURL != "https://gist.github.com/testuser/public456" {
		t.Errorf("Expected PublicGistURL = https://gist.github.com/testuser/public456, got %s", result.PublicGistURL)
	}
	if result.IsUpdate {
		t.Error("Expected IsUpdate = false for new share")
	}

	// Verify calls
	if len(mockClient.GetGistCalls) != 1 || mockClient.GetGistCalls[0] != "private123" {
		t.Errorf("Expected GetGist to be called with private123, got %v", mockClient.GetGistCalls)
	}
	if len(mockClient.CreatePublicGistCalls) != 1 {
		t.Error("Expected CreatePublicGist to be called once")
	}
}

func TestShareManager_SharePrompt_UpdateExistingPublicGist(t *testing.T) {
	ctx := context.Background()

	// Setup mock gist client
	mockClient := &MockGistClient{
		GetGistFunc: func(ctx context.Context, gistID string) (*github.Gist, error) {
			if gistID == "private123" {
				content := `---
name: Test Prompt
author: testuser
category: test
tags: [test]
version: "2.0.0"
---

This is an updated test prompt.`
				return &github.Gist{
					ID:          github.String("private123"),
					HTMLURL:     github.String("https://gist.github.com/testuser/private123"),
					Description: github.String("Test prompt - Updated"),
					Public:      github.Bool(false),
					Files: map[github.GistFilename]github.GistFile{
						"test-prompt.yaml": {
							Content: github.String(content),
						},
					},
				}, nil
			}
			return nil, fmt.Errorf("gist not found")
		},
		ListUserGistsFunc: func(ctx context.Context, username string) ([]*github.Gist, error) {
			// Return existing public gist with parent field
			content := `---
name: Test Prompt
author: testuser
category: test
tags: [test]
version: "1.0.0"
parent: private123
---

This is a test prompt.`
			return []*github.Gist{
				{
					ID:          github.String("public456"),
					HTMLURL:     github.String("https://gist.github.com/testuser/public456"),
					Description: github.String("Test prompt"),
					Public:      github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test-prompt.yaml": {
							Content: github.String(content),
						},
					},
				},
			}, nil
		},
		UpdateGistFunc: func(ctx context.Context, gistID, gistName, description, content string) (string, error) {
			if gistID != "public456" {
				t.Errorf("Expected to update gist public456, got %s", gistID)
			}
			// Verify parent field is preserved
			if !contains(content, "parent: private123") {
				t.Error("Expected content to contain parent field")
			}
			// Verify version is updated
			if !contains(content, "version: 2.0.0") && !contains(content, "version: \"2.0.0\"") {
				t.Error("Expected content to have updated version")
			}
			return "https://gist.github.com/testuser/public456", nil
		},
	}

	mockUI := &MockUI{
		ConfirmFunc: func(message string) (bool, error) {
			// Verify confirmation message
			if !contains(message, "already exists") {
				t.Errorf("Expected confirmation message about existing gist, got: %s", message)
			}
			return true, nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		ui:         mockUI,
		username:   "testuser",
	}

	// Test updating existing public gist
	result, err := manager.SharePrompt(ctx, "private123")
	if err != nil {
		t.Fatalf("SharePrompt() error = %v", err)
	}

	// Verify result
	if result.PublicGistID != "public456" {
		t.Errorf("Expected PublicGistID = public456, got %s", result.PublicGistID)
	}
	if !result.IsUpdate {
		t.Error("Expected IsUpdate = true for update")
	}

	// Verify UI was called for confirmation
	if len(mockUI.ConfirmCalls) != 1 {
		t.Error("Expected Confirm to be called once")
	}
}

func TestShareManager_SharePrompt_PublicGistError(t *testing.T) {
	ctx := context.Background()

	mockClient := &MockGistClient{
		GetGistFunc: func(ctx context.Context, gistID string) (*github.Gist, error) {
			return &github.Gist{
				ID:     github.String("public123"),
				Public: github.Bool(true), // Already public
			}, nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		username:   "testuser",
	}

	// Test sharing an already public gist
	_, err := manager.SharePrompt(ctx, "public123")
	if err == nil {
		t.Fatal("Expected error when sharing public gist")
	}
	if !contains(err.Error(), "already public") {
		t.Errorf("Expected error about gist being public, got: %v", err)
	}
}

func TestShareManager_SharePrompt_GistNotFound(t *testing.T) {
	ctx := context.Background()

	mockClient := &MockGistClient{
		GetGistFunc: func(ctx context.Context, gistID string) (*github.Gist, error) {
			return nil, fmt.Errorf("gist not found")
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		username:   "testuser",
	}

	// Test with non-existent gist
	_, err := manager.SharePrompt(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error when gist not found")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("Expected error about gist not found, got: %v", err)
	}
}

func TestShareManager_SharePrompt_UserCancelsUpdate(t *testing.T) {
	ctx := context.Background()

	mockClient := &MockGistClient{
		GetGistFunc: func(ctx context.Context, gistID string) (*github.Gist, error) {
			return &github.Gist{
				ID:     github.String("private123"),
				Public: github.Bool(false),
				Files: map[github.GistFilename]github.GistFile{
					"test.yaml": {
						Content: github.String(`---
name: Test Prompt
author: testuser
category: test
tags: [test]
---
content`),
					},
				},
			}, nil
		},
		ListUserGistsFunc: func(ctx context.Context, username string) ([]*github.Gist, error) {
			// Return existing public version
			return []*github.Gist{
				{
					ID:     github.String("public456"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String("---\nparent: private123\n---\nold content"),
						},
					},
				},
			}, nil
		},
	}

	mockUI := &MockUI{
		ConfirmFunc: func(message string) (bool, error) {
			// User cancels update
			return false, nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		ui:         mockUI,
		username:   "testuser",
	}

	// Test user cancelling update
	_, err := manager.SharePrompt(ctx, "private123")
	if err == nil {
		t.Fatal("Expected error when user cancels")
	}
	if !contains(err.Error(), "cancelled") {
		t.Errorf("Expected error about cancellation, got: %v", err)
	}
}

func TestShareManager_findExistingPublicGist(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		parentID      string
		userGists     []*github.Gist
		expected      string
		expectError   bool
		errorContains string
	}{
		{
			name:     "finds existing public gist with matching parent",
			parentID: "private123",
			userGists: []*github.Gist{
				{
					ID:          github.String("public456"),
					Description: github.String("Test prompt"),
					Public:      github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
name: Test Prompt
parent: private123
---
Content`),
						},
					},
				},
			},
			expected:    "public456",
			expectError: false,
		},
		{
			name:        "returns empty when no public version exists",
			parentID:    "private123",
			userGists:   []*github.Gist{},
			expected:    "",
			expectError: false,
		},
		{
			name:     "skips gists without parent field",
			parentID: "private123",
			userGists: []*github.Gist{
				{
					ID:     github.String("public456"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
name: Test Prompt
---
Content`),
						},
					},
				},
			},
			expected:    "",
			expectError: false,
		},
		{
			name:     "returns first match when multiple gists have same parent",
			parentID: "private123",
			userGists: []*github.Gist{
				{
					ID:     github.String("public456"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
parent: private123
---`),
						},
					},
				},
				{
					ID:     github.String("public789"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
parent: private123
---`),
						},
					},
				},
			},
			expected:    "public456",
			expectError: false,
		},
		{
			name:     "skips private gists",
			parentID: "private123",
			userGists: []*github.Gist{
				{
					ID:     github.String("private456"),
					Public: github.Bool(false),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
parent: private123
---`),
						},
					},
				},
			},
			expected:    "",
			expectError: false,
		},
		{
			name:     "skips gists with different parent",
			parentID: "private123",
			userGists: []*github.Gist{
				{
					ID:     github.String("public456"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
parent: private999
---`),
						},
					},
				},
			},
			expected:    "",
			expectError: false,
		},
		{
			name:     "handles gists with multiple files",
			parentID: "private123",
			userGists: []*github.Gist{
				{
					ID:     github.String("public456"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"README.md": {
							Content: github.String("# Test"),
						},
						"prompt.yaml": {
							Content: github.String(`---
parent: private123
---`),
						},
					},
				},
			},
			expected:    "public456",
			expectError: false,
		},
		{
			name:     "handles malformed YAML gracefully",
			parentID: "private123",
			userGists: []*github.Gist{
				{
					ID:     github.String("public456"),
					Public: github.Bool(true),
					Files: map[github.GistFilename]github.GistFile{
						"test.yaml": {
							Content: github.String(`---
invalid yaml: [[[
parent: private123
---`),
						},
					},
				},
			},
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGistClient{
				ListUserGistsFunc: func(ctx context.Context, username string) ([]*github.Gist, error) {
					return tt.userGists, nil
				},
			}

			manager := &Manager{
				gistClient: mockClient,
				username:   "testuser",
			}

			result, err := manager.findExistingPublicGist(ctx, tt.parentID)

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
				if result != tt.expected {
					t.Errorf("Expected result %q, got %q", tt.expected, result)
				}
			}

			// Verify ListUserGists was called
			if len(mockClient.ListUserGistsCalls) != 1 || mockClient.ListUserGistsCalls[0] != "testuser" {
				t.Errorf("Expected ListUserGists to be called with 'testuser', got %v", mockClient.ListUserGistsCalls)
			}
		})
	}
}

func TestShareManager_findExistingPublicGist_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Test API error
	mockClient := &MockGistClient{
		ListUserGistsFunc: func(ctx context.Context, username string) ([]*github.Gist, error) {
			return nil, fmt.Errorf("API error: rate limit exceeded")
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		username:   "testuser",
	}

	result, err := manager.findExistingPublicGist(ctx, "private123")
	if err == nil {
		t.Fatal("Expected error when API fails")
	}
	if !contains(err.Error(), "API error") {
		t.Errorf("Expected API error, got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result on error, got: %s", result)
	}
}

func TestShareManager_createPublicGist(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		prompt          *models.Prompt
		createFunc      func(ctx context.Context, gistName, description, content string) (string, string, error)
		expectedID      string
		expectedURL     string
		expectError     bool
		errorContains   string
		validateContent func(t *testing.T, content string)
	}{
		{
			name: "creates public gist with parent field",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:        "Test Prompt",
					Author:      "testuser",
					Category:    "test",
					Tags:        []string{"test"},
					Version:     "1.0.0",
					Description: "Test description",
				},
				GistID:  "private123",
				Content: "This is the prompt content.",
			},
			createFunc: func(ctx context.Context, gistName, description, content string) (string, string, error) {
				return "public456", "https://gist.github.com/testuser/public456", nil
			},
			expectedID:  "public456",
			expectedURL: "https://gist.github.com/testuser/public456",
			expectError: false,
			validateContent: func(t *testing.T, content string) {
				// Should contain parent field
				if !contains(content, "parent: private123") {
					t.Error("Expected content to contain parent field")
				}
				// Should preserve original metadata
				if !contains(content, "name: Test Prompt") {
					t.Error("Expected content to contain name")
				}
				if !contains(content, "version: 1.0.0") && !contains(content, "version: \"1.0.0\"") {
					t.Error("Expected content to contain version")
				}
				// Should contain content
				if !contains(content, "This is the prompt content.") {
					t.Error("Expected content to contain prompt content")
				}
			},
		},
		{
			name: "creates public gist with existing parent field",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:     "Test Prompt",
					Author:   "testuser",
					Category: "test",
					Tags:     []string{"test"},
					Version:  "2.0.0",
					Parent:   "old-parent", // Should be overwritten
				},
				GistID:  "private123",
				Content: "Updated content",
			},
			createFunc: func(ctx context.Context, gistName, description, content string) (string, string, error) {
				return "public789", "https://gist.github.com/testuser/public789", nil
			},
			expectedID:  "public789",
			expectedURL: "https://gist.github.com/testuser/public789",
			expectError: false,
			validateContent: func(t *testing.T, content string) {
				// Should update parent field to current private gist ID
				if !contains(content, "parent: private123") {
					t.Error("Expected content to contain updated parent field")
				}
				if contains(content, "parent: old-parent") {
					t.Error("Expected old parent to be replaced")
				}
			},
		},
		{
			name: "handles API error",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:     "Test Prompt",
					Author:   "testuser",
					Category: "test",
					Tags:     []string{"test"},
				},
				GistID:  "private123",
				Content: "Content",
			},
			createFunc: func(ctx context.Context, gistName, description, content string) (string, string, error) {
				return "", "", fmt.Errorf("API error: rate limit exceeded")
			},
			expectError:   true,
			errorContains: "API error",
		},
		{
			name: "uses gist description from prompt",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:        "Complex Prompt",
					Author:      "testuser",
					Category:    "advanced",
					Tags:        []string{"complex", "advanced"},
					Description: "This is a complex prompt for advanced users",
				},
				GistID:  "private123",
				Content: "Complex content",
			},
			createFunc: func(ctx context.Context, gistName, description, content string) (string, string, error) {
				// Verify description is passed correctly
				if description != "This is a complex prompt for advanced users" {
					t.Errorf("Expected description %q, got %q", "This is a complex prompt for advanced users", description)
				}
				return "public999", "https://gist.github.com/testuser/public999", nil
			},
			expectedID:  "public999",
			expectedURL: "https://gist.github.com/testuser/public999",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGistClient{
				CreatePublicGistFunc: tt.createFunc,
			}

			manager := &Manager{
				gistClient: mockClient,
				username:   "testuser",
			}

			result, err := manager.createPublicGist(ctx, tt.prompt)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			if result.PublicGistID != tt.expectedID {
				t.Errorf("Expected PublicGistID %q, got %q", tt.expectedID, result.PublicGistID)
			}

			if result.PublicGistURL != tt.expectedURL {
				t.Errorf("Expected PublicGistURL %q, got %q", tt.expectedURL, result.PublicGistURL)
			}

			if result.IsUpdate {
				t.Error("Expected IsUpdate to be false for new gist creation")
			}

			// Validate the content that was passed to CreatePublicGist
			if len(mockClient.CreatePublicGistCalls) != 1 {
				t.Fatalf("Expected CreatePublicGist to be called once, got %d calls", len(mockClient.CreatePublicGistCalls))
			}

			call := mockClient.CreatePublicGistCalls[0]
			if tt.validateContent != nil {
				tt.validateContent(t, call.Content)
			}
		})
	}
}

func TestShareManager_createPublicGist_ContentFormatting(t *testing.T) {
	ctx := context.Background()

	prompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:        "Format Test",
			Author:      "testuser",
			Category:    "test",
			Tags:        []string{"format", "test"},
			Version:     "1.0.0",
			Description: "Testing content formatting",
		},
		GistID:  "private123",
		Content: "Line 1\nLine 2\nLine 3",
	}

	var capturedContent string
	mockClient := &MockGistClient{
		CreatePublicGistFunc: func(ctx context.Context, gistName, description, content string) (string, string, error) {
			capturedContent = content
			return "public456", "https://gist.github.com/testuser/public456", nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		username:   "testuser",
	}

	_, err := manager.createPublicGist(ctx, prompt)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify YAML front matter format
	if !strings.HasPrefix(capturedContent, "---\n") {
		t.Error("Expected content to start with YAML delimiter")
	}

	// Verify content ends properly
	if !contains(capturedContent, "\n---\n") {
		t.Error("Expected content to have closing YAML delimiter")
	}

	// Verify all metadata fields are present (order doesn't matter)
	expectedFields := []string{
		"name: Format Test",
		"author: testuser",
		"category: test",
		"description: Testing content formatting",
		"parent: private123",
	}

	for _, field := range expectedFields {
		if !contains(capturedContent, field) {
			t.Errorf("Expected content to contain %q", field)
		}
	}

	// Check tags format (YAML can format arrays differently)
	if !contains(capturedContent, "tags:") {
		t.Error("Expected content to contain tags field")
	}
	if !contains(capturedContent, "- format") && !contains(capturedContent, "[format") {
		t.Error("Expected content to contain 'format' tag")
	}
	if !contains(capturedContent, "- test") && !contains(capturedContent, "test]") {
		t.Error("Expected content to contain 'test' tag")
	}

	// Check version (might not have quotes)
	if !contains(capturedContent, "version: 1.0.0") && !contains(capturedContent, "version: \"1.0.0\"") {
		t.Error("Expected content to contain version field")
	}

	// Verify content is preserved
	if !contains(capturedContent, "Line 1\nLine 2\nLine 3") {
		t.Error("Expected original content to be preserved")
	}
}

func TestShareManager_updatePublicGist(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		gistID         string
		prompt         *models.Prompt
		updateFunc     func(ctx context.Context, gistID, gistName, description, content string) (string, error)
		expectedURL    string
		expectError    bool
		errorContains  string
		validateUpdate func(t *testing.T, gistID, content string)
	}{
		{
			name:   "updates public gist with new content",
			gistID: "public456",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:        "Updated Prompt",
					Author:      "testuser",
					Category:    "test",
					Tags:        []string{"test", "updated"},
					Version:     "2.0.0",
					Description: "Updated description",
				},
				GistID:  "private123",
				Content: "This is updated content.",
			},
			updateFunc: func(ctx context.Context, gistID, gistName, description, content string) (string, error) {
				if gistID != "public456" {
					t.Errorf("Expected gistID %q, got %q", "public456", gistID)
				}
				return "https://gist.github.com/testuser/public456", nil
			},
			expectedURL: "https://gist.github.com/testuser/public456",
			expectError: false,
			validateUpdate: func(t *testing.T, gistID, content string) {
				// Should preserve parent field
				if !contains(content, "parent: private123") {
					t.Error("Expected content to preserve parent field")
				}
				// Should update version
				if !contains(content, "version: 2.0.0") && !contains(content, "version: \"2.0.0\"") {
					t.Error("Expected content to have updated version")
				}
				// Should update content
				if !contains(content, "This is updated content.") {
					t.Error("Expected content to be updated")
				}
			},
		},
		{
			name:   "handles API error during update",
			gistID: "public456",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:     "Test Prompt",
					Author:   "testuser",
					Category: "test",
					Tags:     []string{"test"},
				},
				GistID:  "private123",
				Content: "Content",
			},
			updateFunc: func(ctx context.Context, gistID, gistName, description, content string) (string, error) {
				return "", fmt.Errorf("API error: gist not found")
			},
			expectError:   true,
			errorContains: "API error",
		},
		{
			name:   "preserves parent field during update",
			gistID: "public456",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:     "Test Prompt",
					Author:   "testuser",
					Category: "test",
					Tags:     []string{"test"},
					Parent:   "should-be-overwritten", // This should be replaced with GistID
				},
				GistID:  "private789",
				Content: "New content",
			},
			updateFunc: func(ctx context.Context, gistID, gistName, description, content string) (string, error) {
				return "https://gist.github.com/testuser/public456", nil
			},
			expectedURL: "https://gist.github.com/testuser/public456",
			expectError: false,
			validateUpdate: func(t *testing.T, gistID, content string) {
				// Should use GistID as parent, not the existing Parent field
				if !contains(content, "parent: private789") {
					t.Error("Expected content to have correct parent field")
				}
				if contains(content, "parent: should-be-overwritten") {
					t.Error("Expected old parent to be replaced")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedContent string
			mockClient := &MockGistClient{
				UpdateGistFunc: func(ctx context.Context, gistID, gistName, description, content string) (string, error) {
					capturedContent = content
					if tt.updateFunc != nil {
						return tt.updateFunc(ctx, gistID, gistName, description, content)
					}
					return "", fmt.Errorf("updateFunc not provided")
				},
			}

			manager := &Manager{
				gistClient: mockClient,
				username:   "testuser",
			}

			result, err := manager.updatePublicGist(ctx, tt.gistID, tt.prompt)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			if result.PublicGistID != tt.gistID {
				t.Errorf("Expected PublicGistID %q, got %q", tt.gistID, result.PublicGistID)
			}

			if result.PublicGistURL != tt.expectedURL {
				t.Errorf("Expected PublicGistURL %q, got %q", tt.expectedURL, result.PublicGistURL)
			}

			if !result.IsUpdate {
				t.Error("Expected IsUpdate to be true for gist update")
			}

			// Validate the update
			if tt.validateUpdate != nil {
				tt.validateUpdate(t, tt.gistID, capturedContent)
			}

			// Verify UpdateGist was called
			if len(mockClient.UpdateGistCalls) != 1 {
				t.Fatalf("Expected UpdateGist to be called once, got %d calls", len(mockClient.UpdateGistCalls))
			}
		})
	}
}

func TestShareManager_updatePublicGist_UsesExistingDescription(t *testing.T) {
	ctx := context.Background()

	prompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:     "Test Prompt",
			Author:   "testuser",
			Category: "test",
			Tags:     []string{"test"},
			// No description provided
		},
		GistID:  "private123",
		Content: "Content",
	}

	var capturedDescription string
	mockClient := &MockGistClient{
		UpdateGistFunc: func(ctx context.Context, gistID, gistName, description, content string) (string, error) {
			capturedDescription = description
			return "https://gist.github.com/testuser/public456", nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		username:   "testuser",
	}

	_, err := manager.updatePublicGist(ctx, "public456", prompt)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should create a default description when none provided
	expectedDescription := "Test Prompt - test"
	if capturedDescription != expectedDescription {
		t.Errorf("Expected description %q, got %q", expectedDescription, capturedDescription)
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
