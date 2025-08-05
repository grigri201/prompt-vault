package cmd

import (
	"testing"
)

func TestGet_LooksLikeURL(t *testing.T) {
	g := &get{}
	
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid https URL",
			input:    "https://gist.github.com/user/abc123",
			expected: true,
		},
		{
			name:     "valid http URL",
			input:    "http://gist.github.com/user/abc123",
			expected: true,
		},
		{
			name:     "too short",
			input:    "http",
			expected: false,
		},
		{
			name:     "not a URL",
			input:    "golang",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.looksLikeURL(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikeURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGet_IsGistURL(t *testing.T) {
	g := &get{}
	
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid gist URL with username",
			input:    "https://gist.github.com/user/1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid gist URL without username",
			input:    "https://gist.github.com/1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid gist URL with 20 char ID",
			input:    "https://gist.github.com/user/1234567890abcdef1234",
			expected: true,
		},
		{
			name:     "not a URL",
			input:    "golang",
			expected: false,
		},
		{
			name:     "URL but not gist.github.com",
			input:    "https://github.com/user/repo",
			expected: false,
		},
		{
			name:     "gist URL but no valid ID",
			input:    "https://gist.github.com/user/invalid",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.isGistURL(tt.input)
			if result != tt.expected {
				t.Errorf("isGistURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "substring at beginning",
			s:        "hello world",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring in middle",
			s:        "hello world",
			substr:   "lo wo",
			expected: true,
		},
		{
			name:     "exact match",
			s:        "hello",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring not found",
			s:        "hello world",
			substr:   "xyz",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "hello",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid hex lowercase",
			input:    "1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid hex uppercase",
			input:    "1234567890ABCDEF",
			expected: true,
		},
		{
			name:     "valid hex mixed case",
			input:    "1234567890AbCdEf",
			expected: true,
		},
		{
			name:     "contains non-hex character",
			input:    "123456789g",
			expected: false,
		},
		{
			name:     "contains space",
			input:    "12345 67890",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "contains special characters",
			input:    "123-456",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHexString(tt.input)
			if result != tt.expected {
				t.Errorf("isHexString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		delimiter string
		expected  []string
	}{
		{
			name:      "normal split",
			s:         "a/b/c",
			delimiter: "/",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "no delimiter",
			s:         "abc",
			delimiter: "/",
			expected:  []string{"abc"},
		},
		{
			name:      "empty string",
			s:         "",
			delimiter: "/",
			expected:  []string{},
		},
		{
			name:      "delimiter at start",
			s:         "/a/b",
			delimiter: "/",
			expected:  []string{"", "a", "b"},
		},
		{
			name:      "delimiter at end",
			s:         "a/b/",
			delimiter: "/",
			expected:  []string{"a", "b", ""},
		},
		{
			name:      "multiple consecutive delimiters",
			s:         "a//b",
			delimiter: "/",
			expected:  []string{"a", "", "b"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitString(tt.s, tt.delimiter)
			if len(result) != len(tt.expected) {
				t.Errorf("splitString(%q, %q) length = %d, want %d", tt.s, tt.delimiter, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("splitString(%q, %q)[%d] = %q, want %q", tt.s, tt.delimiter, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestContainsGistID(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid 32-char hex ID",
			url:      "https://gist.github.com/user/1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid 20-char hex ID",
			url:      "https://gist.github.com/user/1234567890abcdef1234",
			expected: true,
		},
		{
			name:     "no valid ID",
			url:      "https://gist.github.com/user/invalid",
			expected: false,
		},
		{
			name:     "short hex string",
			url:      "https://gist.github.com/user/abc",
			expected: false,
		},
		{
			name:     "long but non-hex string",
			url:      "https://gist.github.com/user/this-is-not-hex-but-long-enough",
			expected: false,
		},
		{
			name:     "empty URL",
			url:      "",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsGistID(tt.url)
			if result != tt.expected {
				t.Errorf("containsGistID(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}