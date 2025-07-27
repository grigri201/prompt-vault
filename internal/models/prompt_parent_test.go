package models

import (
	"testing"
	"gopkg.in/yaml.v3"
)

func TestPromptMeta_WithParentField(t *testing.T) {
	tests := []struct {
		name     string
		meta     PromptMeta
		hasParent bool
	}{
		{
			name: "prompt meta with parent field",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test", "example"},
				Parent:   "abc123",
			},
			hasParent: true,
		},
		{
			name: "prompt meta without parent field",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test", "example"},
			},
			hasParent: false,
		},
		{
			name: "prompt meta with empty parent field",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test", "example"},
				Parent:   "",
			},
			hasParent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate still works with parent field
			err := tt.meta.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v", err)
			}

			// Check parent field
			if tt.hasParent && tt.meta.Parent == "" {
				t.Error("Expected parent field to be set")
			}
			if !tt.hasParent && tt.meta.Parent != "" {
				t.Error("Expected parent field to be empty")
			}
		})
	}
}

func TestPromptMeta_MarshalYAML_WithParent(t *testing.T) {
	tests := []struct {
		name         string
		meta         PromptMeta
		wantParent   bool
		expectedYAML string
	}{
		{
			name: "marshal with parent field",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test", "example"},
				Version:  "1.0.0",
				Parent:   "private-gist-123",
			},
			wantParent: true,
			expectedYAML: `name: Test Prompt
author: john
category: test
tags:
    - test
    - example
version: 1.0.0
parent: private-gist-123
`,
		},
		{
			name: "marshal without parent field",
			meta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test"},
				Version:  "1.0.0",
			},
			wantParent: false,
			expectedYAML: `name: Test Prompt
author: john
category: test
tags:
    - test
version: 1.0.0
`,
		},
		{
			name: "marshal with empty parent field",
			meta: PromptMeta{
				Name:        "Test Prompt",
				Author:      "john",
				Category:    "test",
				Tags:        []string{"test"},
				Description: "Test description",
				Parent:      "",
			},
			wantParent: false,
			expectedYAML: `name: Test Prompt
author: john
category: test
tags:
    - test
description: Test description
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.meta)
			if err != nil {
				t.Fatalf("Failed to marshal PromptMeta: %v", err)
			}

			yamlStr := string(data)
			
			// Check if parent field is present when expected
			containsParent := contains(yamlStr, "parent:")
			if tt.wantParent && !containsParent {
				t.Errorf("Expected YAML to contain parent field, got:\n%s", yamlStr)
			}
			if !tt.wantParent && containsParent {
				t.Errorf("Expected YAML to NOT contain parent field, got:\n%s", yamlStr)
			}

			// For exact comparison tests
			if tt.expectedYAML != "" && yamlStr != tt.expectedYAML {
				t.Errorf("YAML mismatch\nExpected:\n%s\nGot:\n%s", tt.expectedYAML, yamlStr)
			}
		})
	}
}

func TestPromptMeta_UnmarshalYAML_WithParent(t *testing.T) {
	tests := []struct {
		name         string
		yamlContent  string
		expectedMeta PromptMeta
		wantError    bool
	}{
		{
			name: "unmarshal with parent field",
			yamlContent: `name: Test Prompt
author: john
category: test
tags:
  - test
  - example
version: "1.0.0"
parent: private-gist-456
`,
			expectedMeta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "john",
				Category: "test",
				Tags:     []string{"test", "example"},
				Version:  "1.0.0",
				Parent:   "private-gist-456",
			},
			wantError: false,
		},
		{
			name: "unmarshal without parent field",
			yamlContent: `name: Test Prompt
author: jane
category: utility
tags:
  - util
version: "2.0.0"
description: A utility prompt
`,
			expectedMeta: PromptMeta{
				Name:        "Test Prompt",
				Author:      "jane",
				Category:    "utility",
				Tags:        []string{"util"},
				Version:     "2.0.0",
				Description: "A utility prompt",
				Parent:      "", // Should be empty
			},
			wantError: false,
		},
		{
			name: "unmarshal with empty parent field",
			yamlContent: `name: Test Prompt
author: bob
category: example
tags: [demo]
parent: ""
`,
			expectedMeta: PromptMeta{
				Name:     "Test Prompt",
				Author:   "bob",
				Category: "example",
				Tags:     []string{"demo"},
				Parent:   "",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var meta PromptMeta
			err := yaml.Unmarshal([]byte(tt.yamlContent), &meta)
			
			if (err != nil) != tt.wantError {
				t.Errorf("Unmarshal() error = %v, wantError %v", err, tt.wantError)
			}
			
			if !tt.wantError {
				// Compare fields
				if meta.Name != tt.expectedMeta.Name {
					t.Errorf("Name mismatch: got %s, want %s", meta.Name, tt.expectedMeta.Name)
				}
				if meta.Author != tt.expectedMeta.Author {
					t.Errorf("Author mismatch: got %s, want %s", meta.Author, tt.expectedMeta.Author)
				}
				if meta.Category != tt.expectedMeta.Category {
					t.Errorf("Category mismatch: got %s, want %s", meta.Category, tt.expectedMeta.Category)
				}
				if meta.Version != tt.expectedMeta.Version {
					t.Errorf("Version mismatch: got %s, want %s", meta.Version, tt.expectedMeta.Version)
				}
				if meta.Description != tt.expectedMeta.Description {
					t.Errorf("Description mismatch: got %s, want %s", meta.Description, tt.expectedMeta.Description)
				}
				if meta.Parent != tt.expectedMeta.Parent {
					t.Errorf("Parent mismatch: got %s, want %s", meta.Parent, tt.expectedMeta.Parent)
				}
				if len(meta.Tags) != len(tt.expectedMeta.Tags) {
					t.Errorf("Tags length mismatch: got %d, want %d", len(meta.Tags), len(tt.expectedMeta.Tags))
				}
			}
		})
	}
}

func TestPromptMeta_BackwardCompatibility(t *testing.T) {
	// Test that existing YAML files without parent field can still be parsed
	oldYAML := `name: Legacy Prompt
author: olduser
category: legacy
tags:
  - old
  - legacy
version: "0.1.0"
description: This is an old prompt without parent field
`

	var meta PromptMeta
	err := yaml.Unmarshal([]byte(oldYAML), &meta)
	if err != nil {
		t.Fatalf("Failed to unmarshal legacy YAML: %v", err)
	}

	// Verify all fields are parsed correctly
	if meta.Name != "Legacy Prompt" {
		t.Errorf("Expected name to be 'Legacy Prompt', got '%s'", meta.Name)
	}
	if meta.Author != "olduser" {
		t.Errorf("Expected author to be 'olduser', got '%s'", meta.Author)
	}
	if meta.Parent != "" {
		t.Errorf("Expected parent to be empty for legacy prompt, got '%s'", meta.Parent)
	}

	// Verify it can be marshaled back
	newYAML, err := yaml.Marshal(&meta)
	if err != nil {
		t.Fatalf("Failed to marshal legacy prompt: %v", err)
	}

	// Parent field should not appear in the output since it's empty
	if contains(string(newYAML), "parent:") {
		t.Error("Empty parent field should not appear in marshaled YAML")
	}
}

func TestPromptMeta_ParentFieldPreservation(t *testing.T) {
	// Test that parent field is preserved through read/write operations
	original := PromptMeta{
		Name:     "Test Prompt",
		Author:   "john",
		Category: "test",
		Tags:     []string{"test"},
		Parent:   "original-parent-123",
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var restored PromptMeta
	err = yaml.Unmarshal(yamlData, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify parent field is preserved
	if restored.Parent != original.Parent {
		t.Errorf("Parent field not preserved: got %s, want %s", restored.Parent, original.Parent)
	}

	// Verify all other fields are also preserved
	if restored.Name != original.Name || restored.Author != original.Author {
		t.Error("Other fields were not preserved correctly")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || index(s, substr) != -1)
}

// Simple index function to find substring
func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}