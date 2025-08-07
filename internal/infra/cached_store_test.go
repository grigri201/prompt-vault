package infra

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/grigri/pv/internal/config"
	appErrors "github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
)

// MockStore implements the Store interface for testing
type MockStore struct {
	listFunc                  func() ([]model.Prompt, error)
	addFunc                   func(model.Prompt) error
	deleteFunc                func(string) error
	updateFunc                func(model.Prompt) error
	getFunc                   func(string) ([]model.Prompt, error)
	getContentFunc            func(string) (string, error)
	createPublicGistFunc      func(model.Prompt) (string, error)
	updateGistFunc            func(string, model.Prompt) error
	getGistInfoFunc           func(string) (*GistInfo, error)
	addExportFunc             func(model.IndexedPrompt) error
	updateExportFunc          func(model.IndexedPrompt) error
	getExportsFunc            func() ([]model.IndexedPrompt, error)
	findExistingPromptByURLFunc func(string) (*model.Prompt, error)
}

func (m *MockStore) List() ([]model.Prompt, error) {
	if m.listFunc != nil {
		return m.listFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) Add(prompt model.Prompt) error {
	if m.addFunc != nil {
		return m.addFunc(prompt)
	}
	return errors.New("not implemented")
}

func (m *MockStore) Delete(keyword string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(keyword)
	}
	return errors.New("not implemented")
}

func (m *MockStore) Update(prompt model.Prompt) error {
	if m.updateFunc != nil {
		return m.updateFunc(prompt)
	}
	return errors.New("not implemented")
}

