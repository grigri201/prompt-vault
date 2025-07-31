package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/imports"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
	"github.com/grigri201/prompt-vault/internal/upload"
)

type InputType int

const (
	InputTypeFile InputType = iota
	InputTypeGistURL
)

type AddCommand struct {
	Force bool
	Input string
}

func NewAddCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add [file|gist-url]",
		Short: "Add a prompt template from file or import from Gist URL",
		Long: `Add a prompt template to your vault. Supports two input types:

1. Local file: Upload a YAML file with prompt template
2. Gist URL: Import a prompt from GitHub Gist URL

Examples:
  pv add prompt.yaml                    # Upload local file
  pv add https://gist.github.com/...    # Import from Gist URL
  pv add prompt.yaml --force            # Force overwrite duplicates`,
		Args: cobra.ExactArgs(1),
		RunE: runAddCommand,
	}

	cmd.Flags().BoolP("force", "f", false, "Force overwrite duplicate prompts")

	// Integrate sync middleware
	return WithSyncMiddleware(cmd, "add")
}

func (c *AddCommand) detectInputType(input string) InputType {
	if strings.HasPrefix(input, "https://gist.github.com/") {
		return InputTypeGistURL
	}
	return InputTypeFile
}

func runAddCommand(cmd *cobra.Command, args []string) error {
	addCmd := &AddCommand{
		Input: args[0],
	}

	var err error
	addCmd.Force, err = cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Detect input type and execute appropriate logic
	inputType := addCmd.detectInputType(addCmd.Input)

	switch inputType {
	case InputTypeFile:
		return addCmd.handleFileUpload(cmd)
	case InputTypeGistURL:
		return addCmd.handleGistImport(cmd)
	default:
		return errors.NewValidationErrorMsg("AddCommand", "unsupported input type")
	}
}

