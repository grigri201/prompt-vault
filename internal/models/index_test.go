package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIndex_WithImportedEntries(t *testing.T) {
	// Test that Index can hold imported entries
	index := Index{
		Username: "testuser",
		Entries: []IndexEntry{
			{
				GistID:   "abc123",
				GistURL:  "https://gist.github.com/testuser/abc123",
				Name:     "My Prompt",
				Author:   "testuser",
				Category: "test",
				Tags:     []string{"test"},
				Version:  "1.0.0",
			},
		},
		ImportedEntries: []IndexEntry{
			{
				GistID:   "def456",
				GistURL:  "https://gist.github.com/otheruser/def456",
				Name:     "Imported Prompt",
				Author:   "otheruser",
				Category: "utility",
				Tags:     []string{"util", "import"},
				Version:  "2.0.0",
			},
		},
		UpdatedAt: time.Now(),
	}

	// Verify the structure
	if len(index.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(index.Entries))
	}
	if len(index.ImportedEntries) != 1 {
		t.Errorf("Expected 1 imported entry, got %d", len(index.ImportedEntries))
	}
	if index.ImportedEntries[0].Author != "otheruser" {
		t.Errorf("Expected imported entry author to be 'otheruser', got '%s'", index.ImportedEntries[0].Author)
	}
}

func TestIndex_JSONMarshaling_WithImportedEntries(t *testing.T) {
	originalIndex := Index{
		Username: "testuser",
		Entries: []IndexEntry{
			{
				GistID:      "abc123",
				GistURL:     "https://gist.github.com/testuser/abc123",
				Name:        "Regular Prompt",
				Author:      "testuser",
				Category:    "test",
				Tags:        []string{"test"},
				Version:     "1.0.0",
				Description: "A regular prompt",
				UpdatedAt:   time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC),
			},
		},
		ImportedEntries: []IndexEntry{
			{
				GistID:      "imported123",
				GistURL:     "https://gist.github.com/alice/imported123",
				Name:        "Alice's Prompt",
				Author:      "alice",
				Category:    "productivity",
				Tags:        []string{"work", "efficiency"},
				Version:     "3.0.0",
				Description: "Imported from Alice",
				UpdatedAt:   time.Date(2024, 1, 21, 15, 30, 0, 0, time.UTC),
			},
			{
				GistID:      "imported456",
				GistURL:     "https://gist.github.com/bob/imported456",
				Name:        "Bob's Utility",
				Author:      "bob",
				Category:    "utility",
				Tags:        []string{"tool"},
				Version:     "1.5.0",
				Description: "Imported from Bob",
				UpdatedAt:   time.Date(2024, 1, 22, 9, 45, 0, 0, time.UTC),
			},
		},
		UpdatedAt: time.Date(2024, 1, 22, 16, 0, 0, 0, time.UTC),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(originalIndex, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal Index: %v", err)
	}

	// Verify JSON contains imported_entries field
	jsonStr := string(data)
	if !indexContains(jsonStr, `"imported_entries"`) {
		t.Error("JSON should contain 'imported_entries' field")
	}

	// Unmarshal back
	var unmarshaledIndex Index
	if err := json.Unmarshal(data, &unmarshaledIndex); err != nil {
		t.Fatalf("Failed to unmarshal Index: %v", err)
	}

	// Verify all fields are preserved
	if unmarshaledIndex.Username != originalIndex.Username {
		t.Errorf("Username mismatch: got %s, want %s", unmarshaledIndex.Username, originalIndex.Username)
	}
	if len(unmarshaledIndex.Entries) != len(originalIndex.Entries) {
		t.Errorf("Entries count mismatch: got %d, want %d", len(unmarshaledIndex.Entries), len(originalIndex.Entries))
	}
	if len(unmarshaledIndex.ImportedEntries) != len(originalIndex.ImportedEntries) {
		t.Errorf("ImportedEntries count mismatch: got %d, want %d", len(unmarshaledIndex.ImportedEntries), len(originalIndex.ImportedEntries))
	}

	// Verify imported entries details
	if len(unmarshaledIndex.ImportedEntries) > 0 {
		firstImported := unmarshaledIndex.ImportedEntries[0]
		if firstImported.Author != "alice" {
			t.Errorf("First imported entry author mismatch: got %s, want alice", firstImported.Author)
		}
		if firstImported.GistID != "imported123" {
			t.Errorf("First imported entry GistID mismatch: got %s, want imported123", firstImported.GistID)
		}
	}
}