func (m *MockStore) Get(keyword string) ([]model.Prompt, error) {
	if m.getFunc != nil {
		return m.getFunc(keyword)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) GetContent(gistID string) (string, error) {
	if m.getContentFunc != nil {
		return m.getContentFunc(gistID)
	}
	return "", errors.New("not implemented")
}

func (m *MockStore) CreatePublicGist(prompt model.Prompt) (string, error) {
	if m.createPublicGistFunc != nil {
		return m.createPublicGistFunc(prompt)
	}
	return "", errors.New("not implemented")
}

func (m *MockStore) UpdateGist(gistURL string, prompt model.Prompt) error {
	if m.updateGistFunc != nil {
		return m.updateGistFunc(gistURL, prompt)
	}
	return errors.New("not implemented")
}

func (m *MockStore) GetGistInfo(gistURL string) (*GistInfo, error) {
	if m.getGistInfoFunc != nil {
		return m.getGistInfoFunc(gistURL)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) AddExport(prompt model.IndexedPrompt) error {
	if m.addExportFunc != nil {
		return m.addExportFunc(prompt)
	}
	return errors.New("not implemented")
}

func (m *MockStore) UpdateExport(prompt model.IndexedPrompt) error {
	if m.updateExportFunc != nil {
		return m.updateExportFunc(prompt)
	}
	return errors.New("not implemented")
}

func (m *MockStore) GetExports() ([]model.IndexedPrompt, error) {
	if m.getExportsFunc != nil {
		return m.getExportsFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockStore) FindExistingPromptByURL(gistURL string) (*model.Prompt, error) {
	if m.findExistingPromptByURLFunc != nil {
		return m.findExistingPromptByURLFunc(gistURL)
	}
	return nil, errors.New("not implemented")
}

// MockConfigStore implements config.Store for testing
type MockConfigStore struct{}

func (m *MockConfigStore) SaveToken(token string) error { return nil }
func (m *MockConfigStore) GetToken() (string, error)    { return "test-token", nil }
func (m *MockConfigStore) DeleteToken() error           { return nil }
func (m *MockConfigStore) GetConfigPath() string        { return "/test/config" }

// Ensure MockConfigStore implements config.Store
var _ config.Store = (*MockConfigStore)(nil)

// Helper function to create test prompts
func createTestPrompts() []model.Prompt {
	return []model.Prompt{
		{
			ID:      "gist1",
			Name:    "Test Prompt 1",
			Author:  "author1",
			GistURL: "https://gist.github.com/user/gist1",
			Content: "Content 1",
		},
		{
			ID:      "gist2",
			Name:    "Test Prompt 2",
			Author:  "author2",
			GistURL: "https://gist.github.com/user/gist2",
			Content: "Content 2",
		},
	}
}

// Helper function to create test index
func createTestIndex() *model.Index {
	return &model.Index{
		Prompts: []model.IndexedPrompt{
			{
				GistURL:     "https://gist.github.com/user/gist1",
				FilePath:    "Test Prompt 1.yaml",
				Author:      "author1",
				Name:        "Test Prompt 1",
				LastUpdated: time.Now(),
			},
			{
				GistURL:     "https://gist.github.com/user/gist2",
				FilePath:    "Test Prompt 2.yaml",
				Author:      "author2",
				Name:        "Test Prompt 2",
				LastUpdated: time.Now(),
			},
		},
		LastUpdated: time.Now(),
	}
}

func TestCachedStore_List_RemoteFirst(t *testing.T) {
	tests := []struct {
		name        string
		forceRemote bool
		remoteError error
		cacheError  error
		expectError bool
		checkFunc   func([]model.Prompt) error
	}{
		{
			name:        "remote success - updates cache",
			forceRemote: false,
			remoteError: nil,
			cacheError:  nil,
			expectError: false,
			checkFunc: func(prompts []model.Prompt) error {
				if len(prompts) != 2 {
					return fmt.Errorf("expected 2 prompts, got %d", len(prompts))
				}
				return nil
			},
		},
		{
			name:        "remote fails, cache succeeds - cache fallback",
			forceRemote: false,
			remoteError: appErrors.NewAppError(appErrors.ErrNetwork, "network error", nil),
			cacheError:  nil,
			expectError: false,
			checkFunc: func(prompts []model.Prompt) error {
				if len(prompts) != 2 {
					return fmt.Errorf("expected 2 prompts from cache, got %d", len(prompts))
				}
				return nil
			},
		},
		{
			name:        "remote fails with force remote - no cache fallback",
			forceRemote: true,
			remoteError: appErrors.NewAppError(appErrors.ErrNetwork, "network error", nil),
			cacheError:  nil,
			expectError: true,
		},
		{
			name:        "remote fails, cache fails - both fail",
			forceRemote: false,
			remoteError: appErrors.NewAppError(appErrors.ErrNetwork, "network error", nil),
			cacheError:  appErrors.NewAppError(appErrors.ErrStorage, "cache error", nil),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for cache manager
			tempDir := t.TempDir()
			
			// Setup mock remote store
			mockRemote := &MockStore{
				listFunc: func() ([]model.Prompt, error) {
					if tt.remoteError != nil {
						return nil, tt.remoteError
					}
					return createTestPrompts(), nil
				},
			}

			// Setup real cache manager with temp directory
			cacheManager := &CacheManager{cacheDir: tempDir}
			
			// Setup mock config store
			mockConfig := &MockConfigStore{}

			// Create cached store
			store := NewCachedStore(mockRemote, cacheManager, mockConfig, tt.forceRemote)

			// Pre-populate cache if test expects cache to work
			if tt.cacheError == nil && tt.remoteError != nil {
				// Save test index to cache for fallback scenarios
				testIndex := createTestIndex()
				cacheManager.SaveIndex(testIndex)
			}

			// Test List operation
			prompts, err := store.List()

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Run additional checks
				if tt.checkFunc != nil {
					if checkErr := tt.checkFunc(prompts); checkErr != nil {
						t.Errorf("Check failed: %v", checkErr)
					}
				}
			}
		})
	}
}

func TestCachedStore_GetContent_RemoteFirst(t *testing.T) {
	tests := []struct {
		name            string
		gistID          string
		forceRemote     bool
		remoteError     error
		remoteContent   string
		cacheError      error
		cachedContent   string
		expectError     bool
		expectedContent string
	}{
		{
			name:            "remote success - returns remote content and caches it",
			gistID:          "test-gist",
			forceRemote:     false,
			remoteError:     nil,
			remoteContent:   "remote content",
			cacheError:      nil,
			cachedContent:   "cached content",
			expectError:     false,
			expectedContent: "remote content",
		},
		{
			name:            "remote fails, cache succeeds - returns cached content",
			gistID:          "test-gist",
			forceRemote:     false,
			remoteError:     appErrors.NewAppError(appErrors.ErrNetwork, "network error", nil),
			remoteContent:   "",
			cacheError:      nil,
			cachedContent:   "cached content",
			expectError:     false,
			expectedContent: "cached content",
		},
		{
			name:            "remote fails with force remote - no cache fallback",
			gistID:          "test-gist",
			forceRemote:     true,
			remoteError:     appErrors.NewAppError(appErrors.ErrNetwork, "network error", nil),
			remoteContent:   "",
			cacheError:      nil,
			cachedContent:   "cached content",
			expectError:     true,
			expectedContent: "",
		},
		{
			name:            "remote fails, cache fails - both fail",
			gistID:          "test-gist",
			forceRemote:     false,
			remoteError:     appErrors.NewAppError(appErrors.ErrNetwork, "network error", nil),
			remoteContent:   "",
			cacheError:      appErrors.NewAppError(appErrors.ErrStorage, "cache error", nil),
			cachedContent:   "",
			expectError:     true,
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for cache manager
			tempDir := t.TempDir()
			
			// Setup mock remote store
			mockRemote := &MockStore{
				getContentFunc: func(gistID string) (string, error) {
					if gistID != tt.gistID {
						return "", fmt.Errorf("unexpected gist ID: %s", gistID)
					}
					if tt.remoteError != nil {
						return "", tt.remoteError
					}
					return tt.remoteContent, nil
				},
			}

			// Setup real cache manager with temp directory
			cacheManager := &CacheManager{cacheDir: tempDir}
			
			// Setup mock config store
			mockConfig := &MockConfigStore{}

			// Pre-populate cache if test expects cache to work
			if tt.cacheError == nil && tt.cachedContent != "" {
				cacheManager.SaveContent(tt.gistID, tt.cachedContent)
			}

			// Create cached store
			store := NewCachedStore(mockRemote, cacheManager, mockConfig, tt.forceRemote)

			// Test GetContent operation
			content, err := store.GetContent(tt.gistID)

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify content
				if content != tt.expectedContent {
					t.Errorf("Expected content %q, got %q", tt.expectedContent, content)
				}
			}
		})
	}
}

