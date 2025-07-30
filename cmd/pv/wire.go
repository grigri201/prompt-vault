//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/container"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/paths"
	pvwire "github.com/grigri201/prompt-vault/internal/wire"
)

// buildContainer creates a new dependency container using wire
func buildContainer() *container.Container {
	wire.Build(
		pvwire.ProvidePathManager,
		pvwire.ProvideCacheManager,
		pvwire.ProvideConfigManager,
		pvwire.ProvideAuthManager,
		pvwire.ProvideGistClient,
		provideContainer,
	)
	return nil
}

// provideContainer creates a Container with all dependencies
func provideContainer(
	pathManager *paths.PathManager,
	cacheManager interfaces.CacheManager,
	configManager *config.Manager,
	authManager interfaces.AuthManager,
	gistClient *gist.Client,
) *container.Container {
	return &container.Container{
		PathManager:   pathManager,
		CacheManager:  cacheManager,
		ConfigManager: configManager,
		AuthManager:   authManager,
		GistClient:    gistClient,
		// initialized will be false by default
	}
}