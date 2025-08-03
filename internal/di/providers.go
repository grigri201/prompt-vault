package di

import (
	"github.com/spf13/cobra"

	"github.com/grigri/pv/cmd"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/service"
)

// ProvideAuthCommands provides all auth-related commands as a single AuthCmd
func ProvideAuthCommands(authService service.AuthService) *cobra.Command {
	loginCmd := cmd.NewAuthLoginCommand(authService)
	statusCmd := cmd.NewAuthStatusCommand(authService)
	logoutCmd := cmd.NewAuthLogoutCommand(authService)

	return cmd.NewAuthCommand(loginCmd, statusCmd, logoutCmd)
}

// Commands holds all the subcommands
type Commands struct {
	ListCmd *cobra.Command
	AuthCmd *cobra.Command
}

// ProvideCommands provides all commands
func ProvideCommands(store infra.Store, authService service.AuthService) Commands {
	listCmd := cmd.NewListCommand(store)
	authCmd := ProvideAuthCommands(authService)
	return Commands{
		ListCmd: listCmd,
		AuthCmd: authCmd,
	}
}

// ProvideRootCommand provides the root command with all subcommands
func ProvideRootCommand(commands Commands) *cobra.Command {
	return cmd.NewRootCommand(commands.ListCmd, commands.AuthCmd)
}
