package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/tui"
)

// MockPromptService is a mock implementation of service.PromptService for testing
type MockPromptService struct {
	// Predefined return values
	addFromFileResult         *model.Prompt
	addFromFileError          error
	deleteByKeywordError      error
	deleteByURLError          error
	listPromptsResult     []model.Prompt
	listPromptsError      error
	filterPromptsResult   []model.Prompt
	filterPromptsError    error

	// Method call tracking
	addFromFileCalls         []string
	deleteByKeywordCalls     []string
	deleteByURLCalls         []string
	listPromptsCalls     int
	filterPromptsCalls   []string

	// Custom function overrides
	deleteByKeywordFunc     func(string) error
	deleteByURLFunc         func(string) error
	listForDeletionFunc     func() ([]model.Prompt, error)
	filterForDeletionFunc   func(string) ([]model.Prompt, error)
}

func NewMockPromptService() *MockPromptService {
	return &MockPromptService{
		addFromFileCalls:       make([]string, 0),
		deleteByKeywordCalls:   make([]string, 0),
		deleteByURLCalls:       make([]string, 0),
		filterPromptsCalls: make([]string, 0),
	}
}

func (m *MockPromptService) AddFromFile(filePath string) (*model.Prompt, error) {
	m.addFromFileCalls = append(m.addFromFileCalls, filePath)
	if m.addFromFileError != nil {
		return nil, m.addFromFileError
	}
	return m.addFromFileResult, nil
}

func (m *MockPromptService) DeleteByKeyword(keyword string) error {
	m.deleteByKeywordCalls = append(m.deleteByKeywordCalls, keyword)
	if m.deleteByKeywordFunc != nil {
		return m.deleteByKeywordFunc(keyword)
	}
	return m.deleteByKeywordError
}

func (m *MockPromptService) DeleteByURL(gistURL string) error {
	m.deleteByURLCalls = append(m.deleteByURLCalls, gistURL)
	if m.deleteByURLFunc != nil {
		return m.deleteByURLFunc(gistURL)
	}
	return m.deleteByURLError
}

func (m *MockPromptService) ListPrompts() ([]model.Prompt, error) {
	m.listPromptsCalls++
	if m.listForDeletionFunc != nil {
		return m.listForDeletionFunc()
	}
	if m.listPromptsError != nil {
		return nil, m.listPromptsError
	}
	return m.listPromptsResult, nil
}

func (m *MockPromptService) FilterPrompts(keyword string) ([]model.Prompt, error) {
	m.filterPromptsCalls = append(m.filterPromptsCalls, keyword)
	if m.filterForDeletionFunc != nil {
		return m.filterForDeletionFunc(keyword)
	}
	if m.filterPromptsError != nil {
		return nil, m.filterPromptsError
	}
	return m.filterPromptsResult, nil
}

// GetPromptByURL implements the missing PromptService method for testing
func (m *MockPromptService) GetPromptByURL(gistURL string) (*model.Prompt, error) {
	// This method is not used by delete command but required by interface
	return nil, nil
}

// GetPromptContent implements the missing PromptService method for testing
func (m *MockPromptService) GetPromptContent(prompt *model.Prompt) (string, error) {
	// This method is not used by delete command but required by interface
	return "", nil
}

func (m *MockPromptService) Reset() {
	m.addFromFileResult = nil
	m.addFromFileError = nil
	m.deleteByKeywordError = nil
	m.deleteByURLError = nil
	m.listPromptsResult = nil
	m.listPromptsError = nil
	m.filterPromptsResult = nil
	m.filterPromptsError = nil
	m.addFromFileCalls = make([]string, 0)
	m.deleteByKeywordCalls = make([]string, 0)
	m.deleteByURLCalls = make([]string, 0)
	m.listPromptsCalls = 0
	m.filterPromptsCalls = make([]string, 0)
	m.deleteByKeywordFunc = nil
	m.deleteByURLFunc = nil
	m.listForDeletionFunc = nil
	m.filterForDeletionFunc = nil
}

// MockStore is a simple mock implementation of infra.Store for testing
type MockStore struct {
	prompts     []model.Prompt
	listError   error
	addError    error
	deleteError error
	getError    error
	updateError error
}

