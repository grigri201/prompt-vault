//go:build wireinject
// +build wireinject

package di

import (
	"pv/cmd"
	"pv/internal/infra"

	"github.com/google/wire"
)

var StoreSet = wire.NewSet(
	infra.NewMemoryStore,
	cmd.NewListCommand,
	cmd.NewRootCommand,
)

func BuildCLI() (cmd.RootCmd, error) {
	wire.Build(StoreSet)
	return nil, nil
}
