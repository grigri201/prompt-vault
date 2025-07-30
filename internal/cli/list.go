package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/errors"
)

// newListCmd creates the list command
func newListCmd() *cobra.Command {
	var page int

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all available prompt templates",
		Long: `List all available prompt templates from your GitHub Gists.
Templates are displayed in a paginated table showing name, author, 
category, and version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, page)
		},
	}

	cmd.Flags().IntVarP(&page, "page", "p", 1, "Page number to display")

	return cmd
}

func runList(cmd *cobra.Command, page int) error {
	// Create cache manager
	cachePath := getCachePathFunc()
	_, cacheManager := createManagersWithPath(cachePath)

	// Get index from cache
	index, err := cacheManager.GetIndex()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to load index")
	}

	// Check if index is empty
	if index == nil || len(index.Entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prompts found.")
		fmt.Fprintln(cmd.OutOrStdout(), "Use 'pv sync' to download prompts from GitHub.")
		return nil
	}

	// Calculate pagination
	const pageSize = 20
	totalItems := len(index.Entries)
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Validate page number
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// Calculate start and end indices
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize
	if endIdx > totalItems {
		endIdx = totalItems
	}

	// Print entries for current page in a detailed format
	for i := startIdx; i < endIdx; i++ {
		entry := index.Entries[i]
		fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s by %s\n", i+1, entry.Name, entry.Author)
		fmt.Fprintf(cmd.OutOrStdout(), "    Category: %s\n", entry.Category)
		if len(entry.Tags) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "    Tags: %s\n", strings.Join(entry.Tags, ", "))
		}
		if entry.Version != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "    Version: %s\n", entry.Version)
		}
		if entry.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "    Description: %s\n", entry.Description)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "    Updated: %s\n", entry.UpdatedAt.Format("2006-01-02"))
		if entry.GistURL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "    Gist URL: %s\n", entry.GistURL)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Print pagination info
	if totalPages > 1 {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintf(cmd.OutOrStdout(), "Page %d of %d (Showing %d-%d of %d)\n",
			page, totalPages, startIdx+1, endIdx, totalItems)

		if page < totalPages {
			fmt.Fprintf(cmd.OutOrStdout(), "Use 'pv list --page %d' to see the next page.\n", page+1)
		}
	}

	// Show last sync time
	if !index.UpdatedAt.IsZero() {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintf(cmd.OutOrStdout(), "Last synced: %s\n",
			index.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}
