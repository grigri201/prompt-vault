package validator

import (
	"bytes"
	"strings"
	"testing"

	"github.com/grigri/pv/internal/errors"
)

func TestYAMLValidator_ValidatePromptFile(t *testing.T) {
	validator := NewYAMLValidator()

	testCases := []struct {
		name                string
		input               []byte
		expectError         bool
		expectedErrorType   interface{}
		expectedName        string
		expectedAuthor      string
		expectedDescription string
		expectedTags        []string
		expectedVersion     string
		expectedContent     string
	}{
		{
			name: "valid YAML with all fields",
			input: []byte(`name: "Test Prompt"
author: "Test Author"
description: "This is a test prompt"
tags:
  - "test"
  - "example"
version: "1.0.0"
---
This is the prompt content
with multiple lines
and various formatting.`),
			expectError:         false,
			expectedName:        "Test Prompt",
			expectedAuthor:      "Test Author",
			expectedDescription: "This is a test prompt",
			expectedTags:        []string{"test", "example"},
			expectedVersion:     "1.0.0",
			expectedContent:     "This is the prompt content\nwith multiple lines\nand various formatting.",
		},
		{
			name: "valid YAML with only required fields",
			input: []byte(`name: "Minimal Prompt"
author: "Minimal Author"
---
Simple prompt content`),
			expectError:     false,
			expectedName:    "Minimal Prompt",
			expectedAuthor:  "Minimal Author",
			expectedContent: "Simple prompt content",
		},
		{
			name: "valid YAML with empty optional fields",
			input: []byte(`name: "Empty Optional Fields"
author: "Author Name"
description: ""
tags: []
version: ""
---
Content with empty optionals`),
			expectError:     false,
			expectedName:    "Empty Optional Fields",
			expectedAuthor:  "Author Name",
			expectedContent: "Content with empty optionals",
		},
		{
			name: "content with multiple YAML separators",
			input: []byte(`name: "Multi Separator"
author: "Author"
---
This content has --- in it
---
And another --- separator
More content here`),
			expectError:     false,
			expectedName:    "Multi Separator",
			expectedAuthor:  "Author",
			expectedContent: "This content has --- in it\n---\nAnd another --- separator\nMore content here",
		},
		{
			name: "content with leading and trailing whitespace",
			input: []byte(`name: "Whitespace Test"
author: "Author"
---


  Content with whitespace  


`),
			expectError:     false,
			expectedName:    "Whitespace Test",
			expectedAuthor:  "Author",
			expectedContent: "Content with whitespace",
		},
		{
			name:              "missing YAML separator",
			input:             []byte(`name: "No Separator"\nauthor: "Author"\nThis should be content but no separator`),
			expectError:       true,
			expectedErrorType: errors.ErrInvalidYAML,
		},
		{
			name: "invalid YAML syntax",
			input: []byte(`name: "Invalid YAML"
author: [invalid yaml structure
description: "Missing closing bracket"
---
Content`),
			expectError:       true,
			expectedErrorType: errors.ErrInvalidYAML,
		},
		{
			name:              "empty input",
			input:             []byte(""),
			expectError:       true,
			expectedErrorType: errors.ErrInvalidYAML,
		},
		{
			name:              "only separator",
			input:             []byte("---"),
			expectError:       true,
			expectedErrorType: errors.ErrInvalidYAML,
		},
		{
			name: "separator at beginning (standard format)",
			input: []byte(`---
name: "Standard Format"
author: "Author"
---
Content`),
			expectError:     false,
			expectedName:    "Standard Format",
			expectedAuthor:  "Author",
			expectedContent: "Content",
		},
		{
			name: "unicode characters in metadata",
			input: []byte(`name: "Unicode Test 🚀"
author: "作者名称"
description: "描述信息 with émojis 🎉"
tags:
  - "标签"
  - "テスト"
version: "1.0.0-α"
---
Unicode content: 你好世界 🌍`),
			expectError:         false,
			expectedName:        "Unicode Test 🚀",
			expectedAuthor:      "作者名称",
			expectedDescription: "描述信息 with émojis 🎉",
			expectedTags:        []string{"标签", "テスト"},
			expectedVersion:     "1.0.0-α",
			expectedContent:     "Unicode content: 你好世界 🌍",
		},
		{
			name: "special characters in content",
			input: []byte(`name: "Special Chars"
author: "Author"
---
Content with special chars: !@#$%^&*()_+-=[]{}|;:,.<>?
And newlines\ttabs\r\n
"Quotes" and 'apostrophes'`),
			expectError:     false,
			expectedName:    "Special Chars",
			expectedAuthor:  "Author",
			expectedContent: "Content with special chars: !@#$%^&*()_+-=[]{}|;:,.<>?\nAnd newlines\\ttabs\\r\\n\n\"Quotes\" and 'apostrophes'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.ValidatePromptFile(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Check error type if specified
				if tc.expectedErrorType != nil {
					appErr, ok := err.(errors.AppError)
					if !ok {
						t.Errorf("Expected AppError, got %T: %v", err, err)
						return
					}

					expectedErr, ok := tc.expectedErrorType.(errors.AppError)
					if ok && appErr.Err != expectedErr {
						t.Errorf("Expected error Err %v, got %v", expectedErr, appErr.Err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			// Validate metadata fields
			if result.Metadata.Name != tc.expectedName {
				t.Errorf("Expected name %q, got %q", tc.expectedName, result.Metadata.Name)
			}

			if result.Metadata.Author != tc.expectedAuthor {
				t.Errorf("Expected author %q, got %q", tc.expectedAuthor, result.Metadata.Author)
			}

			if tc.expectedDescription != "" && result.Metadata.Description != tc.expectedDescription {
				t.Errorf("Expected description %q, got %q", tc.expectedDescription, result.Metadata.Description)
			}

			if tc.expectedVersion != "" && result.Metadata.Version != tc.expectedVersion {
				t.Errorf("Expected version %q, got %q", tc.expectedVersion, result.Metadata.Version)
			}

			if len(tc.expectedTags) > 0 {
				if len(result.Metadata.Tags) != len(tc.expectedTags) {
					t.Errorf("Expected %d tags, got %d", len(tc.expectedTags), len(result.Metadata.Tags))
				} else {
					for i, expectedTag := range tc.expectedTags {
						if i < len(result.Metadata.Tags) && result.Metadata.Tags[i] != expectedTag {
							t.Errorf("Expected tag[%d] %q, got %q", i, expectedTag, result.Metadata.Tags[i])
						}
					}
				}
			}

			// Validate content
			if result.Content != tc.expectedContent {
				t.Errorf("Expected content %q, got %q", tc.expectedContent, result.Content)
			}
		})
	}
}

func TestYAMLValidator_ValidateRequired(t *testing.T) {
	validator := NewYAMLValidator()

	testCases := []struct {
		name           string
		input          *PromptFileContent
		expectError    bool
		expectedField  string
		expectedErrMsg string
	}{
		{
			name: "valid prompt with all required fields",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Prompt",
					Author: "Valid Author",
				},
				Content: "Valid content",
			},
			expectError: false,
		},
		{
			name: "valid prompt with optional fields",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:        "Valid Prompt",
					Author:      "Valid Author",
					Description: "Optional description",
					Tags:        []string{"tag1", "tag2"},
					Version:     "1.0.0",
				},
				Content: "Valid content",
			},
			expectError: false,
		},
		{
			name:           "nil prompt content",
			input:          nil,
			expectError:    true,
			expectedErrMsg: "prompt content cannot be nil",
		},
		{
			name: "empty name",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "",
					Author: "Valid Author",
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "name",
			expectedErrMsg: "name is required and cannot be empty",
		},
		{
			name: "whitespace only name",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "   \t\n   ",
					Author: "Valid Author",
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "name",
			expectedErrMsg: "name is required and cannot be empty",
		},
		{
			name: "name too long",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   strings.Repeat("a", 101), // 101 characters
					Author: "Valid Author",
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "name",
			expectedErrMsg: "name cannot be longer than 100 characters",
		},
		{
			name: "name exactly 100 characters (boundary test)",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   strings.Repeat("a", 100), // exactly 100 characters
					Author: "Valid Author",
				},
				Content: "Content",
			},
			expectError: false,
		},
		{
			name: "empty author",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "",
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "author",
			expectedErrMsg: "author is required and cannot be empty",
		},
		{
			name: "whitespace only author",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "   \t\n   ",
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "author",
			expectedErrMsg: "author is required and cannot be empty",
		},
		{
			name: "author too long",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: strings.Repeat("a", 51), // 51 characters
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "author",
			expectedErrMsg: "author cannot be longer than 50 characters",
		},
		{
			name: "author exactly 50 characters (boundary test)",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: strings.Repeat("a", 50), // exactly 50 characters
				},
				Content: "Content",
			},
			expectError: false,
		},
		{
			name: "description too long",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:        "Valid Name",
					Author:      "Valid Author",
					Description: strings.Repeat("a", 501), // 501 characters
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "description",
			expectedErrMsg: "description cannot be longer than 500 characters",
		},
		{
			name: "description exactly 500 characters (boundary test)",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:        "Valid Name",
					Author:      "Valid Author",
					Description: strings.Repeat("a", 500), // exactly 500 characters
				},
				Content: "Content",
			},
			expectError: false,
		},
		{
			name: "empty tag in tags list",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "Valid Author",
					Tags:   []string{"valid-tag", "", "another-tag"},
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "tags",
			expectedErrMsg: "tags cannot contain empty values",
		},
		{
			name: "whitespace only tag",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "Valid Author",
					Tags:   []string{"valid-tag", "   \t\n   ", "another-tag"},
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "tags",
			expectedErrMsg: "tags cannot contain empty values",
		},
		{
			name: "tag too long",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "Valid Author",
					Tags:   []string{"valid-tag", strings.Repeat("a", 21)}, // 21 characters
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "tags",
			expectedErrMsg: "each tag cannot be longer than 20 characters",
		},
		{
			name: "tag exactly 20 characters (boundary test)",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "Valid Author",
					Tags:   []string{strings.Repeat("a", 20)}, // exactly 20 characters
				},
				Content: "Content",
			},
			expectError: false,
		},
		{
			name: "duplicate tags",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "Valid Author",
					Tags:   []string{"tag1", "tag2", "tag1", "tag3"},
				},
				Content: "Content",
			},
			expectError:    true,
			expectedField:  "tags",
			expectedErrMsg: "duplicate tags are not allowed",
		},
		{
			name: "case-sensitive duplicate tags (should pass)",
			input: &PromptFileContent{
				Metadata: PromptMetadata{
					Name:   "Valid Name",
					Author: "Valid Author",
					Tags:   []string{"Tag1", "tag1", "TAG1"},
				},
				Content: "Content",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateRequired(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Check error type and message
				if tc.expectedField != "" {
					validationErr, ok := err.(errors.ValidationError)
					if !ok {
						t.Errorf("Expected ValidationError, got %T: %v", err, err)
						return
					}

					if validationErr.Field != tc.expectedField {
						t.Errorf("Expected error field %q, got %q", tc.expectedField, validationErr.Field)
					}

					if !strings.Contains(validationErr.Message, tc.expectedErrMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tc.expectedErrMsg, validationErr.Message)
					}
				} else if tc.expectedErrMsg != "" {
					if !strings.Contains(err.Error(), tc.expectedErrMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tc.expectedErrMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestYAMLValidator_EdgeCases(t *testing.T) {
	validator := NewYAMLValidator()

	testCases := []struct {
		name        string
		input       []byte
		expectError bool
		description string
	}{
		{
			name: "very large content",
			input: []byte(`name: "Large Content Test"
author: "Test Author"
---
` + strings.Repeat("This is a very long content line. ", 1000)),
			expectError: false,
			description: "Should handle large content without issues",
		},
		{
			name: "YAML with comments",
			input: []byte(`# This is a comment
name: "Comment Test" # Inline comment
author: "Test Author"
description: "Test with comments"
# Another comment
tags:
  - "test" # Comment on tag
version: "1.0" # Version comment
---
Content with comments`),
			expectError: false,
			description: "Should parse YAML with comments correctly",
		},
		{
			name: "YAML with complex structures",
			input: []byte(`name: "Complex Test"
author: "Test Author"
description: |
  This is a multi-line description
  with YAML literal block style
  and multiple lines
tags:
  - "tag1"
  - "tag2"
metadata:
  custom: "value"
  nested:
    key: "value"
---
Content after complex metadata`),
			expectError: false,
			description: "Should handle complex YAML structures",
		},
		{
			name: "empty content after separator",
			input: []byte(`name: "Empty Content"
author: "Test Author"
---`),
			expectError: false,
			description: "Should handle empty content after separator",
		},
		{
			name: "content with only whitespace after separator",
			input: []byte(`name: "Whitespace Content"
author: "Test Author"
---
   
	
   `),
			expectError: false,
			description: "Content with only whitespace should be trimmed to empty",
		},
		{
			name: "YAML with quoted strings",
			input: []byte(`name: "Quoted String Test"
author: 'Single Quoted Author'
description: "Double \"quoted\" with escapes"
tags:
  - 'single quoted tag'
  - "double quoted tag"
version: "1.0.0"
---
Content with 'mixed' "quotes"`),
			expectError: false,
			description: "Should handle various quoting styles",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.ValidatePromptFile(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for case: %s", tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for case %s: %v", tc.description, err)
					return
				}

				if result == nil {
					t.Errorf("Expected result but got nil for case: %s", tc.description)
					return
				}

				// Basic validation that we got some result
				if result.Metadata.Name == "" {
					t.Errorf("Expected non-empty name for case: %s", tc.description)
				}

				if result.Metadata.Author == "" {
					t.Errorf("Expected non-empty author for case: %s", tc.description)
				}
			}
		})
	}
}

func TestYAMLValidator_Integration(t *testing.T) {
	validator := NewYAMLValidator()

	// Test the complete validation flow
	testInput := []byte(`name: "Integration Test Prompt"
author: "Integration Author"
description: "This is an integration test for the YAML validator"
tags:
  - "integration"
  - "test"
  - "yaml"
version: "2.0.0"
---
This is the prompt content for integration testing.

It includes multiple lines and various formatting:
- Bullet points
- **Bold text** 
- ` + "`code snippets`" + `

And some special characters: !@#$%^&*()`)

	// Step 1: Parse the YAML
	result, err := validator.ValidatePromptFile(testInput)
	if err != nil {
		t.Fatalf("Failed to parse valid YAML: %v", err)
	}

	// Step 2: Validate required fields
	err = validator.ValidateRequired(result)
	if err != nil {
		t.Fatalf("Failed to validate required fields: %v", err)
	}

	// Step 3: Verify all data was parsed correctly
	expected := PromptFileContent{
		Metadata: PromptMetadata{
			Name:        "Integration Test Prompt",
			Author:      "Integration Author",
			Description: "This is an integration test for the YAML validator",
			Tags:        []string{"integration", "test", "yaml"},
			Version:     "2.0.0",
		},
		Content: `This is the prompt content for integration testing.

It includes multiple lines and various formatting:
- Bullet points
- **Bold text** 
- ` + "`code snippets`" + `

And some special characters: !@#$%^&*()`,
	}

	if result.Metadata.Name != expected.Metadata.Name {
		t.Errorf("Name mismatch: got %q, want %q", result.Metadata.Name, expected.Metadata.Name)
	}

	if result.Metadata.Author != expected.Metadata.Author {
		t.Errorf("Author mismatch: got %q, want %q", result.Metadata.Author, expected.Metadata.Author)
	}

	if result.Metadata.Description != expected.Metadata.Description {
		t.Errorf("Description mismatch: got %q, want %q", result.Metadata.Description, expected.Metadata.Description)
	}

	if len(result.Metadata.Tags) != len(expected.Metadata.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(result.Metadata.Tags), len(expected.Metadata.Tags))
	} else {
		for i, tag := range expected.Metadata.Tags {
			if result.Metadata.Tags[i] != tag {
				t.Errorf("Tag[%d] mismatch: got %q, want %q", i, result.Metadata.Tags[i], tag)
			}
		}
	}

	if result.Metadata.Version != expected.Metadata.Version {
		t.Errorf("Version mismatch: got %q, want %q", result.Metadata.Version, expected.Metadata.Version)
	}

	if result.Content != expected.Content {
		t.Errorf("Content mismatch:\nGot:\n%q\nWant:\n%q", result.Content, expected.Content)
	}
}

