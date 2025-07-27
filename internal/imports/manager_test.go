package imports

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/models"
)

// MockGistClient is a mock implementation of the gist client interface
type MockGistClient struct {
	GetGistFunc       func(ctx context.Context, gistID string) (*github.Gist, error)
	GetGistByURLFunc  func(ctx context.Context, gistURL string) (*github.Gist, error)
	ExtractGistIDFunc func(gistURL string) (string, error)
}

func (m *MockGistClient) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	if m.GetGistFunc != nil {
		return m.GetGistFunc(ctx, gistID)
	}
	return nil, fmt.Errorf("GetGist not implemented")
}

func (m *MockGistClient) GetGistByURL(ctx context.Context, gistURL string) (*github.Gist, error) {
	if m.GetGistByURLFunc != nil {
		return m.GetGistByURLFunc(ctx, gistURL)
	}
	return nil, fmt.Errorf("GetGistByURL not implemented")
}

func (m *MockGistClient) ExtractGistID(gistURL string) (string, error) {
	if m.ExtractGistIDFunc != nil {
		return m.ExtractGistIDFunc(gistURL)
	}
	return "", fmt.Errorf("ExtractGistID not implemented")
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

func TestImportManager_ImportPrompt_NewImport(t *testing.T) {
	ctx := context.Background()

	// Test importing a new gist
	mockClient := &MockGistClient{
		ExtractGistIDFunc: func(gistURL string) (string, error) {
			if gistURL == "https://gist.github.com/testuser/abc123" {
				return "abc123", nil
			}
			return "", fmt.Errorf("invalid URL")
		},
		GetGistByURLFunc: func(ctx context.Context, gistURL string) (*github.Gist, error) {
			content := `---
name: Import Test
author: otheruser
category: testing
tags: [import, test]
version: "1.0.0"
description: Test import
---

This is an imported prompt.`
			return &github.Gist{
				ID:          github.String("abc123"),
				HTMLURL:     github.String("https://gist.github.com/otheruser/abc123"),
				Description: github.String("Test import"),
				Public:      github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"import-test.yaml": {
						Content: github.String(content),
					},
				},
			}, nil
		},
	}

	mockUI := &MockUI{}

	manager := &Manager{
		gistClient: mockClient,
		ui:         mockUI,
	}

	// Create an index without any imported entries
	index := &models.Index{
		Username:        "testuser",
		Entries:         []models.IndexEntry{},
		ImportedEntries: []models.IndexEntry{},
		UpdatedAt:       time.Now(),
	}

	result, err := manager.ImportPrompt(ctx, "https://gist.github.com/testuser/abc123", index)
	if err != nil {
		t.Fatalf("ImportPrompt() error = %v", err)
	}

	// Verify result
	if result.GistID != "abc123" {
		t.Errorf("Expected GistID = abc123, got %s", result.GistID)
	}
	if result.IsUpdate {
		t.Error("Expected IsUpdate = false for new import")
	}

	// Verify index was updated
	if len(index.ImportedEntries) != 1 {
		t.Fatalf("Expected 1 imported entry, got %d", len(index.ImportedEntries))
	}

	entry := index.ImportedEntries[0]
	if entry.GistID != "abc123" {
		t.Errorf("Expected entry GistID = abc123, got %s", entry.GistID)
	}
	if entry.Name != "Import Test" {
		t.Errorf("Expected entry Name = Import Test, got %s", entry.Name)
	}
	if entry.Author != "otheruser" {
		t.Errorf("Expected entry Author = otheruser, got %s", entry.Author)
	}
}

