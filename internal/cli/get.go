package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/models"
)


// newGetCmd creates the get command
func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [keyword]",
		Short: "Search and retrieve prompt templates",
		Long: `Search for prompt templates by keyword across names, categories, 
tags, authors, and descriptions. Select a template to fill in variables 
interactively and copy the result to your clipboard.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runGet,
	}
}

func runGet(cmd *cobra.Command, args []string) error {
	// Create cache manager
	cachePath := getCachePathFunc()
	cacheManager := cache.NewManagerWithPath(cachePath)

	// Get index from cache
	index, err := cacheManager.GetIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Check if index is empty
	if index == nil || len(index.Entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prompts found.")
		fmt.Fprintln(cmd.OutOrStdout(), "Use 'pv sync' to download prompts from GitHub.")
		return nil
	}

	// If no keyword provided, show all
	keyword := ""
	if len(args) > 0 {
		keyword = strings.ToLower(args[0])
	}

	// Search entries
	var matches []int
	for i, entry := range index.Entries {
		if keyword == "" || matchesKeyword(entry, keyword) {
			matches = append(matches, i)
		}
	}

	// Check if any matches found
	if len(matches) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No prompts found matching '%s'.\n", keyword)
		return nil
	}

	// Display matches
	fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompt(s):\n\n", len(matches))
	
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
	}

	// For now, just show the search results
	// Interactive selection will be implemented in task 7.2
	fmt.Fprintln(cmd.OutOrStdout(), "Interactive prompt selection will be available soon.")

	return nil
}

// matchesKeyword checks if an entry matches the search keyword
func matchesKeyword(entry models.IndexEntry, keyword string) bool {
	// Search in name
	if strings.Contains(strings.ToLower(entry.Name), keyword) {
		return true
	}
	
	// Search in author
	if strings.Contains(strings.ToLower(entry.Author), keyword) {
		return true
	}
	
	// Search in category
	if strings.Contains(strings.ToLower(entry.Category), keyword) {
		return true
	}
	
	// Search in description
	if strings.Contains(strings.ToLower(entry.Description), keyword) {
		return true
	}
	
	// Search in tags
	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), keyword) {
			return true
		}
	}
	
	return false
}