func TestYAMLValidator_RobustParsing(t *testing.T) {
	validator := NewYAMLValidator()

	testCases := []struct {
		name        string
		input       []byte
		expectError bool
		description string
	}{
		{
			name: "YAML 中包含 --- 分隔符（引用字符串）",
			input: []byte(`name: "Test Prompt"
author: "Test Author"
description: "This description contains --- separator"
---
Content here should be parsed correctly`),
			expectError: false,
			description: "应该正确解析包含 --- 的 YAML 字段值",
		},
		{
			name: "YAML 中包含 --- 分隔符（多行字符串）",
			input: []byte(`name: "Test Prompt"
author: "Test Author"
description: |
  Multi-line description
  with --- separator in the middle
  of the content
---
Actual prompt content`),
			expectError: false,
			description: "应该正确解析多行 YAML 字段中的 --- 分隔符",
		},
		{
			name: "标准 Jekyll/Hugo 格式",
			input: []byte(`---
name: "Jekyll Test"
author: "Hugo Author"
tags:
  - "jekyll"
  - "hugo"
---
Standard front matter format content`),
			expectError: false,
			description: "应该支持标准的前导 --- 格式",
		},
		{
			name: "内容部分包含多个 --- 分隔符",
			input: []byte(`name: "Content Test"
author: "Test Author"
---
This content has multiple --- separators
---
And more content after --- another separator
---
Final content`),
			expectError: false,
			description: "内容部分的 --- 分隔符应该被保留",
		},
		{
			name: "未闭合的 front matter",
			input: []byte(`name: "Unclosed Test"
author: "Test Author"
description: "No closing separator"
This should be treated as content but there's no --- separator`),
			expectError: true,
			description: "缺少闭合 --- 分隔符应该报错",
		},
		{
			name: "空的 front matter",
			input: []byte(`---
---
Only content here`),
			expectError: false,
			description: "空的标准格式 front matter 应该被允许",
		},
		{
			name:        "过长的文件",
			input:       bytes.Repeat([]byte("a"), 11<<20), // 11MB
			expectError: true,
			description: "超过大小限制的文件应该被拒绝",
		},
		{
			name:        "非 UTF-8 内容",
			input:       []byte{0xff, 0xfe, 0xfd}, // 无效的 UTF-8 字节
			expectError: true,
			description: "非 UTF-8 内容应该被拒绝",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.ValidatePromptFile(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for case: %s", tc.description)
				}
				if result != nil {
					t.Errorf("Expected nil result but got: %+v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for case %s: %v", tc.description, err)
					return
				}

				if result == nil {
					t.Errorf("Expected result but got nil for case: %s", tc.description)
					return
				}

				// 验证基本字段（跳过有空 front matter 的情况）
				if !strings.Contains(tc.description, "空的标准格式") {
					if result.Metadata.Name == "" {
						t.Errorf("Expected non-empty name for case: %s", tc.description)
					}

					if result.Metadata.Author == "" {
						t.Errorf("Expected non-empty author for case: %s", tc.description)
					}
				}

				// 验证内容部分
				if result.Content == "" {
					t.Errorf("Expected non-empty content for case: %s", tc.description)
				}
			}
		})
	}
}

