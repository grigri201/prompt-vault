package search

import (
	"testing"
	"time"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestSearcher_MatchesKeyword(t *testing.T) {
	searcher := NewSearcher()

	entry := models.IndexEntry{
		GistID:      "abc123",
		GistURL:     "https://gist.github.com/abc123",
		Name:        "Test Prompt Template",
		Author:      "john_doe",
		Category:    "Development",
		Tags:        []string{"testing", "golang", "prompt"},
		Version:     "1.0.0",
		Description: "A sample prompt for unit testing",
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name     string
		keyword  string
		expected bool
	}{
		// Name matching
		{
			name:     "matches name exact",
			keyword:  "Test Prompt Template",
			expected: true,
		},
		{
			name:     "matches name partial",
			keyword:  "prompt",
			expected: true,
		},
		{
			name:     "matches name case insensitive",
			keyword:  "TEMPLATE",
			expected: true,
		},
		// Author matching
		{
			name:     "matches author exact",
			keyword:  "john_doe",
			expected: true,
		},
		{
			name:     "matches author partial",
			keyword:  "john",
			expected: true,
		},
		// Category matching
		{
			name:     "matches category",
			keyword:  "development",
			expected: true,
		},
		{
			name:     "matches category case insensitive",
			keyword:  "DEVELOPMENT",
			expected: true,
		},
		// Description matching
		{
			name:     "matches description",
			keyword:  "unit testing",
			expected: true,
		},
		{
			name:     "matches description partial",
			keyword:  "sample",
			expected: true,
		},
		// Tag matching
		{
			name:     "matches tag exact",
			keyword:  "golang",
			expected: true,
		},
		{
			name:     "matches tag case insensitive",
			keyword:  "GOLANG",
			expected: true,
		},
		{
			name:     "matches another tag",
			keyword:  "testing",
			expected: true,
		},
		// Non-matching cases
		{
			name:     "does not match unrelated keyword",
			keyword:  "python",
			expected: false,
		},
		{
			name:     "does not match partial non-match",
			keyword:  "xyz",
			expected: false,
		},
		{
			name:     "does not match empty string",
			keyword:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searcher.MatchesKeyword(entry, tt.keyword)
			if result != tt.expected {
				t.Errorf("MatchesKeyword(%q) = %v, want %v", tt.keyword, result, tt.expected)
			}
		})
	}
}

func TestSearcher_SearchEntries(t *testing.T) {
	searcher := NewSearcher()

	entries := []models.IndexEntry{
		{
			Name:        "Go Testing Guide",
			Author:      "alice",
			Category:    "Testing",
			Tags:        []string{"go", "testing", "guide"},
			Description: "Comprehensive guide for Go testing",
		},
		{
			Name:        "Python Development",
			Author:      "bob",
			Category:    "Development",
			Tags:        []string{"python", "dev"},
			Description: "Python development best practices",
		},
		{
			Name:        "Testing Strategies",
			Author:      "charlie",
			Category:    "Testing",
			Tags:        []string{"testing", "qa"},
			Description: "General testing strategies and patterns",
		},
		{
			Name:        "Go Web Development",
			Author:      "alice",
			Category:    "Web",
			Tags:        []string{"go", "web", "api"},
			Description: "Building web applications with Go",
		},
	}

	tests := []struct {
		name            string
		keyword         string
		expectedIndices []int
	}{
		{
			name:            "empty keyword returns all entries",
			keyword:         "",
			expectedIndices: []int{0, 1, 2, 3},
		},
		{
			name:            "matches single entry by name",
			keyword:         "Python Development",
			expectedIndices: []int{1},
		},
		{
			name:            "matches multiple entries by keyword in name",
			keyword:         "testing",
			expectedIndices: []int{0, 2},
		},
		{
			name:            "matches by author",
			keyword:         "alice",
			expectedIndices: []int{0, 3},
		},
		{
			name:            "matches by category",
			keyword:         "Testing",
			expectedIndices: []int{0, 2},
		},
		{
			name:            "matches by tag",
			keyword:         "go",
			expectedIndices: []int{0, 3},
		},
		{
			name:            "matches by description keyword",
			keyword:         "practices",
			expectedIndices: []int{1},
		},
		{
			name:            "case insensitive search",
			keyword:         "GO",
			expectedIndices: []int{0, 3},
		},
		{
			name:            "partial match in multiple fields",
			keyword:         "web",
			expectedIndices: []int{3},
		},
		{
			name:            "no matches",
			keyword:         "javascript",
			expectedIndices: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searcher.SearchEntries(entries, tt.keyword)

			if len(result) != len(tt.expectedIndices) {
				t.Errorf("SearchEntries(%q) returned %d indices, want %d",
					tt.keyword, len(result), len(tt.expectedIndices))
				return
			}

			for i, idx := range result {
				if i >= len(tt.expectedIndices) || idx != tt.expectedIndices[i] {
					t.Errorf("SearchEntries(%q) index[%d] = %d, want %d",
						tt.keyword, i, idx, tt.expectedIndices[i])
				}
			}
		})
	}
}

func TestSearcher_SearchEntries_EdgeCases(t *testing.T) {
	searcher := NewSearcher()

	t.Run("empty entries list", func(t *testing.T) {
		result := searcher.SearchEntries([]models.IndexEntry{}, "test")
		if len(result) != 0 {
			t.Errorf("SearchEntries with empty list returned %d results, want 0", len(result))
		}
	})

	t.Run("nil entries list", func(t *testing.T) {
		result := searcher.SearchEntries(nil, "test")
		if len(result) != 0 {
			t.Errorf("SearchEntries with nil list returned %d results, want 0", len(result))
		}
	})

	t.Run("entries with empty fields", func(t *testing.T) {
		entries := []models.IndexEntry{
			{
				Name:        "",
				Author:      "",
				Category:    "",
				Tags:        nil,
				Description: "",
			},
			{
				Name:        "Valid Entry",
				Author:      "author",
				Category:    "category",
				Tags:        []string{"tag"},
				Description: "description",
			},
		}

		result := searcher.SearchEntries(entries, "valid")
		if len(result) != 1 || result[0] != 1 {
			t.Errorf("SearchEntries did not handle empty fields correctly")
		}
	})

	t.Run("special characters in keyword", func(t *testing.T) {
		entries := []models.IndexEntry{
			{
				Name:        "Test (with parentheses)",
				Description: "Has special chars: @#$%",
			},
		}

		// Test parentheses
		result := searcher.SearchEntries(entries, "(with")
		if len(result) != 1 {
			t.Errorf("SearchEntries did not match special characters correctly")
		}

		// Test other special chars
		result = searcher.SearchEntries(entries, "@#$")
		if len(result) != 1 {
			t.Errorf("SearchEntries did not match special characters in description")
		}
	})
}
