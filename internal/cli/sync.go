package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/models"
)

// newSyncCmd creates the sync command
func newSyncCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize local cache with GitHub Gists",
		Long: `Synchronize your local prompt cache with GitHub Gists.
This downloads all prompts from your online index and updates the local cache.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd, verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress")

	return cmd
}

func runSync(cmd *cobra.Command, verbose bool) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Starting synchronization...")

	// Create cache manager
	cachePath := getCachePathFunc()
	_, cacheManager := createManagersWithPath(cachePath)

	// Check if cache directory exists before initialization
	_, dirExistsBefore := os.Stat(cachePath)
	cacheDirCreated := os.IsNotExist(dirExistsBefore)

	// Initialize cache manager to ensure directories exist
	if err := cacheManager.Initialize(cmd.Context()); err != nil {
		return errors.WrapWithMessage(err, "failed to initialize cache")
	}

	// If cache directory was created, notify the user
	if cacheDirCreated {
		fmt.Fprintf(cmd.OutOrStdout(), "Cache directory created at %s\n", cachePath)
	}

	// Get config to get username and token
	cfgManager, _ := createManagers()
	cfg, err := cfgManager.GetConfig()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get config")
	}

	if cfg.Token == "" {
		return errors.NewAuthErrorMsg("sync", "not authenticated. Please run 'pv login' first")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Fetching prompts from GitHub for user: %s\n", cfg.Username)

	// Create GitHub client
	client, err := gist.NewClient(cfg.Token)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to create GitHub client")
	}

	// Create GistOperations for fetching prompts with retry logic
	gistOps := gist.NewGistOperations(gist.GistOperationsConfig{
		Client:     client,
		RetryCount: 3,
	})

	// Get the index gist from GitHub
	index, _, err := client.GetIndexGist(cmd.Context(), cfg.Username)
	if err != nil {
		if strings.Contains(err.Error(), "index gist not found") {
			// No index gist exists yet
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo prompts found online.")
			fmt.Fprintln(cmd.OutOrStdout(), "Upload your first prompt with 'pv upload <file>' to get started.")

			// Create empty local index
			emptyIndex := &models.Index{
				Username:  cfg.Username,
				Entries:   []models.IndexEntry{},
				UpdatedAt: time.Now(),
			}
			if err := cacheManager.SaveIndex(emptyIndex); err != nil {
				return errors.WrapWithMessage(err, "failed to save empty index")
			}
			return nil
		}
		return errors.WrapWithMessage(err, "failed to get index from GitHub")
	}

	if len(index.Entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\nNo prompts found in index.")
		// Save the empty index locally
		if err := cacheManager.SaveIndex(index); err != nil {
			return errors.WrapWithMessage(err, "failed to save index")
		}
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Found index with %d prompts\n", len(index.Entries))

	// Process each entry in the index
	var validEntries []models.IndexEntry
	processedCount := 0
	deletedCount := 0

	fmt.Fprintln(cmd.OutOrStdout(), "\nDownloading prompts...")

	for i, entry := range index.Entries {
		// Show progress
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "\n[%d/%d] Fetching prompt: %s (gist %s)\n",
				i+1, len(index.Entries), entry.Name, entry.GistID)
		}

		// Fetch the prompt gist
		prompt, err := gistOps.FetchPromptGist(cmd.Context(), entry.GistID)
		if err != nil {
			// Check if it's a 404 error (gist deleted)
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found") {
				fmt.Fprintf(cmd.OutOrStderr(), "⚠ Prompt '%s' (gist %s) has been deleted from GitHub\n",
					entry.Name, entry.GistID)
				deletedCount++
				continue // Skip this entry
			}

			// For other errors, report but continue
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to fetch prompt '%s': %v\n", entry.Name, err)
			if verbose {
				fmt.Fprintf(cmd.OutOrStderr(), "  Continuing with sync...\n")
			}
			// Still add to valid entries as it might be a temporary issue
			validEntries = append(validEntries, entry)
			continue
		}

		// Save the prompt to local cache
		if err := cacheManager.SavePrompt(prompt); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to cache prompt '%s': %v\n", entry.Name, err)
		} else {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓ Cached prompt: %s\n", prompt.Name)
			} else if (processedCount+1)%10 == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Processed %d prompts...\n", processedCount+1)
			}
			processedCount++
		}

		// Add to valid entries (not deleted)
		validEntries = append(validEntries, entry)
	}

	// Update the index if any entries were deleted
	if deletedCount > 0 {
		index.Entries = validEntries
		index.UpdatedAt = time.Now()

		// Update the online index
		_, err = client.UpdateIndexGist(cmd.Context(), cfg.Username, index)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: Failed to update online index: %v\n", err)
			fmt.Fprintln(cmd.OutOrStderr(), "The deleted entries will remain in the online index until manually updated.")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed %d deleted entries from online index\n", deletedCount)
		}
	}

	// Save the index locally (either updated or original)
	if err := cacheManager.SaveIndex(index); err != nil {
		return errors.WrapWithMessage(err, "failed to save index")
	}

	// Display summary
	fmt.Fprintln(cmd.OutOrStdout(), "\nSync completed successfully!")
	fmt.Fprintf(cmd.OutOrStdout(), "  Total entries in index: %d\n", len(index.Entries))
	fmt.Fprintf(cmd.OutOrStdout(), "  Prompts synced: %d\n", processedCount)
	if deletedCount > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Deleted entries removed: %d\n", deletedCount)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nLocal cache updated at: %s\n", cacheManager.GetCacheDir())

	return nil
}