func NewMockStore() *MockStore {
	return &MockStore{
		prompts: make([]model.Prompt, 0),
	}
}

func (m *MockStore) List() ([]model.Prompt, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.prompts, nil
}

func (m *MockStore) Add(prompt model.Prompt) error {
	if m.addError != nil {
		return m.addError
	}
	m.prompts = append(m.prompts, prompt)
	return nil
}

func (m *MockStore) Delete(keyword string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	// Remove matching prompts
	for i := len(m.prompts) - 1; i >= 0; i-- {
		if m.prompts[i].ID == keyword {
			m.prompts = append(m.prompts[:i], m.prompts[i+1:]...)
		}
	}
	return nil
}

func (m *MockStore) Update(prompt model.Prompt) error {
	if m.updateError != nil {
		return m.updateError
	}
	// Find and update matching prompt
	for i, p := range m.prompts {
		if p.ID == prompt.ID {
			m.prompts[i] = prompt
			return nil
		}
	}
	return nil
}

func (m *MockStore) Get(keyword string) ([]model.Prompt, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	var matches []model.Prompt
	for _, prompt := range m.prompts {
		if strings.Contains(strings.ToLower(prompt.Name), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(prompt.Author), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(prompt.Description), strings.ToLower(keyword)) ||
			prompt.ID == keyword ||
			prompt.GistURL == keyword {
			matches = append(matches, prompt)
		}
	}
	return matches, nil
}

// GetContent implements the missing Store method for testing
func (m *MockStore) GetContent(gistID string) (string, error) {
	// This method is not used by delete command but required by interface
	return "", nil
}

// Test data helpers
func createTestPrompts() []model.Prompt {
	return []model.Prompt{
		{
			ID:          "gist1234567890abcdef1234",
			Name:        "Go Code Review Prompt",
			Author:      "John Doe",
			Description: "A comprehensive guide for Go code review",
			Tags:        []string{"go", "review", "best-practice"},
			Version:     "1.0",
			Content:     "Go code review guidelines...",
			GistURL:     "https://gist.github.com/johndoe/gist1234567890abcdef1234",
		},
		{
			ID:          "gist4567890123abcdef4567",
			Name:        "SQL Query Optimizer",
			Author:      "Jane Smith",
			Description: "Tips for optimizing SQL queries",
			Tags:        []string{"sql", "performance"},
			Version:     "2.1",
			Content:     "SQL optimization techniques...",
			GistURL:     "https://gist.github.com/janesmith/gist4567890123abcdef4567",
		},
		{
			ID:          "gist7890123456abcdef7890",
			Name:        "Docker Best Practices",
			Author:      "John Doe",
			Description: "Best practices for Docker containerization",
			Tags:        []string{"docker", "containers"},
			Version:     "1.5",
			Content:     "Docker containerization guidelines...",
			GistURL:     "https://gist.github.com/johndoe/gist7890123456abcdef7890",
		},
	}
}

// captureOutput captures stdout during command execution for testing
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestNewDeleteCommand tests the command creation and configuration
func TestNewDeleteCommand(t *testing.T) {
	testCases := []struct {
		name         string
		store        infra.Store
		promptService service.PromptService
		expectPanic  bool
	}{
		{
			name:          "successful command creation with valid dependencies",
			store:         NewMockStore(),
			promptService: NewMockPromptService(),
			expectPanic:   false,
		},
		{
			name:          "command creation with nil store (should not panic)",
			store:         nil,
			promptService: NewMockPromptService(),
			expectPanic:   false,
		},
		{
			name:          "command creation with nil prompt service (should not panic)",
			store:         NewMockStore(),
			promptService: nil,
			expectPanic:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.expectPanic {
					t.Errorf("Unexpected panic: %v", r)
				} else if r == nil && tc.expectPanic {
					t.Errorf("Expected panic but none occurred")
				}
			}()

			cmd := NewDeleteCommand(tc.store, tc.promptService)

			if !tc.expectPanic {
				// Verify command structure
				if cmd == nil {
					t.Fatal("Expected command but got nil")
				}

				// Check command configuration
				if cmd.Use != "delete [keyword|gist-url]" {
					t.Errorf("Expected Use 'delete [keyword|gist-url]', got %q", cmd.Use)
				}

				if cmd.Short != "删除存储的提示" {
					t.Errorf("Expected Short '删除存储的提示', got %q", cmd.Short)
				}

				if !strings.Contains(cmd.Long, "删除存储在 Prompt Vault 中的提示") {
					t.Errorf("Expected Long description to contain '删除存储在 Prompt Vault 中的提示'")
				}

				if !strings.Contains(cmd.Example, "pv delete") {
					t.Errorf("Expected Example to contain 'pv delete'")
				}

				// Check Args configuration
				if cmd.Args == nil {
					t.Errorf("Expected Args to be set")
				}

				// Check Run function is set
				if cmd.Run == nil {
					t.Errorf("Expected Run function to be set")
				}
			}
		})
	}
}

