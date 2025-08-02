package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	apperrors "github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/service"
)

type AddCmd *cobra.Command

type add struct {
	promptService service.PromptService
}

func (ac *add) execute(cmd *cobra.Command, args []string) {
	// Validate arguments - we need exactly one argument (file path or URL)
	if len(args) != 1 {
		fmt.Println("❌ Error: Please provide exactly one file path or gist URL")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  pv add <file-path>")
		fmt.Println("  pv add <gist-url>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  pv add my-prompt.yaml")
		fmt.Println("  pv add https://gist.github.com/user/abc123")
		return
	}

	argument := args[0]
	var prompt *model.Prompt
	var err error

	// Determine if the argument is a URL or file path
	if ac.isGistURL(argument) {
		fmt.Printf("正在从 URL 导入提示词: %s\n", argument)
		prompt, err = ac.handleURLMode(argument)
	} else {
		fmt.Printf("正在从文件添加提示词: %s\n", argument)
		prompt, err = ac.handleFileMode(argument)
	}
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

// handleFileMode 处理文件路径模式
func (ac *add) handleFileMode(filePath string) (*model.Prompt, error) {
	return ac.promptService.AddFromFile(filePath)
}

// handleURLMode 处理 gist URL 模式
func (ac *add) handleURLMode(gistURL string) (*model.Prompt, error) {
	return ac.promptService.AddFromURL(gistURL)
}

// isGistURL 判断字符串是否为 gist URL
func (ac *add) isGistURL(str string) bool {
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
		return false
	}
	
	parsedURL, err := url.Parse(str)
	if err != nil {
		return false
	}
	
	// 检查是否为 GitHub Gist URL
	if parsedURL.Host != "gist.github.com" {
		return false
	}
	
	// 检查路径格式: /username/gist_id
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return false
	}
	
	// 验证 gist ID 格式（32位十六进制字符）
	gistID := pathParts[len(pathParts)-1]
	if len(gistID) != 32 {
		return false
	}
	
	for _, char := range gistID {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	
	return true
}

func NewAddCommand(promptService service.PromptService) AddCmd {
	ac := &add{promptService: promptService}
	return &cobra.Command{
		Use:   "add <file-path|gist-url>",
		Short: "Add a prompt from a YAML file or public gist URL to your vault",
		Long: `Add a prompt from a YAML file or public GitHub Gist URL to your vault.

For YAML files, the file should contain prompt metadata and content in the following format:

  metadata:
    name: "My Prompt"
    author: "Author Name"
    description: "A brief description"
    tags: ["tag1", "tag2"]
    version: "1.0.0"
  content: |
    Your prompt content here...

For gist URLs, the gist must be public and contain a valid prompt in YAML format.

The prompt will be uploaded to GitHub Gists (if from file) or imported to your local index (if from URL).`,
		Args: cobra.ExactArgs(1),
		Run:  ac.execute,
	}
}
