package cli

import (
	"github.com/spf13/cobra"
)

var (
	// Version is set during build
	Version = "dev"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pv",
		Short: "Prompt Vault - manage and reuse prompt templates",
		Long: `Prompt Vault is a command-line tool that enables you to manage, 
store, and reuse prompt templates through GitHub Gists.

Store your prompt templates privately in GitHub Gists, search and retrieve 
templates, fill in variables interactively, and copy the final prompt to 
your clipboard.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
	}

	// Add subcommands
	rootCmd.AddCommand(
		newLoginCmd(),
		newUploadCmd(),
		newListCmd(),
		newGetCmd(),
		newDeleteCmd(),
		newSyncCmd(),
	)

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCmd().Execute()
}

// Placeholder commands - will be implemented in subsequent tasks

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"del"},
		Short:   "Delete a prompt template",
		Long: `Delete a prompt template from your GitHub Gists.
This action requires confirmation and can only be performed on your own templates.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Delete command not yet implemented")
			return nil
		},
	}
}

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Synchronize local cache with GitHub Gists",
		Long: `Synchronize your local prompt cache with GitHub Gists.
This downloads all prompts from your index and updates the local cache.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Sync command not yet implemented")
			return nil
		},
	}
}