// TestDeleteCommand_CommandConfiguration tests command parameters and flags
func TestDeleteCommand_CommandConfiguration(t *testing.T) {
	mockStore := NewMockStore()
	mockPromptService := NewMockPromptService()
	cmd := NewDeleteCommand(mockStore, mockPromptService)

	// Test command properties
	expectedUse := "delete [keyword|gist-url]"
	if cmd.Use != expectedUse {
		t.Errorf("Expected Use %q, got %q", expectedUse, cmd.Use)
	}

	expectedShort := "删除存储的提示"
	if cmd.Short != expectedShort {
		t.Errorf("Expected Short %q, got %q", expectedShort, cmd.Short)
	}

	// Test help text contains expected information
	longDescriptionKeywords := []string{
		"删除存储在 Prompt Vault 中的提示",
		"交互式删除",
		"关键字筛选删除",
		"直接URL删除",
		"确认",
	}

	for _, keyword := range longDescriptionKeywords {
		if !strings.Contains(cmd.Long, keyword) {
			t.Errorf("Expected Long description to contain %q", keyword)
		}
	}

	// Test example contains expected commands
	exampleCommands := []string{
		"pv delete",
		"pv delete golang",
		"pv delete https://gist.github.com/user/abc123",
		"pv del golang",
	}

	for _, command := range exampleCommands {
		if !strings.Contains(cmd.Example, command) {
			t.Errorf("Expected Example to contain %q", command)
		}
	}

	// Test args validation - should accept 0-1 arguments
	if cmd.Args == nil {
		t.Error("Expected Args validation function to be set")
	}

	// Test Args function directly (cobra.MaximumNArgs(1))
	testCases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, false},
		{"one arg", []string{"test"}, false},
		{"two args", []string{"test", "extra"}, true},
		{"three args", []string{"test", "extra", "more"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cmd.Args(cmd, tc.args)
			if tc.wantErr && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestDeleteCommand_ParameterParsing tests parameter parsing logic
func TestDeleteCommand_ParameterParsing(t *testing.T) {
	testCases := []struct {
		name             string
		args             []string
		expectedMode     string
		expectError      bool
		expectedOutput   string
	}{
		{
			name:         "no arguments - interactive mode",
			args:         []string{},
			expectedMode: "interactive",
			expectError:  false,
		},
		{
			name:         "keyword argument - filter mode",
			args:         []string{"golang"},
			expectedMode: "filter",
			expectError:  false,
		},
		{
			name:         "valid gist URL - direct mode",
			args:         []string{"https://gist.github.com/user/1234567890abcdef1234567890abcdef"},
			expectedMode: "direct",
			expectError:  false,
		},
		{
			name:         "invalid gist URL - error",
			args:         []string{"https://gist.github.com/user/invalid"},
			expectedMode: "error",
			expectError:  false,
			expectedOutput: "Invalid GitHub Gist URL format",
		},
		{
			name:         "non-URL argument - filter mode", 
			args:         []string{"sql"},
			expectedMode: "filter",
			expectError:  false,
		},
		{
			name:           "too many arguments - error",
			args:           []string{"arg1", "arg2", "arg3"},
			expectedMode:   "error",
			expectError:    false,
			expectedOutput: "Too many arguments provided",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			// Set up mock responses based on expected mode
			switch tc.expectedMode {
			case "interactive":
				mockPromptService.listPromptsResult = createTestPrompts()
			case "filter":
				mockPromptService.filterPromptsResult = createTestPrompts()[:1]
			case "direct":
				mockStore.prompts = createTestPrompts()
			}

			cmd := NewDeleteCommand(mockStore, mockPromptService)
			
			// Create a mock TUI that always returns the first prompt and confirms
			mockTUI := tui.NewMockTUI()
			mockTUI.SetupInteractiveDeleteScenario(createTestPrompts()[0], true)
			
			// Replace the TUI creation with our mock (this would require dependency injection in real implementation)
			// For now, we'll test the parameter parsing by checking error output
			
			// Capture output
			output := captureOutput(func() {
				cmd.Run(cmd, tc.args)
			})

			if tc.expectedOutput != "" {
				if !strings.Contains(output, tc.expectedOutput) {
					t.Errorf("Expected output to contain %q, got %q", tc.expectedOutput, output)
				}
			}

			// Verify correct service method calls based on mode
			switch tc.expectedMode {
			case "interactive":
				if mockPromptService.listPromptsCalls != 1 {
					t.Errorf("Expected 1 ListPrompts call, got %d", mockPromptService.listPromptsCalls)
				}
			case "filter":
				if len(mockPromptService.filterPromptsCalls) == 0 {
					t.Error("Expected FilterPrompts to be called")
				} else if mockPromptService.filterPromptsCalls[0] != tc.args[0] {
					t.Errorf("Expected FilterPrompts to be called with %q, got %q", 
						tc.args[0], mockPromptService.filterPromptsCalls[0])
				}
			case "direct":
				// Note: Direct mode calls store.Get first, then DeleteByURL
				// We can't easily test this without more complex mocking
			}
		})
	}
}

// TestDeleteCommand_URLValidation tests URL parameter validation
func TestDeleteCommand_URLValidation(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expectValid bool
	}{
		{
			name:        "valid gist URL with user",
			url:         "https://gist.github.com/user/1234567890abcdef1234567890abcdef",
			expectValid: true,
		},
		{
			name:        "valid gist URL without user",
			url:         "https://gist.github.com/1234567890abcdef1234567890abcdef",
			expectValid: true,
		},
		{
			name:        "valid gist URL with 20 char ID",
			url:         "https://gist.github.com/user/1234567890abcdef1234",
			expectValid: true,
		},
		{
			name:        "valid gist URL with 32 char ID",
			url:         "https://gist.github.com/user/1234567890abcdef1234567890abcdef",
			expectValid: true,
		},
		{
			name:        "invalid - not gist.github.com",
			url:         "https://github.com/user/repo",
			expectValid: false,
		},
		{
			name:        "invalid - too short ID",
			url:         "https://gist.github.com/user/abc123",
			expectValid: false,
		},
		{
			name:        "invalid - not hex characters",
			url:         "https://gist.github.com/user/1234567890abcdefghij",
			expectValid: false,
		},
		{
			name:        "invalid - not URL format",
			url:         "not-a-url",
			expectValid: false,
		},
		{
			name:        "invalid - empty string",
			url:         "",
			expectValid: false,
		},
		{
			name:        "invalid - only spaces",
			url:         "   ",
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			deleteCmd := &delete{
				store:         mockStore,
				promptService: mockPromptService,
			}

			isValid := deleteCmd.isGistURL(tc.url)
			if isValid != tc.expectValid {
				t.Errorf("Expected isGistURL(%q) = %v, got %v", tc.url, tc.expectValid, isValid)
			}
		})
	}
}

