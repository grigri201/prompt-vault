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
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
	"github.com/grigri201/prompt-vault/internal/upload"
)

// newUploadCmd creates the upload command
func newUploadCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "upload [file]",
		Aliases: []string{"up"},
		Short:   "Upload a prompt template to GitHub Gist",
		Long: `Upload a prompt template file to GitHub Gist.
The file should be in YAML format with front matter containing metadata.

Example file format:
---
name: API Documentation Generator
author: yourname
category: documentation
tags: [api, docs, swagger]
description: Generate API documentation from OpenAPI specs
---
Generate {format} documentation for the following API:
{api_spec}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpload(cmd, args, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Automatically overwrite existing prompts without confirmation")

	// Apply auto-sync middleware
	WrapWithAutoSync(cmd)

	return cmd
}

func runUpload(cmd *cobra.Command, args []string, force bool) error {
	filename := args[0]

	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileSystemErrorMsg("upload", fmt.Sprintf("file not found: %s", filename))
		}
		return errors.WrapWithMessage(err, "failed to access file")
	}

	// Check if it's a regular file
	if info.IsDir() {
		return errors.NewValidationErrorMsg("upload", fmt.Sprintf("%s is a directory, not a file", filename))
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
		return errors.NewAuthErrorMsg("upload", "not authenticated. Please run 'pv login' first")
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
	if err := performPreSync(cmd, force); err != nil {
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
		if !force {
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
				return errors.NewValidationErrorMsg("upload", "upload cancelled by user")
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

	// GitHub index update will be handled by auto-sync middleware

	// Display success message
	fmt.Fprintf(cmd.OutOrStdout(), "\n✓ Successfully uploaded '%s' by %s\n", prompt.Name, prompt.Author)
	fmt.Fprintf(cmd.OutOrStdout(), "  Category: %s\n", prompt.Category)
	fmt.Fprintf(cmd.OutOrStdout(), "  Tags: %v\n", strings.Join(prompt.Tags, ", "))
	fmt.Fprintf(cmd.OutOrStdout(), "  Version: %s\n", prompt.Version)
	if prompt.Description != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Description: %s\n", prompt.Description)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\n  View at: %s\n", prompt.GistURL)

	return nil
}
