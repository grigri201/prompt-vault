package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/errors"
)

// NewSyncCommand creates the sync command
func NewSyncCommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Manually trigger complete synchronization",
		Long: `Manually trigger complete synchronization between local and remote data.

The sync command compares timestamps and synchronizes data in both directions:
- If remote is newer: downloads remote index and prompts
- If local is newer: uploads local changes to remote
- If timestamps are equal: no synchronization needed

Examples:
  pv sync                    # Standard synchronization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSyncCommand(cmd, args, verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress")

	return cmd
}

func runSyncCommand(cmd *cobra.Command, args []string, verbose bool) error {
	// Get command context
	cmdContext := GetCommandContext()
	if cmdContext == nil {
		return errors.NewValidationErrorMsg("runSyncCommand", "command context not initialized")
	}

	container := cmdContext.Container

	// Initialize container if needed
	if !container.IsInitialized() {
		if err := container.Initialize(cmd.Context()); err != nil {
			return errors.WrapWithMessage(err, "failed to initialize container")
		}
	}

	// Check authentication
	if container.GistClient == nil {
		return errors.NewAuthErrorMsg("sync", "not authenticated. Please run 'pv login' first")
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Starting synchronization...")

	ctx := context.Background()

	// Get sync manager and initialize if needed
	syncManager := container.SyncManager
	if !syncManager.IsInitialized() {
		if err := syncManager.Initialize(ctx); err != nil {
			return errors.WrapWithMessage(err, "failed to initialize sync manager")
		}
	}

	// Perform synchronization
	err := syncManager.SynchronizeData(ctx)
	if err != nil {
		return errors.WrapError("runSyncCommand", err)
	}

	// Get sync status and display results
	status := syncManager.GetSyncStatus()

	fmt.Fprintln(cmd.OutOrStdout(), "\nSync completed successfully!")
	fmt.Fprintf(cmd.OutOrStdout(), "  Local time: %v\n", status.LocalTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(cmd.OutOrStdout(), "  Remote time: %v\n", status.RemoteTime.Format("2006-01-02 15:04:05"))

	if status.NeedsSync {
		fmt.Fprintf(cmd.OutOrStdout(), "  Status: %s\n", status.DisplayString())
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "  Status: %s\n", status.DisplayString())
	}

	if verbose && status.Progress.Total > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Progress: %d/%d items processed\n",
			status.Progress.Completed, status.Progress.Total)
	}

	return nil
}