func TestFrontMatterParser_SpecificCases(t *testing.T) {
	parser := &FrontMatterParser{}

	testCases := []struct {
		name            string
		input           []byte
		expectedYAML    string
		expectedContent string
		expectError     bool
	}{
		{
			name: "YAML 字段包含 --- 的正确解析",
			input: []byte(`name: "Test"
description: "Has --- inside"
author: "Author"
---
Content`),
			expectedYAML: `name: "Test"
description: "Has --- inside"
author: "Author"`,
			expectedContent: "Content",
			expectError:     false,
		},
		{
			name: "多行 YAML 包含 --- 的正确解析",
			input: []byte(`name: "Test"
description: |
  Multi-line
  with --- separator
  in content
author: "Author"
---
Actual content`),
			expectedYAML: `name: "Test"
description: |
  Multi-line
  with --- separator
  in content
author: "Author"`,
			expectedContent: "Actual content",
			expectError:     false,
		},
		{
			name: "标准格式的正确解析",
			input: []byte(`---
name: "Standard"
author: "Author"
---
Content`),
			expectedYAML: `name: "Standard"
author: "Author"`,
			expectedContent: "Content",
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.Parse(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.YAMLContent != tc.expectedYAML {
				t.Errorf("YAML content mismatch:\nExpected: %q\nGot: %q", tc.expectedYAML, result.YAMLContent)
			}

			if result.BodyContent != tc.expectedContent {
				t.Errorf("Body content mismatch:\nExpected: %q\nGot: %q", tc.expectedContent, result.BodyContent)
			}
		})
	}
}
