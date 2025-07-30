package container

import (
	"context"
	"fmt"

	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/paths"
)

// Container holds all application dependencies
type Container struct {
	PathManager   *paths.PathManager
	CacheManager  interfaces.CacheManager
	ConfigManager *config.Manager // Keep concrete type to avoid circular dependency
	AuthManager   interfaces.AuthManager
	GistClient    *gist.Client
	initialized   bool
}

// NewContainer creates a new dependency container
func NewContainer() *Container {
	pathManager := paths.NewPathManager()

	return &Container{
		PathManager:   pathManager,
		CacheManager:  cache.NewManagerWithPathManager(pathManager),
		ConfigManager: config.NewManagerWithPathManager(pathManager),
		AuthManager:   auth.NewManagerWithPath(pathManager.GetConfigPath()),
	}
}

// NewTestContainer creates a container for testing with custom home directory
func NewTestContainer(homeDir string) *Container {
	pathManager := paths.NewPathManagerWithHome(homeDir)

	return &Container{
		PathManager:   pathManager,
		CacheManager:  cache.NewManagerWithPathManager(pathManager),
		ConfigManager: config.NewManagerWithPathManager(pathManager),
		AuthManager:   auth.NewManagerWithPath(pathManager.GetConfigPath()),
	}
}

// Initialize initializes all components in the container
func (c *Container) Initialize(ctx context.Context) error {
	if c.initialized {
		return nil
	}

	// Ensure directories exist
	if err := c.PathManager.EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to ensure cache directory: %w", err)
	}

	if err := c.PathManager.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	// Initialize config manager
	cfg, err := c.ConfigManager.GetConfig()
	if err != nil {
		// If config doesn't exist, create a default one
		cfg = &config.Config{}
		if err := c.ConfigManager.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save default config: %w", err)
		}
	}

	// Initialize Gist client if token is available
	if cfg.Token != "" {
		client, err := gist.NewClient(cfg.Token)
		if err != nil {
			return fmt.Errorf("failed to create gist client: %w", err)
		}
		c.GistClient = client
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
		return fmt.Errorf("failed to create gist client: %w", err)
	}
	c.GistClient = client

	// Save the token to config
	cfg, err := c.ConfigManager.GetConfig()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.Token = token

	if err := c.ConfigManager.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save token to config: %w", err)
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
		return nil, fmt.Errorf("not authenticated. Please run 'pv login' first")
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
