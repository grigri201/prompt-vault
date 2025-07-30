package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/grigri201/prompt-vault/internal/errors"
)

// PromptMeta contains the metadata for a prompt template
type PromptMeta struct {
	Name        string   `yaml:"name" json:"name"`
	Author      string   `yaml:"author" json:"author"`
	Category    string   `yaml:"category" json:"category"`
	Tags        []string `yaml:"tags" json:"tags"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Parent      string   `yaml:"parent,omitempty" json:"parent,omitempty"`
	ID          string   `yaml:"id,omitempty" json:"id,omitempty"`
}

// ValidateID checks if the ID contains only valid characters
func (m *PromptMeta) ValidateID() error {
	if m.ID == "" {
		return nil // ID is optional
	}
	
	// Check minimum and maximum length first
	if len(m.ID) < 3 {
		return errors.NewValidationErrorMsg("PromptMeta.ValidateID", "ID must be at least 3 characters long")
	}
	if len(m.ID) > 100 {
		return errors.NewValidationErrorMsg("PromptMeta.ValidateID", "ID must not exceed 100 characters")
	}
	
	// ID can only contain alphanumeric characters, hyphens, and underscores
	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validID.MatchString(m.ID) {
		return errors.NewValidationErrorMsg("PromptMeta.ValidateID", 
			"ID can only contain letters, numbers, hyphens, and underscores")
	}
	
	return nil
}

// Validate checks if all required fields are present and valid
func (m *PromptMeta) Validate() error {
	if m.Name == "" {
		return errors.NewValidationErrorMsg("PromptMeta.Validate", "name is required")
	}
	if m.Author == "" {
		return errors.NewValidationErrorMsg("PromptMeta.Validate", "author is required")
	}
	if m.Category == "" {
		return errors.NewValidationErrorMsg("PromptMeta.Validate", "category is required")
	}
	if len(m.Tags) == 0 {
		return errors.NewValidationErrorMsg("PromptMeta.Validate", "at least one tag is required")
	}
	
	// Validate ID if present
	if err := m.ValidateID(); err != nil {
		return err
	}
	
	return nil
}

// SetDefaultVersion sets the version to current timestamp if not provided
func (m *PromptMeta) SetDefaultVersion() {
	if m.Version == "" {
		m.Version = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
}

// Prompt represents a complete prompt with metadata and content
type Prompt struct {
	PromptMeta
	GistID    string    `json:"gist_id"`
	GistURL   string    `json:"gist_url"`
	UpdatedAt time.Time `json:"updated_at"`
	Content   string    `json:"-"` // Not serialized to index
}

// ToIndexEntry converts a Prompt to an IndexEntry for storage in the index
func (p *Prompt) ToIndexEntry() IndexEntry {
	return IndexEntry{
		GistID:      p.GistID,
		GistURL:     p.GistURL,
		Name:        p.Name,
		Author:      p.Author,
		Category:    p.Category,
		Tags:        p.Tags,
		Version:     p.Version,
		Description: p.Description,
		ID:          p.ID,
		UpdatedAt:   p.UpdatedAt,
	}
}

// IndexEntry represents a prompt entry in the index file
type IndexEntry struct {
	GistID      string    `json:"gist_id"`
	GistURL     string    `json:"gist_url"`
	Name        string    `json:"name"`
	Author      string    `json:"author"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	ID          string    `json:"id,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Index represents the complete index file with all prompt entries
type Index struct {
	Username        string       `json:"username"`
	Entries         []IndexEntry `json:"entries"`
	ImportedEntries []IndexEntry `json:"imported_entries"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// AddImportedEntry adds a new entry to ImportedEntries
func (idx *Index) AddImportedEntry(entry IndexEntry) {
	if idx.ImportedEntries == nil {
		idx.ImportedEntries = []IndexEntry{}
	}
	idx.ImportedEntries = append(idx.ImportedEntries, entry)
}

// UpdateImportedEntry updates an existing imported entry by GistID
// Returns true if the entry was found and updated, false otherwise
func (idx *Index) UpdateImportedEntry(entry IndexEntry) bool {
	for i, e := range idx.ImportedEntries {
		if e.GistID == entry.GistID {
			idx.ImportedEntries[i] = entry
			return true
		}
	}
	return false
}

// FindImportedEntry finds an imported entry by GistID
// Returns the entry and true if found, empty entry and false otherwise
func (idx *Index) FindImportedEntry(gistID string) (IndexEntry, bool) {
	for _, entry := range idx.ImportedEntries {
		if entry.GistID == gistID {
			return entry, true
		}
	}
	return IndexEntry{}, false
}
