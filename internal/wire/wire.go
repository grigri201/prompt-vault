//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/grigri201/prompt-vault/internal/container"
)

// InitializeContainer creates a new dependency container using wire
func InitializeContainer() (*container.Container, error) {
	wire.Build(
		ProvidePathManager,
		ProvideCacheManager,
		ProvideConfigManager,
		ProvideAuthManager,
		ProvideGistClient,
		wire.Struct(new(container.Container), "*"),
	)
	return nil, nil
}

// InitializeTestContainer creates a container for testing with custom home directory
func InitializeTestContainer(homeDir string) (*container.Container, error) {
	wire.Build(
		ProvideTestPathManager,
		ProvideCacheManager,
		ProvideConfigManager,
		ProvideAuthManager,
		ProvideGistClient,
		wire.Struct(new(container.Container), "*"),
	)
	return nil, nil
}

// ProvideTestPathManager provides a PathManager instance for testing
func ProvideTestPathManager(homeDir string) *paths.PathManager {
	return paths.NewPathManagerWithHome(homeDir)
}