func (c *AddCommand) handleFileUpload(cmd *cobra.Command) error {
	filename := c.Input

	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileSystemErrorMsg("add", fmt.Sprintf("file not found: %s", filename))
		}
		return errors.WrapWithMessage(err, "failed to access file")
	}

	// Check if it's a regular file
	if info.IsDir() {
		return errors.NewValidationErrorMsg("add", fmt.Sprintf("%s is a directory, not a file", filename))
	}

	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to read file")
	}

	// Parse the prompt file using the unified parser
	fmt.Fprintf(cmd.OutOrStdout(), "Parsing prompt file: %s\n", filename)

	// Create a parser with strict validation for uploads
	yamlParser := parser.NewYAMLParser(parser.YAMLParserConfig{
		Strict: true,
	})

	prompt, err := yamlParser.ParsePromptFile(string(content))
	if err != nil {
		return errors.WrapWithMessage(err, "failed to parse prompt file")
	}

	// Set default version if not provided
	prompt.PromptMeta.SetDefaultVersion()

	// Create auth manager
	authManager := auth.NewManager()

	// Get current user
	ctx := context.Background()
	username, err := authManager.GetCurrentUser(ctx)
	if err != nil {
		return errors.NewAuthErrorMsg("add", "not authenticated. Please run 'pv login' first")
	}

	// Create GitHub client
	token, err := authManager.GetToken()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get authentication token")
	}

	client, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to create GitHub client")
	}

	// Create GistOperations wrapper
	gistOps := gist.NewGistOperations(gist.GistOperationsConfig{
		Client:     client,
		RetryCount: 3,
	})

	// Perform pre-sync to ensure we have latest data for duplicate detection
	if err := performPreSync(cmd, c.Force); err != nil {
		return err
	}

	// Create cache manager after sync
	_, cacheManager := createManagers()

	// Read current index (should be updated after sync)
	index, err := cacheManager.GetIndex()
	if err != nil {
		// If index doesn't exist, create a new one
		index = &models.Index{
			Username:  username,
			Entries:   []models.IndexEntry{},
			UpdatedAt: time.Now(),
		}
	}

	// Use duplicate detector to check for existing prompts
	duplicateDetector := upload.NewDuplicateDetector()
	duplicateMatch, err := duplicateDetector.FindDuplicate(prompt, index)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to check for duplicates")
	}

	var existingGistID string

	if duplicateMatch != nil {
		existingGistID = duplicateMatch.Entry.GistID

		// Log the type of match found
		switch duplicateMatch.MatchType {
		case upload.MatchByID:
			fmt.Fprintf(cmd.OutOrStdout(), "Found existing prompt with ID '%s'\n", prompt.ID)
		case upload.MatchByNameAuthor:
			fmt.Fprintf(cmd.OutOrStdout(), "Found existing prompt '%s' by %s\n", prompt.Name, prompt.Author)
		case upload.MatchByGistID:
			fmt.Fprintf(cmd.OutOrStdout(), "Found existing gist %s\n", prompt.GistID)
		}

		// Handle duplicate based on force flag
		if !c.Force {
			// Interactive mode - ask for confirmation
			fmt.Fprintf(cmd.OutOrStdout(), "Update existing prompt '%s'? (Y/n): ", duplicateMatch.Entry.Name)

			reader := bufio.NewReader(cmd.InOrStdin())
			response, err := reader.ReadString('\n')
			if err != nil {
				return errors.WrapWithMessage(err, "failed to read confirmation")
			}

			response = strings.TrimSpace(strings.ToLower(response))
			// Default is yes (empty response or 'y'/'yes')
			if response != "" && response != "y" && response != "yes" {
				fmt.Fprintf(cmd.OutOrStdout(), "Upload cancelled.\n")
				return errors.NewValidationErrorMsg("add", "upload cancelled by user")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updating existing prompt...\n")
		} else {
			// Force mode - automatically update existing
			fmt.Fprintf(cmd.OutOrStdout(), "Force mode: updating existing prompt\n")
		}
	}

	// Generate gist filename
	gistName := prompt.Name

	// Reconstruct the full content with front matter
	fullContent := parser.FormatPromptFile(&prompt.PromptMeta, prompt.Content)

	// Prepare gist data
	gistData := &gist.GistData{
		Name:        gistName,
		Description: prompt.Description,
		Content:     fullContent,
		Public:      false,
	}

	// Use CreateOrUpdate which handles 404 cases automatically
	var result *gist.GistResult
	if existingGistID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Checking existing prompt '%s' by %s...\n", prompt.Name, prompt.Author)
		result, err = gistOps.CreateOrUpdateWithRetry(ctx, existingGistID, gistData)

		// If it was created (not updated), we need to clean up the stale entry
		if err == nil && result.Created && index != nil {
			for i, entry := range index.Entries {
				if entry.GistID == existingGistID {
					// Remove the stale entry
					index.Entries = append(index.Entries[:i], index.Entries[i+1:]...)
					break
				}
			}
		}
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Creating new prompt '%s' by %s...\n", prompt.Name, prompt.Author)
		result, err = gistOps.CreateOrUpdateWithRetry(ctx, "", gistData)
	}

	if err != nil {
		return errors.WrapWithMessage(err, "failed to upload prompt")
	}

	// Update operation message based on result
	if result.Created {
		if existingGistID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Previous gist not found, created new prompt\n")
		}
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Updated existing prompt\n")
	}

	// Update prompt with gist info
	prompt.GistID = result.ID
	prompt.GistURL = result.URL
	prompt.UpdatedAt = time.Now()

	// Cache the prompt locally
	if err := cacheManager.SavePrompt(prompt); err != nil {
		// Don't fail the upload if caching fails
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to cache prompt locally: %v\n", err)
	}

	// Update the index
	indexEntry := prompt.ToIndexEntry()
	updated := false

	// If we found a duplicate, update that entry
	if duplicateMatch != nil {
		for i, entry := range index.Entries {
			if entry.GistID == duplicateMatch.Entry.GistID {
				index.Entries[i] = indexEntry
				updated = true
				break
			}
		}
	}

	// If not updated yet, look for name+author match (backward compatibility)
	if !updated {
		for i, entry := range index.Entries {
			if entry.Name == prompt.Name && entry.Author == prompt.Author {
				index.Entries[i] = indexEntry
				updated = true
				break
			}
		}
	}

	// If still not updated, it's a new entry
	if !updated {
		index.Entries = append(index.Entries, indexEntry)
	}
	index.UpdatedAt = time.Now()

	// Save the updated index
	if err := cacheManager.SaveIndex(index); err != nil {
		// Don't fail the upload if index saving fails
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update local index: %v\n", err)
	}

	// Display success message
	fmt.Fprintf(cmd.OutOrStdout(), "\n✓ Successfully uploaded '%s' by %s\n", prompt.Name, prompt.Author)
	fmt.Fprintf(cmd.OutOrStdout(), "  Tags: %v\n", strings.Join(prompt.Tags, ", "))
	fmt.Fprintf(cmd.OutOrStdout(), "  Version: %s\n", prompt.Version)
	if prompt.Description != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Description: %s\n", prompt.Description)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\n  View at: %s\n", prompt.GistURL)

	return nil
}

