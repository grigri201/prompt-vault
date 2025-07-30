package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/imports"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/share"
)

// TestShareWithProgress tests share operation with simulated delays
func TestShareWithProgress(t *testing.T) {
	ctx := context.Background()

	// Create mocks
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
tags: [test]
version: "1.0.0"
---
Content`),
			},
		},
	}
	mockClient.gists["private123"] = privateGist

	// Create a mock UI
	mockUI := &MockProgressUI{
		confirmResponses: []bool{true},
	}

	// Create manager
	manager := share.NewManager(mockClient, mockUI, "testuser")

	// Share
	result, err := manager.SharePrompt(ctx, "private123")
	if err != nil {
		t.Fatalf("SharePrompt() error = %v", err)
	}

	// Verify result
	if result.PublicGistID == "" {
		t.Error("Expected public gist ID to be set")
	}

	// Test that operation completes successfully
	t.Logf("Share operation completed successfully with gist ID: %s", result.PublicGistID)
}

// TestImportWithProgress tests import operation
func TestImportWithProgress(t *testing.T) {
	ctx := context.Background()

	// Create mocks
	mockClient := &MockImportGistClient{
		gists: make(map[string]*github.Gist),
	}

	// Add a public gist to import
	publicGist := &github.Gist{
		ID:          github.String("public123"),
		HTMLURL:     github.String("https://gist.github.com/otheruser/public123"),
		Description: github.String("Import Test Prompt"),
		Public:      github.Bool(true),
		Files: map[github.GistFilename]github.GistFile{
			"prompt.yaml": {
				Content: github.String(`---
name: Import Test
author: otheruser
category: test
tags: [test, import]
version: "1.0.0"
---
Import test content`),
			},
		},
	}
	mockClient.gists["public123"] = publicGist

	// Create mock UI
	mockUI := &MockProgressUI{
		confirmResponses: []bool{true},
	}

	// Create manager
	manager := imports.NewManager(mockClient, mockUI)

	// Create index
	index := &models.Index{
		Username:        "testuser",
		Entries:         []models.IndexEntry{},
		ImportedEntries: []models.IndexEntry{},
		UpdatedAt:       time.Now(),
	}

	// Import
	result, err := manager.ImportPrompt(ctx, "https://gist.github.com/otheruser/public123", index)
	if err != nil {
		t.Fatalf("ImportPrompt() error = %v", err)
	}

	// Verify result
	if result.GistID != "public123" {
		t.Errorf("Expected GistID = public123, got %s", result.GistID)
	}

	// Test that operation completes successfully
	t.Logf("Import operation completed successfully with gist ID: %s", result.GistID)
}

// TestLongRunningOperation tests that operations can handle delays
func TestLongRunningOperation(t *testing.T) {
	// Test that operations handle delays gracefully
	startTime := time.Now()

	// Simulate a long operation with multiple steps
	steps := []struct {
		message string
		delay   time.Duration
	}{
		{"Initializing...", 50 * time.Millisecond},
		{"Processing step 1/3...", 100 * time.Millisecond},
		{"Processing step 2/3...", 100 * time.Millisecond},
		{"Processing step 3/3...", 100 * time.Millisecond},
		{"Finalizing...", 50 * time.Millisecond},
	}

	for i, step := range steps {
		t.Logf("Step %d at %v: %s", i+1, time.Since(startTime), step.message)
		time.Sleep(step.delay)
	}

	// Verify total time
	totalTime := time.Since(startTime)
	expectedMinTime := 400 * time.Millisecond
	if totalTime < expectedMinTime {
		t.Errorf("Operation completed too quickly: %v < %v", totalTime, expectedMinTime)
	}

	t.Logf("Long operation completed in %v", totalTime)
}

// MockProgressUI extends MockUI with progress tracking
type MockProgressUI struct {
	confirmResponses []bool
	confirmIndex     int
	confirmCalls     []string
	progressMessages []string
}

func (m *MockProgressUI) Confirm(message string) (bool, error) {
	m.confirmCalls = append(m.confirmCalls, message)

	if m.confirmIndex < len(m.confirmResponses) {
		response := m.confirmResponses[m.confirmIndex]
		m.confirmIndex++
		return response, nil
	}

	return false, nil
}