func TestCachedStore_NetworkFailureScenarios(t *testing.T) {
	tests := []struct {
		name         string
		networkError error
		cacheData    bool
		expectError  bool
		description  string
	}{
		{
			name:         "temporary network error with cache available",
			networkError: appErrors.NewAppError(appErrors.ErrNetwork, "temporary network issue", nil),
			cacheData:    true,
			expectError:  false,
			description:  "Should fallback to cache when network is temporarily unavailable",
		},
		{
			name:         "permanent network error with cache available",
			networkError: appErrors.NewAppError(appErrors.ErrNetwork, "no internet connection", nil),
			cacheData:    true,
			expectError:  false,
			description:  "Should work offline using cached data",
		},
		{
			name:         "network error with no cache available",
			networkError: appErrors.NewAppError(appErrors.ErrNetwork, "network error", nil),
			cacheData:    false,
			expectError:  true,
			description:  "Should fail when both network and cache are unavailable",
		},
		{
			name:         "auth error with cache available",
			networkError: appErrors.NewAppError(appErrors.ErrAuth, "authentication failed", nil),
			cacheData:    true,
			expectError:  false,
			description:  "Should fallback to cache when authentication fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for cache manager
			tempDir := t.TempDir()
			
			// Setup mock remote store that fails
			mockRemote := &MockStore{
				listFunc: func() ([]model.Prompt, error) {
					return nil, tt.networkError
				},
			}

			// Setup real cache manager with temp directory
			cacheManager := &CacheManager{cacheDir: tempDir}
			
			// Pre-populate cache if test expects cache to work
			if tt.cacheData {
				testIndex := createTestIndex()
				cacheManager.SaveIndex(testIndex)
			}

			// Setup mock config store
			mockConfig := &MockConfigStore{}

			// Create cached store (not force remote to allow cache fallback)
			store := NewCachedStore(mockRemote, cacheManager, mockConfig, false)

			// Test List operation
			prompts, err := store.List()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none - %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v - %s", err, tt.description)
				}
				if len(prompts) == 0 {
					t.Errorf("Expected cached prompts but got none - %s", tt.description)
				}
			}
		})
	}
}

