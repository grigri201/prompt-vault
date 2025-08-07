package di

import (
	"github.com/spf13/cobra"

	"github.com/grigri/pv/cmd"
	"github.com/grigri/pv/internal/clipboard"
	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/tui"
	"github.com/grigri/pv/internal/validator"
	"github.com/grigri/pv/internal/variable"
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
	ListCmd   *cobra.Command
	AddCmd    *cobra.Command
	DeleteCmd *cobra.Command
	GetCmd    *cobra.Command
	SyncCmd   *cobra.Command
	AuthCmd   *cobra.Command
	ShareCmd  *cobra.Command
}

// ProvideCommands provides all commands
func ProvideCommands(
	store infra.Store, 
	configStore config.Store,
	authService service.AuthService, 
	promptService service.PromptService,
	clipboardUtil clipboard.Util,
	variableParser variable.Parser,
	tuiInterface tui.TUIInterface,
) Commands {
	listCmd := cmd.NewListCommand(store, configStore)
	addCmd := cmd.NewAddCommand(promptService)
	deleteCmd := cmd.NewDeleteCommand(store, promptService)
	getCmd := cmd.NewGetCommand(promptService, clipboardUtil, variableParser, tuiInterface)
	syncCmd := cmd.NewSyncCommand(promptService)
	authCmd := ProvideAuthCommands(authService)
	shareCmd := cmd.NewShareCommand(promptService, tuiInterface)
	return Commands{
		ListCmd:   listCmd,
		AddCmd:    addCmd,
		DeleteCmd: deleteCmd,
		GetCmd:    getCmd,
		SyncCmd:   syncCmd,
		AuthCmd:   authCmd,
		ShareCmd:  shareCmd,
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

// ProvideClipboardUtil provides a clipboard utility instance
func ProvideClipboardUtil() clipboard.Util {
	return clipboard.NewUtil()
}

// ProvideVariableParser provides a variable parser instance
func ProvideVariableParser() variable.Parser {
	return variable.NewParser()
}

// ProvideTUIInterface provides a TUI interface instance
func ProvideTUIInterface() tui.TUIInterface {
	return tui.NewBubbleTeaTUI()
}

// ProvideCacheManager provides a CacheManager instance
func ProvideCacheManager() (*infra.CacheManager, error) {
	return infra.NewCacheManager()
}

// ProvideGitHubStore provides a GitHubStore instance (not bound to Store interface)
func ProvideGitHubStore(configStore config.Store) *infra.GitHubStore {
	// Use the existing NewGitHubStore and perform type assertion
	// This is safe because we know NewGitHubStore returns *GitHubStore wrapped in Store interface
	return infra.NewGitHubStore(configStore).(*infra.GitHubStore)
}

// ProvideCachedStore provides a CachedStore instance that wraps GitHubStore with caching
func ProvideCachedStore(gitHubStore *infra.GitHubStore, cacheManager *infra.CacheManager, configStore config.Store) infra.Store {
	// Default to forceRemote = false for normal operations
	// Commands can override this behavior as needed
	return infra.NewCachedStore(gitHubStore, cacheManager, configStore, false)
}

// ProvideRootCommand provides the root command with all subcommands
func ProvideRootCommand(commands Commands) *cobra.Command {
	return cmd.NewRootCommand(commands.ListCmd, commands.AddCmd, commands.DeleteCmd, commands.GetCmd, commands.SyncCmd, commands.AuthCmd, commands.ShareCmd)
}
