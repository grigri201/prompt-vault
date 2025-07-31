package cli

import (
	"context"
	"log"

	"github.com/grigri201/prompt-vault/internal/container"
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
		SilenceUsage:     true,
		SilenceErrors:    true,
		Version:          Version,
		PersistentPreRun: setupVerboseLogging,
	}

	// Add global persistent flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Help for any command")

	// Add subcommands
	rootCmd.AddCommand(
		NewLoginCommand(),
		NewAddCommand(),
		NewGetCommand(),
		NewShareCommand(),
		NewDelCommand(),
		NewSyncCommand(),
		NewConfigCommand(),
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

// setupVerboseLogging configures verbose logging based on the flag
func setupVerboseLogging(cmd *cobra.Command, args []string) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		// Enable verbose logging
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose mode enabled")
	}
}

// Placeholder commands - will be implemented in subsequent tasks
