package parser

import (
	"reflect"
	"testing"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestParseYAMLFrontMatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantMeta    *models.PromptMeta
		wantContent string
		wantError   bool
	}{
		{
			name: "valid front matter with all fields",
			content: `---
name: "API Documentation"
author: "john"
category: "docs"
tags: ["api", "swagger", "documentation"]
version: "1.0"
description: "Generate API documentation from OpenAPI specs"
---
Generate {format} documentation for the following API endpoint:
{endpoint}`,
			wantMeta: &models.PromptMeta{
				Name:        "API Documentation",
				Author:      "john",
				Category:    "docs",
				Tags:        []string{"api", "swagger", "documentation"},
				Version:     "1.0",
				Description: "Generate API documentation from OpenAPI specs",
			},
			wantContent: "Generate {format} documentation for the following API endpoint:\n{endpoint}",
			wantError:   false,
		},
		{
			name: "valid front matter with required fields only",
			content: `---
name: "Simple Prompt"
author: "jane"
category: "general"
tags: ["basic"]
---
This is a simple {type} prompt.`,
			wantMeta: &models.PromptMeta{
				Name:     "Simple Prompt",
				Author:   "jane",
				Category: "general",
				Tags:     []string{"basic"},
			},
			wantContent: "This is a simple {type} prompt.",
			wantError:   false,
		},
		{
			name: "missing front matter",
			content: `This is just plain content without front matter.
It has {variables} but no metadata.`,
			wantMeta:    nil,
			wantContent: "",
			wantError:   true,
		},
		{
			name: "empty front matter",
			content: `---
---
Content after empty front matter.`,
			wantMeta:    nil,
			wantContent: "",
			wantError:   true,
		},
		{
			name: "missing required field - name",
			content: `---
author: "john"
category: "test"
tags: ["test"]
---
Content`,
			wantMeta:    nil,
			wantContent: "",
			wantError:   true,
		},
		{
			name: "missing required field - author",
			content: `---
name: "Test"
category: "test"
tags: ["test"]
---
Content`,
			wantMeta:    nil,
			wantContent: "",
			wantError:   true,
		},
		{
			name: "missing required field - category",
			content: `---
name: "Test"
author: "john"
tags: ["test"]
---
Content`,
			wantMeta:    nil,
			wantContent: "",
			wantError:   true,
		},
		{
			name: "missing required field - tags",
			content: `---
name: "Test"
author: "john"
category: "test"
---
Content`,
			wantMeta:    nil,
			wantContent: "",
			wantError:   true,
		},
		{
			name: "invalid YAML syntax",
			content: `---
name: "Test
author: john"
category: test
tags: [test]
---
Content`,
			wantMeta:    nil,
			wantContent: "",
			wantError:   true,
		},
		{
			name: "no content after front matter",
			content: `---
name: "Empty Content"
author: "john"
category: "test"
tags: ["test"]
---`,
			wantMeta: &models.PromptMeta{
				Name:     "Empty Content",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
			},
			wantContent: "",
			wantError:   false,
		},
		{
			name: "front matter with extra fields",
			content: `---
name: "Test"
author: "john"
category: "test"
tags: ["test"]
extra_field: "ignored"
---
Content`,
			wantMeta: &models.PromptMeta{
				Name:     "Test",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
			},
			wantContent: "Content",
			wantError:   false,
		},
		{
			name: "multiline content",
			content: `---
name: "Multiline"
author: "john"
category: "test"
tags: ["test"]
---
Line 1 with {var1}
Line 2 with {var2}

Line 4 after empty line`,
			wantMeta: &models.PromptMeta{
				Name:     "Multiline",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
			},
			wantContent: "Line 1 with {var1}\nLine 2 with {var2}\n\nLine 4 after empty line",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, content, err := ParseYAMLFrontMatter(tt.content)

			if (err != nil) != tt.wantError {
				t.Errorf("ParseYAMLFrontMatter() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if !reflect.DeepEqual(meta, tt.wantMeta) {
					t.Errorf("ParseYAMLFrontMatter() meta = %+v, want %+v", meta, tt.wantMeta)
				}
				if content != tt.wantContent {
					t.Errorf("ParseYAMLFrontMatter() content = %q, want %q", content, tt.wantContent)
				}
			}
		})
	}
}

func TestParsePromptFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantName  string
		wantError bool
	}{
		{
			name: "valid prompt file",
			content: `---
name: "Test Prompt"
author: "john"
category: "test"
tags: ["test"]
---
Test content with {variable}`,
			wantName:  "Test Prompt",
			wantError: false,
		},
		{
			name:      "invalid prompt file",
			content:   "No front matter here",
			wantName:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := ParsePromptFile(tt.content)

			if (err != nil) != tt.wantError {
				t.Errorf("ParsePromptFile() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && prompt.Name != tt.wantName {
				t.Errorf("ParsePromptFile() prompt.Name = %s, want %s", prompt.Name, tt.wantName)
			}
		})
	}
}

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single variable",
			content:  "Hello {name}!",
			expected: []string{"name"},
		},
		{
			name:     "multiple variables",
			content:  "Generate {format} documentation for {endpoint} API",
			expected: []string{"format", "endpoint"},
		},
		{
			name:     "duplicate variables",
			content:  "{name} is a {type}. {name} likes {food}.",
			expected: []string{"name", "type", "food"},
		},
		{
			name:     "no variables",
			content:  "This is a plain text without any variables.",
			expected: []string{},
		},
		{
			name:     "variables with spaces inside braces",
			content:  "Invalid { var } and {var with spaces} are ignored, but {valid} is extracted",
			expected: []string{"valid"},
		},
		{
			name:     "nested braces",
			content:  "This {{nested}} and {valid} variables",
			expected: []string{"valid"},
		},
		{
			name:     "empty braces",
			content:  "Empty {} braces and {valid} variable",
			expected: []string{"valid"},
		},
		{
			name:     "variables with underscores and numbers",
			content:  "Use {user_name} and {var2} and {test_var_3}",
			expected: []string{"user_name", "var2", "test_var_3"},
		},
		{
			name:     "multiline content",
			content:  "Line 1: {var1}\nLine 2: {var2}\n\nLine 4: {var3}",
			expected: []string{"var1", "var2", "var3"},
		},
		{
			name:     "variables with hyphens",
			content:  "Use {api-key} and {user-id}",
			expected: []string{"api-key", "user-id"},
		},
		{
			name:     "adjacent variables",
			content:  "{greeting}{punctuation} {name}!",
			expected: []string{"greeting", "punctuation", "name"},
		},
		{
			name:     "variable at start and end",
			content:  "{start} some text {end}",
			expected: []string{"start", "end"},
		},
		{
			name:     "special characters around variables",
			content:  "({var1}), [{var2}], <{var3}>",
			expected: []string{"var1", "var2", "var3"},
		},
		{
			name:     "case sensitivity",
			content:  "Use {userName} and {UserName} and {USERNAME}",
			expected: []string{"userName", "UserName", "USERNAME"},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "only braces",
			content:  "{}{}{}",
			expected: []string{},
		},
		{
			name:     "unclosed brace",
			content:  "This {unclosed and {valid} variable",
			expected: []string{"valid"},
		},
		{
			name:     "unopened brace",
			content:  "This unopened} and {valid} variable",
			expected: []string{"valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVariables(tt.content)
			
			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractVariables() returned %d variables, want %d", len(result), len(tt.expected))
				t.Errorf("Got: %v, Want: %v", result, tt.expected)
				return
			}
			
			// Check each variable
			for i, v := range tt.expected {
				if i >= len(result) || result[i] != v {
					t.Errorf("ExtractVariables()[%d] = %v, want %v", i, result[i], v)
				}
			}
		})
	}
}

func TestExtractVariables_OrderPreservation(t *testing.T) {
	content := "First {var1}, then {var2}, then {var1} again, finally {var3}"
	expected := []string{"var1", "var2", "var3"}
	
	result := ExtractVariables(content)
	
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ExtractVariables() = %v, want %v (order matters)", result, expected)
	}
}