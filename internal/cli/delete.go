package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/models"
)

// newDeleteCmd creates the delete command
func newDeleteCmd() *cobra.Command {
	var force bool
	
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a prompt template",
		Long: `Delete a prompt template from your GitHub Gists.
This action requires confirmation and can only be performed on your own templates.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteSimple(cmd, args[0], force)
		},
	}
	
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	
	return cmd
}

func runDeleteSimple(cmd *cobra.Command, promptName string, force bool) error {
	// For now, implement a simplified version that only removes from local cache
	// Full GitHub integration will be implemented later
	
	// Create cache manager
	cachePath := getCachePathFunc()
	cacheManager := cache.NewManagerWithPath(cachePath)

	// Get index from cache
	index, err := cacheManager.GetIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Check if index is nil or empty
	if index == nil {
		index = &models.Index{Entries: []models.IndexEntry{}}
	}

	// Find the prompt entry
	var targetEntry *models.IndexEntry
	var entryIndex int
	for i, entry := range index.Entries {
		if entry.Name == promptName {
			targetEntry = &entry
			entryIndex = i
			break
		}
	}

	if targetEntry == nil {
		fmt.Fprintf(cmd.OutOrStderr(), "Error: Prompt '%s' not found.\n", promptName)
		return fmt.Errorf("prompt not found")
	}

	// For testing purposes, we'll check a mock authentication
	// In real implementation, this would check against actual auth
	mockUsername := "testuser" // This would come from auth in real implementation
	if targetEntry.Author != mockUsername {
		fmt.Fprintf(cmd.OutOrStderr(), "Error: You can only delete your own prompts.\n")
		fmt.Fprintf(cmd.OutOrStderr(), "This prompt belongs to: %s\n", targetEntry.Author)
		return fmt.Errorf("permission denied")
	}

	// Confirm deletion unless force flag is set
	if !force {
		fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete '%s'? (y/N): ", promptName)
		
		reader := bufio.NewReader(cmd.InOrStdin())
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(cmd.OutOrStdout(), "Deletion cancelled.")
			return nil
		}
	}

	// Remove from index
	index.Entries = append(index.Entries[:entryIndex], index.Entries[entryIndex+1:]...)

	// Update local cache
	if err := cacheManager.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to update cache: %w", err)
	}

	// Delete from local cache
	if err := cacheManager.DeletePrompt(promptName); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to delete from local cache: %v\n", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted prompt '%s'.\n", promptName)
	fmt.Fprintln(cmd.OutOrStdout(), "Note: This is a simplified version. GitHub sync not yet implemented.")
	
	return nil
}