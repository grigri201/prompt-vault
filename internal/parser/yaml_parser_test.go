package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestNewYAMLParser(t *testing.T) {
	tests := []struct {
		name   string
		config YAMLParserConfig
		want   *YAMLParser
	}{
		{
			name:   "default parser",
			config: YAMLParserConfig{},
			want: &YAMLParser{
				strict:         false,
				requireVersion: false,
				defaultAuthor:  "",
			},
		},
		{
			name: "strict parser with defaults",
			config: YAMLParserConfig{
				Strict:         true,
				RequireVersion: true,
				DefaultAuthor:  "test-user",
			},
			want: &YAMLParser{
				strict:         true,
				requireVersion: true,
				defaultAuthor:  "test-user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewYAMLParser(tt.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewYAMLParser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYAMLParser_ParseFrontMatter(t *testing.T) {
	tests := []struct {
		name        string
		parser      *YAMLParser
		content     string
		wantMeta    *models.PromptMeta
		wantBody    string
		wantErr     bool
		errContains string
	}{
		{
			name:   "valid front matter with content",
			parser: NewYAMLParser(YAMLParserConfig{}),
			content: `---
name: Test Prompt
author: test-user
category: testing
tags: [test, example]
---
This is the prompt content`,
			wantMeta: &models.PromptMeta{
				Name:     "Test Prompt",
				Author:   "test-user",
				Category: "testing",
				Tags:     []string{"test", "example"},
			},
			wantBody: "This is the prompt content",
			wantErr:  false,
		},
		{
			name:    "valid front matter with Windows line endings",
			parser:  NewYAMLParser(YAMLParserConfig{}),
			content: "---\r\nname: Test Prompt\r\nauthor: test-user\r\ncategory: testing\r\ntags: [test]\r\n---\r\nContent",
			wantMeta: &models.PromptMeta{
				Name:     "Test Prompt",
				Author:   "test-user",
				Category: "testing",
				Tags:     []string{"test"},
			},
			wantBody: "Content",
			wantErr:  false,
		},
		{
			name:   "front matter only (no content)",
			parser: NewYAMLParser(YAMLParserConfig{}),
			content: `---
name: Test Prompt
author: test-user
category: testing
tags: [test]
---`,
			wantMeta: &models.PromptMeta{
				Name:     "Test Prompt",
				Author:   "test-user",
				Category: "testing",
				Tags:     []string{"test"},
			},
			wantBody: "",
			wantErr:  false,
		},
		{
			name:   "missing closing delimiter in lenient mode",
			parser: NewYAMLParser(YAMLParserConfig{Strict: false}),
			content: `---
name: Test Prompt
author: test-user
category: testing
tags: [test]`,
			wantMeta: &models.PromptMeta{
				Name:     "Test Prompt",
				Author:   "test-user",
				Category: "testing",
				Tags:     []string{"test"},
			},
			wantBody: "",
			wantErr:  false,
		},
		{
			name:   "missing closing delimiter in strict mode",
			parser: NewYAMLParser(YAMLParserConfig{Strict: true}),
			content: `---
name: Test Prompt
author: test-user
category: testing
tags: [test]`,
			wantErr:     true,
			errContains: "unclosed YAML front matter",
		},
		{
			name:     "no front matter in lenient mode",
			parser:   NewYAMLParser(YAMLParserConfig{Strict: false}),
			content:  "Just plain content",
			wantMeta: &models.PromptMeta{},
			wantBody: "Just plain content",
			wantErr:  false,
		},
		{
			name:        "no front matter in strict mode",
			parser:      NewYAMLParser(YAMLParserConfig{Strict: true}),
			content:     "Just plain content",
			wantErr:     true,
			errContains: "missing YAML front matter",
		},
		{
			name:   "default author applied",
			parser: NewYAMLParser(YAMLParserConfig{DefaultAuthor: "default-user"}),
			content: `---
name: Test Prompt
category: testing
tags: [test]
---
Content`,
			wantMeta: &models.PromptMeta{
				Name:     "Test Prompt",
				Author:   "default-user",
				Category: "testing",
				Tags:     []string{"test"},
			},
			wantBody: "Content",
			wantErr:  false,
		},
		{
			name:   "version required but missing",
			parser: NewYAMLParser(YAMLParserConfig{RequireVersion: true}),
			content: `---
name: Test Prompt
author: test-user
category: testing
tags: [test]
---
Content`,
			wantErr:     true,
			errContains: "version is required",
		},
		{
			name:   "version required and present",
			parser: NewYAMLParser(YAMLParserConfig{RequireVersion: true}),
			content: `---
name: Test Prompt
author: test-user
category: testing
tags: [test]
version: "1.0"
---
Content`,
			wantMeta: &models.PromptMeta{
				Name:     "Test Prompt",
				Author:   "test-user",
				Category: "testing",
				Tags:     []string{"test"},
				Version:  "1.0",
			},
			wantBody: "Content",
			wantErr:  false,
		},
		{
			name:   "strict validation - missing required fields",
			parser: NewYAMLParser(YAMLParserConfig{Strict: true}),
			content: `---
name: Test Prompt
---
Content`,
			wantErr:     true,
			errContains: "author is required",
		},
		{
			name:   "lenient validation - only name required",
			parser: NewYAMLParser(YAMLParserConfig{Strict: false}),
			content: `---
name: Test Prompt
---
Content`,
			wantMeta: &models.PromptMeta{
				Name: "Test Prompt",
			},
			wantBody: "Content",
			wantErr:  false,
		},
		{
			name:   "invalid YAML",
			parser: NewYAMLParser(YAMLParserConfig{}),
			content: `---
name: Test Prompt
invalid yaml syntax [
---
Content`,
			wantErr:     true,
			errContains: "failed to parse YAML",
		},
		{
			name:   "empty name in lenient mode",
			parser: NewYAMLParser(YAMLParserConfig{Strict: false}),
			content: `---
author: test-user
category: testing
tags: [test]
---
Content`,
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:   "optional fields",
			parser: NewYAMLParser(YAMLParserConfig{}),
			content: `---
name: Test Prompt
author: test-user
category: testing
tags: [test]
description: "This is a test prompt"
parent: "parent-gist-id"
---
Content`,
			wantMeta: &models.PromptMeta{
				Name:        "Test Prompt",
				Author:      "test-user",
				Category:    "testing",
				Tags:        []string{"test"},
				Description: "This is a test prompt",
				Parent:      "parent-gist-id",
			},
			wantBody: "Content",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMeta, gotBody, err := tt.parser.ParseFrontMatter(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFrontMatter() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseFrontMatter() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFrontMatter() unexpected error = %v", err)
				return
			}

			if !reflect.DeepEqual(gotMeta, tt.wantMeta) {
				t.Errorf("ParseFrontMatter() gotMeta = %v, want %v", gotMeta, tt.wantMeta)
			}

			if gotBody != tt.wantBody {
				t.Errorf("ParseFrontMatter() gotBody = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}

func TestYAMLParser_ParsePromptFile(t *testing.T) {
	parser := NewYAMLParser(YAMLParserConfig{})

	content := `---
name: Test Prompt
author: test-user
category: testing
tags: [test]
---
This is the prompt content`

	prompt, err := parser.ParsePromptFile(content)
	if err != nil {
		t.Fatalf("ParsePromptFile() unexpected error = %v", err)
	}

	if prompt.Name != "Test Prompt" {
		t.Errorf("ParsePromptFile() Name = %v, want %v", prompt.Name, "Test Prompt")
	}
	if prompt.Author != "test-user" {
		t.Errorf("ParsePromptFile() Author = %v, want %v", prompt.Author, "test-user")
	}
	if prompt.Category != "testing" {
		t.Errorf("ParsePromptFile() Category = %v, want %v", prompt.Category, "testing")
	}
	if !reflect.DeepEqual(prompt.Tags, []string{"test"}) {
		t.Errorf("ParsePromptFile() Tags = %v, want %v", prompt.Tags, []string{"test"})
	}
	if prompt.Content != "This is the prompt content" {
		t.Errorf("ParsePromptFile() Content = %v, want %v", prompt.Content, "This is the prompt content")
	}
}

func TestYAMLParser_ValidateMetadata(t *testing.T) {
	tests := []struct {
		name        string
		parser      *YAMLParser
		meta        *models.PromptMeta
		wantErr     bool
		errContains string
	}{
		{
			name:   "valid metadata in strict mode",
			parser: NewYAMLParser(YAMLParserConfig{Strict: true}),
			meta: &models.PromptMeta{
				Name:     "Test",
				Author:   "user",
				Category: "test",
				Tags:     []string{"tag"},
			},
			wantErr: false,
		},
		{
			name:   "missing tags in strict mode",
			parser: NewYAMLParser(YAMLParserConfig{Strict: true}),
			meta: &models.PromptMeta{
				Name:     "Test",
				Author:   "user",
				Category: "test",
			},
			wantErr:     true,
			errContains: "at least one tag is required",
		},
		{
			name:   "only name required in lenient mode",
			parser: NewYAMLParser(YAMLParserConfig{Strict: false}),
			meta: &models.PromptMeta{
				Name: "Test",
			},
			wantErr: false,
		},
		{
			name:   "version required",
			parser: NewYAMLParser(YAMLParserConfig{RequireVersion: true}),
			meta: &models.PromptMeta{
				Name: "Test",
			},
			wantErr:     true,
			errContains: "version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parser.ValidateMetadata(tt.meta)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateMetadata() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateMetadata() error = %v, want error containing %v", err, tt.errContains)
				}
			} else if err != nil {
				t.Errorf("ValidateMetadata() unexpected error = %v", err)
			}
		})
	}
}
