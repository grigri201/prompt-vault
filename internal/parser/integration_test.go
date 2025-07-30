package parser_test

import (
	"testing"

	"github.com/grigri201/prompt-vault/internal/parser"
)

// TestIntegration_UnifiedParsing verifies all parsing modes work correctly
func TestIntegration_UnifiedParsing(t *testing.T) {
	// Test content with full metadata
	fullContent := `---
name: Integration Test Prompt
author: testuser
category: testing
tags: [integration, test]
version: "1.0.0"
description: Testing unified components
---

This is a test prompt for {subject}.
It demonstrates the integration of:
- Unified YAML parsing
- Consistent error handling
- Shared components`

	// Test strict parsing (used in upload)
	t.Run("strict parsing", func(t *testing.T) {
		strictParser := parser.NewYAMLParser(parser.YAMLParserConfig{
			Strict:         true,
			RequireVersion: true,
		})

		prompt, err := strictParser.ParsePromptFile(fullContent)
		if err != nil {
			t.Fatalf("Failed to parse prompt in strict mode: %v", err)
		}

		// Verify all fields are parsed correctly
		if prompt.Name != "Integration Test Prompt" {
			t.Errorf("Expected name 'Integration Test Prompt', got '%s'", prompt.Name)
		}
		if prompt.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", prompt.Version)
		}
		if len(prompt.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(prompt.Tags))
		}
	})

	// Test lenient parsing (used in sync)
	t.Run("lenient parsing", func(t *testing.T) {
		lenientParser := parser.NewYAMLParser(parser.YAMLParserConfig{
			Strict: false,
		})

		// Minimal content that would fail strict parsing
		minimalContent := `---
name: Minimal Prompt
---
Just the content`

		prompt, err := lenientParser.ParsePromptFile(minimalContent)
		if err != nil {
			t.Fatalf("Failed to parse minimal prompt in lenient mode: %v", err)
		}

		if prompt.Name != "Minimal Prompt" {
			t.Errorf("Expected name 'Minimal Prompt', got '%s'", prompt.Name)
		}
		if prompt.Content != "Just the content" {
			t.Errorf("Expected content 'Just the content', got '%s'", prompt.Content)
		}
	})

	// Test backward compatibility
	t.Run("backward compatibility", func(t *testing.T) {
		// Test old ParseYAMLFrontMatter function
		meta, _, err := parser.ParseYAMLFrontMatter(fullContent)
		if err != nil {
			t.Fatalf("ParseYAMLFrontMatter failed: %v", err)
		}

		if meta.Name != "Integration Test Prompt" {
			t.Errorf("Expected name 'Integration Test Prompt', got '%s'", meta.Name)
		}

		// Test old ParsePromptFile function
		prompt, err := parser.ParsePromptFile(fullContent)
		if err != nil {
			t.Fatalf("ParsePromptFile failed: %v", err)
		}

		if prompt.Name != "Integration Test Prompt" {
			t.Errorf("Expected name 'Integration Test Prompt', got '%s'", prompt.Name)
		}
	})
}

// TestIntegration_ParsingModes tests different parsing configurations
func TestIntegration_ParsingModes(t *testing.T) {
	tests := []struct {
		name       string
		config     parser.YAMLParserConfig
		content    string
		shouldFail bool
		checkName  string
	}{
		{
			name: "strict mode with valid prompt",
			config: parser.YAMLParserConfig{
				Strict: true,
			},
			content: `---
name: Valid Prompt
author: user
category: test
tags: [test]
---
Content`,
			shouldFail: false,
			checkName:  "Valid Prompt",
		},
		{
			name: "strict mode missing required fields",
			config: parser.YAMLParserConfig{
				Strict: true,
			},
			content: `---
name: Missing Author
---
Content`,
			shouldFail: true,
		},
		{
			name: "lenient mode missing fields",
			config: parser.YAMLParserConfig{
				Strict: false,
			},
			content: `---
name: Lenient Prompt
---
Content`,
			shouldFail: false,
			checkName:  "Lenient Prompt",
		},
		{
			name: "require version mode",
			config: parser.YAMLParserConfig{
				Strict:         false,
				RequireVersion: true,
			},
			content: `---
name: No Version
author: user
---
Content`,
			shouldFail: true,
		},
		{
			name: "default author mode",
			config: parser.YAMLParserConfig{
				Strict:        false,
				DefaultAuthor: "default-user",
			},
			content: `---
name: Default Author Test
---
Content`,
			shouldFail: false,
			checkName:  "Default Author Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := parser.NewYAMLParser(tt.config)
			prompt, err := parser.ParsePromptFile(tt.content)

			if tt.shouldFail {
				if err == nil {
					t.Error("Expected parsing to fail but it succeeded")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected parsing to succeed but got error: %v", err)
				}
				if tt.checkName != "" && prompt.Name != tt.checkName {
					t.Errorf("Expected name '%s', got '%s'", tt.checkName, prompt.Name)
				}
				if tt.config.DefaultAuthor != "" && prompt.Author != tt.config.DefaultAuthor {
					t.Errorf("Expected default author '%s', got '%s'", tt.config.DefaultAuthor, prompt.Author)
				}
			}
		})
	}
}

