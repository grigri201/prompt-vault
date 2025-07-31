package share

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
	"gopkg.in/yaml.v3"
)

// GistClient defines the interface for gist operations
type GistClient interface {
	GetGist(ctx context.Context, gistID string) (*github.Gist, error)
	CreatePublicGist(ctx context.Context, gistName, description, content string) (string, string, error)
	UpdateGist(ctx context.Context, gistID, gistName, description, content string) (string, error)
	ListUserGists(ctx context.Context, username string) ([]*github.Gist, error)
}

// UI defines the interface for user interactions
type UI interface {
	Confirm(message string) (bool, error)
}

// Manager handles sharing prompts
type Manager struct {
	gistClient GistClient
	ui         UI
	username   string
}

// ShareResult contains the result of sharing a prompt
type ShareResult struct {
	PublicGistID  string
	PublicGistURL string
	IsUpdate      bool
}

// NewManager creates a new share manager
func NewManager(gistClient GistClient, ui UI, username string) *Manager {
	return &Manager{
		gistClient: gistClient,
		ui:         ui,
		username:   username,
	}
}

// SharePrompt shares a private prompt as a public gist
func (m *Manager) SharePrompt(ctx context.Context, privateGistID string) (*ShareResult, error) {
	// Get the private gist
	gist, err := m.gistClient.GetGist(ctx, privateGistID)
	if err != nil {
		return nil, errors.WrapError("SharePrompt", err)
	}

	// Validate it's a private gist
	if gist.Public != nil && *gist.Public {
		return nil, errors.NewValidationErrorMsg("SharePrompt", fmt.Sprintf("cannot share: gist %s is already public", privateGistID))
	}

	// Extract prompt content from gist
	var promptContent string
	var promptFile string
	for filename, file := range gist.Files {
		if file.Content != nil {
			promptContent = *file.Content
			promptFile = string(filename)
			break
		}
	}

	if promptContent == "" {
		return nil, errors.NewValidationErrorMsg("SharePrompt", "no content found in gist")
	}

	// Parse the prompt
	prompt, err := parsePromptFromGist(privateGistID, promptFile, promptContent)
	if err != nil {
		return nil, errors.WrapError("SharePrompt", err)
	}

	// Check if a public version already exists
	existingPublicID, err := m.findExistingPublicGist(ctx, privateGistID)
	if err != nil {
		return nil, errors.WrapError("SharePrompt", err)
	}

	// If public version exists, ask user if they want to update it
	if existingPublicID != "" {
		message := fmt.Sprintf("A public version of this prompt already exists (ID: %s). Do you want to update it?", existingPublicID)
		confirmed, err := m.ui.Confirm(message)
		if err != nil {
			return nil, errors.WrapError("SharePrompt", err)
		}
		if !confirmed {
			return nil, errors.NewValidationErrorMsg("SharePrompt", "update cancelled by user")
		}
		// Update existing public gist
		return m.updatePublicGist(ctx, existingPublicID, prompt)
	}

	// Create new public gist
	return m.createPublicGist(ctx, prompt)
}

// parsePromptFromGist parses prompt content from a gist
func parsePromptFromGist(gistID, filename, content string) (*models.Prompt, error) {
	// Create lenient parser for share operations
	yamlParser := parser.NewYAMLParser(parser.YAMLParserConfig{
		Strict: false, // Be lenient for parsing shared prompts
	})

	// Parse the prompt file
	prompt, err := yamlParser.ParsePromptFile(content)
	if err != nil {
		return nil, err
	}

	// Set the gist ID
	prompt.GistID = gistID

	return prompt, nil
}

// extractParentFromContent extracts the parent field from YAML front matter without full validation
func extractParentFromContent(content string) (string, error) {
	// Check if content starts with front matter delimiter
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return "", errors.NewParsingErrorMsg("extractParentFromContent", "no YAML front matter")
	}

	// Find the closing delimiter
	content = strings.TrimPrefix(content, "---\n")
	content = strings.TrimPrefix(content, "---\r\n")

	endIndex := strings.Index(content, "\n---")
	if endIndex == -1 {
		endIndex = strings.Index(content, "\r\n---")
		if endIndex == -1 {
			return "", errors.NewParsingErrorMsg("extractParentFromContent", "unclosed YAML front matter")
		}
	}

	frontMatter := content[:endIndex]

	// Parse YAML just to get the parent field
	var meta struct {
		Parent string `yaml:"parent"`
	}
	if err := yaml.Unmarshal([]byte(frontMatter), &meta); err != nil {
		return "", err
	}

	return meta.Parent, nil
}

