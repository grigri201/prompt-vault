package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/search"
	"github.com/grigri201/prompt-vault/internal/ui"
)

// newDeleteCmd creates the delete command
func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete [<name>|<keyword>|<gist-url>]",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a prompt template",
		Long: `Delete a prompt template from your GitHub Gists.
This action requires confirmation and can only be performed on your own templates.

Usage:
  pv delete                     # List all prompts and select one to delete
  pv delete <keyword>           # Search for prompts matching keyword and select one to delete
  pv delete <gist-url>          # Delete a specific prompt by its gist URL
  pv delete <gist-id>           # Delete a specific prompt by its gist ID`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runDeleteWithSelection(cmd, "", force)
			}
			return runDeleteWithSelection(cmd, args[0], force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

// extractGistIDFromURL extracts the gist ID from a GitHub gist URL
func extractGistIDFromURL(url string) string {
	// Handle URLs like https://gist.github.com/username/gistid
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return url
}

// isGistIDOrURL checks if the input is a gist ID or URL
func isGistIDOrURL(input string) bool {
	// Check if it's a URL
	if strings.Contains(input, "gist.github.com") {
		return true
	}
	// Check if it's a 32-character hex string (gist ID)
	if len(input) == 32 {
		for _, c := range input {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
		return true
	}
	return false
}

func runDeleteWithSelection(cmd *cobra.Command, input string, force bool) error {
	// Create cache manager
	cachePath := getCachePathFunc()
	cacheManager := cache.NewManagerWithPath(cachePath)

	// Get index from cache
	index, err := cacheManager.GetIndex()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to load index")
	}

	// Check if index is nil or empty
	if index == nil || len(index.Entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prompts found.")
		fmt.Fprintln(cmd.OutOrStdout(), "Use 'pv sync' to download prompts from GitHub.")
		return nil
	}

	var targetEntry *models.IndexEntry
	var entryIndex int

	// If no input provided, show all prompts for selection
	if input == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompt(s) to delete:\n\n", len(index.Entries))
		
		// Create selector items
		selectorItems := make([]string, len(index.Entries))
		for i, entry := range index.Entries {
			fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s by %s\n", i+1, entry.Name, entry.Author)
			fmt.Fprintf(cmd.OutOrStdout(), "    Category: %s\n", entry.Category)
			if len(entry.Tags) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "    Tags: %s\n", strings.Join(entry.Tags, ", "))
			}
			if entry.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "    Description: %s\n", entry.Description)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			
			selectorItems[i] = fmt.Sprintf("%s by %s", entry.Name, entry.Author)
		}
		
		// Create and run selector
		selector := ui.NewSelector(selectorItems)
		fmt.Fprintln(cmd.OutOrStdout())
		
		p := tea.NewProgram(selector)
		finalModel, err := p.Run()
		if err != nil {
			return errors.WrapWithMessage(err, "failed to run selector")
		}
		
		// Check if user made a selection
		selectorModel := finalModel.(ui.SelectorModel)
		if !selectorModel.IsConfirmed() {
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo selection made.")
			return nil
		}
		
		entryIndex = selectorModel.Selected
		targetEntry = &index.Entries[entryIndex]
	} else if isGistIDOrURL(input) {
		// Handle gist ID or URL
		gistID := extractGistIDFromURL(input)
		
		// Find by gist ID
		for i, entry := range index.Entries {
			if entry.GistID == gistID {
				targetEntry = &entry
				entryIndex = i
				break
			}
		}
		
		if targetEntry == nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: Prompt with gist ID '%s' not found.\n", gistID)
			return errors.NewValidationErrorMsg("delete", "prompt not found")
		}
	} else {
		// Search by keyword
		searcher := search.NewSearcher()
		matches := searcher.SearchEntries(index.Entries, input)
		
		if len(matches) == 0 {
			fmt.Fprintf(cmd.OutOrStderr(), "No prompts found matching '%s'.\n", input)
			return errors.NewValidationErrorMsg("delete", "prompt not found")
		}
		
		// Always show selection (even for single match)
		if len(matches) >= 1 {
			fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompt(s) matching '%s':\n\n", len(matches), input)
			
			// Create selector items
			selectorItems := make([]string, len(matches))
			for i, idx := range matches {
				entry := index.Entries[idx]
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s by %s\n", i+1, entry.Name, entry.Author)
				fmt.Fprintf(cmd.OutOrStdout(), "    Category: %s\n", entry.Category)
				if len(entry.Tags) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "    Tags: %s\n", strings.Join(entry.Tags, ", "))
				}
				if entry.Description != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "    Description: %s\n", entry.Description)
				}
				fmt.Fprintln(cmd.OutOrStdout())
				
				selectorItems[i] = fmt.Sprintf("%s by %s", entry.Name, entry.Author)
			}
			
			// Create and run selector
			selector := ui.NewSelector(selectorItems)
			fmt.Fprintln(cmd.OutOrStdout())
			
			p := tea.NewProgram(selector)
			finalModel, err := p.Run()
			if err != nil {
				return errors.WrapWithMessage(err, "failed to run selector")
			}
			
			// Check if user made a selection
			selectorModel := finalModel.(ui.SelectorModel)
			if !selectorModel.IsConfirmed() {
				fmt.Fprintln(cmd.OutOrStdout(), "\nNo selection made.")
				return nil
			}
			
			entryIndex = matches[selectorModel.Selected]
			targetEntry = &index.Entries[entryIndex]
		}
	}

	return runDeleteSimple(cmd, targetEntry, entryIndex, index, cacheManager, force)
}

func runDeleteSimple(cmd *cobra.Command, targetEntry *models.IndexEntry, entryIndex int, index *models.Index, cacheManager *cache.Manager, force bool) error {
	// Confirm deletion unless force flag is set
	if !force {
		fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete '%s'? (y/N): ", targetEntry.Name)

		reader := bufio.NewReader(cmd.InOrStdin())
		response, err := reader.ReadString('\n')
		if err != nil {
			return errors.WrapWithMessage(err, "failed to read confirmation")
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(cmd.OutOrStdout(), "Deletion cancelled.")
			return nil
		}
	}

	// Get auth token
	authManager := auth.NewManager()
	token, username, err := authManager.GetToken()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get authentication token")
	}

	// Create GitHub client
	client, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to create GitHub client")
	}

	// Delete from GitHub Gist
	ctx := context.Background()
	if err := client.DeleteGist(ctx, targetEntry.GistID); err != nil {
		// If gist is already deleted or not found, continue with local cleanup
		if !strings.Contains(err.Error(), "not found") {
			return errors.WrapWithMessage(err, "failed to delete gist from GitHub")
		}
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: Gist not found on GitHub, cleaning up local cache only\n")
	}

	// Remove from index
	index.Entries = append(index.Entries[:entryIndex], index.Entries[entryIndex+1:]...)

	// Update local cache
	if err := cacheManager.SaveIndex(index); err != nil {
		return errors.WrapWithMessage(err, "failed to update cache")
	}

	// Delete from local cache
	if err := cacheManager.DeletePrompt(targetEntry.Name); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to delete from local cache: %v\n", err)
	}

	// Update index on GitHub
	if _, err := client.UpdateIndexGist(ctx, username, index); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to update index on GitHub: %v\n", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted prompt '%s'.\n", targetEntry.Name)

	return nil
}
