package parser

import (
	"fmt"
	"strings"

	"github.com/grigri201/prompt-vault/internal/models"
	"gopkg.in/yaml.v3"
)

// ParseYAMLFrontMatter parses the YAML front matter from a prompt file
// Returns the metadata, content, and any error
func ParseYAMLFrontMatter(content string) (*models.PromptMeta, string, error) {
	// Check if content starts with front matter delimiter
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, "", fmt.Errorf("missing YAML front matter")
	}

	// Find the closing delimiter
	// Handle both Unix and Windows line endings
	content = strings.TrimPrefix(content, "---\n")
	content = strings.TrimPrefix(content, "---\r\n")
	
	// Find the end of front matter
	endIndex := strings.Index(content, "\n---\n")
	endIndexWin := strings.Index(content, "\r\n---\r\n")
	
	var frontMatter string
	var promptContent string
	
	if endIndex == -1 && endIndexWin == -1 {
		// Check if the entire content is just front matter
		if strings.HasSuffix(content, "\n---") || strings.HasSuffix(content, "\r\n---") {
			// Remove the trailing delimiter
			frontMatter = strings.TrimSuffix(content, "\n---")
			frontMatter = strings.TrimSuffix(frontMatter, "\r\n---")
			promptContent = ""
		} else {
			return nil, "", fmt.Errorf("unclosed YAML front matter")
		}
	} else {
		// Determine which line ending style is used
		if endIndexWin != -1 && (endIndex == -1 || endIndexWin < endIndex) {
			frontMatter = content[:endIndexWin]
			promptContent = content[endIndexWin+8:] // Skip "\r\n---\r\n"
		} else {
			frontMatter = content[:endIndex]
			promptContent = content[endIndex+5:] // Skip "\n---\n"
		}
	}

	// Parse YAML front matter
	var meta models.PromptMeta
	if err := yaml.Unmarshal([]byte(frontMatter), &meta); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML front matter: %w", err)
	}

	// Validate required fields
	if err := meta.Validate(); err != nil {
		return nil, "", fmt.Errorf("invalid front matter: %w", err)
	}

	// Trim any leading/trailing whitespace from content
	promptContent = strings.TrimSpace(promptContent)

	return &meta, promptContent, nil
}

// ParsePromptFile parses a complete prompt file and returns a Prompt object
func ParsePromptFile(content string) (*models.Prompt, error) {
	meta, promptContent, err := ParseYAMLFrontMatter(content)
	if err != nil {
		return nil, err
	}

	prompt := &models.Prompt{
		PromptMeta: *meta,
		Content:    promptContent,
	}

	return prompt, nil
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