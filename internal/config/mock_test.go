package config

import (
	"errors"
	"testing"
	"time"
)

func TestMockManager_GetConfig(t *testing.T) {
	// Test with custom function
	expectedConfig := &Config{
		Token:    "custom-token",
		Username: "custom-user",
		LastSync: time.Now(),
	}
	mock := &MockManager{
		GetConfigFunc: func() (*Config, error) {
			return expectedConfig, nil
		},
	}

	config, err := mock.GetConfig()
	if err != nil {
		t.Errorf("GetConfig() error = %v", err)
	}
	if config.Token != expectedConfig.Token {
		t.Error("GetConfigFunc did not return expected config")
	}

	// Test default behavior
	mock2 := &MockManager{}
	config2, err := mock2.GetConfig()
	if err != nil {
		t.Errorf("GetConfig() with default behavior error = %v", err)
	}
	if config2.Token != "mock-token" {
		t.Error("Default GetConfig should return mock config")
	}
}

func TestMockManager_SaveConfig(t *testing.T) {
	cfg := &Config{
		Token:    "test-token",
		Username: "test-user",
	}

	// Test with custom function
	called := false
	mock := &MockManager{
		SaveConfigFunc: func(c *Config) error {
			called = true
			if c.Token != cfg.Token {
				t.Error("SaveConfigFunc called with wrong config")
			}
			return nil
		},
	}

	err := mock.SaveConfig(cfg)
	if err != nil {
		t.Errorf("SaveConfig() error = %v", err)
	}
	if !called {
		t.Error("SaveConfigFunc was not called")
	}
	if !mock.SaveConfigCalled {
		t.Error("SaveConfigCalled should be true")
	}
	if mock.SaveConfigCallCount != 1 {
		t.Errorf("SaveConfigCallCount = %d, want 1", mock.SaveConfigCallCount)
	}
	if mock.LastSavedConfig != cfg {
		t.Error("LastSavedConfig should be the saved config")
	}

	// Test error case
	mock2 := &MockManager{
		SaveConfigFunc: func(c *Config) error {
			return errors.New("save failed")
		},
	}
	err = mock2.SaveConfig(cfg)
	if err == nil {
		t.Error("SaveConfig() should return error when func returns error")
	}
}

func TestMockManager_UpdateLastSync(t *testing.T) {
	// Test with custom function
	called := false
	mock := &MockManager{
		UpdateLastSyncFunc: func() error {
			called = true
			return nil
		},
	}

	err := mock.UpdateLastSync()
	if err != nil {
		t.Errorf("UpdateLastSync() error = %v", err)
	}
	if !called {
		t.Error("UpdateLastSyncFunc was not called")
	}

	// Test default behavior
	mock2 := &MockManager{}
	beforeTime := time.Now()
	err = mock2.UpdateLastSync()
	if err != nil {
		t.Errorf("UpdateLastSync() with default behavior error = %v", err)
	}
	if !mock2.SaveConfigCalled {
		t.Error("Default UpdateLastSync should call SaveConfig")
	}
	if mock2.LastSavedConfig == nil {
		t.Error("Default UpdateLastSync should save config")
	}
	if mock2.LastSavedConfig.LastSync.Before(beforeTime) {
		t.Error("Default UpdateLastSync should update LastSync time")
	}
}

func TestMockManager_CallTracking(t *testing.T) {
	mock := &MockManager{}
	cfg := &Config{Token: "test"}

	// Test call tracking
	mock.SaveConfig(cfg)
	mock.SaveConfig(cfg)
	mock.SaveConfig(cfg)

	if mock.SaveConfigCallCount != 3 {
		t.Errorf("SaveConfigCallCount = %d, want 3", mock.SaveConfigCallCount)
	}
}