func TestImportManager_ImportPrompt_UpdateExisting(t *testing.T) {
	ctx := context.Background()

	// Test updating an existing import
	mockClient := &MockGistClient{
		ExtractGistIDFunc: func(gistURL string) (string, error) {
			return "abc123", nil
		},
		GetGistByURLFunc: func(ctx context.Context, gistURL string) (*github.Gist, error) {
			content := `---
name: Import Test
author: otheruser
category: testing
tags: [import, test, updated]
version: "2.0.0"
description: Updated test import
---

This is an updated imported prompt.`
			return &github.Gist{
				ID:          github.String("abc123"),
				HTMLURL:     github.String("https://gist.github.com/otheruser/abc123"),
				Description: github.String("Updated test import"),
				Public:      github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"import-test.yaml": {
						Content: github.String(content),
					},
				},
			}, nil
		},
	}

	mockUI := &MockUI{
		ConfirmFunc: func(message string) (bool, error) {
			// Verify confirmation message mentions versions
			if !contains(message, "1.0.0") || !contains(message, "2.0.0") {
				t.Errorf("Expected confirmation message to mention versions, got: %s", message)
			}
			return true, nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
		ui:         mockUI,
	}

	// Create an index with existing imported entry
	index := &models.Index{
		Username: "testuser",
		Entries:  []models.IndexEntry{},
		ImportedEntries: []models.IndexEntry{
			{
				GistID:      "abc123",
				Name:        "Import Test",
				Author:      "otheruser",
				Category:    "testing",
				Tags:        []string{"import", "test"},
				Version:     "1.0.0",
				Description: "Test import",
			},
		},
		UpdatedAt: time.Now(),
	}

	result, err := manager.ImportPrompt(ctx, "https://gist.github.com/otheruser/abc123", index)
	if err != nil {
		t.Fatalf("ImportPrompt() error = %v", err)
	}

	// Verify result
	if result.GistID != "abc123" {
		t.Errorf("Expected GistID = abc123, got %s", result.GistID)
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

	// Verify confirmation was requested
	if len(mockUI.ConfirmCalls) != 1 {
		t.Error("Expected Confirm to be called once")
	}

	// Verify index was updated
	if len(index.ImportedEntries) != 1 {
		t.Fatalf("Expected 1 imported entry, got %d", len(index.ImportedEntries))
	}

	entry := index.ImportedEntries[0]
	if entry.Version != "2.0.0" {
		t.Errorf("Expected updated version 2.0.0, got %s", entry.Version)
	}
	if len(entry.Tags) != 3 || entry.Tags[2] != "updated" {
		t.Errorf("Expected updated tags, got %v", entry.Tags)
	}
}

func TestImportManager_ImportPrompt_PrivateGistError(t *testing.T) {
	ctx := context.Background()

	mockClient := &MockGistClient{
		ExtractGistIDFunc: func(gistURL string) (string, error) {
			return "private123", nil
		},
		GetGistByURLFunc: func(ctx context.Context, gistURL string) (*github.Gist, error) {
			return &github.Gist{
				ID:     github.String("private123"),
				Public: github.Bool(false), // Private gist
			}, nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
	}

	index := &models.Index{
		Username:        "testuser",
		ImportedEntries: []models.IndexEntry{},
	}

	// Test importing a private gist
	_, err := manager.ImportPrompt(ctx, "https://gist.github.com/testuser/private123", index)
	if err == nil {
		t.Fatal("Expected error when importing private gist")
	}
	if !contains(err.Error(), "private") {
		t.Errorf("Expected error about private gist, got: %v", err)
	}
}

func TestImportManager_ImportPrompt_InvalidPromptFormat(t *testing.T) {
	ctx := context.Background()

	mockClient := &MockGistClient{
		ExtractGistIDFunc: func(gistURL string) (string, error) {
			return "invalid123", nil
		},
		GetGistByURLFunc: func(ctx context.Context, gistURL string) (*github.Gist, error) {
			// Return gist with invalid content (no YAML front matter)
			return &github.Gist{
				ID:     github.String("invalid123"),
				Public: github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"invalid.txt": {
						Content: github.String("This is not a valid prompt file"),
					},
				},
			}, nil
		},
	}

	manager := &Manager{
		gistClient: mockClient,
	}

	index := &models.Index{
		Username:        "testuser",
		ImportedEntries: []models.IndexEntry{},
	}

	// Test importing invalid prompt
	_, err := manager.ImportPrompt(ctx, "https://gist.github.com/testuser/invalid123", index)
	if err == nil {
		t.Fatal("Expected error when importing invalid prompt")
	}
	if !contains(err.Error(), "valid prompt") || !contains(err.Error(), "YAML") {
		t.Errorf("Expected error about invalid prompt format, got: %v", err)
	}
}

func TestImportManager_ImportPrompt_UserCancelsUpdate(t *testing.T) {
	ctx := context.Background()

	mockClient := &MockGistClient{
		ExtractGistIDFunc: func(gistURL string) (string, error) {
			return "abc123", nil
		},
		GetGistByURLFunc: func(ctx context.Context, gistURL string) (*github.Gist, error) {
			content := `---
name: Import Test
author: otheruser
category: testing
tags: [import, test]
version: "2.0.0"
---
Updated content`
			return &github.Gist{
				ID:     github.String("abc123"),
				Public: github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"test.yaml": {
						Content: github.String(content),
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
	}

	// Create index with existing entry
	index := &models.Index{
		Username: "testuser",
		ImportedEntries: []models.IndexEntry{
			{
				GistID:  "abc123",
				Name:    "Import Test",
				Version: "1.0.0",
			},
		},
	}

	// Test user cancelling update
	_, err := manager.ImportPrompt(ctx, "https://gist.github.com/otheruser/abc123", index)
	if err == nil {
		t.Fatal("Expected error when user cancels")
	}
	if !contains(err.Error(), "cancelled") {
		t.Errorf("Expected error about cancellation, got: %v", err)
	}

	// Verify index was not updated
	if index.ImportedEntries[0].Version != "1.0.0" {
		t.Error("Expected index to remain unchanged when update is cancelled")
	}
}

func TestImportManager_extractGistID(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedID    string
		expectError   bool
		errorContains string
	}{
		{
			name:       "standard GitHub gist URL",
			url:        "https://gist.github.com/testuser/abc123",
			expectedID: "abc123",
		},
		{
			name:       "gist URL with revision",
			url:        "https://gist.github.com/testuser/abc123/revision",
			expectedID: "abc123",
		},
		{
			name:       "gist URL with file anchor",
			url:        "https://gist.github.com/testuser/abc123#file-test-yaml",
			expectedID: "abc123",
		},
		{
			name:          "non-GitHub URL",
			url:           "https://example.com/gist/123",
			expectError:   true,
			errorContains: "not a GitHub gist URL",
		},
		{
			name:          "invalid URL format",
			url:           "not-a-url",
			expectError:   true,
			errorContains: "invalid URL",
		},
		{
			name:          "GitHub URL but not gist",
			url:           "https://github.com/user/repo",
			expectError:   true,
			errorContains: "not a GitHub gist URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{}
			
			id, err := manager.extractGistID(tt.url)
			
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
				if id != tt.expectedID {
					t.Errorf("Expected ID %q, got %q", tt.expectedID, id)
				}
			}
		})
	}
}

func TestImportManager_validatePromptGist(t *testing.T) {
	tests := []struct {
		name          string
		gist          *github.Gist
		expectError   bool
		errorContains string
		validatePrompt func(t *testing.T, prompt *models.Prompt)
	}{
		{
			name: "valid prompt gist",
			gist: &github.Gist{
				ID:     github.String("abc123"),
				Public: github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"test.yaml": {
						Content: github.String(`---
name: Valid Prompt
author: testuser
category: testing
tags: [test]
version: "1.0.0"
---
Prompt content`),
					},
				},
			},
			expectError: false,
			validatePrompt: func(t *testing.T, prompt *models.Prompt) {
				if prompt.Name != "Valid Prompt" {
					t.Errorf("Expected name = Valid Prompt, got %s", prompt.Name)
				}
				if prompt.Content != "Prompt content" {
					t.Errorf("Expected content = Prompt content, got %s", prompt.Content)
				}
			},
		},
		{
			name: "missing required fields",
			gist: &github.Gist{
				ID:     github.String("abc123"),
				Public: github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"test.yaml": {
						Content: github.String(`---
name: Missing Fields
author: testuser
---
Content`),
					},
				},
			},
			expectError:   true,
			errorContains: "required",
		},
		{
			name: "no YAML front matter",
			gist: &github.Gist{
				ID:     github.String("abc123"),
				Public: github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"test.txt": {
						Content: github.String("Just plain text"),
					},
				},
			},
			expectError:   true,
			errorContains: "YAML front matter",
		},
		{
			name: "empty gist",
			gist: &github.Gist{
				ID:     github.String("abc123"),
				Public: github.Bool(true),
				Files:  map[github.GistFilename]github.GistFile{},
			},
			expectError:   true,
			errorContains: "no files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{}
			
			prompt, err := manager.validatePromptGist(tt.gist)
			
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
				if tt.validatePrompt != nil {
					tt.validatePrompt(t, prompt)
				}
			}
		})
	}
}