// TestDeleteCommand_InvalidParameterHandling tests handling of invalid parameters
func TestDeleteCommand_InvalidParameterHandling(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name: "too many arguments",
			args: []string{"arg1", "arg2"},
			expectedOutput: []string{
				"❌ Error: Too many arguments provided",
				"Usage:",
				"pv delete",
				"pv delete <keyword>",
				"pv delete <gist-url>",
			},
		},
		{
			name: "invalid URL that looks like URL",
			args: []string{"https://gist.github.com/user/invalid"},
			expectedOutput: []string{
				"❌ Error: Invalid GitHub Gist URL format",
				"Valid GitHub Gist URL formats:",
				"https://gist.github.com/username/gist-id",
			},
		},
		{
			name: "invalid non-gist URL",
			args: []string{"https://github.com/user/repo"},
			expectedOutput: []string{
				"❌ Error: Invalid GitHub Gist URL format",
				"This appears to be a URL but not a GitHub Gist URL",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			cmd := NewDeleteCommand(mockStore, mockPromptService)

			output := captureOutput(func() {
				cmd.Run(cmd, tc.args)
			})

			for _, expectedText := range tc.expectedOutput {
				if !strings.Contains(output, expectedText) {
					t.Errorf("Expected output to contain %q, got:\n%s", expectedText, output)
				}
			}
		})
	}
}

