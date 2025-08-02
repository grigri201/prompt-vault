package cmd

import "github.com/spf13/cobra"

// AuthCmd is the auth command type
type AuthCmd = *cobra.Command

// NewAuthCommand creates the auth parent command
func NewAuthCommand(loginCmd AuthLoginCmd, statusCmd AuthStatusCmd, logoutCmd AuthLogoutCmd) AuthCmd {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage GitHub authentication",
		Long: `Manage GitHub authentication for accessing GitHub Gist API.

Use 'pv auth login' to authenticate with a GitHub Personal Access Token.
Use 'pv auth status' to check your authentication status.
Use 'pv auth logout' to clear your authentication.`,
	}

	// Add subcommands
	cmd.AddCommand(loginCmd, statusCmd, logoutCmd)

	return cmd
}
