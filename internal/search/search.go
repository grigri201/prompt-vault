package search

import (
	"strings"

	"github.com/grigri201/prompt-vault/internal/models"
)

// Searcher provides search functionality for prompt entries
type Searcher struct{}

// NewSearcher creates a new searcher instance
func NewSearcher() *Searcher {
	return &Searcher{}
}

// SearchEntries searches through entries based on keyword and returns matching indices
func (s *Searcher) SearchEntries(entries []models.IndexEntry, keyword string) []int {
	if keyword == "" {
		// Return all indices if no keyword provided
		indices := make([]int, len(entries))
		for i := range entries {
			indices[i] = i
		}
		return indices
	}

	// Convert keyword to lowercase for case-insensitive search
	keyword = strings.ToLower(keyword)

	var matches []int
	for i, entry := range entries {
		if s.MatchesKeyword(entry, keyword) {
			matches = append(matches, i)
		}
	}

	return matches
}

// MatchesKeyword checks if an entry matches the search keyword
func (s *Searcher) MatchesKeyword(entry models.IndexEntry, keyword string) bool {
	// Empty keyword doesn't match anything
	if keyword == "" {
		return false
	}

	// Convert keyword to lowercase for case-insensitive comparison
	keyword = strings.ToLower(keyword)

	// Search in name
	if strings.Contains(strings.ToLower(entry.Name), keyword) {
		return true
	}

	// Search in author
	if strings.Contains(strings.ToLower(entry.Author), keyword) {
		return true
	}

	// Search in category
	if strings.Contains(strings.ToLower(entry.Category), keyword) {
		return true
	}

	// Search in description
	if strings.Contains(strings.ToLower(entry.Description), keyword) {
		return true
	}

	// Search in tags
	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), keyword) {
			return true
		}
	}

	return false
}
