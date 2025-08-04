package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	apperrors "github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/service"
)

type AddCmd *cobra.Command

type add struct {
	promptService service.PromptService
}

func (ac *add) execute(cmd *cobra.Command, args []string) {
	// Validate arguments - we need exactly one file path
	if len(args) != 1 {
		fmt.Println("❌ Error: Please provide exactly one file path")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  pv add <file-path>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  pv add my-prompt.yaml")
		return
	}

	filePath := args[0]

	// Call the prompt service to add the prompt from file
	prompt, err := ac.promptService.AddFromFile(filePath)
	if err != nil {
		// Handle different types of errors with user-friendly messages
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			switch appErr.Type {
			case apperrors.ErrValidation:
				fmt.Printf("❌ Validation Error: %s\n", appErr.Message)
				return
			case apperrors.ErrStorage:
				fmt.Printf("❌ Storage Error: %s\n", appErr.Message)
				return
			case apperrors.ErrAuth:
				fmt.Printf("❌ Authentication Error: %s\n", appErr.Message)
				fmt.Println()
				fmt.Println("Please run 'pv auth login' to authenticate with GitHub.")
				return
			case apperrors.ErrNetwork:
				fmt.Printf("❌ Network Error: %s\n", appErr.Message)
				fmt.Println()
				fmt.Println("Please check your internet connection and try again.")
				return
			default:
				fmt.Printf("❌ Error: %s\n", appErr.Message)
				return
			}
		}

		// For other unexpected errors, show the original error and exit
		log.Fatalf("unexpected error adding prompt: %v", err)
	}

	// Display success message
	fmt.Println("✅ Prompt added successfully!")
	fmt.Println()
	fmt.Printf("  Name: %s\n", prompt.Name)
	fmt.Printf("  Author: %s\n", prompt.Author)
	if prompt.Description != "" {
		fmt.Printf("  Description: %s\n", prompt.Description)
	}
	if prompt.GistURL != "" {
		fmt.Printf("  GitHub Gist: %s\n", prompt.GistURL)
	}
	fmt.Println()
	fmt.Println("Run 'pv list' to see all your prompts.")
}

func NewAddCommand(promptService service.PromptService) AddCmd {
	ac := &add{promptService: promptService}
	return &cobra.Command{
		Use:   "add <file-path>",
		Short: "Add a prompt from a YAML file to your vault",
		Long: `Add a prompt from a YAML file to your vault.

The YAML file should contain prompt metadata and content in the following format:

  metadata:
    name: "My Prompt"
    author: "Author Name"
    description: "A brief description"
    tags: ["tag1", "tag2"]
    version: "1.0.0"
  content: |
    Your prompt content here...

The prompt will be uploaded to GitHub Gists and indexed in your prompt vault.`,
		Args: cobra.ExactArgs(1),
		Run:  ac.execute,
	}
}