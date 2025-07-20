package config

import (
	"context"
	"time"

	"github.com/grigri201/prompt-vault/internal/managers"
)

// MockManager is a mock implementation of config.Manager for testing
type MockManager struct {
	managers.BaseManager

	// Function fields for customizable behavior
	InitializeFunc     func(ctx context.Context) error
	CleanupFunc        func() error
	GetConfigFunc      func() (*Config, error)
	SaveConfigFunc     func(*Config) error
	UpdateLastSyncFunc func() error

	// State for tracking calls
	SaveConfigCalled    bool
	SaveConfigCallCount int
	LastSavedConfig     *Config
}

// Initialize mocks the Initialize method
func (m *MockManager) Initialize(ctx context.Context) error {
	if m.InitializeFunc != nil {
		return m.InitializeFunc(ctx)
	}
	m.SetInitialized(true)
	return nil
}

// Cleanup mocks the Cleanup method
func (m *MockManager) Cleanup() error {
	if m.CleanupFunc != nil {
		return m.CleanupFunc()
	}
	m.SetInitialized(false)
	return nil
}

// GetConfig mocks the GetConfig method
func (m *MockManager) GetConfig() (*Config, error) {
	if m.GetConfigFunc != nil {
		return m.GetConfigFunc()
	}
	// Return a default config
	return &Config{
		Token:    "mock-token",
		Username: "mock-user",
		LastSync: time.Now(),
	}, nil
}

// SaveConfig mocks the SaveConfig method
func (m *MockManager) SaveConfig(config *Config) error {
	m.SaveConfigCalled = true
	m.SaveConfigCallCount++
	m.LastSavedConfig = config

	if m.SaveConfigFunc != nil {
		return m.SaveConfigFunc(config)
	}
	return nil
}

// UpdateLastSync mocks the UpdateLastSync method
func (m *MockManager) UpdateLastSync() error {
	if m.UpdateLastSyncFunc != nil {
		return m.UpdateLastSyncFunc()
	}

	// Default behavior: get config, update time, save
	cfg, err := m.GetConfig()
	if err != nil {
		return err
	}
	cfg.LastSync = time.Now()
	return m.SaveConfig(cfg)
}
