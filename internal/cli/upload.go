package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
)

// newUploadCmd creates the upload command
func newUploadCmd() *cobra.Command {
	return &cobra.Command{
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
		RunE: runUpload,
	}
}

func runUpload(cmd *cobra.Command, args []string) error {
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

	// Parse the prompt file
	fmt.Fprintf(cmd.OutOrStdout(), "Parsing prompt file: %s\n", filename)
	prompt, err := parser.ParsePromptFile(string(content))
	if err != nil {
		return errors.WrapWithMessage(err, "failed to parse prompt file")
	}

	// Validate prompt metadata
	if err := prompt.PromptMeta.Validate(); err != nil {
		return errors.WrapWithMessage(err, "invalid prompt metadata")
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
	token, _, err := authManager.GetToken()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get authentication token")
	}

	client, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to create GitHub client")
	}

	// Generate gist name
	gistName := fmt.Sprintf("%s-%s", prompt.Author, prompt.Name)

	// Check if prompt already exists in the index
	cacheManager := cache.NewManager()
	index, _ := cacheManager.GetIndex()
	
	var existingGistID string
	
	if index != nil {
		// Look for existing prompt with same name and author
		for _, entry := range index.Entries {
			if entry.Name == prompt.Name && entry.Author == prompt.Author {
				existingGistID = entry.GistID
				break
			}
		}
	}

	// Reconstruct the full content with front matter
	fullContent := parser.FormatPromptFile(&prompt.PromptMeta, prompt.Content)

	var gistID, gistURL string

	if existingGistID != "" {
		// Check if the gist actually exists
		fmt.Fprintf(cmd.OutOrStdout(), "Checking existing prompt '%s' by %s...\n", prompt.Name, prompt.Author)
		
		_, err := client.GetGist(ctx, existingGistID)
		if err != nil {
			// If gist doesn't exist, create a new one instead
			if strings.Contains(err.Error(), "gist not found") {
				fmt.Fprintf(cmd.OutOrStdout(), "Previous gist not found, creating new prompt...\n")
				
				newGistID, newGistURL, err := client.CreateGist(ctx, gistName, prompt.Description, fullContent)
				if err != nil {
					return errors.WrapWithMessage(err, "failed to create gist")
				}
				gistID = newGistID
				gistURL = newGistURL
				
				// Clear the stale entry from cache
				if index != nil {
					for i, entry := range index.Entries {
						if entry.GistID == existingGistID {
							// Remove the stale entry
							index.Entries = append(index.Entries[:i], index.Entries[i+1:]...)
							break
						}
					}
				}
			} else {
				return errors.WrapWithMessage(err, "failed to check existing gist")
			}
		} else {
			// Gist exists, update it
			fmt.Fprintf(cmd.OutOrStdout(), "Updating existing prompt '%s' by %s...\n", prompt.Name, prompt.Author)
			
			updatedURL, err := client.UpdateGist(ctx, existingGistID, gistName, prompt.Description, fullContent)
			if err != nil {
				// If update fails with 404, try creating new gist
				if strings.Contains(err.Error(), "gist not found") {
					fmt.Fprintf(cmd.OutOrStdout(), "Gist not found during update, creating new prompt...\n")
					
					newGistID, newGistURL, err := client.CreateGist(ctx, gistName, prompt.Description, fullContent)
					if err != nil {
						return errors.WrapWithMessage(err, "failed to create gist")
					}
					gistID = newGistID
					gistURL = newGistURL
					
					// Clear the stale entry from cache
					if index != nil {
						for i, entry := range index.Entries {
							if entry.GistID == existingGistID {
								// Remove the stale entry
								index.Entries = append(index.Entries[:i], index.Entries[i+1:]...)
								break
							}
						}
					}
				} else {
					return errors.WrapWithMessage(err, "failed to update gist")
				}
			} else {
				gistID = existingGistID
				gistURL = updatedURL
			}
		}
	} else {
		// Create new gist
		fmt.Fprintf(cmd.OutOrStdout(), "Creating new prompt '%s' by %s...\n", prompt.Name, prompt.Author)
		
		newGistID, newGistURL, err := client.CreateGist(ctx, gistName, prompt.Description, fullContent)
		if err != nil {
			return errors.WrapWithMessage(err, "failed to create gist")
		}
		gistID = newGistID
		gistURL = newGistURL
	}

	// Update the prompt with gist information
	prompt.GistID = gistID
	prompt.GistURL = gistURL
	prompt.UpdatedAt = time.Now()

	// Update the index
	fmt.Fprintln(cmd.OutOrStdout(), "Updating index...")
	if err := updateIndex(ctx, client, username, prompt); err != nil {
		return errors.WrapWithMessage(err, "failed to update index")
	}

	// Cache the prompt locally
	if err := cacheManager.InitializeCache(); err != nil {
		// Log but don't fail
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: failed to initialize cache: %v\n", err)
	} else {
		if err := cacheManager.SavePrompt(prompt); err != nil {
			// Log but don't fail
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: failed to cache prompt: %v\n", err)
		}
	}

	// Success message
	fmt.Fprintln(cmd.OutOrStdout(), "\nPrompt uploaded successfully!")
	fmt.Fprintf(cmd.OutOrStdout(), "Gist URL: %s\n", gistURL)
	fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", prompt.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Version: %s\n", prompt.Version)

	return nil
}

// updateIndex updates the user's index gist with the new prompt
func updateIndex(ctx context.Context, client *gist.Client, username string, prompt *models.Prompt) error {
	// Get existing index from cache
	cacheManager := cache.NewManager()
	index, err := cacheManager.GetIndex()
	if err != nil || index == nil {
		// Create new index if none exists
		index = &models.Index{
			Username:  username,
			Entries:   []models.IndexEntry{},
			UpdatedAt: time.Now(),
		}
	}

	// Check if prompt already exists in index
	found := false
	for i, entry := range index.Entries {
		if entry.GistID == prompt.GistID ||
			(entry.Name == prompt.Name && entry.Author == prompt.Author) {
			// Update existing entry
			index.Entries[i] = prompt.ToIndexEntry()
			found = true
			break
		}
	}

	// Add new entry if not found
	if !found {
		index.Entries = append(index.Entries, prompt.ToIndexEntry())
	}

	// Update timestamp
	index.UpdatedAt = time.Now()

	// Update the index gist
	_, err = client.UpdateIndexGist(ctx, username, index)
	if err != nil {
		return err
	}

	// Cache the updated index
	if err := cacheManager.SaveIndex(index); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to cache index: %v\n", err)
	}

	return nil
}
