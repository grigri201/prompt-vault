package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildContainer(t *testing.T) {
	container := buildContainer()
	assert.NotNil(t, container)

	// Verify all components are properly initialized
	assert.NotNil(t, container.PathManager)
	assert.NotNil(t, container.CacheManager)
	assert.NotNil(t, container.ConfigManager)
	assert.NotNil(t, container.AuthManager)
	assert.NotNil(t, container.SyncManager)
	assert.NotNil(t, container.SyncMiddleware)

	// GistClient may be nil initially (requires authentication)
	// This is expected behavior
}

func TestContainerStructure(t *testing.T) {
	container := buildContainer()

	// Verify sync manager implements the interface
	syncManager := container.SyncManager
	assert.NotNil(t, syncManager)
	assert.NotNil(t, syncManager.GetSyncStatus())

	// Verify sync middleware implements the interface
	syncMiddleware := container.SyncMiddleware
	assert.NotNil(t, syncMiddleware)
}
