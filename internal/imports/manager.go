package imports

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/models"
	"gopkg.in/yaml.v3"
)

// GistClient defines the interface for gist operations
type GistClient interface {
	GetGist(ctx context.Context, gistID string) (*github.Gist, error)
	GetGistByURL(ctx context.Context, gistURL string) (*github.Gist, error)
	ExtractGistID(gistURL string) (string, error)
}

// UI defines the interface for user interactions
type UI interface {
	Confirm(message string) (bool, error)
}

// Manager handles importing prompts
type Manager struct {
	gistClient GistClient
	ui         UI
}

// ImportResult contains the result of importing a prompt
type ImportResult struct {
	GistID     string
	IsUpdate   bool
	OldVersion string
	NewVersion string
}

// NewManager creates a new import manager
func NewManager(gistClient GistClient, ui UI) *Manager {
	return &Manager{
		gistClient: gistClient,
		ui:         ui,
	}
}

// ImportPrompt imports a prompt from a gist URL
func (m *Manager) ImportPrompt(ctx context.Context, gistURL string, index *models.Index) (*ImportResult, error) {
	// Extract gist ID from URL
	gistID, err := m.extractGistID(gistURL)
	if err != nil {
		return nil, errors.WrapError("Manager.ImportPrompt", err)
	}

	// Get the gist
	gist, err := m.gistClient.GetGistByURL(ctx, gistURL)
	if err != nil {
		return nil, errors.WrapError("Manager.ImportPrompt", err)
	}

	// Check if it's a public gist
	if gist.Public != nil && !*gist.Public {
		return nil, errors.NewValidationErrorMsg("Manager.ImportPrompt", "cannot import private gist")
	}

	// Validate the gist contains a valid prompt
	prompt, err := m.validatePromptGist(gist)
	if err != nil {
		return nil, errors.WrapError("Manager.ImportPrompt", err)
	}

	// Create index entry for the prompt
	newEntry := models.IndexEntry{
		GistID:      gistID,
		Name:        prompt.Name,
		Author:      prompt.Author,
		Tags:        prompt.Tags,
		Version:     prompt.Version,
		Description: prompt.Description,
		Parent:      prompt.Parent,
	}

	// Check if already imported
	existingEntry, exists := m.checkExistingImport(index, gistID)

	result := &ImportResult{
		GistID: gistID,
	}

	if exists {
		// Check if versions differ
		if existingEntry.Version == newEntry.Version {
			// Same version, no update needed
			return result, nil
		}

		// Different versions, ask user to confirm update
		confirmed, err := m.confirmVersionUpdate(existingEntry, &newEntry)
		if err != nil {
			return nil, errors.WrapError("Manager.ImportPrompt", err)
		}

		if !confirmed {
			return nil, errors.NewValidationErrorMsg("Manager.ImportPrompt", "update cancelled by user")
		}

		// Update the existing entry
		result.IsUpdate = true
		result.OldVersion = existingEntry.Version
		result.NewVersion = newEntry.Version

		// Update in index
		index.UpdateImportedEntry(newEntry)
	} else {
		// New import
		result.NewVersion = newEntry.Version

		// Add to index
		index.AddImportedEntry(newEntry)
	}

	return result, nil
}

// extractGistID extracts the gist ID from a GitHub gist URL
func (m *Manager) extractGistID(gistURL string) (string, error) {
	// Parse the URL
	u, err := url.Parse(gistURL)
	if err != nil {
		return "", errors.NewValidationError("Manager.extractGistID", err)
	}

	// Check if it's a GitHub gist URL
	if u.Host != "gist.github.com" {
		if u.Host == "" {
			return "", errors.NewValidationErrorMsg("Manager.extractGistID", "invalid URL")
		}
		return "", errors.NewValidationErrorMsg("Manager.extractGistID", "not a GitHub gist URL")
	}

	// Extract path components
	// Format: /username/gistid[/revision][#anchor]
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return "", errors.NewValidationErrorMsg("Manager.extractGistID", "invalid gist URL format")
	}

	// The gist ID is the second part
	gistID := parts[1]

	// Validate gist ID (should be alphanumeric)
	if gistID == "" || !isValidGistID(gistID) {
		return "", errors.NewValidationErrorMsg("Manager.extractGistID", "invalid gist ID")
	}

	return gistID, nil
}

