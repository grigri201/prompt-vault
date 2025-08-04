package service

import "github.com/grigri/pv/internal/model"

// PromptService defines the interface for prompt business logic operations
type PromptService interface {
	// AddFromFile adds a prompt from a YAML file to the vault
	AddFromFile(filePath string) (*model.Prompt, error)
}