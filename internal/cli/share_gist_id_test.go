package cli

import (
	"testing"
)

func TestIsGistID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid gist IDs
		{
			name:     "valid 32-char hex string",
			input:    "1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid gist ID with all numbers",
			input:    "12345678901234567890123456789012",
			expected: true,
		},
		{
			name:     "valid gist ID with all lowercase letters",
			input:    "abcdefabcdefabcdefabcdefabcdefab",
			expected: true,
		},
		{
			name:     "valid mixed hex string",
			input:    "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			expected: true,
		},
		// Invalid gist IDs
		{
			name:     "too short",
			input:    "1234567890abcdef",
			expected: false,
		},
		{
			name:     "too long",
			input:    "1234567890abcdef1234567890abcdef1",
			expected: false,
		},
		{
			name:     "contains uppercase letters",
			input:    "1234567890ABCDEF1234567890abcdef",
			expected: false,
		},
		{
			name:     "contains non-hex characters",
			input:    "1234567890abcdef1234567890abcdeg",
			expected: false,
		},
		{
			name:     "contains special characters",
			input:    "1234567890abcdef-234567890abcdef",
			expected: false,
		},
		{
			name:     "contains spaces",
			input:    "1234567890abcdef 234567890abcdef",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "regular word",
			input:    "testing",
			expected: false,
		},
		{
			name:     "URL",
			input:    "https://gist.github.com/user/abc123",
			expected: false,
		},
		{
			name:     "partial gist ID",
			input:    "abc123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGistID(tt.input)
			if result != tt.expected {
				t.Errorf("isGistID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