func TestImportManager_checkExistingImport(t *testing.T) {
	tests := []struct {
		name         string
		index        *models.Index
		gistID       string
		expectFound  bool
		expectedEntry models.IndexEntry
	}{
		{
			name: "finds existing import",
			index: &models.Index{
				ImportedEntries: []models.IndexEntry{
					{
						GistID:  "abc123",
						Name:    "Test Import",
						Version: "1.0.0",
					},
					{
						GistID:  "def456",
						Name:    "Another Import",
						Version: "2.0.0",
					},
				},
			},
			gistID:      "abc123",
			expectFound: true,
			expectedEntry: models.IndexEntry{
				GistID:  "abc123",
				Name:    "Test Import",
				Version: "1.0.0",
			},
		},
		{
			name: "not found in imports",
			index: &models.Index{
				ImportedEntries: []models.IndexEntry{
					{
						GistID:  "def456",
						Name:    "Different Import",
						Version: "1.0.0",
					},
				},
			},
			gistID:      "abc123",
			expectFound: false,
		},
		{
			name: "empty imported entries",
			index: &models.Index{
				ImportedEntries: []models.IndexEntry{},
			},
			gistID:      "abc123",
			expectFound: false,
		},
		{
			name: "nil imported entries",
			index: &models.Index{
				ImportedEntries: nil,
			},
			gistID:      "abc123",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{}
			
			entry, found := manager.checkExistingImport(tt.index, tt.gistID)
			
			if found != tt.expectFound {
				t.Errorf("Expected found = %v, got %v", tt.expectFound, found)
			}
			
			if tt.expectFound {
				if entry == nil {
					t.Fatal("Expected entry but got nil")
				}
				if entry.GistID != tt.expectedEntry.GistID {
					t.Errorf("Expected GistID %s, got %s", tt.expectedEntry.GistID, entry.GistID)
				}
				if entry.Name != tt.expectedEntry.Name {
					t.Errorf("Expected Name %s, got %s", tt.expectedEntry.Name, entry.Name)
				}
				if entry.Version != tt.expectedEntry.Version {
					t.Errorf("Expected Version %s, got %s", tt.expectedEntry.Version, entry.Version)
				}
			} else {
				if entry != nil {
					t.Errorf("Expected nil entry, got %+v", entry)
				}
			}
		})
	}
}

