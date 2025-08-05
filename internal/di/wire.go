//go:build wireinject
// +build wireinject

package di

import (
	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/auth"
	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/validator"

	"github.com/google/wire"
)

// InfraSet provides infrastructure components
var InfraSet = wire.NewSet(
	infra.NewGitHubStore,
	config.NewFileStore,
)

// AuthSet provides authentication related components
var AuthSet = wire.NewSet(
	auth.NewGitHubClient,
	auth.NewTokenValidator,
	service.NewAuthService,
)

// ServiceSet provides service layer components
var ServiceSet = wire.NewSet(
	validator.NewYAMLValidator,
	service.NewPromptService,
)

// GetCommandSet provides components specific to the get command
var GetCommandSet = wire.NewSet(
	ProvideClipboardUtil,
	ProvideVariableParser,
	ProvideTUIInterface,
)

// CommandSet provides CLI commands
var CommandSet = wire.NewSet(
	ProvideCommands,
	ProvideRootCommand,
)

func BuildCLI() (*cobra.Command, error) {
	wire.Build(InfraSet, AuthSet, ServiceSet, GetCommandSet, CommandSet)
	return nil, nil
}
