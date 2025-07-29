package cli

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/search"
	"github.com/grigri201/prompt-vault/internal/share"
	"github.com/grigri201/prompt-vault/internal/ui"
	"github.com/spf13/cobra"
)

// isGistID checks if the input string is a valid GitHub gist ID
// GitHub gist IDs are 32-character hexadecimal strings
func isGistID(input string) bool {
	if len(input) != 32 {
		return false
	}
	
	for _, c := range input {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	
	return true
}

// formatTags formats a slice of tags into a comma-separated string
func formatTags(tags []string) string {
	return strings.Join(tags, ", ")
}

type shareOptions struct {
	gistID string
}

func newShareCommand() *cobra.Command {
	opts := &shareOptions{}

	cmd := &cobra.Command{
		Use:   "share [<gist-id>|<keyword>]",
		Short: "Share a private prompt as a public gist",
		Long: `Share a private prompt as a public gist. This creates a public copy of your private prompt
that can be shared with others. The original private gist remains unchanged.

Usage:
  pv share                   # List all prompts and select one to share
  pv share <keyword>         # Search for prompts matching keyword and select one to share
  pv share <gist-id>         # Share a specific prompt by its gist ID

If a public version already exists, you'll be prompted to update it.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.gistID = args[0]
			}
			return runShare(cmd, opts)
		},
	}

	return cmd
}

// newShareCmd creates a share command with injected dependencies (for testing)
func newShareCmd(manager shareManager) *cobra.Command {
	opts := &shareOptions{}

	cmd := &cobra.Command{
		Use:   "share [<gist-id>|<keyword>]",
		Short: "Share a private prompt as a public gist",
		Long: `Share a private prompt as a public gist. This creates a public copy of your private prompt
that can be shared with others. The original private gist remains unchanged.

Usage:
  pv share                   # List all prompts and select one to share
  pv share <keyword>         # Search for prompts matching keyword and select one to share
  pv share <gist-id>         # Share a specific prompt by its gist ID

If a public version already exists, you'll be prompted to update it.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.gistID = args[0]
			}
			return runShareWithManager(cmd, opts, manager)
		},
	}

	return cmd
}

// shareManager interface for dependency injection
type shareManager interface {
	SharePrompt(ctx context.Context, privateGistID string) (*share.ShareResult, error)
}

func runShare(cmd *cobra.Command, opts *shareOptions) error {
	// Initialize dependencies
	authManager := auth.NewManager()
	token, username, err := authManager.GetToken()
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgAuthRequired)
	}

	gistClient, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgShareNetworkError)
	}

	// Create UI adapter
	uiAdapter := &cliUI{cmd: cmd}

	// Create share manager
	manager := share.NewManager(gistClient, uiAdapter, username)

	return runShareWithManager(cmd, opts, manager)
}

// cliUI implements the UI interface for share.Manager
type cliUI struct {
	cmd *cobra.Command
}

func (u *cliUI) Confirm(message string) (bool, error) {
	// Print the message
	u.cmd.Println(message)
	
	// Create a selector model for confirmation
	choices := []string{"Yes", "No"}
	model := ui.NewSelector(choices)

	// Run the bubble tea program
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("failed to run confirmation prompt: %w", err)
	}

	// Check if user selected "Yes"
	selectorModel := finalModel.(*ui.SelectorModel)
	return selectorModel.IsConfirmed() && selectorModel.GetSelection() == "Yes", nil
}

func runShareWithManager(cmd *cobra.Command, opts *shareOptions, manager shareManager) error {
	ctx := context.Background()

	// If no gist ID provided, show prompt selection
	if opts.gistID == "" {
		// Get cache manager
		cachePath := getCachePathFunc()
		cacheManager := cache.NewManagerWithPath(cachePath)
		
		// Load index
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
		
		// Display all prompts for selection
		fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompt(s) to share:\n\n", len(index.Entries))
		
		// Create display items
		items := make([]string, len(index.Entries))
		for i, entry := range index.Entries {
			items[i] = fmt.Sprintf("[%d] %s by %s", i+1, entry.Name, entry.Author)
		}
		
		// Show details for each prompt
		for i, entry := range index.Entries {
			fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s by %s\n", i+1, entry.Name, entry.Author)
			fmt.Fprintf(cmd.OutOrStdout(), "   Category: %s\n", entry.Category)
			if len(entry.Tags) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "   Tags: %s\n", formatTags(entry.Tags))
			}
			if entry.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "   Description: %s\n", entry.Description)
			}
			if i < len(index.Entries)-1 {
				fmt.Fprintln(cmd.OutOrStdout())
			}
		}
		
		// Create selector for prompts
		selector := ui.NewSelector(items)
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
		
		// Get the selected prompt's gist ID
		selectedIdx := selectorModel.Selected
		opts.gistID = index.Entries[selectedIdx].GistID
	} else if !isGistID(opts.gistID) {
		// If argument is not a gist ID, treat it as a keyword search
		cachePath := getCachePathFunc()
		cacheManager := cache.NewManagerWithPath(cachePath)
		
		// Load index
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
		
		// Search for matching prompts
		searcher := search.Searcher{}
		matches := searcher.SearchEntries(index.Entries, opts.gistID)
		
		if len(matches) == 0 {
			return fmt.Errorf("No prompts found matching '%s'", opts.gistID)
		}
		
		// Show all matches with selector (even for single match)
		if len(matches) >= 1 {
			fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompt(s) matching '%s':\n\n", len(matches), opts.gistID)
			
			// Create display items
			items := make([]string, len(matches))
			for i, matchIdx := range matches {
				entry := index.Entries[matchIdx]
				items[i] = fmt.Sprintf("[%d] %s by %s", i+1, entry.Name, entry.Author)
			}
			
			// Show details for each matching prompt
			for i, matchIdx := range matches {
				entry := index.Entries[matchIdx]
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s by %s\n", i+1, entry.Name, entry.Author)
				fmt.Fprintf(cmd.OutOrStdout(), "   Category: %s\n", entry.Category)
				if len(entry.Tags) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "   Tags: %s\n", formatTags(entry.Tags))
				}
				if entry.Description != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "   Description: %s\n", entry.Description)
				}
				if i < len(matches)-1 {
					fmt.Fprintln(cmd.OutOrStdout())
				}
			}
			
			// Create selector for prompts
			selector := ui.NewSelector(items)
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
			
			// Get the selected prompt's gist ID
			selectedIdx := selectorModel.Selected
			opts.gistID = index.Entries[matches[selectedIdx]].GistID
		}
	}

	// Share the prompt
	result, err := manager.SharePrompt(ctx, opts.gistID)
	if err != nil {
		errMsg := errors.GetShareErrorMessage(err)
		cmd.PrintErrf("%s\n", errMsg)
		return err
	}

	// Display result
	if result.IsUpdate {
		cmd.Printf("Successfully updated public gist: %s\n", result.PublicGistURL)
	} else {
		cmd.Printf("Successfully created public gist: %s\n", result.PublicGistURL)
	}

	return nil
}