// TestIntegration_EdgeCases tests edge cases in parsing
func TestIntegration_EdgeCases(t *testing.T) {
	lenientParser := parser.NewYAMLParser(parser.YAMLParserConfig{
		Strict: false,
	})

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty content",
			content: "",
			wantErr: false, // Lenient mode should handle this
		},
		{
			name:    "only frontmatter delimiters",
			content: "---\n---",
			wantErr: true, // Even lenient mode requires at least a name
		},
		{
			name: "frontmatter without closing delimiter",
			content: `---
name: Unclosed`,
			wantErr: false, // Lenient mode treats as all frontmatter
		},
		{
			name: "multiple yaml documents",
			content: `---
name: First
---
---
name: Second
---`,
			wantErr: false, // Should parse first document
		},
		{
			name: "invalid yaml syntax",
			content: `---
name: [[[
---`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lenientParser.ParsePromptFile(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePromptFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIntegration_FormattingConsistency ensures formatting functions work correctly
func TestIntegration_FormattingConsistency(t *testing.T) {
	// Test content with all fields
	originalContent := `---
name: Format Test
author: testuser
category: testing
tags: [format, test]
version: "1.0.0"
description: Testing formatting
---
This is the prompt content`

	// Parse it first
	strictParser := parser.NewYAMLParser(parser.YAMLParserConfig{
		Strict: true,
	})

	parsedPrompt, err := strictParser.ParsePromptFile(originalContent)
	if err != nil {
		t.Fatalf("Failed to parse original content: %v", err)
	}

	// Format using the function
	formatted := parser.FormatPromptFile(&parsedPrompt.PromptMeta, parsedPrompt.Content)

	// Parse it back
	reparsedPrompt, err := strictParser.ParsePromptFile(formatted)
	if err != nil {
		t.Fatalf("Failed to parse formatted content: %v", err)
	}

	// Verify round-trip consistency
	if reparsedPrompt.Name != parsedPrompt.Name {
		t.Errorf("Name mismatch: expected '%s', got '%s'", parsedPrompt.Name, reparsedPrompt.Name)
	}
	if reparsedPrompt.Author != parsedPrompt.Author {
		t.Errorf("Author mismatch: expected '%s', got '%s'", parsedPrompt.Author, reparsedPrompt.Author)
	}
	if reparsedPrompt.Category != parsedPrompt.Category {
		t.Errorf("Category mismatch: expected '%s', got '%s'", parsedPrompt.Category, reparsedPrompt.Category)
	}
	if reparsedPrompt.Version != parsedPrompt.Version {
		t.Errorf("Version mismatch: expected '%s', got '%s'", parsedPrompt.Version, reparsedPrompt.Version)
	}
	if reparsedPrompt.Content != parsedPrompt.Content {
		t.Errorf("Content mismatch: expected '%s', got '%s'", parsedPrompt.Content, reparsedPrompt.Content)
	}
}

// TestIntegration_AllCommandUsages verifies parser works for all command scenarios
func TestIntegration_AllCommandUsages(t *testing.T) {
	// Upload command scenario - strict parsing
	t.Run("upload command usage", func(t *testing.T) {
		uploadParser := parser.NewYAMLParser(parser.YAMLParserConfig{
			Strict: true,
		})

		uploadContent := `---
name: Upload Test
author: uploader
category: commands
tags: [upload, test]
version: "1.0.0"
---
Upload command test content`

		prompt, err := uploadParser.ParsePromptFile(uploadContent)
		if err != nil {
			t.Fatalf("Upload parsing failed: %v", err)
		}

		// Verify strict validation passed
		if prompt.Author != "uploader" {
			t.Error("Upload should require author in strict mode")
		}
	})

	// Sync command scenario - lenient parsing
	t.Run("sync command usage", func(t *testing.T) {
		syncParser := parser.NewYAMLParser(parser.YAMLParserConfig{
			Strict: false,
		})

		// Sync might encounter prompts with missing fields
		syncContent := `---
name: Sync Test
category: commands
---
Sync command test content`

		prompt, err := syncParser.ParsePromptFile(syncContent)
		if err != nil {
			t.Fatalf("Sync parsing failed: %v", err)
		}

		// Should succeed even without author
		if prompt.Name != "Sync Test" {
			t.Error("Sync should parse prompts with missing fields")
		}
	})

	// Share command scenario - lenient parsing with parent field
	t.Run("share command usage", func(t *testing.T) {
		shareParser := parser.NewYAMLParser(parser.YAMLParserConfig{
			Strict: false,
		})

		shareContent := `---
name: Share Test
author: sharer
category: commands
parent: private123
---
Share command test content`

		prompt, err := shareParser.ParsePromptFile(shareContent)
		if err != nil {
			t.Fatalf("Share parsing failed: %v", err)
		}

		// Should preserve parent field
		if prompt.Parent != "private123" {
			t.Error("Share should preserve parent field")
		}
	})
}
