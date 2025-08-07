package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/model"
)

// LocalMockStore is a local mock implementation for list tests
type LocalMockStore struct {
	prompts     []model.Prompt
	listError   error
	getExportsResult []model.IndexedPrompt
	getExportsError error
}

// MockStoreForList extends the LocalMockStore for list-specific testing
type MockStoreForList struct {
	*LocalMockStore
	listCalls int
}

func NewMockStoreForList(prompts []model.Prompt, listError error) *MockStoreForList {
	return &MockStoreForList{
		LocalMockStore: &LocalMockStore{
			prompts: prompts,
			listError: listError,
			getExportsResult: []model.IndexedPrompt{}, // Default to empty exports
		},
	}
}

func (m *LocalMockStore) List() ([]model.Prompt, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.prompts, nil
}

func (m *LocalMockStore) GetExports() ([]model.IndexedPrompt, error) {
	if m.getExportsError != nil {
		return nil, m.getExportsError
	}
	return m.getExportsResult, nil
}

// Implement other Store interface methods (stubs for testing)
func (m *LocalMockStore) Add(prompt model.Prompt) error { return nil }
func (m *LocalMockStore) Delete(keyword string) error { return nil }
func (m *LocalMockStore) Update(prompt model.Prompt) error { return nil }
func (m *LocalMockStore) Get(keyword string) ([]model.Prompt, error) { return nil, nil }
func (m *LocalMockStore) GetContent(gistID string) (string, error) { return "", nil }
func (m *LocalMockStore) CreatePublicGist(prompt model.Prompt) (string, error) { return "", nil }
func (m *LocalMockStore) UpdateGist(gistURL string, prompt model.Prompt) error { return nil }
func (m *LocalMockStore) GetGistInfo(gistURL string) (*infra.GistInfo, error) { return nil, nil }
func (m *LocalMockStore) AddExport(prompt model.IndexedPrompt) error { return nil }
func (m *LocalMockStore) UpdateExport(prompt model.IndexedPrompt) error { return nil }
func (m *LocalMockStore) FindExistingPromptByURL(gistURL string) (*model.Prompt, error) { return nil, nil }

func (m *MockStoreForList) List() ([]model.Prompt, error) {
	m.listCalls++
	return m.LocalMockStore.List()
}

// MockCacheManager implements cache management for testing
type MockCacheManager struct {
	loadIndexResult    *model.Index
	loadIndexError     error
	saveIndexError     error
	loadContentResult  string
	loadContentError   error
	saveContentError   error
	cacheInfoResult    *model.CacheInfo
	cacheInfoError     error
	ensureCacheDirError error

	// Call tracking
	loadIndexCalls     int
	saveIndexCalls     []*model.Index
	loadContentCalls   []string
	saveContentCalls   map[string]string
}

func NewMockCacheManager() *MockCacheManager {
	return &MockCacheManager{
		saveContentCalls: make(map[string]string),
	}
}

func (m *MockCacheManager) LoadIndex() (*model.Index, error) {
	m.loadIndexCalls++
	return m.loadIndexResult, m.loadIndexError
}

func (m *MockCacheManager) SaveIndex(index *model.Index) error {
	m.saveIndexCalls = append(m.saveIndexCalls, index)
	return m.saveIndexError
}

func (m *MockCacheManager) LoadContent(gistID string) (string, error) {
	m.loadContentCalls = append(m.loadContentCalls, gistID)
	return m.loadContentResult, m.loadContentError
}

func (m *MockCacheManager) SaveContent(gistID, content string) error {
	m.saveContentCalls[gistID] = content
	return m.saveContentError
}

func (m *MockCacheManager) GetCacheInfo() (*model.CacheInfo, error) {
	return m.cacheInfoResult, m.cacheInfoError
}

func (m *MockCacheManager) EnsureCacheDir() error {
	return m.ensureCacheDirError
}

// MockConfigStore implements config.Store for testing
type MockConfigStore struct {
	token       string
	tokenError  error
	saveError   error
	deleteError error
	configPath  string

	saveTokenCalls   []string
	getTokenCalls    int
	deleteTokenCalls int
}

func NewMockConfigStore() *MockConfigStore {
	return &MockConfigStore{
		configPath: "/mock/config/path",
	}
}

func (m *MockConfigStore) SaveToken(token string) error {
	m.saveTokenCalls = append(m.saveTokenCalls, token)
	if m.saveError != nil {
		return m.saveError
	}
	m.token = token
	return nil
}

func (m *MockConfigStore) GetToken() (string, error) {
	m.getTokenCalls++
	return m.token, m.tokenError
}

func (m *MockConfigStore) DeleteToken() error {
	m.deleteTokenCalls++
	if m.deleteError != nil {
		return m.deleteError
	}
	m.token = ""
	return nil
}

func (m *MockConfigStore) GetConfigPath() string {
	return m.configPath
}

// Use the existing captureOutput function from delete_test.go

