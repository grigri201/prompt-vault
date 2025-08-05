package service

import "github.com/grigri/pv/internal/model"

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
}