func TestImportManager_confirmVersionUpdate(t *testing.T) {
	tests := []struct {
		name          string
		oldEntry      models.IndexEntry
		newEntry      models.IndexEntry
		confirmResult bool
		confirmError  error
		expectError   bool
		expectedMsg   string
	}{
		{
			name: "user confirms update",
			oldEntry: models.IndexEntry{
				Name:    "Test Import",
				Version: "1.0.0",
			},
			newEntry: models.IndexEntry{
				Name:    "Test Import",
				Version: "2.0.0",
			},
			confirmResult: true,
			expectError:   false,
		},
		{
			name: "user declines update",
			oldEntry: models.IndexEntry{
				Name:    "Test Import",
				Version: "1.0.0",
			},
			newEntry: models.IndexEntry{
				Name:    "Test Import",
				Version: "2.0.0",
			},
			confirmResult: false,
			expectError:   false,
		},
		{
			name: "confirmation error",
			oldEntry: models.IndexEntry{
				Name:    "Test Import",
				Version: "1.0.0",
			},
			newEntry: models.IndexEntry{
				Name:    "Test Import",
				Version: "2.0.0",
			},
			confirmError: fmt.Errorf("UI error"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUI := &MockUI{
				ConfirmFunc: func(message string) (bool, error) {
					// Verify message contains versions
					if !contains(message, tt.oldEntry.Version) || !contains(message, tt.newEntry.Version) {
						t.Errorf("Expected message to contain versions, got: %s", message)
					}
					return tt.confirmResult, tt.confirmError
				},
			}

			manager := &Manager{
				ui: mockUI,
			}

			result, err := manager.confirmVersionUpdate(&tt.oldEntry, &tt.newEntry)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != tt.confirmResult {
					t.Errorf("Expected result %v, got %v", tt.confirmResult, result)
				}
			}

			// Verify confirm was called
			if len(mockUI.ConfirmCalls) != 1 {
				t.Errorf("Expected Confirm to be called once, got %d calls", len(mockUI.ConfirmCalls))
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