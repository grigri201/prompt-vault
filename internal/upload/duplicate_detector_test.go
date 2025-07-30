package upload

import (
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestDuplicateDetector_FindDuplicate(t *testing.T) {
	detector := NewDuplicateDetector()
	
	// Create test index
	testIndex := &models.Index{
		Username: "testuser",
		Entries: []models.IndexEntry{
			{
				GistID:   "gist123",
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
				Version:  "1.0",
				ID:       "john-test-prompt",
			},
			{
				GistID:   "gist456",
				Name:     "Another Prompt",
				Author:   "jane",
				Category: "demo",
				Tags:     []string{"demo"},
				Version:  "1.0",
				ID:       "jane-another-prompt",
			},
			{
				GistID:   "gist789",
				Name:     "No ID Prompt",
				Author:   "bob",
				Category: "legacy",
				Tags:     []string{"old"},
				Version:  "1.0",
				// No ID field - testing backward compatibility
			},
		},
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		prompt        *models.Prompt
		index         *models.Index
		wantMatch     bool
		wantMatchType MatchType
		wantGistID    string
	}{
		{
			name: "match by custom ID",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:   "Test Prompt Updated",
					Author: "john",
					ID:     "john-test-prompt",
				},
			},
			index:         testIndex,
			wantMatch:     true,
			wantMatchType: MatchByID,
			wantGistID:    "gist123",
		},
		{
			name: "match by name and author",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:   "No ID Prompt",
					Author: "bob",
					// No ID provided
				},
			},
			index:         testIndex,
			wantMatch:     true,
			wantMatchType: MatchByNameAuthor,
			wantGistID:    "gist789",
		},
		{
			name: "match by gist ID",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:   "Different Name",
					Author: "different",
				},
				GistID: "gist456",
			},
			index:         testIndex,
			wantMatch:     true,
			wantMatchType: MatchByGistID,
			wantGistID:    "gist456",
		},
		{
			name: "no match - new prompt",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:   "New Prompt",
					Author: "newuser",
					ID:     "newuser-new-prompt",
				},
			},
			index:     testIndex,
			wantMatch: false,
		},
		{
			name: "no match - empty index",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:   "Any Prompt",
					Author: "anyone",
				},
			},
			index:     &models.Index{},
			wantMatch: false,
		},
		{
			name: "no match - nil index",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:   "Any Prompt",
					Author: "anyone",
				},
			},
			index:     nil,
			wantMatch: false,
		},
		{
			name: "priority test - ID over name/author",
			prompt: &models.Prompt{
				PromptMeta: models.PromptMeta{
					Name:   "Another Prompt", // Matches gist456 by name/author
					Author: "jane",
					ID:     "john-test-prompt", // But ID matches gist123
				},
			},
			index:         testIndex,
			wantMatch:     true,
			wantMatchType: MatchByID,
			wantGistID:    "gist123", // Should match by ID, not name/author
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := detector.FindDuplicate(tt.prompt, tt.index)
			if err != nil {
				t.Fatalf("FindDuplicate() returned error: %v", err)
			}

			if tt.wantMatch && match == nil {
				t.Error("FindDuplicate() expected match but got nil")
				return
			}

			if !tt.wantMatch && match != nil {
				t.Errorf("FindDuplicate() expected no match but got: %+v", match)
				return
			}

			if tt.wantMatch {
				if match.MatchType != tt.wantMatchType {
					t.Errorf("FindDuplicate() match type = %v, want %v", match.MatchType, tt.wantMatchType)
				}
				if match.Entry.GistID != tt.wantGistID {
					t.Errorf("FindDuplicate() matched GistID = %v, want %v", match.Entry.GistID, tt.wantGistID)
				}
			}
		})
	}
}

func TestMatchType_String(t *testing.T) {
	tests := []struct {
		matchType MatchType
		want      string
	}{
		{MatchByID, "MatchByID"},
		{MatchByNameAuthor, "MatchByNameAuthor"},
		{MatchByGistID, "MatchByGistID"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			// Simple test to ensure constants are defined correctly
			if int(tt.matchType) < 0 || int(tt.matchType) > 2 {
				t.Errorf("Invalid MatchType value: %d", tt.matchType)
			}
		})
	}
}