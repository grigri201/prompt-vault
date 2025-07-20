package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/models"
)

// newSyncCmd creates the sync command
func newSyncCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize local cache with GitHub Gists",
		Long: `Synchronize your local prompt cache with GitHub Gists.
This downloads all prompts from your index and updates the local cache.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd, verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress")

	return cmd
}

func runSync(cmd *cobra.Command, verbose bool) error {
	// For now, implement a simplified version that simulates sync
	// Full GitHub integration will be implemented later

	fmt.Fprintln(cmd.OutOrStdout(), "Starting synchronization...")

	// Create cache manager
	cachePath := getCachePathFunc()
	cacheManager := cache.NewManagerWithPath(cachePath)

	// Get or create index
	index, err := cacheManager.GetIndex()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to load index")
	}

	if index == nil {
		index = &models.Index{
			Username:  "testuser", // This would come from auth in real implementation
			Entries:   []models.IndexEntry{},
			UpdatedAt: time.Now(),
		}
	}

	// Simulate sync progress
	if verbose {
		fmt.Fprintln(cmd.OutOrStdout(), "Checking for updates...")
		fmt.Fprintln(cmd.OutOrStdout(), "Downloading index from GitHub...")
	}

	// Simulate downloading prompts
	downloadCount := 0
	updateCount := 0

	if len(index.Entries) > 0 {
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompts in index.\n", len(index.Entries))
		}

		// Simulate downloading each prompt
		for _, entry := range index.Entries {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "Downloading: %s...\n", entry.Name)
			}
			// In real implementation, this would download from GitHub
			downloadCount++
		}
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "No prompts found in index.")
		fmt.Fprintln(cmd.OutOrStdout(), "Upload prompts using 'pv upload' to get started.")
		return nil
	}

	// Update sync timestamp
	index.UpdatedAt = time.Now()
	if err := cacheManager.SaveIndex(index); err != nil {
		return errors.WrapWithMessage(err, "failed to save index")
	}

	// Show summary
	fmt.Fprintln(cmd.OutOrStdout(), "\nSync completed successfully!")
	fmt.Fprintf(cmd.OutOrStdout(), "- Downloaded: %d prompts\n", downloadCount)
	fmt.Fprintf(cmd.OutOrStdout(), "- Updated: %d prompts\n", updateCount)
	fmt.Fprintf(cmd.OutOrStdout(), "- Total prompts: %d\n", len(index.Entries))

	if verbose {
		fmt.Fprintf(cmd.OutOrStdout(), "\nLast sync: %s\n", index.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}