func TestCachedStore_CacheHitMissScenarios(t *testing.T) {
	testGistID := "cache-test"
	remoteContent := "remote content"
	cachedContent := "cached content"

	tests := []struct {
		name            string
		cacheAvailable  bool
		remoteAvailable bool
		expectedContent string
		expectError     bool
		description     string
	}{
		{
			name:            "cache hit with remote available - remote priority",
			cacheAvailable:  true,
			remoteAvailable: true,
			expectedContent: remoteContent,
			expectError:     false,
			description:     "Should prefer remote data even when cache is available (remote-first strategy)",
		},
		{
			name:            "cache hit with remote unavailable - cache fallback",
			cacheAvailable:  true,
			remoteAvailable: false,
			expectedContent: cachedContent,
			expectError:     false,
			description:     "Should use cached data when remote is unavailable",
		},
		{
			name:            "cache miss with remote available - remote success",
			cacheAvailable:  false,
			remoteAvailable: true,
			expectedContent: remoteContent,
			expectError:     false,
			description:     "Should use remote data when cache is not available",
		},
		{
			name:            "cache miss with remote unavailable - both fail",
			cacheAvailable:  false,
			remoteAvailable: false,
			expectedContent: "",
			expectError:     true,
			description:     "Should fail when both remote and cache are unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for cache manager
			tempDir := t.TempDir()
			
			// Setup mock remote store
			mockRemote := &MockStore{
				getContentFunc: func(gistID string) (string, error) {
					if !tt.remoteAvailable {
						return "", appErrors.NewAppError(appErrors.ErrNetwork, "remote unavailable", nil)
					}
					return remoteContent, nil
				},
			}

			// Setup real cache manager with temp directory
			cacheManager := &CacheManager{cacheDir: tempDir}
			
			// Pre-populate cache if test expects cache to be available
			if tt.cacheAvailable {
				cacheManager.SaveContent(testGistID, cachedContent)
			}

			// Setup mock config store
			mockConfig := &MockConfigStore{}

			// Create cached store
			store := NewCachedStore(mockRemote, cacheManager, mockConfig, false)

			// Test GetContent operation
			content, err := store.GetContent(testGistID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none - %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v - %s", err, tt.description)
				}
				if content != tt.expectedContent {
					t.Errorf("Expected content %q, got %q - %s", tt.expectedContent, content, tt.description)
				}
			}
		})
	}
}

func TestCachedStore_DataConsistency(t *testing.T) {
	t.Run("cache update failure does not affect remote operation success", func(t *testing.T) {
		testPrompts := createTestPrompts()

		// Setup mock remote store that succeeds
		mockRemote := &MockStore{
			listFunc: func() ([]model.Prompt, error) {
				return testPrompts, nil
			},
		}

		// Create cache manager that will fail to save (by using a bad directory path)
		// This simulates cache write failures without affecting remote operations
		badCacheManager := &CacheManager{cacheDir: "/invalid/path/that/does/not/exist"}

		// Setup mock config store
		mockConfig := &MockConfigStore{}

		// Create cached store
		store := NewCachedStore(mockRemote, badCacheManager, mockConfig, false)

		// Test List operation - should succeed despite cache failure
		prompts, err := store.List()

		// Should not return error even if cache update fails
		if err != nil {
			t.Errorf("Operation should succeed even if cache update fails: %v", err)
		}

		// Should return remote data
		if len(prompts) != len(testPrompts) {
			t.Errorf("Expected %d prompts, got %d", len(testPrompts), len(prompts))
		}
	})

	t.Run("cache data integrity - index matches content", func(t *testing.T) {
		// Create a temporary directory for cache manager
		tempDir := t.TempDir()
		
		testGistID := "test-consistency"
		testContent := "consistent content"

		// Setup mock remote store
		mockRemote := &MockStore{
			getContentFunc: func(gistID string) (string, error) {
				if gistID == testGistID {
					return testContent, nil
				}
				return "", fmt.Errorf("gist not found")
			},
		}

		// Setup real cache manager with temp directory
		cacheManager := &CacheManager{cacheDir: tempDir}

		// Setup mock config store
		mockConfig := &MockConfigStore{}

		// Create cached store
		store := NewCachedStore(mockRemote, cacheManager, mockConfig, false)

		// Test GetContent operation
		content, err := store.GetContent(testGistID)

		// Verify operation succeeded
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify data consistency
		if content != testContent {
			t.Errorf("Expected content %q, got %q", testContent, content)
		}

		// Verify content was cached properly
		cachedContent, err := cacheManager.LoadContent(testGistID)
		if err != nil {
			t.Errorf("Failed to load cached content: %v", err)
		}

		if cachedContent != testContent {
			t.Errorf("Expected cached content %q, got %q", testContent, cachedContent)
		}
	})
}