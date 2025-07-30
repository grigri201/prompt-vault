package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPromptMeta_Validation(t *testing.T) {
	tests := []struct {
		name      string
		meta      PromptMeta
		wantError bool
	}{
		{
			name: "valid prompt meta with all required fields",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test", "example"},
			},
			wantError: false,
		},
		{
			name: "valid prompt meta with optional fields",
			meta: PromptMeta{
				Name:        "Test Prompt",
				Author:      "john",
				Category:    "test",
				Tags:        []string{"test"},
				Version:     "1.0",
				Description: "A test prompt",
			},
			wantError: false,
		},
		{
			name: "valid prompt meta with valid ID",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
				ID:       "test-prompt-123",
			},
			wantError: false,
		},
		{
			name: "invalid ID with spaces",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
				ID:       "test prompt 123",
			},
			wantError: true,
		},
		{
			name: "invalid ID with special characters",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
				ID:       "test@prompt!123",
			},
			wantError: true,
		},
		{
			name: "ID too short",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
				ID:       "ab",
			},
			wantError: true,
		},
		{
			name: "missing name",
			meta: PromptMeta{
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
			},
			wantError: true,
		},
		{
			name: "missing author",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Category: "test",
				Tags:     []string{"test"},
			},
			wantError: true,
		},
		{
			name: "missing category",
			meta: PromptMeta{
				Name:   "Test Prompt",
				Author: "john",
				Tags:   []string{"test"},
			},
			wantError: true,
		},
		{
			name: "missing tags",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
			},
			wantError: true,
		},
		{
			name: "empty tags",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestPromptMeta_ValidateID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty ID is valid",
			id:        "",
			wantError: false,
		},
		{
			name:      "valid ID with letters",
			id:        "testprompt",
			wantError: false,
		},
		{
			name:      "valid ID with numbers",
			id:        "test123",
			wantError: false,
		},
		{
			name:      "valid ID with hyphens",
			id:        "test-prompt-123",
			wantError: false,
		},
		{
			name:      "valid ID with underscores",
			id:        "test_prompt_123",
			wantError: false,
		},
		{
			name:      "valid ID with mixed characters",
			id:        "Test_Prompt-123",
			wantError: false,
		},
		{
			name:      "invalid ID with spaces",
			id:        "test prompt",
			wantError: true,
			errorMsg:  "ID can only contain letters, numbers, hyphens, and underscores",
		},
		{
			name:      "invalid ID with special characters",
			id:        "test@prompt",
			wantError: true,
			errorMsg:  "ID can only contain letters, numbers, hyphens, and underscores",
		},
		{
			name:      "ID too short",
			id:        "ab",
			wantError: true,
			errorMsg:  "ID must be at least 3 characters long",
		},
		{
			name:      "ID too long",
			id:        strings.Repeat("a", 101),
			wantError: true,
			errorMsg:  "ID must not exceed 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := PromptMeta{ID: tt.id}
			err := meta.ValidateID()
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateID() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateID() error message = %v, want containing %v", err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && s[len(s)-len(substr):] == substr || len(s) > len(substr) && s[len(s)-len(substr)-1:len(s)-1] == ": "+substr)
}

func TestPromptMeta_DefaultVersion(t *testing.T) {
	meta := PromptMeta{
		Name:     "Test",
		Author:   "john",
		Category: "test",
		Tags:     []string{"test"},
	}

	// Set default version
	beforeTime := time.Now().UnixMilli()
	meta.SetDefaultVersion()
	afterTime := time.Now().UnixMilli()

	// Parse version as int64
	version := meta.Version
	if version == "" {
		t.Error("Version should not be empty after SetDefaultVersion")
	}

	// Check if version is a timestamp within reasonable bounds
	var versionTime int64
	if _, err := fmt.Sscanf(version, "%d", &versionTime); err != nil {
		t.Errorf("Version should be a valid timestamp number, got %s", version)
	}

	if versionTime < beforeTime || versionTime > afterTime {
		t.Errorf("Version timestamp %d should be between %d and %d", versionTime, beforeTime, afterTime)
	}
}

func TestIndexEntry_JSONMarshaling(t *testing.T) {
	entry := IndexEntry{
		GistID:      "abc123",
		GistURL:     "https://gist.github.com/user/abc123",
		Name:        "Test Prompt",
		Author:      "john",
		Category:    "test",
		Tags:        []string{"test", "example"},
		Version:     "1.0",
		Description: "Test description",
		UpdatedAt:   time.Date(2024, 1, 19, 10, 30, 0, 0, time.UTC),
	}

	// Test marshaling
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal IndexEntry: %v", err)
	}

	// Test unmarshaling
	var decoded IndexEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal IndexEntry: %v", err)
	}

	// Verify fields
	if decoded.GistID != entry.GistID {
		t.Errorf("GistID mismatch: got %s, want %s", decoded.GistID, entry.GistID)
	}
	if decoded.Name != entry.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, entry.Name)
	}
	if len(decoded.Tags) != len(entry.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(decoded.Tags), len(entry.Tags))
	}
	if !decoded.UpdatedAt.Equal(entry.UpdatedAt) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", decoded.UpdatedAt, entry.UpdatedAt)
	}
}

func TestPrompt_ToIndexEntry(t *testing.T) {
	prompt := Prompt{
		PromptMeta: PromptMeta{
			Name:        "Test Prompt",
			Author:      "john",
			Category:    "test",
			Tags:        []string{"test"},
			Version:     "1.0",
			Description: "Test description",
		},
		GistID:    "abc123",
		GistURL:   "https://gist.github.com/user/abc123",
		UpdatedAt: time.Date(2024, 1, 19, 10, 30, 0, 0, time.UTC),
		Content:   "Test content with {variable}",
	}

	entry := prompt.ToIndexEntry()

	if entry.GistID != prompt.GistID {
		t.Errorf("GistID mismatch: got %s, want %s", entry.GistID, prompt.GistID)
	}
	if entry.Name != prompt.Name {
		t.Errorf("Name mismatch: got %s, want %s", entry.Name, prompt.Name)
	}
	if entry.Author != prompt.Author {
		t.Errorf("Author mismatch: got %s, want %s", entry.Author, prompt.Author)
	}
	if entry.Description != prompt.Description {
		t.Errorf("Description mismatch: got %s, want %s", entry.Description, prompt.Description)
	}
}

func TestIndexEntry_JSONFields(t *testing.T) {
	// Test that JSON tags are properly set
	jsonStr := `{
		"gist_id": "test123",
		"gist_url": "https://gist.github.com/test/test123",
		"name": "Test",
		"author": "john",
		"category": "test",
		"tags": ["tag1", "tag2"],
		"version": "1.0",
		"description": "Test description",
		"updated_at": "2024-01-19T10:30:00Z"
	}`

	var entry IndexEntry
	if err := json.Unmarshal([]byte(jsonStr), &entry); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if entry.GistID != "test123" {
		t.Errorf("Expected gist_id to be 'test123', got '%s'", entry.GistID)
	}
	if entry.Name != "Test" {
		t.Errorf("Expected name to be 'Test', got '%s'", entry.Name)
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "tag1" {
		t.Errorf("Expected tags to be [tag1, tag2], got %v", entry.Tags)
	}
}
