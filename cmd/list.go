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
	var prompts, err = lc.store.List()
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
		fmt.Printf("  %s (by %s)\n", prompt.Name, prompt.Author)
	}
}

func NewListCommand(store infra.Store) ListCmd {
	lc := &list{store: store}
	return &cobra.Command{
		Use: "list",
		Run: lc.execute,
	}
}