// TestDeleteCommand_RoutingLogic tests the three-mode routing logic
func TestDeleteCommand_RoutingLogic(t *testing.T) {
	testCases := []struct {
		name                 string
		args                 []string
		setupMocks          func(*MockStore, *MockPromptService)
		expectedServiceCalls map[string]int
		expectedMode        string
	}{
		{
			name: "no args routes to interactive mode",
			args: []string{},
			setupMocks: func(store *MockStore, service *MockPromptService) {
				service.listPromptsResult = createTestPrompts()
			},
			expectedServiceCalls: map[string]int{
				"ListPrompts": 1,
			},
			expectedMode: "interactive",
		},
		{
			name: "keyword routes to filter mode",
			args: []string{"golang"},
			setupMocks: func(store *MockStore, service *MockPromptService) {
				service.filterPromptsResult = createTestPrompts()[:1]
			},
			expectedServiceCalls: map[string]int{
				"FilterPrompts": 1,
			},
			expectedMode: "filter",
		},
		{
			name: "valid URL routes to direct mode",
			args: []string{"https://gist.github.com/user/1234567890abcdef1234567890abcdef"},
			setupMocks: func(store *MockStore, service *MockPromptService) {
				store.prompts = createTestPrompts()
			},
			expectedServiceCalls: map[string]int{
				"StoreGet": 1,
			},
			expectedMode: "direct",
		},
		{
			name: "invalid URL shows error",
			args: []string{"https://gist.github.com/user/invalid"},
			setupMocks: func(store *MockStore, service *MockPromptService) {
				// No setup needed for error case
			},
			expectedServiceCalls: map[string]int{},
			expectedMode:         "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			// Set up mocks according to test case
			tc.setupMocks(mockStore, mockPromptService)
			
			cmd := NewDeleteCommand(mockStore, mockPromptService)
			
			// Execute command
			output := captureOutput(func() {
				cmd.Run(cmd, tc.args)
			})

			// Verify service calls
			if expectedCalls, exists := tc.expectedServiceCalls["ListPrompts"]; exists {
				if mockPromptService.listPromptsCalls != expectedCalls {
					t.Errorf("Expected %d ListPrompts calls, got %d", 
						expectedCalls, mockPromptService.listPromptsCalls)
				}
			}

			if expectedCalls, exists := tc.expectedServiceCalls["FilterPrompts"]; exists {
				if len(mockPromptService.filterPromptsCalls) != expectedCalls {
					t.Errorf("Expected %d FilterPrompts calls, got %d", 
						expectedCalls, len(mockPromptService.filterPromptsCalls))
				}
			}

			// Verify mode-specific behavior by checking output patterns
			switch tc.expectedMode {
			case "interactive":
				if !strings.Contains(output, "Interactive mode") {
					t.Errorf("Expected interactive mode output, got: %s", output)
				}
			case "filter":
				if !strings.Contains(output, "Filter mode") {
					t.Errorf("Expected filter mode output, got: %s", output)
				}
			case "direct":
				if !strings.Contains(output, "Direct mode") {
					t.Errorf("Expected direct mode output, got: %s", output)
				}
			case "error":
				if !strings.Contains(output, "❌ Error") {
					t.Errorf("Expected error output, got: %s", output)
				}
			}
		})
	}
}

