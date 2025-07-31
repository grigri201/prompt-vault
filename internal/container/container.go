package container

import (
	"context"

	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/paths"
	"github.com/grigri201/prompt-vault/internal/sync"
)

// Container holds all application dependencies
type Container struct {
	PathManager    *paths.PathManager
	CacheManager   interfaces.CacheManager
	ConfigManager  *config.Manager // Keep concrete type to avoid circular dependency
	AuthManager    interfaces.AuthManager
	GistClient     *gist.Client
	SyncManager    interfaces.SyncManager
	SyncMiddleware interfaces.SyncMiddleware
	initialized    bool
}

// NewContainer creates a new dependency container
func NewContainer() *Container {
	pathManager := paths.NewPathManager()

	container := &Container{
		PathManager:   pathManager,
		CacheManager:  cache.NewManagerWithPathManager(pathManager),
		ConfigManager: config.NewManagerWithPathManager(pathManager),
		AuthManager:   auth.NewManagerWithPath(pathManager.GetConfigPath()),
	}

	// Initialize sync components (they will be properly initialized later)
	syncManager := sync.NewManager(container.CacheManager, container.AuthManager, nil)
	container.SyncManager = syncManager
	container.SyncMiddleware = sync.NewSyncMiddleware(syncManager)

	return container
}

// NewTestContainer creates a container for testing with custom home directory
func NewTestContainer(homeDir string) *Container {
	pathManager := paths.NewPathManagerWithHome(homeDir)

	container := &Container{
		PathManager:   pathManager,
		CacheManager:  cache.NewManagerWithPathManager(pathManager),
		ConfigManager: config.NewManagerWithPathManager(pathManager),
		AuthManager:   auth.NewManagerWithPath(pathManager.GetConfigPath()),
	}

	// Initialize sync components (they will be properly initialized later)
	syncManager := sync.NewManager(container.CacheManager, container.AuthManager, nil)
	container.SyncManager = syncManager
	container.SyncMiddleware = sync.NewSyncMiddleware(syncManager)

	return container
}

// Initialize initializes all components in the container
func (c *Container) Initialize(ctx context.Context) error {
	if c.initialized {
		return nil
	}

	// Ensure directories exist
	if err := c.PathManager.EnsureCacheDir(); err != nil {
		return errors.NewFileSystemError("Container.Initialize", err)
	}

	if err := c.PathManager.EnsureConfigDir(); err != nil {
		return errors.NewFileSystemError("Container.Initialize", err)
	}

	// Initialize config manager
	cfg, err := c.ConfigManager.GetConfig()
	if err != nil {
		// If config doesn't exist, create a default one
		cfg = &config.Config{}
		if err := c.ConfigManager.SaveConfig(cfg); err != nil {
			return errors.WrapError("Container.Initialize", err)
		}
	}

	// Initialize Gist client if token is available
	if cfg.Token != "" {
		client, err := gist.NewClient(cfg.Token)
		if err != nil {
			return errors.WrapError("Container.Initialize", err)
		}
		c.GistClient = client

		// Update sync manager with gist client
		syncManager := sync.NewManager(c.CacheManager, c.AuthManager, client)
		c.SyncManager = syncManager
		c.SyncMiddleware = sync.NewSyncMiddleware(syncManager)
	}

	c.initialized = true
	return nil
}

// InitializeWithToken initializes the container and sets up the Gist client with the given token
func (c *Container) InitializeWithToken(ctx context.Context, token string) error {
	if err := c.Initialize(ctx); err != nil {
		return err
	}

	// Update the Gist client with the new token
	client, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapError("Container.InitializeWithToken", err)
	}
	c.GistClient = client

	// Update sync manager with new gist client
	syncManager := sync.NewManager(c.CacheManager, c.AuthManager, client)
	c.SyncManager = syncManager
	c.SyncMiddleware = sync.NewSyncMiddleware(syncManager)

	// Save the token to config
	cfg, err := c.ConfigManager.GetConfig()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.Token = token

	if err := c.ConfigManager.SaveConfig(cfg); err != nil {
		return errors.WrapError("Container.InitializeWithToken", err)
	}

	return nil
}

// IsInitialized returns whether the container has been initialized
func (c *Container) IsInitialized() bool {
	return c.initialized
}

// RequireGistClient returns the Gist client or an error if not available
func (c *Container) RequireGistClient() (*gist.Client, error) {
	if c.GistClient == nil {
		return nil, errors.NewAuthErrorMsg("Container.RequireGistClient", "not authenticated. Please run 'pv login' first")
	}
	return c.GistClient, nil
}

// Cleanup performs cleanup operations for all components
func (c *Container) Cleanup() error {
	// Currently, no cleanup is needed for these components
	// This method is here for future extensibility
	c.initialized = false
	return nil
}
