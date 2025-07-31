package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIndex_AddImportedEntry(t *testing.T) {
	t.Run("Add to empty index", func(t *testing.T) {
		index := &Index{}
		entry := IndexEntry{
			GistID: "test123",
			Name:   "Test Entry",
		}

		index.AddImportedEntry(entry)

		assert.Len(t, index.ImportedEntries, 1)
		assert.Equal(t, "test123", index.ImportedEntries[0].GistID)
		assert.Equal(t, "Test Entry", index.ImportedEntries[0].Name)
	})

	t.Run("Add to existing entries", func(t *testing.T) {
		index := &Index{
			ImportedEntries: []IndexEntry{
				{GistID: "existing", Name: "Existing Entry"},
			},
		}
		entry := IndexEntry{
			GistID: "test123",
			Name:   "New Entry",
		}

		index.AddImportedEntry(entry)

		assert.Len(t, index.ImportedEntries, 2)
		assert.Equal(t, "existing", index.ImportedEntries[0].GistID)
		assert.Equal(t, "test123", index.ImportedEntries[1].GistID)
	})
}

func TestIndex_UpdateImportedEntry(t *testing.T) {
	t.Run("Update existing entry", func(t *testing.T) {
		index := &Index{
			ImportedEntries: []IndexEntry{
				{GistID: "test123", Name: "Original", Version: "1.0"},
			},
		}

		updated := IndexEntry{
			GistID:  "test123",
			Name:    "Updated",
			Version: "2.0",
		}

		result := index.UpdateImportedEntry(updated)

		assert.True(t, result)
		assert.Equal(t, "Updated", index.ImportedEntries[0].Name)
		assert.Equal(t, "2.0", index.ImportedEntries[0].Version)
	})

	t.Run("Update non-existent entry", func(t *testing.T) {
		index := &Index{
			ImportedEntries: []IndexEntry{
				{GistID: "other", Name: "Other Entry"},
			},
		}

		updated := IndexEntry{
			GistID: "nonexistent",
			Name:   "Updated",
		}

		result := index.UpdateImportedEntry(updated)

		assert.False(t, result)
		assert.Equal(t, "Other Entry", index.ImportedEntries[0].Name)
	})

	t.Run("Update in multiple entries", func(t *testing.T) {
		index := &Index{
			ImportedEntries: []IndexEntry{
				{GistID: "first", Name: "First"},
				{GistID: "target", Name: "Target"},
				{GistID: "third", Name: "Third"},
			},
		}

		updated := IndexEntry{
			GistID: "target",
			Name:   "Updated Target",
		}

		result := index.UpdateImportedEntry(updated)

		assert.True(t, result)
		assert.Equal(t, "First", index.ImportedEntries[0].Name)
		assert.Equal(t, "Updated Target", index.ImportedEntries[1].Name)
		assert.Equal(t, "Third", index.ImportedEntries[2].Name)
	})
}

func TestIndex_FindImportedEntry(t *testing.T) {
	index := &Index{
		ImportedEntries: []IndexEntry{
			{GistID: "test123", Name: "Test Entry", Author: "testuser"},
			{GistID: "test456", Name: "Another Entry", Author: "otheruser"},
		},
	}

	t.Run("Find existing entry", func(t *testing.T) {
		entry, found := index.FindImportedEntry("test123")
		assert.True(t, found)
		assert.Equal(t, "Test Entry", entry.Name)
		assert.Equal(t, "testuser", entry.Author)
	})

	t.Run("Find non-existent entry", func(t *testing.T) {
		entry, found := index.FindImportedEntry("nonexistent")
		assert.False(t, found)
		assert.Equal(t, IndexEntry{}, entry)
	})

	t.Run("Find in empty index", func(t *testing.T) {
		emptyIndex := &Index{}
		entry, found := emptyIndex.FindImportedEntry("test123")
		assert.False(t, found)
		assert.Equal(t, IndexEntry{}, entry)
	})
}

func TestIndex_CompleteWorkflow(t *testing.T) {
	t.Run("Add, update, and find workflow", func(t *testing.T) {
		index := &Index{}
		now := time.Now()

		// Add initial entry
		entry := IndexEntry{
			GistID:    "workflow-test",
			Name:      "Workflow Entry",
			Author:    "testuser",
			Tags:      []string{"workflow", "test"},
			UpdatedAt: now,
		}

		index.AddImportedEntry(entry)
		assert.Len(t, index.ImportedEntries, 1)

		// Find the entry
		found, exists := index.FindImportedEntry("workflow-test")
		assert.True(t, exists)
		assert.Equal(t, "Workflow Entry", found.Name)

		// Update the entry
		updated := IndexEntry{
			GistID:      "workflow-test",
			Name:        "Updated Workflow Entry",
			Author:      "testuser",
			Tags:        []string{"workflow", "test", "updated"},
			Version:     "2.0",
			Description: "Updated description",
			UpdatedAt:   now.Add(time.Hour),
		}

		result := index.UpdateImportedEntry(updated)
		assert.True(t, result)

		// Verify update
		found, exists = index.FindImportedEntry("workflow-test")
		assert.True(t, exists)
		assert.Equal(t, "Updated Workflow Entry", found.Name)
		assert.Equal(t, "2.0", found.Version)
		assert.Equal(t, "Updated description", found.Description)
		assert.Contains(t, found.Tags, "updated")
	})
}

func TestIndexEntry_Serialization(t *testing.T) {
	t.Run("IndexEntry with all fields", func(t *testing.T) {
		now := time.Now()
		entry := IndexEntry{
			GistID:      "gist123",
			GistURL:     "https://gist.github.com/user/gist123",
			Name:        "Complete Entry",
			Author:      "testuser",
			Tags:        []string{"complete", "test"},
			Version:     "1.0",
			Description: "Complete test entry",
			Parent:      "parent-id",
			ID:          "entry-id",
			UpdatedAt:   now,
		}

		// Test that all fields are properly set
		assert.Equal(t, "gist123", entry.GistID)
		assert.Equal(t, "https://gist.github.com/user/gist123", entry.GistURL)
		assert.Equal(t, "Complete Entry", entry.Name)
		assert.Equal(t, "testuser", entry.Author)
		assert.Equal(t, []string{"complete", "test"}, entry.Tags)
		assert.Equal(t, "1.0", entry.Version)
		assert.Equal(t, "Complete test entry", entry.Description)
		assert.Equal(t, "parent-id", entry.Parent)
		assert.Equal(t, "entry-id", entry.ID)
		assert.Equal(t, now, entry.UpdatedAt)
	})
}
