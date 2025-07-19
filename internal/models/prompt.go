package models

import (
	"fmt"
	"time"
)

// PromptMeta contains the metadata for a prompt template
type PromptMeta struct {
	Name        string   `yaml:"name" json:"name"`
	Author      string   `yaml:"author" json:"author"`
	Category    string   `yaml:"category" json:"category"`
	Tags        []string `yaml:"tags" json:"tags"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
}

// Validate checks if all required fields are present and valid
func (m *PromptMeta) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Author == "" {
		return fmt.Errorf("author is required")
	}
	if m.Category == "" {
		return fmt.Errorf("category is required")
	}
	if len(m.Tags) == 0 {
		return fmt.Errorf("at least one tag is required")
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
	UpdatedAt   time.Time `json:"updated_at"`
}