// TestDeleteCommand_InteractiveMode tests interactive mode (non-TUI parts)
func TestDeleteCommand_InteractiveMode(t *testing.T) {
	testCases := []struct {
		name           string
		prompts        []model.Prompt
		serviceError   error
		expectedOutput []string
		expectError    bool
	}{
		{
			name:    "empty prompt list",
			prompts: []model.Prompt{},
			expectedOutput: []string{
				"📭 No prompts found in your vault",
				"To add prompts to your vault, use:",
				"pv add <path-to-yaml-file>",
			},
		},
		{
			name:         "list for deletion service error",
			prompts:      nil,
			serviceError: errors.NewAppError(errors.ErrStorage, "database error", nil),
			expectedOutput: []string{
				"❌ Error loading prompts",
			},
			expectError: true,
		},
		{
			name:    "interactive mode starts with prompts",
			prompts: createTestPrompts(),
			expectedOutput: []string{
				"🔄 Interactive mode - loading all prompts",
				"📋 Found 3 prompt(s) in your vault",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			// Set up service responses
			if tc.serviceError != nil {
				mockPromptService.listPromptsError = tc.serviceError
			} else {
				mockPromptService.listPromptsResult = tc.prompts
			}

			cmd := NewDeleteCommand(mockStore, mockPromptService)
			
			// Execute with no arguments (interactive mode)
			output := captureOutput(func() {
				cmd.Run(cmd, []string{})
			})

			// Verify expected output
			for _, expectedText := range tc.expectedOutput {
				if !strings.Contains(output, expectedText) {
					t.Errorf("Expected output to contain %q, got:\n%s", expectedText, output)
				}
			}
		})
	}
}

// TestDeleteCommand_FilterMode tests filter mode functionality
func TestDeleteCommand_FilterMode(t *testing.T) {
	testCases := []struct {
		name           string
		keyword        string
		filteredPrompts []model.Prompt
		serviceError   error
		expectedOutput []string
	}{
		{
			name:            "successful filter with matches",
			keyword:         "golang",
			filteredPrompts: createTestPrompts()[:1],
			expectedOutput: []string{
				"Filter mode - searching for prompts matching 'golang'",
				"Found 1 prompt(s) matching 'golang'",
				"Matching prompts:",
			},
		},
		{
			name:            "no matches found",
			keyword:         "nonexistent",
			filteredPrompts: []model.Prompt{},
			expectedOutput: []string{
				"📭 No prompts found matching 'nonexistent'",
				"Tips for better search results:",
				"Try a shorter or more general keyword",
			},
		},
		{
			name:         "filter service error",
			keyword:      "test",
			serviceError: errors.NewAppError(errors.ErrStorage, "network error", nil),
			expectedOutput: []string{
				"❌ Error filtering prompts",
				"This could be due to:",
				"Network connectivity issues",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			// Set up service responses
			if tc.serviceError != nil {
				mockPromptService.filterPromptsError = tc.serviceError
			} else {
				mockPromptService.filterPromptsResult = tc.filteredPrompts
			}

			cmd := NewDeleteCommand(mockStore, mockPromptService)
			
			// Execute with keyword argument (filter mode)
			output := captureOutput(func() {
				cmd.Run(cmd, []string{tc.keyword})
			})

			// Verify expected output
			for _, expectedText := range tc.expectedOutput {
				if !strings.Contains(output, expectedText) {
					t.Errorf("Expected output to contain %q, got:\n%s", expectedText, output)
				}
			}

			// Verify service was called with correct keyword
			if tc.serviceError == nil {
				if len(mockPromptService.filterPromptsCalls) != 1 {
					t.Errorf("Expected 1 FilterPrompts call, got %d", 
						len(mockPromptService.filterPromptsCalls))
				} else if mockPromptService.filterPromptsCalls[0] != tc.keyword {
					t.Errorf("Expected FilterPrompts called with %q, got %q", 
						tc.keyword, mockPromptService.filterPromptsCalls[0])
				}
			}
		})
	}
}

