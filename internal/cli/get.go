package cli

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/clipboard"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/parser"
	"github.com/grigri201/prompt-vault/internal/search"
	"github.com/grigri201/prompt-vault/internal/ui"
)

// NewGetCommand creates the get command
func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [keyword]",
		Short: "Search and retrieve prompt templates",
		Long: `Search for prompt templates by keyword across names, categories, 
tags, authors, and descriptions. Select a template to fill in variables 
interactively and copy the result to your clipboard.

Examples:
  pv get "code review"                    # Search and display
  pv get "deployment" --output deploy.md  # Output to file`,
		Args: cobra.MaximumNArgs(1),
		RunE: runGet,
	}

	cmd.Flags().StringP("output", "o", "", "Output to file instead of clipboard")

	// Integrate sync middleware
	return WithSyncMiddleware(cmd, "get")
}

func runGet(cmd *cobra.Command, args []string) error {
	// Get output file flag
	outputFile, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
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

	// If no keyword provided, show all
	keyword := ""
	if len(args) > 0 {
		keyword = args[0]
	}

	// Use search package to find matches
	searcher := search.NewSearcher()
	matches := searcher.SearchEntries(index.Entries, keyword)

	// Check if any matches found
	if len(matches) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No prompts found matching '%s'.\n", keyword)
		return nil
	}

	// Display matches
	fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompt(s):\n\n", len(matches))

	// Create selector items
	selectorItems := make([]string, len(matches))
	for i, idx := range matches {
		entry := index.Entries[idx]
		fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s by %s\n", i+1, entry.Name, entry.Author)
		if len(entry.Tags) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "    Tags: %s\n", strings.Join(entry.Tags, ", "))
		}
		if entry.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "    Description: %s\n", entry.Description)
		}
		if entry.GistURL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "    Gist URL: %s\n", entry.GistURL)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "    Gist URL: ")
		}
		fmt.Fprintln(cmd.OutOrStdout())

		// Build selector item
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

	// Get the selected prompt entry
	selectedIdx := matches[selectorModel.Selected]
	selectedEntry := index.Entries[selectedIdx]

	// Load the prompt content
	prompt, err := cacheManager.GetPrompt(selectedEntry.GistID)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to load prompt")
	}

	// Display selected prompt details with Gist URL
	fmt.Fprintln(cmd.OutOrStdout(), "\nSelected prompt:")
	fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", selectedEntry.Name)
	if selectedEntry.GistURL != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Gist URL: %s\n", selectedEntry.GistURL)
	}

	// Parse variables in the prompt
	vars := parser.ExtractVariables(prompt.Content)

	// If there are variables, show form to collect values
	var finalContent string
	if len(vars) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\nPrompt contains %d variable(s) to fill:\n", len(vars))

		// Create and run form
		form := ui.NewForm("Fill in the variables", vars)
		formProgram := tea.NewProgram(form)
		formModel, err := formProgram.Run()
		if err != nil {
			return errors.WrapWithMessage(err, "failed to run form")
		}

		// Check if form was submitted
		finalForm := formModel.(ui.FormModel)
		if !finalForm.IsSubmitted() {
			fmt.Fprintln(cmd.OutOrStdout(), "\nForm cancelled.")
			return nil
		}

		// Replace variables with values
		values := finalForm.GetValues()
		finalContent = parser.FillVariables(prompt.Content, values)
	} else {
		finalContent = prompt.Content
	}

	// Handle output based on flag
	if outputFile != "" {
		// Write to file
		err = os.WriteFile(outputFile, []byte(finalContent), 0644)
		if err != nil {
			return errors.WrapWithMessage(err, fmt.Sprintf("failed to write to file %s", outputFile))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n✓ Prompt written to file: %s\n", outputFile)
		if selectedEntry.GistURL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Gist URL: %s\n", selectedEntry.GistURL)
		}
	} else {
		// Copy to clipboard
		err = clipboard.Copy(finalContent)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "\nWarning: Failed to copy to clipboard: %v\n", err)
			fmt.Fprintln(cmd.OutOrStdout(), "\nPrompt content:")
			fmt.Fprintln(cmd.OutOrStdout(), finalContent)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "\n✓ Prompt copied to clipboard!")
			if selectedEntry.GistURL != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Gist URL: %s\n", selectedEntry.GistURL)
			}
		}
	}

	return nil
}
