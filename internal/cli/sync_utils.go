package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/sync"
)

// For testing - common cache path function
var getCachePathFunc = cache.GetCachePath

// createManagers creates config and cache managers with consistent setup
func createManagers() (*config.Manager, interfaces.CacheManager) {
	ctx := GetCommandContext()
	return ctx.Container.ConfigManager, ctx.Container.CacheManager
}

// createManagersWithPath creates config and cache managers with specific cache path
func createManagersWithPath(cachePath string) (*config.Manager, interfaces.CacheManager) {
	// For now, just use the standard managers since Wire doesn't support dynamic paths easily
	return createManagers()
}

// performPreSync performs full synchronization before commands that need current data
// This is extracted from upload.go to be reused by import and other commands
func performPreSync(cmd *cobra.Command, force bool) error {
	// Create managers
	configManager, cacheManager := createManagers()

	// Sync with GitHub before operation
	fmt.Fprintf(cmd.OutOrStdout(), "Syncing with GitHub...\n")
	// Type assert cache manager for sync service
	cacheManagerImpl, ok := cacheManager.(*cache.Manager)
	if !ok {
		return errors.NewValidationErrorMsg("presync", "invalid cache manager type")
	}
	syncService := sync.NewService(configManager, cacheManagerImpl)

	syncCtx := context.Background()
	syncResult, err := syncService.SyncWithTimeout(syncCtx, 30*time.Second)
	if err != nil {
		// Handle sync failure
		if !force {
			// In interactive mode, ask user if they want to continue
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Failed to sync with GitHub: %v\n", err)
			fmt.Fprintf(cmd.OutOrStdout(), "Continue with potentially outdated data? (y/N): ")

			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				return errors.NewValidationErrorMsg("presync", "operation cancelled by user")
			}
		} else {
			// In force mode, just warn and continue
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Failed to sync with GitHub: %v\n", err)
			fmt.Fprintf(cmd.ErrOrStderr(), "Continuing with potentially outdated data due to --force flag\n")
		}
	} else {
		// Display sync results
		fmt.Fprintf(cmd.OutOrStdout(), "Synced %d prompts\n", syncResult.TotalPrompts)
	}

	return nil
}

// performFullSync performs full bidirectional synchronization
// This replaces the old performAutoSync to provide real sync instead of just pushing to GitHub
func performFullSync(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create managers
	configManager, cacheManager := createManagers()

	// Get config to check authentication
	cfg, err := configManager.GetConfig()
	if err != nil {
		return err
	}

	if cfg.Token == "" {
		return nil // Not authenticated, skip sync
	}

	// Perform full bidirectional sync using sync service
	// Type assert cache manager for sync service
	cacheManagerImpl, ok := cacheManager.(*cache.Manager)
	if !ok {
		return errors.NewValidationErrorMsg("presync", "invalid cache manager type")
	}
	syncService := sync.NewService(configManager, cacheManagerImpl)
	_, err = syncService.SyncWithTimeout(ctx, 30*time.Second)
	if err != nil {
		// Check if it's just a GitHub index update failure
		if strings.Contains(err.Error(), "failed to update GitHub index (local sync completed successfully)") {
			// Local sync succeeded, just couldn't update GitHub
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: %v\n", err)
			fmt.Fprintln(cmd.OutOrStderr(), "✓ Local sync completed (GitHub index will be updated on next sync)")
			return nil
		}
		return err
	}

	fmt.Fprintln(cmd.OutOrStderr(), "✓ Full sync completed")
	return nil
}