// TestDeleteCommand_DirectMode tests direct URL deletion mode
func TestDeleteCommand_DirectMode(t *testing.T) {
	testCases := []struct {
		name           string
		gistURL        string
		setupStore     func(*MockStore)
		serviceError   error
		expectedOutput []string
	}{
		{
			name:    "successful direct deletion",
			gistURL: "https://gist.github.com/johndoe/1234567890abcdef1234567890abcdef",
			setupStore: func(store *MockStore) {
				// Update store prompts to use valid 32-char gist IDs
				store.prompts = []model.Prompt{
					{
						ID:      "1234567890abcdef1234567890abcdef",
						Name:    "Test Prompt",
						Author:  "johndoe",
						GistURL: "https://gist.github.com/johndoe/1234567890abcdef1234567890abcdef",
					},
				}
			},
			expectedOutput: []string{
				"Direct mode - processing URL",
				"Found prompt:",
				"Direct URL deletion mode",
			},
		},
		{
			name:    "prompt not found for URL",
			gistURL: "https://gist.github.com/user/abcdef1234567890abcdef1234567890",
			setupStore: func(store *MockStore) {
				store.prompts = createTestPrompts() // URL doesn't match any prompt
			},
			expectedOutput: []string{
				"❌ Prompt not found for URL",
				"Possible reasons:",
				"The Gist URL is not in your Prompt Vault",
			},
		},
		{
			name:    "invalid gist URL format",
			gistURL: "https://gist.github.com/user/invalid",
			expectedOutput: []string{
				"❌ Error: Invalid GitHub Gist URL format",
				"Valid GitHub Gist URL formats:",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			// Set up store if needed
			if tc.setupStore != nil {
				tc.setupStore(mockStore)
			}

			cmd := NewDeleteCommand(mockStore, mockPromptService)
			
			// Execute with URL argument (direct mode)
			output := captureOutput(func() {
				cmd.Run(cmd, []string{tc.gistURL})
			})

			// Verify expected output
			for _, expectedText := range tc.expectedOutput {
				if !strings.Contains(output, expectedText) {
					t.Errorf("Expected output to contain %q, got:\n%s", expectedText, output)
				}
			}
		})
	}
}

// TestDeleteCommand_EdgeCases tests edge cases and error scenarios
func TestDeleteCommand_EdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		setupMocks     func(*MockStore, *MockPromptService)
		expectedOutput []string
	}{
		{
			name: "special characters in keyword",
			args: []string{"特殊字符 with émojis 🚀"},
			setupMocks: func(store *MockStore, service *MockPromptService) {
				service.filterPromptsResult = []model.Prompt{}
			},
			expectedOutput: []string{
				"No prompts found matching",
			},
		},
		{
			name: "very long keyword",
			args: []string{strings.Repeat("long", 50)},
			setupMocks: func(store *MockStore, service *MockPromptService) {
				service.filterPromptsResult = []model.Prompt{}
			},
			expectedOutput: []string{
				"No prompts found matching",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			tc.setupMocks(mockStore, mockPromptService)
			
			cmd := NewDeleteCommand(mockStore, mockPromptService)
			
			output := captureOutput(func() {
				cmd.Run(cmd, tc.args)
			})

			for _, expectedText := range tc.expectedOutput {
				if !strings.Contains(output, expectedText) {
					t.Errorf("Expected output to contain %q, got:\n%s", expectedText, output)
				}
			}
		})
	}
}

// TestDeleteCommand_URLExtractionLogic tests URL parsing and ID extraction
func TestDeleteCommand_URLExtractionLogic(t *testing.T) {
	testCases := []struct {
		name       string
		url        string
		expectedID string
	}{
		{
			name:       "standard gist URL with user",
			url:        "https://gist.github.com/user/1234567890abcdef1234567890abcdef",
			expectedID: "1234567890abcdef1234567890abcdef",
		},
		{
			name:       "gist URL without user",
			url:        "https://gist.github.com/1234567890abcdef1234567890abcdef",
			expectedID: "1234567890abcdef1234567890abcdef",
		},
		{
			name:       "20 character gist ID",
			url:        "https://gist.github.com/user/1234567890abcdef1234",
			expectedID: "1234567890abcdef1234",
		},
		{
			name:       "URL with additional path components",
			url:        "https://gist.github.com/user/1234567890abcdef1234/raw/file.txt",
			expectedID: "1234567890abcdef1234",
		},
		{
			name:       "invalid URL - no valid ID",
			url:        "https://gist.github.com/user/invalid",
			expectedID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStore := NewMockStore()
			mockPromptService := NewMockPromptService()
			
			deleteCmd := &delete{
				store:         mockStore,
				promptService: mockPromptService,
			}

			extractedID := deleteCmd.extractGistID(tc.url)
			if extractedID != tc.expectedID {
				t.Errorf("Expected extractGistID(%q) = %q, got %q", tc.url, tc.expectedID, extractedID)
			}
		})
	}
}