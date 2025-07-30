package cli

import (
	"fmt"

	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/spf13/cobra"
)

// AutoSyncCommands are commands that should trigger auto-sync after execution
var AutoSyncCommands = map[string]bool{
	"upload": true,
	"delete": true,
	"import": true,
}

// WrapWithAutoSync wraps a command to perform sync after execution
func WrapWithAutoSync(cmd *cobra.Command) {
	originalRunE := cmd.RunE

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Run the original command
		err := originalRunE(cmd, args)

		// Check if this command should trigger auto-sync
		if AutoSyncCommands[cmd.Name()] && err == nil {
			// Get auto-sync setting from config
			configManager := config.NewManager()
			cfg, configErr := configManager.GetConfig()
			if configErr == nil && cfg != nil {
				// Check if auto-sync is enabled (we can add this to config later)
				// For now, always sync after these commands
				if shouldAutoSync(cfg) {
					fmt.Fprintln(cmd.OutOrStderr(), "\nSyncing with GitHub...")
					syncErr := performFullSync(cmd)
					if syncErr != nil {
						fmt.Fprintf(cmd.OutOrStderr(), "Warning: Auto-sync failed: %v\n", syncErr)
					}
				}
			}
		}

		return err
	}
}

// shouldAutoSync checks if auto-sync is enabled
func shouldAutoSync(cfg *config.Config) bool {
	// TODO: Add auto_sync field to config
	// For now, always return true
	return true
}

// performAutoSync is deprecated - use performFullSync from sync_utils.go
// This function is kept for backward compatibility but should not be used
func performAutoSync(cmd *cobra.Command) error {
	return performFullSync(cmd)
}
