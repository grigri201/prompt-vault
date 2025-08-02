package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/infra"
)

type ListCmd *cobra.Command

type list struct {
	store infra.Store
}

func (lc *list) execute(cmd *cobra.Command, args []string) {
	prompts, err := lc.store.List()
	if err != nil {
		// Handle friendly error messages for empty/missing index
		if errors.Is(err, infra.ErrNoIndex) {
			fmt.Println("📝 Welcome to Prompt Vault!")
			fmt.Println()
			fmt.Println("It looks like this is your first time using pv. Your prompt collection is empty.")
			fmt.Println()
			fmt.Println("To get started:")
			fmt.Println("  • Create prompts directly in GitHub Gists")
			fmt.Println("  • Use 'pv add <name>' to create a new prompt")
			fmt.Println("  • Run 'pv list' again to see your prompts")
			return
		}

		if errors.Is(err, infra.ErrEmptyIndex) {
			fmt.Println("📝 Your prompt collection is currently empty.")
			fmt.Println()
			fmt.Println("To add prompts:")
			fmt.Println("  • Create prompts directly in GitHub Gists")
			fmt.Println("  • Use 'pv add <name>' to create a new prompt")
			fmt.Println("  • Run 'pv list' again to see your prompts")
			return
		}

		// For other errors, show the original error
		log.Fatalf("error in get prompts: %v", err)
	}

	// Display prompts if we have any
	if len(prompts) == 0 {
		fmt.Println("📝 No prompts found in your collection.")
		return
	}

	fmt.Printf("📝 Found %d prompt(s):\n\n", len(prompts))
	for i := range prompts {
		var prompt = prompts[i]
		fmt.Printf("  %s - author: %s : %s\n", prompt.Name, prompt.Author, prompt.GistURL)
	}
}

func NewListCommand(store infra.Store) ListCmd {
	lc := &list{store: store}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all prompts in your collection",
		Long: `List all prompts in your collection.

This command always fetches data from GitHub to ensure the latest state.`,
		Run: lc.execute,
	}

	return listCmd
}
