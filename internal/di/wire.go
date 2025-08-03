//go:build wireinject
// +build wireinject

package di

import (
	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/auth"
	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/service"

	"github.com/google/wire"
)

// InfraSet provides infrastructure components
var InfraSet = wire.NewSet(
	infra.NewMemoryStore,
	config.NewFileStore,
)

// AuthSet provides authentication related components
var AuthSet = wire.NewSet(
	auth.NewGitHubClient,
	auth.NewTokenValidator,
	service.NewAuthService,
)

// CommandSet provides CLI commands
var CommandSet = wire.NewSet(
	ProvideCommands,
	ProvideRootCommand,
)

func BuildCLI() (*cobra.Command, error) {
	wire.Build(InfraSet, AuthSet, CommandSet)
	return nil, nil
}
