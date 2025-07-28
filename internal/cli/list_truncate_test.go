package cli

import (
	"testing"
)

func TestTruncateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		maxLength int
		expected  string
	}{
		{
			name:      "short URL not truncated",
			url:       "https://gist.github.com/user/abc123",
			maxLength: 50,
			expected:  "https://gist.github.com/user/abc123",
		},
		{
			name:      "long URL truncated",
			url:       "https://gist.github.com/verylongusername123456789/abcdefghijklmnopqrstuvwxyz012345678901234567890",
			maxLength: 50,
			expected:  "...abcdefghijklmnopqrstuvwxyz012345678901234567890",
		},
		{
			name:      "empty URL",
			url:       "",
			maxLength: 50,
			expected:  "",
		},
		{
			name:      "URL exactly at max length",
			url:       "https://gist.github.com/user/12345678901234567890",
			maxLength: 50,
			expected:  "https://gist.github.com/user/12345678901234567890",
		},
		{
			name:      "very short max length",
			url:       "https://gist.github.com/user/abc123",
			maxLength: 5,
			expected:  "https://gist.github.com/user/abc123", // Not truncated if maxLength < 10
		},
		{
			name:      "truncate to 20 chars",
			url:       "https://gist.github.com/user/abc123def456",
			maxLength: 20,
			expected:  "...user/abc123def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateURL(tt.url, tt.maxLength)
			if result != tt.expected {
				t.Errorf("truncateURL(%q, %d) = %q, want %q", 
					tt.url, tt.maxLength, result, tt.expected)
			}
		})
	}
}