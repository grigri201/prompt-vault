package di

import (
	"github.com/spf13/cobra"

	"github.com/grigri/pv/cmd"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/validator"
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
	AddCmd  *cobra.Command
	AuthCmd *cobra.Command
}

// ProvideCommands provides all commands
func ProvideCommands(store infra.Store, authService service.AuthService, promptService service.PromptService) Commands {
	listCmd := cmd.NewListCommand(store)
	addCmd := cmd.NewAddCommand(promptService)
	authCmd := ProvideAuthCommands(authService)
	return Commands{
		ListCmd: listCmd,
		AddCmd:  addCmd,
		AuthCmd: authCmd,
	}
}

// ProvideYAMLValidator provides a YAML validator instance
func ProvideYAMLValidator() validator.YAMLValidator {
	return validator.NewYAMLValidator()
}

// ProvidePromptService provides a PromptService instance with dependencies
func ProvidePromptService(store infra.Store, validator validator.YAMLValidator) service.PromptService {
	return service.NewPromptService(store, validator)
}

// ProvideRootCommand provides the root command with all subcommands
func ProvideRootCommand(commands Commands) *cobra.Command {
	return cmd.NewRootCommand(commands.ListCmd, commands.AddCmd, commands.AuthCmd)
}
