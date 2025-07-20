package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/grigri201/prompt-vault/internal/config"
)

// newConfigCmd creates the config command
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display configuration information",
		Long:  "Display current configuration values and configuration file location",
		RunE:  runConfig,
	}

	return cmd
}

func runConfig(cmd *cobra.Command, args []string) error {
	// Create config manager
	cfgManager := config.NewManager()

	// Get configuration
	cfg, err := cfgManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get configuration paths
	configPath := config.GetConfigPath()
	configDir := filepath.Dir(configPath)
	
	// Get cache path
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "."
	}
	cacheDir := filepath.Join(homeDir, ".cache", "prompt-vault")

	fmt.Fprintln(cmd.OutOrStdout(), "Configuration Information:")
	fmt.Fprintln(cmd.OutOrStdout(), "==========================")
	fmt.Fprintf(cmd.OutOrStdout(), "Config Directory: %s\n", configDir)
	fmt.Fprintf(cmd.OutOrStdout(), "Config File: %s\n", configPath)
	fmt.Fprintf(cmd.OutOrStdout(), "Cache Directory: %s\n", cacheDir)
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Current Settings:")
	fmt.Fprintf(cmd.OutOrStdout(), "  GitHub Username: %s\n", cfg.Username)
	if cfg.Token != "" {
		fmt.Fprintln(cmd.OutOrStdout(), "  GitHub Token: ****** (set)")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "  GitHub Token: (not set)")
	}
	
	// Display last sync time
	if !cfg.LastSync.IsZero() {
		fmt.Fprintf(cmd.OutOrStdout(), "  Last Sync: %s\n", cfg.LastSync.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "  Last Sync: Never")
	}

	return nil
}