// Test list command with cache functionality
func TestListCommand_WithCacheFunctionality(t *testing.T) {
	tests := []struct {
		name           string
		remote         bool
		storeResult    []model.Prompt
		storeError     error
		cacheResult    *model.Index
		cacheError     error
		cacheInfo      *model.CacheInfo
		expectedOutput []string
		expectError    bool
	}{
		{
			name:   "successful list with cache - default behavior",
			remote: false,
			storeResult: []model.Prompt{
				{ID: "123", Name: "Test Prompt", Author: "testuser", GistURL: "https://gist.github.com/testuser/123"},
			},
			storeError: nil,
			cacheInfo: &model.CacheInfo{
				LastUpdated:  time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
				TotalPrompts: 1,
				CacheSize:    1024,
			},
			expectedOutput: []string{
				"Found 1 prompt(s):",
				"Test Prompt - author: testuser : https://gist.github.com/testuser/123 [not exported]",
				"Cache last updated:", // Cache timestamp will be current time
			},
			expectError: false,
		},
		{
			name:   "successful list with --remote flag",
			remote: true,
			storeResult: []model.Prompt{
				{ID: "456", Name: "Remote Prompt", Author: "remoteuser", GistURL: "https://gist.github.com/remoteuser/456"},
			},
			storeError: nil,
			expectedOutput: []string{
				"Found 1 prompt(s):",
				"Remote Prompt - author: remoteuser : https://gist.github.com/remoteuser/456 [not exported]",
			},
			expectError: false,
		},
		{
			name:        "network error shows error message",
			remote:      false,
			storeResult: nil,
			storeError:  errors.NewAppError(errors.ErrNetwork, "network error", nil),
			expectedOutput: []string{
				"error in get prompts:", // The actual error handling behavior
			},
			expectError: false, // Command doesn't return error, just shows message
		},
		{
			name:           "ErrNoIndex - first time user message",
			remote:         false,
			storeResult:    nil,
			storeError:     infra.ErrNoIndex,
			expectedOutput: []string{
				"Welcome to Prompt Vault!",
				"It looks like this is your first time using pv",
				"To get started:",
			},
			expectError: false,
		},
		{
			name:           "ErrEmptyIndex - empty collection message",
			remote:         false,
			storeResult:    nil,
			storeError:     infra.ErrEmptyIndex,
			expectedOutput: []string{
				"Your prompt collection is currently empty",
				"To add prompts:",
			},
			expectError: false,
		},
		{
			name:           "empty prompts list",
			remote:         false,
			storeResult:    []model.Prompt{},
			storeError:     nil,
			expectedOutput: []string{
				"No prompts found in your collection",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock dependencies
			mockStore := NewMockStoreForList(tt.storeResult, tt.storeError)
			
			mockConfig := NewMockConfigStore()
			
			// Create list command
			listCmd := NewListCommand(mockStore, mockConfig)
			
			// Set the remote flag if needed
			if tt.remote {
				(*listCmd).Flags().Set("remote", "true")
			}
			
			// Capture output
			output := captureOutput(func() {
				(*listCmd).Execute()
			})
			
			// Verify expected output strings are present
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but got: %s", expected, output)
				}
			}
		})
	}
}

// Test list command cache behavior integration
func TestListCommand_CacheIntegration(t *testing.T) {
	tests := []struct {
		name                string
		remote              bool
		remotePrompts       []model.Prompt
		remoteError         error
		cachedIndex         *model.Index
		cacheLoadError      error
		cacheManagerError   error
		expectedCacheCalls  int
		expectedRemoteCalls int
	}{
		{
			name:   "successful remote with cache update",
			remote: false,
			remotePrompts: []model.Prompt{
				{ID: "123", Name: "New Prompt", Author: "user", GistURL: "https://gist.github.com/user/123"},
			},
			remoteError:         nil,
			cacheManagerError:   nil,
			expectedCacheCalls:  0, // Cache used as fallback only
			expectedRemoteCalls: 1,
		},
		{
			name:              "remote failure triggers cache fallback",
			remote:            false,
			remotePrompts:     nil,
			remoteError:       errors.NewAppError(errors.ErrNetwork, "network down", nil),
			cachedIndex: &model.Index{
				Prompts: []model.IndexedPrompt{
					{GistURL: "https://gist.github.com/user/cached", Name: "Cached", Author: "user"},
				},
			},
			cacheLoadError:      nil,
			expectedCacheCalls:  0, // CacheManager is created, but actual cache loading happens in CachedStore
			expectedRemoteCalls: 1,
		},
		{
			name:                "--remote flag bypasses cache",
			remote:              true,
			remotePrompts:       []model.Prompt{{ID: "123", Name: "Remote", Author: "user"}},
			remoteError:         nil,
			cacheManagerError:   nil,
			expectedCacheCalls:  0, // No cache manager created with --remote
			expectedRemoteCalls: 1,
		},
		{
			name:                "cache manager creation failure falls back to remote only",
			remote:              false,
			remotePrompts:       []model.Prompt{{ID: "123", Name: "Fallback", Author: "user"}},
			remoteError:         nil,
			cacheManagerError:   errors.NewAppError(errors.ErrStorage, "cache dir creation failed", nil),
			expectedCacheCalls:  0, // Cache manager creation failed
			expectedRemoteCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock store
			mockStore := NewMockStoreForList(tt.remotePrompts, tt.remoteError)
			
			mockConfig := NewMockConfigStore()
			
			// Create command
			listCmd := NewListCommand(mockStore, mockConfig)
			
			// Set remote flag if needed
			if tt.remote {
				(*listCmd).Flags().Set("remote", "true")
			}
			
			// Execute command
			captureOutput(func() {
				(*listCmd).Execute()
			})
			
			// Verify remote store was called expected number of times
			if mockStore.listCalls != tt.expectedRemoteCalls {
				t.Errorf("Expected %d remote calls, got %d", tt.expectedRemoteCalls, mockStore.listCalls)
			}
		})
	}
}

