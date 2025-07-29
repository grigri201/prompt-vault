package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
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
	fmt.Fprintln(cmd.OutOrStdout(), "Starting synchronization...")

	// Create cache manager
	cachePath := getCachePathFunc()
	cacheManager := cache.NewManagerWithPath(cachePath)

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
	cfgManager := config.NewManager()
	cfg, err := cfgManager.GetConfig()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to load config")
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgAuthRequired)
	}

	// Create GitHub client
	gistClient, err := gist.NewClient(cfg.Token)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to create GitHub client")
	}

	// Get or create index
	index, err := cacheManager.GetIndex()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to load index")
	}

	// Debug: Print existing index info
	if verbose {
		if index != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "\nExisting index info:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "- Username: %s\n", index.Username)
			fmt.Fprintf(cmd.OutOrStdout(), "- Entries: %d\n", len(index.Entries))
			fmt.Fprintf(cmd.OutOrStdout(), "- Last updated: %s\n", index.UpdatedAt.Format(time.RFC3339))
			if len(index.Entries) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "- Existing prompts:\n")
				for _, entry := range index.Entries {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s (GistID: %s)\n", entry.Name, entry.GistID)
				}
			}
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo existing index found (will create new)")
		}
	}

	// Fetch all gists from GitHub
	if verbose {
		fmt.Fprintln(cmd.OutOrStdout(), "Fetching prompts from GitHub...")
	}

	gists, err := gistClient.ListUserGists(cmd.Context(), cfg.Username)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to fetch gists from GitHub")
	}

	// Debug: Print GitHub gists info
	if verbose {
		fmt.Fprintf(cmd.OutOrStdout(), "\nFetched %d gists from GitHub for user: %s\n", len(gists), cfg.Username)
	}

	// Build new index from GitHub gists
	newIndex := &models.Index{
		Username:  cfg.Username,
		Entries:   []models.IndexEntry{},
		UpdatedAt: time.Now(),
	}

	// Count for statistics
	downloadCount := 0
	updateCount := 0
	newCount := 0

	// Process each gist
	for _, g := range gists {
		// Debug: Print gist info
		if verbose {
			gistDesc := "no description"
			if g.Description != nil {
				gistDesc = *g.Description
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nProcessing gist: %s (%s)\n", *g.ID, gistDesc)
			fmt.Fprintf(cmd.OutOrStdout(), "  Files: %d\n", len(g.Files))
		}

		// Skip if gist has no files
		if len(g.Files) == 0 {
			if verbose {
				fmt.Fprintln(cmd.OutOrStdout(), "  Skipped: No files")
			}
			continue
		}

		// Get the full gist with content
		fullGist, err := gistClient.GetGist(cmd.Context(), *g.ID)
		if err != nil {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  Error fetching full gist: %v\n", err)
			}
			continue
		}

		// Get the first file (prompt-vault gists typically have one file)
		var fileName string
		var fileContent string
		for name, file := range fullGist.Files {
			fileName = string(name)
			if file.Content != nil {
				fileContent = *file.Content
			}
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  File: %s (size: %d bytes)\n", fileName, len(fileContent))
			}
			break // Take first file
		}

		// Skip if no content
		if fileContent == "" {
			if verbose {
				fmt.Fprintln(cmd.OutOrStdout(), "  Skipped: No content")
			}
			continue
		}

		// Skip index.json files
		if strings.HasSuffix(fileName, "-promptvault-index.json") || fileName == "index.json" {
			if verbose {
				fmt.Fprintln(cmd.OutOrStdout(), "  Skipped: Index file")
			}
			continue
		}

		// Check if it's a prompt-vault prompt (has YAML frontmatter)
		// Look for YAML frontmatter
		if !strings.HasPrefix(fileContent, "---\n") {
			if verbose {
				fmt.Fprintln(cmd.OutOrStdout(), "  Skipped: No YAML frontmatter")
			}
			continue
		}
		
		if verbose {
			// Show first 100 chars of content
			preview := fileContent
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Content preview: %s\n", strings.ReplaceAll(preview, "\n", "\\n"))
		}

		// Parse the prompt file to extract metadata and content
		prompt, err := parser.ParsePromptFile(fileContent)
		if err != nil {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  Error parsing prompt: %v\n", err)
			}
			continue
		}

		// Set gist-specific fields
		prompt.GistID = *g.ID
		prompt.UpdatedAt = g.UpdatedAt.Time

		// Create index entry with all metadata
		entry := models.IndexEntry{
			Name:        fileName,
			Author:      cfg.Username,
			GistID:      *g.ID,
			UpdatedAt:   g.UpdatedAt.Time,
			Category:    prompt.Category,
			Tags:        prompt.Tags,
			Version:     prompt.Version,
			Description: prompt.Description,
		}

		// Override with gist description if available
		if g.Description != nil {
			entry.Description = *g.Description
		}
		
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "  Parsed metadata - Category: %s, Tags: %v, Version: %s\n", 
				prompt.Category, prompt.Tags, prompt.Version)
		}

		// Check if this is a new or updated prompt
		isNew := true
		if index != nil {
			for _, oldEntry := range index.Entries {
				if oldEntry.GistID == entry.GistID {
					isNew = false
					if oldEntry.UpdatedAt.Before(entry.UpdatedAt) {
						updateCount++
					}
					break
				}
			}
		}
		if isNew {
			newCount++
		}

		// Add to new index
		newIndex.Entries = append(newIndex.Entries, entry)

		// Save to cache
		if err := cacheManager.SavePrompt(prompt); err != nil {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  Error saving to cache: %v\n", err)
			}
		} else {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  Saved to cache: %s\n", prompt.GistID)
			}
		}

		downloadCount++

		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "Found: %s\n", entry.Name)
		}
	}

	if len(newIndex.Entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prompts found in your GitHub account.")
		fmt.Fprintln(cmd.OutOrStdout(), "Upload prompts using 'pv upload' to get started.")
	}

	// Save the new index
	if err := cacheManager.SaveIndex(newIndex); err != nil {
		return errors.WrapWithMessage(err, "failed to save index")
	}

	// Show summary
	fmt.Fprintln(cmd.OutOrStdout(), "\nSync completed successfully!")
	fmt.Fprintf(cmd.OutOrStdout(), "- Found: %d prompts\n", downloadCount)
	fmt.Fprintf(cmd.OutOrStdout(), "- New: %d prompts\n", newCount)
	fmt.Fprintf(cmd.OutOrStdout(), "- Updated: %d prompts\n", updateCount)
	fmt.Fprintf(cmd.OutOrStdout(), "- Total prompts: %d\n", len(newIndex.Entries))

	// Always show where files were synced to
	fmt.Fprintf(cmd.OutOrStdout(), "\nSync cache files to %s\n", cachePath)

	if verbose {
		fmt.Fprintf(cmd.OutOrStdout(), "Last sync: %s\n", index.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}
