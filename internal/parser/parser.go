package parser

import (
	"fmt"
	"strings"

	"github.com/grigri201/prompt-vault/internal/models"
	"gopkg.in/yaml.v3"
)

// ParseYAMLFrontMatter parses the YAML front matter from a prompt file
// Returns the metadata, content, and any error
// This function maintains backward compatibility while using the new YAMLParser internally
func ParseYAMLFrontMatter(content string) (*models.PromptMeta, string, error) {
	// Create a parser with strict mode to maintain backward compatibility
	parser := NewYAMLParser(YAMLParserConfig{
		Strict: true,
	})

	return parser.ParseFrontMatter(content)
}

// ParsePromptFile parses a complete prompt file and returns a Prompt object
// This function maintains backward compatibility while using the new YAMLParser internally
func ParsePromptFile(content string) (*models.Prompt, error) {
	// Create a parser with strict mode to maintain backward compatibility
	parser := NewYAMLParser(YAMLParserConfig{
		Strict: true,
	})

	return parser.ParsePromptFile(content)
}

// ExtractVariables extracts all unique variable names from a template content
// Variables are defined as text within curly braces: {variable}
// Returns a slice of unique variable names in the order they first appear
func ExtractVariables(content string) []string {
	// Use a map to track unique variables and a slice to maintain order
	seen := make(map[string]bool)
	var variables []string

	i := 0
	for i < len(content) {
		// Find the start of a potential variable
		start := strings.IndexByte(content[i:], '{')
		if start == -1 {
			break
		}
		start += i

		// Check if this is a nested brace (previous char is also '{')
		if start > 0 && content[start-1] == '{' {
			// This is part of {{ pattern, skip it
			i = start + 1
			continue
		}

		// Find the matching closing brace
		end := strings.IndexByte(content[start+1:], '}')
		if end == -1 {
			// No closing brace found, move past this opening brace
			i = start + 1
			continue
		}
		end += start + 1

		// Check if the closing brace is followed by another '}'
		if end+1 < len(content) && content[end+1] == '}' {
			// This is part of }} pattern, skip it
			i = start + 1
			continue
		}

		// Extract the content between braces
		varContent := content[start+1 : end]

		// Check if there are any braces in the content
		if strings.ContainsAny(varContent, "{}") {
			// Skip variables with braces in the content
			i = start + 1
			continue
		}

		// Validate the variable name
		if varContent != "" && isValidVariableName(varContent) {
			if !seen[varContent] {
				seen[varContent] = true
				variables = append(variables, varContent)
			}
		}

		// Move past this variable
		i = end + 1
	}

	return variables
}

// isValidVariableName checks if a string is a valid variable name
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	// Check if all characters are valid (letters, numbers, underscores, hyphens)
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_' || ch == '-') {
			return false
		}
	}

	return true
}

// FillVariables replaces all variable placeholders in the content with their corresponding values
// Variables not found in the values map are left unchanged
// Returns a new string with replacements made
func FillVariables(content string, values map[string]string) string {
	if values == nil {
		return content
	}

	result := content

	// Replace each variable with its value
	for varName, value := range values {
		placeholder := "{" + varName + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// FormatPromptFile formats a prompt into YAML front matter format
func FormatPromptFile(meta *models.PromptMeta, content string) string {
	// Create metadata map
	metaMap := map[string]interface{}{
		"name":   meta.Name,
		"author": meta.Author,
		"tags":   meta.Tags,
	}

	// Add optional fields
	if meta.Version != "" {
		metaMap["version"] = meta.Version
	}
	if meta.Description != "" {
		metaMap["description"] = meta.Description
	}

	// Marshal to YAML
	metaYAML, err := yaml.Marshal(metaMap)
	if err != nil {
		// Fallback to simple format
		return fmt.Sprintf("---\nname: %s\nauthor: %s\ntags: %v\n---\n%s",
			meta.Name, meta.Author, meta.Tags, content)
	}

	return fmt.Sprintf("---\n%s---\n%s", string(metaYAML), content)
}