func TestIndex_AddImportedEntry(t *testing.T) {
	index := &Index{
		Username:        "testuser",
		Entries:         []IndexEntry{},
		ImportedEntries: []IndexEntry{},
		UpdatedAt:       time.Now(),
	}

	// Add first imported entry
	newEntry := IndexEntry{
		GistID:    "import1",
		GistURL:   "https://gist.github.com/user1/import1",
		Name:      "First Import",
		Author:    "user1",
		Category:  "test",
		Tags:      []string{"import"},
		Version:   "1.0.0",
		UpdatedAt: time.Now(),
	}

	index.AddImportedEntry(newEntry)

	if len(index.ImportedEntries) != 1 {
		t.Errorf("Expected 1 imported entry after add, got %d", len(index.ImportedEntries))
	}
	if index.ImportedEntries[0].GistID != "import1" {
		t.Errorf("Expected imported entry GistID to be 'import1', got '%s'", index.ImportedEntries[0].GistID)
	}

	// Add second imported entry
	secondEntry := IndexEntry{
		GistID:    "import2",
		GistURL:   "https://gist.github.com/user2/import2",
		Name:      "Second Import",
		Author:    "user2",
		Category:  "utility",
		Tags:      []string{"tool"},
		Version:   "2.0.0",
		UpdatedAt: time.Now(),
	}

	index.AddImportedEntry(secondEntry)

	if len(index.ImportedEntries) != 2 {
		t.Errorf("Expected 2 imported entries after second add, got %d", len(index.ImportedEntries))
	}
}

func TestIndex_UpdateImportedEntry(t *testing.T) {
	index := &Index{
		Username: "testuser",
		Entries:  []IndexEntry{},
		ImportedEntries: []IndexEntry{
			{
				GistID:    "update-test",
				GistURL:   "https://gist.github.com/original/update-test",
				Name:      "Original Name",
				Author:    "original",
				Category:  "test",
				Tags:      []string{"old"},
				Version:   "1.0.0",
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		UpdatedAt: time.Now(),
	}

	// Update the entry
	updatedEntry := IndexEntry{
		GistID:      "update-test",
		GistURL:     "https://gist.github.com/original/update-test",
		Name:        "Updated Name",
		Author:      "original",
		Category:    "test",
		Tags:        []string{"new", "updated"},
		Version:     "2.0.0",
		Description: "Now with description",
		UpdatedAt:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
	}

	updated := index.UpdateImportedEntry(updatedEntry)

	if !updated {
		t.Error("Expected UpdateImportedEntry to return true for existing entry")
	}

	if len(index.ImportedEntries) != 1 {
		t.Errorf("Expected 1 imported entry after update, got %d", len(index.ImportedEntries))
	}

	// Verify the entry was updated
	entry := index.ImportedEntries[0]
	if entry.Name != "Updated Name" {
		t.Errorf("Expected name to be 'Updated Name', got '%s'", entry.Name)
	}
	if entry.Version != "2.0.0" {
		t.Errorf("Expected version to be '2.0.0', got '%s'", entry.Version)
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "new" {
		t.Errorf("Expected tags to be updated, got %v", entry.Tags)
	}
	if entry.Description != "Now with description" {
		t.Errorf("Expected description to be 'Now with description', got '%s'", entry.Description)
	}

	// Try to update non-existent entry
	nonExistent := IndexEntry{
		GistID:  "not-found",
		GistURL: "https://gist.github.com/none/not-found",
		Name:    "Not Found",
	}

	updated = index.UpdateImportedEntry(nonExistent)
	if updated {
		t.Error("Expected UpdateImportedEntry to return false for non-existent entry")
	}
}

func TestIndex_FindImportedEntry(t *testing.T) {
	index := &Index{
		Username: "testuser",
		ImportedEntries: []IndexEntry{
			{
				GistID:   "find1",
				GistURL:  "https://gist.github.com/user/find1",
				Name:     "First",
				Author:   "user1",
				Category: "test",
				Tags:     []string{"test"},
			},
			{
				GistID:   "find2",
				GistURL:  "https://gist.github.com/user/find2",
				Name:     "Second",
				Author:   "user2",
				Category: "test",
				Tags:     []string{"test"},
			},
		},
	}

	// Find existing entry
	entry, found := index.FindImportedEntry("find2")
	if !found {
		t.Error("Expected to find entry with GistID 'find2'")
	}
	if entry.Name != "Second" {
		t.Errorf("Expected found entry name to be 'Second', got '%s'", entry.Name)
	}

	// Try to find non-existent entry
	_, found = index.FindImportedEntry("not-exist")
	if found {
		t.Error("Expected not to find entry with GistID 'not-exist'")
	}
}


// Helper function to check string contains substring
func indexContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
