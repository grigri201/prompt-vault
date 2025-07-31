package sync

import (
	"context"

	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/interfaces"
)

// SyncMiddleware defines the interface for sync middleware operations
type SyncMiddleware interface {
	PreSync(ctx context.Context, cmd string) error
	PostSync(ctx context.Context, cmd string) error
}

// SyncConfig defines sync behavior for a command
type SyncConfig struct {
	Pre  bool // Execute pre-sync
	Post bool // Execute post-sync
}

// Middleware implements the SyncMiddleware interface
type Middleware struct {
	syncManager interfaces.SyncManager
	config      map[string]SyncConfig
}

// NewSyncMiddleware creates a new sync middleware
func NewSyncMiddleware(syncManager interfaces.SyncManager) *Middleware {
	return &Middleware{
		syncManager: syncManager,
		config: map[string]SyncConfig{
			"login":  {Pre: false, Post: true},
			"add":    {Pre: true, Post: true},
			"get":    {Pre: true, Post: false},
			"share":  {Pre: true, Post: false},
			"del":    {Pre: true, Post: true},
			"sync":   {Pre: false, Post: false}, // Manual sync command doesn't use middleware
			"config": {Pre: false, Post: false},
		},
	}
}

// PreSync executes pre-command synchronization if configured
func (m *Middleware) PreSync(ctx context.Context, cmd string) error {
	config, exists := m.config[cmd]
	if !exists || !config.Pre {
		return nil // No pre-sync needed for this command
	}

	if !m.syncManager.IsInitialized() {
		if err := m.syncManager.Initialize(ctx); err != nil {
			return errors.WrapWithMessage(err, "failed to initialize sync manager for pre-sync")
		}
	}

	return m.syncManager.SynchronizeData(ctx)
}

// PostSync executes post-command synchronization if configured
func (m *Middleware) PostSync(ctx context.Context, cmd string) error {
	config, exists := m.config[cmd]
	if !exists || !config.Post {
		return nil // No post-sync needed for this command
	}

	if !m.syncManager.IsInitialized() {
		if err := m.syncManager.Initialize(ctx); err != nil {
			return errors.WrapWithMessage(err, "failed to initialize sync manager for post-sync")
		}
	}

	return m.syncManager.SynchronizeData(ctx)
}

// SetCommandConfig allows customizing sync behavior for specific commands
func (m *Middleware) SetCommandConfig(cmd string, config SyncConfig) {
	m.config[cmd] = config
}

// GetCommandConfig returns the sync configuration for a command
func (m *Middleware) GetCommandConfig(cmd string) (SyncConfig, bool) {
	config, exists := m.config[cmd]
	return config, exists
}
