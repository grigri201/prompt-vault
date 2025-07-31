package wire

import (
	"github.com/google/wire"
	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/paths"
	"github.com/grigri201/prompt-vault/internal/sync"
)

// ProvidePathManager provides a PathManager instance
func ProvidePathManager() *paths.PathManager {
	return paths.NewPathManager()
}

// ProvideCacheManager provides a CacheManager instance
func ProvideCacheManager(pathManager *paths.PathManager) interfaces.CacheManager {
	return cache.NewManagerWithPathManager(pathManager)
}

// ProvideConfigManager provides a ConfigManager instance
func ProvideConfigManager(pathManager *paths.PathManager) *config.Manager {
	return config.NewManagerWithPathManager(pathManager)
}

// ProvideAuthManager provides an AuthManager instance
func ProvideAuthManager(pathManager *paths.PathManager) interfaces.AuthManager {
	return auth.NewManagerWithPath(pathManager.GetConfigPath())
}

// ProvideGistClient provides a GistClient instance
// This will return nil if no token is configured
func ProvideGistClient(configManager *config.Manager) *gist.Client {
	cfg, err := configManager.GetConfig()
	if err != nil {
		// Return nil client if config doesn't exist
		return nil
	}

	if cfg.Token == "" {
		// Return nil client if no token
		return nil
	}

	client, err := gist.NewClient(cfg.Token)
	if err != nil {
		// Return nil if client creation fails
		return nil
	}

	return client
}

// ProvideSyncManager provides a SyncManager instance
func ProvideSyncManager(cacheManager interfaces.CacheManager, authManager interfaces.AuthManager, gistClient *gist.Client) *sync.Manager {
	return sync.NewManager(cacheManager, authManager, gistClient)
}

// ProvideSyncMiddleware provides a SyncMiddleware instance
func ProvideSyncMiddleware(syncManager *sync.Manager) *sync.Middleware {
	return sync.NewSyncMiddleware(syncManager)
}

// Wire Sets for organizing Providers
var SyncSet = wire.NewSet(
	ProvideSyncManager,
	ProvideSyncMiddleware,
	wire.Bind(new(interfaces.SyncManager), new(*sync.Manager)),
	wire.Bind(new(interfaces.SyncMiddleware), new(*sync.Middleware)),
)
