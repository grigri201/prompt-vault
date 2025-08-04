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

	// ListForDeletion retrieves all prompts available for deletion.
	// This method is used to display the complete list of prompts to users
	// in interactive deletion modes. Returns an error if the listing fails.
	ListForDeletion() ([]model.Prompt, error)

	// FilterForDeletion retrieves prompts that match the given keyword for deletion.
	// The keyword is used to filter prompts by name, author, or description.
	// Returns an empty slice if no prompts match the keyword, or an error if filtering fails.
	FilterForDeletion(keyword string) ([]model.Prompt, error)
}