// findExistingPublicGist searches for an existing public version of the prompt
func (m *Manager) findExistingPublicGist(ctx context.Context, parentID string) (string, error) {
	// List all user's gists
	gists, err := m.gistClient.ListUserGists(ctx, m.username)
	if err != nil {
		return "", errors.WrapError("findExistingPublicGist", err)
	}

	// Search for a public gist with matching parent field
	for _, gist := range gists {
		// Skip if not public
		if gist.Public == nil || !*gist.Public {
			continue
		}

		// Check each file for YAML metadata with parent field
		for _, file := range gist.Files {
			if file.Content == nil {
				continue
			}

			content := *file.Content
			// Extract parent field without full validation
			parent, err := extractParentFromContent(content)
			if err != nil {
				// Skip files that don't have valid front matter
				continue
			}

			// Check if parent matches
			if parent == parentID {
				if gist.ID == nil {
					continue
				}
				return *gist.ID, nil
			}
		}
	}

	// No matching public gist found
	return "", nil
}

// createPublicGist creates a new public gist from a prompt
func (m *Manager) createPublicGist(ctx context.Context, prompt *models.Prompt) (*ShareResult, error) {
	// Clone the metadata and add parent field
	meta := prompt.PromptMeta
	meta.Parent = prompt.GistID

	// Format the content with updated metadata
	fullContent := formatPromptWithParent(&meta, prompt.Content)

	// Create gist name from prompt name
	gistName := strings.ReplaceAll(strings.ToLower(meta.Name), " ", "-")

	// Use description from prompt or create a default one
	description := meta.Description
	if description == "" {
		description = meta.Name
	}

	// Create the public gist
	gistID, gistURL, err := m.gistClient.CreatePublicGist(ctx, gistName, description, fullContent)
	if err != nil {
		return nil, errors.WrapError("createPublicGist", err)
	}

	return &ShareResult{
		PublicGistID:  gistID,
		PublicGistURL: gistURL,
		IsUpdate:      false,
	}, nil
}

// formatPromptWithParent formats a prompt with parent field included
func formatPromptWithParent(meta *models.PromptMeta, content string) string {
	// Create metadata map with all fields including parent
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
	if meta.Parent != "" {
		metaMap["parent"] = meta.Parent
	}

	// Marshal to YAML
	metaYAML, err := yaml.Marshal(metaMap)
	if err != nil {
		// Fallback to simple format
		parts := []string{
			"---",
			fmt.Sprintf("name: %s", meta.Name),
			fmt.Sprintf("author: %s", meta.Author),
			fmt.Sprintf("tags: %v", meta.Tags),
		}
		if meta.Version != "" {
			parts = append(parts, fmt.Sprintf("version: %q", meta.Version))
		}
		if meta.Description != "" {
			parts = append(parts, fmt.Sprintf("description: %s", meta.Description))
		}
		if meta.Parent != "" {
			parts = append(parts, fmt.Sprintf("parent: %s", meta.Parent))
		}
		parts = append(parts, "---", content)
		return strings.Join(parts, "\n")
	}

	return fmt.Sprintf("---\n%s---\n%s", string(metaYAML), content)
}

// updatePublicGist updates an existing public gist with new content
func (m *Manager) updatePublicGist(ctx context.Context, gistID string, prompt *models.Prompt) (*ShareResult, error) {
	// Clone the metadata and ensure parent field points to private gist
	meta := prompt.PromptMeta
	meta.Parent = prompt.GistID

	// Format the content with updated metadata
	fullContent := formatPromptWithParent(&meta, prompt.Content)

	// Create gist name from prompt name
	gistName := strings.ReplaceAll(strings.ToLower(meta.Name), " ", "-")

	// Use description from prompt or create a default one
	description := meta.Description
	if description == "" {
		description = meta.Name
	}

	// Update the public gist
	gistURL, err := m.gistClient.UpdateGist(ctx, gistID, gistName, description, fullContent)
	if err != nil {
		return nil, errors.WrapError("updatePublicGist", err)
	}

	return &ShareResult{
		PublicGistID:  gistID,
		PublicGistURL: gistURL,
		IsUpdate:      true,
	}, nil
}