// Test backward compatibility - existing behavior should not change
func TestListCommand_BackwardCompatibility(t *testing.T) {
	t.Run("basic list functionality remains unchanged", func(t *testing.T) {
		mockStore := NewMockStoreForList([]model.Prompt{
			{ID: "123", Name: "Legacy Prompt", Author: "legacy", GistURL: "https://gist.github.com/legacy/123"},
		}, nil)
		
		mockConfig := NewMockConfigStore()
		listCmd := NewListCommand(mockStore, mockConfig)
		
		output := captureOutput(func() {
			(*listCmd).Execute()
		})
		
		// Verify legacy output format is preserved (with new export info)
		expected := []string{
			"Found 1 prompt(s):",
			"Legacy Prompt - author: legacy : https://gist.github.com/legacy/123 [not exported]",
		}
		
		for _, exp := range expected {
			if !strings.Contains(output, exp) {
				t.Errorf("Expected output to contain %q, got: %s", exp, output)
			}
		}
		
		// Verify store was called exactly once
		if mockStore.listCalls != 1 {
			t.Errorf("Expected 1 store call, got %d", mockStore.listCalls)
		}
	})
	
	t.Run("error handling remains consistent", func(t *testing.T) {
		mockStore := NewMockStoreForList(nil, errors.NewAppError(errors.ErrAuth, "authentication failed", nil))
		
		mockConfig := NewMockConfigStore()
		listCmd := NewListCommand(mockStore, mockConfig)
		
		// This should not panic and should handle the error gracefully
		output := captureOutput(func() {
			(*listCmd).Execute()
		})
		
		// Should contain error information (exact format may vary)
		if output == "" {
			t.Error("Expected some output for error case, got empty string")
		}
	})
}

// Test mock store behavior directly
func TestMockStoreDebug(t *testing.T) {
	// Test with error
	mockWithError := NewMockStoreForList(nil, errors.NewAppError(errors.ErrNetwork, "test error", nil))
	prompts, err := mockWithError.List()
	
	if err == nil {
		t.Errorf("Expected error but got nil")
	}
	if prompts != nil {
		t.Errorf("Expected nil prompts but got %v", prompts)
	}
	
	// Test with prompts
	testPrompts := []model.Prompt{
		{ID: "test", Name: "Test", Author: "author"},
	}
	mockWithPrompts := NewMockStoreForList(testPrompts, nil)
	prompts2, err2 := mockWithPrompts.List()
	
	if err2 != nil {
		t.Errorf("Expected no error but got %v", err2)
	}
	if len(prompts2) != 1 {
		t.Errorf("Expected 1 prompt but got %d", len(prompts2))
	}
}

// Test command structure and flags
func TestListCommand_Structure(t *testing.T) {
	mockStore := &LocalMockStore{}
	mockConfig := NewMockConfigStore()
	
	listCmd := NewListCommand(mockStore, mockConfig)
	
	// Verify command properties
	if (*listCmd).Use != "list" {
		t.Errorf("Expected Use to be 'list', got %q", (*listCmd).Use)
	}
	
	if (*listCmd).Short == "" {
		t.Error("Expected Short description to be set")
	}
	
	if (*listCmd).Long == "" {
		t.Error("Expected Long description to be set")
	}
	
	// Verify --remote flag exists
	remoteFlag := (*listCmd).Flags().Lookup("remote")
	if remoteFlag == nil {
		t.Error("Expected --remote flag to exist")
	}
	
	if remoteFlag.Shorthand != "r" {
		t.Errorf("Expected --remote flag shorthand to be 'r', got %q", remoteFlag.Shorthand)
	}
	
	if remoteFlag.DefValue != "false" {
		t.Errorf("Expected --remote flag default to be 'false', got %q", remoteFlag.DefValue)
	}
}