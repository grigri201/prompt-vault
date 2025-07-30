package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIndexEntry_BackwardCompatibility(t *testing.T) {
	// Test that old index entries without ID field can still be loaded
	oldIndexJSON := `{
		"gist_id": "abc123",
		"gist_url": "https://gist.github.com/user/abc123",
		"name": "Old Prompt",
		"author": "olduser",
		"category": "legacy",
		"tags": ["old", "test"],
		"version": "1.0",
		"description": "Legacy prompt",
		"updated_at": "2024-01-01T00:00:00Z"
	}`
	
	var entry IndexEntry
	err := json.Unmarshal([]byte(oldIndexJSON), &entry)
	if err != nil {
		t.Fatalf("Failed to unmarshal old index entry: %v", err)
	}
	
	// Verify all fields except ID
	if entry.GistID != "abc123" {
		t.Errorf("Expected GistID 'abc123', got '%s'", entry.GistID)
	}
	if entry.Name != "Old Prompt" {
		t.Errorf("Expected Name 'Old Prompt', got '%s'", entry.Name)
	}
	if entry.ID != "" {
		t.Errorf("Expected ID to be empty for old entries, got '%s'", entry.ID)
	}
}

func TestIndex_BackwardCompatibility(t *testing.T) {
	// Test that old index file without ID fields can still be loaded
	oldIndexJSON := `{
		"username": "testuser",
		"entries": [
			{
				"gist_id": "gist1",
				"gist_url": "https://gist.github.com/user/gist1",
				"name": "Prompt 1",
				"author": "user1",
				"category": "test",
				"tags": ["test"],
				"version": "1.0",
				"updated_at": "2024-01-01T00:00:00Z"
			},
			{
				"gist_id": "gist2",
				"gist_url": "https://gist.github.com/user/gist2",
				"name": "Prompt 2",
				"author": "user2",
				"category": "demo",
				"tags": ["demo"],
				"version": "2.0",
				"updated_at": "2024-01-02T00:00:00Z"
			}
		],
		"updated_at": "2024-01-02T00:00:00Z"
	}`
	
	var index Index
	err := json.Unmarshal([]byte(oldIndexJSON), &index)
	if err != nil {
		t.Fatalf("Failed to unmarshal old index: %v", err)
	}
	
	if len(index.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(index.Entries))
	}
	
	// Verify entries loaded correctly without ID
	for i, entry := range index.Entries {
		if entry.ID != "" {
			t.Errorf("Entry %d: Expected ID to be empty, got '%s'", i, entry.ID)
		}
		if entry.GistID == "" {
			t.Errorf("Entry %d: GistID should not be empty", i)
		}
	}
}

func TestPromptMeta_BackwardCompatibility(t *testing.T) {
	// Test that prompts without ID field still validate
	meta := PromptMeta{
		Name:     "Legacy Prompt",
		Author:   "legacyuser",
		Category: "legacy",
		Tags:     []string{"old"},
		// No ID field
	}
	
	err := meta.Validate()
	if err != nil {
		t.Errorf("Legacy prompt without ID should still validate: %v", err)
	}
	
	// ID validation should pass for empty ID
	err = meta.ValidateID()
	if err != nil {
		t.Errorf("ValidateID should pass for empty ID: %v", err)
	}
}

func TestIndexEntry_MixedEntries(t *testing.T) {
	// Test index with mix of old (no ID) and new (with ID) entries
	mixedIndexJSON := `{
		"username": "testuser",
		"entries": [
			{
				"gist_id": "gist1",
				"gist_url": "https://gist.github.com/user/gist1",
				"name": "Old Prompt",
				"author": "user1",
				"category": "test",
				"tags": ["test"],
				"version": "1.0",
				"updated_at": "2024-01-01T00:00:00Z"
			},
			{
				"gist_id": "gist2",
				"gist_url": "https://gist.github.com/user/gist2",
				"name": "New Prompt",
				"author": "user2",
				"category": "demo",
				"tags": ["demo"],
				"version": "2.0",
				"id": "user2-new-prompt",
				"updated_at": "2024-01-02T00:00:00Z"
			}
		],
		"updated_at": "2024-01-02T00:00:00Z"
	}`
	
	var index Index
	err := json.Unmarshal([]byte(mixedIndexJSON), &index)
	if err != nil {
		t.Fatalf("Failed to unmarshal mixed index: %v", err)
	}
	
	if len(index.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(index.Entries))
	}
	
	// First entry should have no ID
	if index.Entries[0].ID != "" {
		t.Errorf("First entry should have empty ID, got '%s'", index.Entries[0].ID)
	}
	
	// Second entry should have ID
	if index.Entries[1].ID != "user2-new-prompt" {
		t.Errorf("Second entry should have ID 'user2-new-prompt', got '%s'", index.Entries[1].ID)
	}
}

func TestPrompt_ToIndexEntry_PreservesAllFields(t *testing.T) {
	// Ensure ToIndexEntry doesn't break for prompts without ID
	prompt := Prompt{
		PromptMeta: PromptMeta{
			Name:        "Test Prompt",
			Author:      "testuser",
			Category:    "test",
			Tags:        []string{"tag1", "tag2"},
			Version:     "1.0",
			Description: "Test description",
			// No ID
		},
		GistID:    "gist123",
		GistURL:   "https://gist.github.com/test/gist123",
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Content:   "Test content",
	}
	
	entry := prompt.ToIndexEntry()
	
	// Verify all fields are preserved
	if entry.GistID != prompt.GistID {
		t.Errorf("GistID not preserved")
	}
	if entry.Name != prompt.Name {
		t.Errorf("Name not preserved")
	}
	if entry.ID != "" {
		t.Errorf("Empty ID should remain empty")
	}
}