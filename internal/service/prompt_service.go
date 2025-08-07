package service

import "github.com/grigri/pv/internal/model"

// GistInfo 包含 Gist 的基本信息
type GistInfo struct {
	ID          string
	URL         string
	IsPublic    bool
	HasAccess   bool
	Description string
	Owner       string
}

// PromptService defines the interface for prompt business logic operations
type PromptService interface {
	// AddFromFile adds a prompt from a YAML file to the vault
	AddFromFile(filePath string) (*model.Prompt, error)

	// DeleteByKeyword deletes prompts that match the given keyword.
	// The keyword is used to search across prompt names, authors, and descriptions.
	// Returns an error if no matching prompts are found or if the deletion fails.
	DeleteByKeyword(keyword string) error

	// DeleteByURL deletes a prompt identified by its GitHub Gist URL.
	// The gistURL should be a valid GitHub Gist URL (e.g., https://gist.github.com/user/gist_id).
	// Returns an error if the URL format is invalid, the prompt is not found, or deletion fails.
	DeleteByURL(gistURL string) error

	// ListPrompts retrieves all prompts available in the vault.
	// This method is used to display the complete list of prompts to users
	// in various interactive modes (list, delete, get). Returns an error if the listing fails.
	ListPrompts() ([]model.Prompt, error)

	// FilterPrompts retrieves prompts that match the given keyword.
	// The keyword is used to filter prompts by name, author, or description.
	// Returns an empty slice if no prompts match the keyword, or an error if filtering fails.
	FilterPrompts(keyword string) ([]model.Prompt, error)

	// GetPromptByURL retrieves a specific prompt by its GitHub Gist URL.
	// The gistURL should be a valid GitHub Gist URL (e.g., https://gist.github.com/user/gist_id).
	// Returns the prompt if found, or an error if the URL format is invalid or the prompt is not found.
	GetPromptByURL(gistURL string) (*model.Prompt, error)

	// GetPromptContent retrieves the actual content (YAML file content) of a prompt from GitHub Gist.
	// The prompt parameter should contain a valid ID that corresponds to a GitHub Gist ID.
	// Returns the raw content string of the prompt or an error if the content cannot be retrieved.
	GetPromptContent(prompt *model.Prompt) (string, error)

	// SharePrompt shares a private prompt by creating a public gist.
	// If the prompt has already been shared, updates the existing public gist.
	// Returns the shared prompt with the public gist URL, or an error if sharing fails.
	SharePrompt(prompt *model.Prompt) (*model.Prompt, error)

	// AddFromURL adds a prompt from a public GitHub Gist URL.
	// Validates that the gist is public and contains valid prompt format.
	// Returns the added prompt or an error if the URL is invalid or gist is not accessible.
	AddFromURL(gistURL string) (*model.Prompt, error)

	// ValidateGistAccess validates user access to a gist and returns its information.
	// Checks if the user has read/write access and whether the gist is public or private.
	// Returns gist information or an error if access validation fails.
	ValidateGistAccess(gistURL string) (*GistInfo, error)

	// ListPrivatePrompts retrieves all private prompts available in the vault.
	// Filters the complete prompt list to show only private gists.
	// Used in share command interactive mode. Returns an error if listing fails.
	ListPrivatePrompts() ([]model.Prompt, error)

	// FilterPrivatePrompts retrieves private prompts that match the given keyword.
	// Similar to FilterPrompts but only returns private gists matching the keyword.
	// Used in share command keyword filtering mode. Returns an error if filtering fails.
	// Sync synchronizes the local cache with GitHub, downloading the raw index.json file.
	// This ensures the local cache has the complete and up-to-date index from GitHub.
	// Returns an error if the synchronization fails.
	Sync() error

	FilterPrivatePrompts(keyword string) ([]model.Prompt, error)
}
