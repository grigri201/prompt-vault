package wire

import (
	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/paths"
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