// validatePromptGist validates that a gist contains a valid prompt
func (m *Manager) validatePromptGist(gist *github.Gist) (*models.Prompt, error) {
	// Check if gist has files
	if len(gist.Files) == 0 {
		return nil, errors.NewValidationErrorMsg("Manager.validatePromptGist", "gist has no files")
	}

	// Get the first file content
	var content string
	for _, file := range gist.Files {
		if file.Content != nil {
			content = *file.Content
			break
		}
	}

	if content == "" {
		return nil, errors.NewValidationErrorMsg("Manager.validatePromptGist", "no content found in gist")
	}

	// Parse YAML front matter
	meta, promptContent, err := parseYAMLFrontMatter(content)
	if err != nil {
		return nil, errors.WrapError("Manager.validatePromptGist", err)
	}

	// Validate required fields
	if err := meta.Validate(); err != nil {
		return nil, errors.WrapError("Manager.validatePromptGist", err)
	}

	// Create prompt
	prompt := &models.Prompt{
		PromptMeta: *meta,
		GistID:     *gist.ID,
		Content:    promptContent,
	}

	if gist.HTMLURL != nil {
		prompt.GistURL = *gist.HTMLURL
	}

	return prompt, nil
}

// checkExistingImport checks if a gist is already imported
func (m *Manager) checkExistingImport(index *models.Index, gistID string) (*models.IndexEntry, bool) {
	if index == nil || index.ImportedEntries == nil {
		return nil, false
	}

	for i := range index.ImportedEntries {
		if index.ImportedEntries[i].GistID == gistID {
			// Return a copy of the entry
			entry := index.ImportedEntries[i]
			return &entry, true
		}
	}

	return nil, false
}

// confirmVersionUpdate asks user to confirm version update
func (m *Manager) confirmVersionUpdate(oldEntry, newEntry *models.IndexEntry) (bool, error) {
	if m.ui == nil {
		return false, errors.NewValidationErrorMsg("Manager.confirmVersionUpdate", "UI not configured")
	}

	message := fmt.Sprintf(
		"Prompt '%s' already exists with version %s. Update to version %s?",
		oldEntry.Name,
		oldEntry.Version,
		newEntry.Version,
	)

	return m.ui.Confirm(message)
}

// parseYAMLFrontMatter parses YAML front matter from content
func parseYAMLFrontMatter(content string) (*models.PromptMeta, string, error) {
	// Check if content starts with front matter delimiter
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, "", errors.NewParsingErrorMsg("parseYAMLFrontMatter", "missing YAML front matter")
	}

	// Find the closing delimiter
	content = strings.TrimPrefix(content, "---\n")
	content = strings.TrimPrefix(content, "---\r\n")

	endIndex := strings.Index(content, "\n---")
	if endIndex == -1 {
		endIndex = strings.Index(content, "\r\n---")
		if endIndex == -1 {
			return nil, "", errors.NewParsingErrorMsg("parseYAMLFrontMatter", "unclosed YAML front matter")
		}
	}

	frontMatter := content[:endIndex]
	promptContent := content[endIndex+4:] // Skip "\n---"
	if strings.HasPrefix(promptContent, "\n") {
		promptContent = promptContent[1:]
	}

	// Parse YAML
	var meta models.PromptMeta
	if err := yaml.Unmarshal([]byte(frontMatter), &meta); err != nil {
		return nil, "", errors.NewParsingError("parseYAMLFrontMatter", err)
	}

	return &meta, strings.TrimSpace(promptContent), nil
}

// isValidGistID checks if a string is a valid gist ID
func isValidGistID(id string) bool {
	// GitHub gist IDs are alphanumeric strings (can contain letters and numbers)
	if len(id) == 0 {
		return false
	}
	for _, ch := range id {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			return false
		}
	}
	return true
}
