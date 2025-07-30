package upload

import (
	"github.com/grigri201/prompt-vault/internal/models"
)

// MatchType represents the type of duplicate match found
type MatchType int

const (
	// MatchByID indicates a match by custom ID
	MatchByID MatchType = iota
	// MatchByNameAuthor indicates a match by name and author combination
	MatchByNameAuthor
	// MatchByGistID indicates a match by Gist ID
	MatchByGistID
)

// DuplicateMatch contains information about a found duplicate
type DuplicateMatch struct {
	Entry     models.IndexEntry
	MatchType MatchType
}

// DuplicateDetector detects duplicate prompts in the index
type DuplicateDetector interface {
	// FindDuplicate checks for existing prompts matching the given criteria
	FindDuplicate(prompt *models.Prompt, index *models.Index) (*DuplicateMatch, error)
}

// duplicateDetectorImpl is the default implementation of DuplicateDetector
type duplicateDetectorImpl struct{}

// NewDuplicateDetector creates a new duplicate detector
func NewDuplicateDetector() DuplicateDetector {
	return &duplicateDetectorImpl{}
}

// FindDuplicate implements the DuplicateDetector interface
func (d *duplicateDetectorImpl) FindDuplicate(prompt *models.Prompt, index *models.Index) (*DuplicateMatch, error) {
	if index == nil || len(index.Entries) == 0 {
		return nil, nil
	}

	// Priority 1: Check by custom ID if provided
	if prompt.ID != "" {
		for _, entry := range index.Entries {
			// Extract ID from the entry if it exists
			// Since we store prompts in the index, we need to check if the name matches the ID
			// or if there's an ID field stored somehow
			if entryHasID(entry, prompt.ID) {
				return &DuplicateMatch{
					Entry:     entry,
					MatchType: MatchByID,
				}, nil
			}
		}
	}

	// Priority 2: Check by Name + Author combination
	for _, entry := range index.Entries {
		if entry.Name == prompt.Name && entry.Author == prompt.Author {
			return &DuplicateMatch{
				Entry:     entry,
				MatchType: MatchByNameAuthor,
			}, nil
		}
	}

	// Priority 3: Check by GistID (for updates)
	if prompt.GistID != "" {
		for _, entry := range index.Entries {
			if entry.GistID == prompt.GistID {
				return &DuplicateMatch{
					Entry:     entry,
					MatchType: MatchByGistID,
				}, nil
			}
		}
	}

	// No duplicate found
	return nil, nil
}

// entryHasID checks if an index entry has the given ID
func entryHasID(entry models.IndexEntry, id string) bool {
	return entry.ID != "" && entry.ID == id
}