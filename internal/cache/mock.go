package cache

import (
	"context"

	"github.com/grigri201/prompt-vault/internal/managers"
	"github.com/grigri201/prompt-vault/internal/models"
)

// MockManager is a mock implementation of cache.Manager for testing
type MockManager struct {
	managers.BaseManager

	// Function fields for customizable behavior
	InitializeFunc      func(ctx context.Context) error
	CleanupFunc         func() error
	InitializeCacheFunc func() error
	CleanFunc           func() error
	SavePromptFunc      func(*models.Prompt) error
	GetPromptFunc       func(string) (*models.Prompt, error)
	SaveIndexFunc       func(*models.Index) error
	GetIndexFunc        func() (*models.Index, error)
	DeletePromptFunc    func(string) error
	GetCacheDirFunc     func() string
	GetIndexPathFunc    func() string
	GetMetadataPathFunc func() string
	GetPromptPathFunc   func(string) string
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

// InitializeCache mocks the InitializeCache method
func (m *MockManager) InitializeCache() error {
	if m.InitializeCacheFunc != nil {
		return m.InitializeCacheFunc()
	}
	return m.Initialize(context.Background())
}

// Clean mocks the Clean method
func (m *MockManager) Clean() error {
	if m.CleanFunc != nil {
		return m.CleanFunc()
	}
	return nil
}

// SavePrompt mocks the SavePrompt method
func (m *MockManager) SavePrompt(prompt *models.Prompt) error {
	if m.SavePromptFunc != nil {
		return m.SavePromptFunc(prompt)
	}
	return nil
}

// GetPrompt mocks the GetPrompt method
func (m *MockManager) GetPrompt(gistID string) (*models.Prompt, error) {
	if m.GetPromptFunc != nil {
		return m.GetPromptFunc(gistID)
	}
	return &models.Prompt{
		GistID: gistID,
		PromptMeta: models.PromptMeta{
			Name:     "Mock Prompt",
			Author:   "Mock Author",
			Category: "Mock Category",
			Tags:     []string{"mock"},
		},
	}, nil
}

// SaveIndex mocks the SaveIndex method
func (m *MockManager) SaveIndex(index *models.Index) error {
	if m.SaveIndexFunc != nil {
		return m.SaveIndexFunc(index)
	}
	return nil
}

// GetIndex mocks the GetIndex method
func (m *MockManager) GetIndex() (*models.Index, error) {
	if m.GetIndexFunc != nil {
		return m.GetIndexFunc()
	}
	return &models.Index{
		Username: "mock-user",
		Entries:  []models.IndexEntry{},
	}, nil
}

// DeletePrompt mocks the DeletePrompt method
func (m *MockManager) DeletePrompt(name string) error {
	if m.DeletePromptFunc != nil {
		return m.DeletePromptFunc(name)
	}
	return nil
}

// GetCacheDir mocks the GetCacheDir method
func (m *MockManager) GetCacheDir() string {
	if m.GetCacheDirFunc != nil {
		return m.GetCacheDirFunc()
	}
	return "/tmp/mock-cache"
}

// GetIndexPath mocks the GetIndexPath method
func (m *MockManager) GetIndexPath() string {
	if m.GetIndexPathFunc != nil {
		return m.GetIndexPathFunc()
	}
	return "/tmp/mock-cache/index.json"
}

// GetMetadataPath mocks the GetMetadataPath method
func (m *MockManager) GetMetadataPath() string {
	if m.GetMetadataPathFunc != nil {
		return m.GetMetadataPathFunc()
	}
	return "/tmp/mock-cache/metadata.json"
}

// GetPromptPath mocks the GetPromptPath method
func (m *MockManager) GetPromptPath(gistID string) string {
	if m.GetPromptPathFunc != nil {
		return m.GetPromptPathFunc(gistID)
	}
	return "/tmp/mock-cache/" + gistID + ".yaml"
}
