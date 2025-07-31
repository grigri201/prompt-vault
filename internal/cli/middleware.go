package cli

import (
	"context"
	"fmt"

	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/spf13/cobra"
)

// WithSyncMiddleware wraps a command with the new sync middleware system
func WithSyncMiddleware(cmd *cobra.Command, cmdName string) *cobra.Command {
	originalRunE := cmd.RunE

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get command context
		cmdContext := GetCommandContext()
		if cmdContext == nil {
			return errors.NewValidationErrorMsg("WithSyncMiddleware", "command context not initialized")
		}

		container := cmdContext.Container

		// Initialize container if needed
		if !container.IsInitialized() {
			if err := container.Initialize(cmd.Context()); err != nil {
				return errors.WrapWithMessage(err, "failed to initialize container")
			}
		}

		syncMiddleware := container.SyncMiddleware
		ctx := context.Background()

		// Pre-sync: Execute before the main command
		if err := syncMiddleware.PreSync(ctx, cmdName); err != nil {
			// Pre-sync failure should be reported but not block the command
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: Pre-sync failed: %v\n", err)
		}

		// Execute the original command
		err := originalRunE(cmd, args)

		// Post-sync: Execute after the main command (even if it failed)
		if syncErr := syncMiddleware.PostSync(ctx, cmdName); syncErr != nil {
			if err != nil {
				// If the original command also failed, report sync error but return original error
				fmt.Fprintf(cmd.OutOrStderr(), "Warning: Post-sync failed: %v\n", syncErr)
			} else {
				fmt.Println("syncErr:", syncErr)
				// Check if this is the "no prompts found" error for new users
				if isNoPromptsFoundError(syncErr) {
					// Show this as a helpful suggestion, not an error
					fmt.Fprintf(cmd.OutOrStdout(), "💡 Suggestion: %s\n", extractErrorMessage(syncErr))
				} else {
					// If the original command succeeded but post-sync failed, return sync error
					err = errors.WrapError("middleware.PostSync", syncErr)
				}
			}
		}

		return err
	}

	return cmd
}

// isNoPromptsFoundError checks if the error is the "no prompts found" validation error
func isNoPromptsFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "no prompts found") && contains(errStr, "pv add")
}

// extractErrorMessage extracts the user-friendly message from an error
func extractErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	errStr := err.Error()

	// Look for the pattern "ValidationError: <component>: <message>"
	if idx := findSubstring(errStr, ": "); idx >= 0 {
		remaining := errStr[idx+2:]
		if nextIdx := findSubstring(remaining, ": "); nextIdx >= 0 {
			return remaining[nextIdx+2:]
		}
		return remaining
	}

	return errStr
}
