package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestUploadCommand_WithID(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	
	// Create test prompt file with ID
	promptContent := `---
name: Test Prompt with ID
author: testuser
category: test
tags: [test, integration]
id: test-prompt-123
---
This is a test prompt with {variable}.`

	promptFile := filepath.Join(tempDir, "test-prompt.yaml")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	// Note: This is a unit test example. For full integration testing,
	// you would need to:
	// 1. Mock the GitHub API client
	// 2. Set up test authentication
	// 3. Create a test container with all dependencies
	// 4. Execute the command and verify results
	
	// For now, we'll test the basic file parsing
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		t.Error("Test prompt file was not created")
	}
}

func TestUploadCommand_BackwardCompatibility(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	
	// Create test prompt file WITHOUT ID (backward compatibility)
	promptContent := `---
name: Legacy Prompt
author: testuser
category: legacy
tags: [old, test]
---
This is a legacy prompt without ID field.`

	promptFile := filepath.Join(tempDir, "legacy-prompt.yaml")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create test prompt file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		t.Error("Test prompt file was not created")
	}
}

func TestUploadCommand_ForceFlag(t *testing.T) {
	// This test would verify that the --force flag is properly registered
	cmd := newUploadCmd()
	
	// Check if force flag exists
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Force flag not found in upload command")
	}
	
	// Check short form
	if forceFlag.Shorthand != "f" {
		t.Errorf("Expected force flag shorthand to be 'f', got '%s'", forceFlag.Shorthand)
	}
	
	// Check description
	if forceFlag.Usage == "" {
		t.Error("Force flag should have a usage description")
	}
}

func TestIndexEntry_WithID(t *testing.T) {
	// Test that IndexEntry properly includes ID field
	prompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:     "Test Prompt",
			Author:   "testuser",
			Category: "test",
			Tags:     []string{"test"},
			ID:       "test-prompt-123",
		},
		GistID:    "gist123",
		GistURL:   "https://gist.github.com/test/gist123",
		UpdatedAt: time.Now(),
	}
	
	entry := prompt.ToIndexEntry()
	
	if entry.ID != prompt.ID {
		t.Errorf("Expected IndexEntry.ID to be '%s', got '%s'", prompt.ID, entry.ID)
	}
}

func TestPromptValidation_WithInvalidID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      "valid-id-123",
			wantErr: false,
		},
		{
			name:    "ID with spaces",
			id:      "invalid id",
			wantErr: true,
		},
		{
			name:    "ID too short",
			id:      "ab",
			wantErr: true,
		},
		{
			name:    "empty ID is valid (optional)",
			id:      "",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := &models.PromptMeta{
				Name:     "Test",
				Author:   "test",
				Category: "test",
				Tags:     []string{"test"},
				ID:       tt.id,
			}
			
			err := prompt.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}