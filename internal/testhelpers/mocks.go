package testhelpers

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/grigri201/prompt-vault/internal/models"
)

// MockCacheManager mock implementation for cache operations
type MockCacheManager struct {
	mock.Mock
}

func (m *MockCacheManager) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheManager) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheManager) GetIndex() (*models.Index, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Index), args.Error(1)
}

func (m *MockCacheManager) SaveIndex(index *models.Index) error {
	args := m.Called(index)
	return args.Error(0)
}

func (m *MockCacheManager) GetPrompt(gistID string) (*models.Prompt, error) {
	args := m.Called(gistID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockCacheManager) SavePrompt(prompt *models.Prompt) error {
	args := m.Called(prompt)
	return args.Error(0)
}

func (m *MockCacheManager) DeletePrompt(gistID string) error {
	args := m.Called(gistID)
	return args.Error(0)
}

func (m *MockCacheManager) ClearCache() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheManager) GetCachePath() string {
	args := m.Called()
	return args.String(0)
}

// MockGistClient mock implementation for GitHub Gist operations
type MockGistClient struct {
	mock.Mock
}

func (m *MockGistClient) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGistClient) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGistClient) GetRemoteIndex(ctx context.Context) (*models.Index, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Index), args.Error(1)
}

func (m *MockGistClient) UploadIndex(ctx context.Context, index *models.Index) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

func (m *MockGistClient) UploadPrompt(ctx context.Context, prompt *models.Prompt) (*models.Prompt, error) {
	args := m.Called(ctx, prompt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockGistClient) DownloadPrompt(ctx context.Context, gistID string) (*models.Prompt, error) {
	args := m.Called(ctx, gistID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func (m *MockGistClient) DeletePrompt(ctx context.Context, gistID string) error {
	args := m.Called(ctx, gistID)
	return args.Error(0)
}

func (m *MockGistClient) ListUserGists(ctx context.Context) ([]*models.Prompt, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Prompt), args.Error(1)
}

func (m *MockGistClient) ValidateToken(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// MockConfigManager mock implementation for configuration operations
type MockConfigManager struct {
	mock.Mock
}

func (m *MockConfigManager) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigManager) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigManager) GetToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConfigManager) SaveToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockConfigManager) ClearToken() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigManager) GetUsername() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConfigManager) SaveUsername(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MockConfigManager) GetConfigPath() string {
	args := m.Called()
	return args.String(0)
}

// MockSyncManager mock implementation for sync operations
type MockSyncManager struct {
	mock.Mock
}

func (m *MockSyncManager) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSyncManager) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSyncManager) IsInitialized() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSyncManager) SynchronizeData(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSyncManager) GetSyncStatus() SyncStatus {
	args := m.Called()
	return args.Get(0).(SyncStatus)
}

// SyncStatus represents the status of synchronization for testing
type SyncStatus struct {
	LocalTime  time.Time
	RemoteTime time.Time
	NeedsSync  bool
	SyncAction string
	Progress   SyncProgress
}

// SyncProgress represents sync progress for testing
type SyncProgress struct {
	Completed int
	Total     int
}

// DisplayString returns a human-readable status string
func (s SyncStatus) DisplayString() string {
	if !s.NeedsSync {
		return "Up to date"
	}
	return s.SyncAction
}

// MockAuthManager mock implementation for authentication operations
type MockAuthManager struct {
	mock.Mock
}

func (m *MockAuthManager) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthManager) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthManager) IsAuthenticated() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthManager) Login(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAuthManager) Logout() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthManager) GetToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAuthManager) ValidateToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}