func (c *AddCommand) handleGistImport(cmd *cobra.Command) error {
	// Perform pre-sync to ensure we have latest index and avoid conflicts
	if err := performPreSync(cmd, false); err != nil {
		return err
	}

	// Initialize dependencies
	authManager := auth.NewManager()
	token, err := authManager.GetToken()
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgAuthRequired)
	}
	username, err := authManager.GetUsername()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get username")
	}

	gistClient, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgNetworkTimeout)
	}

	// Create UI adapter for import confirmation
	uiAdapter := &addUIAdapter{cmd: cmd, force: c.Force}

	// Create import manager
	manager := imports.NewManager(gistClient, uiAdapter)

	ctx := context.Background()

	// Create cache manager to work with local index
	_, cacheManager := createManagers()

	// Get current local index after preSync
	index, err := cacheManager.GetIndex()
	if err != nil {
		// If index doesn't exist, create a new one
		index = &models.Index{
			Username:        username,
			Entries:         []models.IndexEntry{},
			ImportedEntries: []models.IndexEntry{},
			UpdatedAt:       time.Now(),
		}
	}

	// Import the prompt (this will modify the index's ImportedEntries)
	result, err := manager.ImportPrompt(ctx, c.Input, index)
	if err != nil {
		errMsg := errors.GetImportErrorMessage(err)
		cmd.PrintErrf("%s\n", errMsg)
		return err
	}

	// Update the local index timestamp
	index.UpdatedAt = time.Now()

	// Save the updated local index
	if err := cacheManager.SaveIndex(index); err != nil {
		// Don't fail the import if index saving fails
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update local index: %v\n", err)
	}

	// Display result
	if result.IsUpdate {
		cmd.Printf("Successfully updated prompt (gist: %s) from version %s to %s\n",
			result.GistID, result.OldVersion, result.NewVersion)
	} else {
		cmd.Printf("Successfully imported prompt (gist: %s)\n", result.GistID)
	}

	return nil
}

// addUIAdapter implements the UI interface for imports.Manager with force support
type addUIAdapter struct {
	cmd   *cobra.Command
	force bool
}

func (u *addUIAdapter) Confirm(message string) (bool, error) {
	// If force is enabled, automatically confirm
	if u.force {
		u.cmd.Printf("%s (force mode: automatically confirmed)\n", message)
		return true, nil
	}

	// Interactive mode - ask for confirmation
	u.cmd.Printf("%s (Y/n): ", message)

	reader := bufio.NewReader(u.cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, errors.WrapWithMessage(err, "failed to read confirmation")
	}

	response = strings.TrimSpace(strings.ToLower(response))
	// Default is yes (empty response or 'y'/'yes')
	return response == "" || response == "y" || response == "yes", nil
}
