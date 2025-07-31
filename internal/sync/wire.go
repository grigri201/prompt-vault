package sync

import (
	"github.com/google/wire"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
)

// ProvideManager provides a sync manager instance
func ProvideManager(
	cache interfaces.CacheManager,
	authManager interfaces.AuthManager,
	gistClient *gist.Client,
) *Manager {
	return NewManager(cache, authManager, gistClient)
}

// ProvideMiddleware provides a sync middleware instance
func ProvideMiddleware(manager *Manager) *Middleware {
	return NewSyncMiddleware(manager)
}

// ProviderSet is the Wire provider set for sync module
var ProviderSet = wire.NewSet(
	ProvideManager,
	ProvideMiddleware,
	wire.Bind(new(interfaces.SyncManager), new(*Manager)),
	wire.Bind(new(interfaces.SyncMiddleware), new(*Middleware)),
)
