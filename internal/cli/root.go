package cli

import (
	"context"
	
	"github.com/spf13/cobra"
	"github.com/grigri201/prompt-vault/internal/container"
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
		newConfigCmd(),
		newShareCommand(),
		newImportCommand(),
	)

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	c := container.NewContainer()
	if err := c.Initialize(context.Background()); err != nil {
		return err
	}
	defer c.Cleanup()
	
	// Set global command context
	SetCommandContext(NewCommandContext(c))
	
	return NewRootCmd().Execute()
}

// Placeholder commands - will be implemented in subsequent tasks
