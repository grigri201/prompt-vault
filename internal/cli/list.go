package cli

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/errors"
)

// For testing
var getCachePathFunc = cache.GetCachePath

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
	cacheManager := cache.NewManagerWithPath(cachePath)

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

	// Create table writer
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "Name\tAuthor\tCategory\tVersion\tUpdated")
	fmt.Fprintln(w, "----\t------\t--------\t-------\t-------")

	// Print entries for current page
	for i := startIdx; i < endIdx; i++ {
		entry := index.Entries[i]
		updated := entry.UpdatedAt.Format("2006-01-02")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			entry.Name,
			entry.Author,
			entry.Category,
			entry.Version,
			updated,
		)
	}

	// Flush table before pagination info
	w.Flush()

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
