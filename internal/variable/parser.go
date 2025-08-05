package variable

import (
	"regexp"
	"sort"
)

// Parser defines the interface for variable parsing and replacement operations
// in prompt content. It handles the extraction and substitution of {variable}
// placeholders within prompt text for the get command functionality.
type Parser interface {
	// ExtractVariables extracts all unique variable names from the given content string.
	// Variables are identified by the {variable_name} syntax. Returns a sorted slice
	// of unique variable names found in the content, or an empty slice if no variables
	// are found. The returned slice does not include the curly braces.
	//
	// Example:
	//   content := "Hello {name}, your {role} is {role}"
	//   variables := parser.ExtractVariables(content)
	//   // returns: ["name", "role"]
	ExtractVariables(content string) []string

	// ReplaceVariables replaces all occurrences of {variable_name} in the content
	// with the corresponding values from the provided map. If a variable in the
	// content is not found in the values map, it remains unchanged in the output.
	//
	// Example:
	//   content := "Hello {name}, your role is {role}"
	//   values := map[string]string{"name": "Alice", "role": "developer"}
	//   result := parser.ReplaceVariables(content, values)
	//   // returns: "Hello Alice, your role is developer"
	ReplaceVariables(content string, values map[string]string) string

	// HasVariables checks whether the given content contains any variables
	// in the {variable_name} syntax. Returns true if at least one variable
	// is found, false otherwise. This is useful for determining whether
	// to show the variable input form to users.
	//
	// Example:
	//   hasVars := parser.HasVariables("Hello {name}!")  // returns: true
	//   hasVars := parser.HasVariables("Hello world!")   // returns: false
	HasVariables(content string) bool
}

// parser is the concrete implementation of the Parser interface.
// It uses regular expressions to identify and manipulate variable placeholders
// in prompt content.
type parser struct {
	// variableRegex is the compiled regular expression for matching {variable} patterns
	variableRegex *regexp.Regexp
}

// NewParser creates a new instance of the Parser interface.
// It initializes the regular expression pattern for variable matching.
func NewParser() Parser {
	return &parser{
		variableRegex: regexp.MustCompile(`\{([^}]+)\}`),
	}
}

// ExtractVariables extracts all unique variable names from the given content string.
// It uses regex pattern \{([^}]+)\} to find {variable} patterns, deduplicates the results,
// and returns them in sorted order. Empty content or content without variables returns
// an empty slice.
func (p *parser) ExtractVariables(content string) []string {
	if content == "" {
		return []string{}
	}

	// Find all matches using the compiled regex
	matches := p.variableRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return []string{}
	}

	// Use a map to deduplicate variable names
	variables := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			variables[match[1]] = true
		}
	}

	// Convert map keys to sorted slice
	result := make([]string, 0, len(variables))
	for variable := range variables {
		result = append(result, variable)
	}
	sort.Strings(result)

	return result
}

// ReplaceVariables replaces all occurrences of {variable_name} in the content
// with the corresponding values from the provided map. If a variable in the content
// is not found in the values map, it remains unchanged in the output. Handles
// multiple occurrences of the same variable by replacing all instances.
func (p *parser) ReplaceVariables(content string, values map[string]string) string {
	if content == "" {
		return content
	}

	if values == nil {
		return content
	}

	// Use ReplaceAllStringFunc to replace all matches
	return p.variableRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name from match (remove { and })
		variableName := match[1 : len(match)-1]

		// Look up replacement value
		if replacement, exists := values[variableName]; exists {
			return replacement
		}

		// If variable not found in values map, leave unchanged
		return match
	})
}

// HasVariables checks whether the given content contains any variables
// in the {variable_name} syntax. Returns true if at least one variable
// is found, false otherwise. This method is optimized for detection
// and can return early on the first match, making it more efficient
// than extracting all variables when only presence detection is needed.
func (p *parser) HasVariables(content string) bool {
	// Handle empty content - no variables possible
	if content == "" {
		return false
	}

	// Use FindStringSubmatch which respects the capture group requirement
	// The regex \{([^}]+)\} requires at least one non-} character inside braces
	// This ensures consistency with ExtractVariables behavior
	match := p.variableRegex.FindStringSubmatch(content)
	return len(match) > 1 && match[1] != ""
}
