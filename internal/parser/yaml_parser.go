package parser

import (
	"strings"

	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/models"
	"gopkg.in/yaml.v3"
)

// YAMLParser provides configurable parsing for YAML frontmatter
type YAMLParser struct {
	strict         bool   // Enable strict validation
	requireVersion bool   // Require version field
	defaultAuthor  string // Default author if not specified
}

// YAMLParserConfig configures the parser behavior
type YAMLParserConfig struct {
	Strict         bool
	RequireVersion bool
	DefaultAuthor  string
}

// NewYAMLParser creates a parser with the given configuration
func NewYAMLParser(config YAMLParserConfig) *YAMLParser {
	return &YAMLParser{
		strict:         config.Strict,
		requireVersion: config.RequireVersion,
		defaultAuthor:  config.DefaultAuthor,
	}
}

// ParseFrontMatter extracts metadata and content from a prompt file
func (p *YAMLParser) ParseFrontMatter(content string) (*models.PromptMeta, string, error) {
	// Check if content starts with front matter delimiter
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		if p.strict {
			return nil, "", errors.NewParsingErrorMsg("ParseFrontMatter", "missing YAML front matter")
		}
		// In lenient mode, return empty metadata
		return &models.PromptMeta{}, content, nil
	}

	// Remove the opening delimiter
	content = strings.TrimPrefix(content, "---\n")
	content = strings.TrimPrefix(content, "---\r\n")

	// Find the closing delimiter
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
			if p.strict {
				return nil, "", errors.NewParsingErrorMsg("ParseFrontMatter", "unclosed YAML front matter")
			}
			// In lenient mode, treat everything as front matter
			frontMatter = content
			promptContent = ""
		}
	} else {
		// Determine which line ending style is used
		if endIndexWin != -1 && (endIndex == -1 || endIndexWin < endIndex) {
			frontMatter = content[:endIndexWin]
			promptContent = content[endIndexWin+6:] // Skip "\r\n---" and look for next content
			// Skip the remaining "\r\n" if present
			if strings.HasPrefix(promptContent, "\r\n") {
				promptContent = promptContent[2:]
			}
		} else {
			frontMatter = content[:endIndex]
			promptContent = content[endIndex+4:] // Skip "\n---" and look for next content
			// Skip the remaining "\n" if present
			if strings.HasPrefix(promptContent, "\n") {
				promptContent = promptContent[1:]
			}
		}
	}

	// Parse YAML front matter
	var meta models.PromptMeta
	if err := yaml.Unmarshal([]byte(frontMatter), &meta); err != nil {
		return nil, "", errors.WrapWithMessage(err, "failed to parse YAML front matter")
	}

	// Apply defaults if configured
	if p.defaultAuthor != "" && meta.Author == "" {
		meta.Author = p.defaultAuthor
	}

	// Validate metadata according to parser configuration
	if err := p.ValidateMetadata(&meta); err != nil {
		return nil, "", err
	}

	// Trim any leading/trailing whitespace from content
	promptContent = strings.TrimSpace(promptContent)

	return &meta, promptContent, nil
}

// ParsePromptFile parses a complete prompt file into a Prompt object
func (p *YAMLParser) ParsePromptFile(content string) (*models.Prompt, error) {
	meta, promptContent, err := p.ParseFrontMatter(content)
	if err != nil {
		return nil, err
	}

	prompt := &models.Prompt{
		PromptMeta: *meta,
		Content:    promptContent,
	}

	return prompt, nil
}

// ValidateMetadata validates metadata according to parser configuration
func (p *YAMLParser) ValidateMetadata(meta *models.PromptMeta) error {
	if p.strict {
		// Use the existing validation method in strict mode
		if err := meta.Validate(); err != nil {
			return err
		}
	} else {
		// In lenient mode, only check critical fields
		if meta.Name == "" {
			return errors.NewValidationErrorMsg("ValidateMetadata", "name is required")
		}
	}

	// Check version requirement if configured
	if p.requireVersion && meta.Version == "" {
		return errors.NewValidationErrorMsg("ValidateMetadata", "version is required")
	}

	return nil
}
