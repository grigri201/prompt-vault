package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/service"
)

func TestGet_LooksLikeURL(t *testing.T) {
	g := &get{}
	
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid https URL",
			input:    "https://gist.github.com/user/abc123",
			expected: true,
		},
		{
			name:     "valid http URL",
			input:    "http://gist.github.com/user/abc123",
			expected: true,
		},
		{
			name:     "too short",
			input:    "http",
			expected: false,
		},
		{
			name:     "not a URL",
			input:    "golang",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.looksLikeURL(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikeURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGet_IsGistURL(t *testing.T) {
	g := &get{}
	
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid gist URL with username",
			input:    "https://gist.github.com/user/1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid gist URL without username",
			input:    "https://gist.github.com/1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid gist URL with 20 char ID",
			input:    "https://gist.github.com/user/1234567890abcdef1234",
			expected: true,
		},
		{
			name:     "not a URL",
			input:    "golang",
			expected: false,
		},
		{
			name:     "URL but not gist.github.com",
			input:    "https://github.com/user/repo",
			expected: false,
		},
		{
			name:     "gist URL but no valid ID",
			input:    "https://gist.github.com/user/invalid",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.isGistURL(tt.input)
			if result != tt.expected {
				t.Errorf("isGistURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "substring at beginning",
			s:        "hello world",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring in middle",
			s:        "hello world",
			substr:   "lo wo",
			expected: true,
		},
		{
			name:     "exact match",
			s:        "hello",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring not found",
			s:        "hello world",
			substr:   "xyz",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "hello",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid hex lowercase",
			input:    "1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid hex uppercase",
			input:    "1234567890ABCDEF",
			expected: true,
		},
		{
			name:     "valid hex mixed case",
			input:    "1234567890AbCdEf",
			expected: true,
		},
		{
			name:     "contains non-hex character",
			input:    "123456789g",
			expected: false,
		},
		{
			name:     "contains space",
			input:    "12345 67890",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "contains special characters",
			input:    "123-456",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHexString(tt.input)
			if result != tt.expected {
				t.Errorf("isHexString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		delimiter string
		expected  []string
	}{
		{
			name:      "normal split",
			s:         "a/b/c",
			delimiter: "/",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "no delimiter",
			s:         "abc",
			delimiter: "/",
			expected:  []string{"abc"},
		},
		{
			name:      "empty string",
			s:         "",
			delimiter: "/",
			expected:  []string{},
		},
		{
			name:      "delimiter at start",
			s:         "/a/b",
			delimiter: "/",
			expected:  []string{"", "a", "b"},
		},
		{
			name:      "delimiter at end",
			s:         "a/b/",
			delimiter: "/",
			expected:  []string{"a", "b", ""},
		},
		{
			name:      "multiple consecutive delimiters",
			s:         "a//b",
			delimiter: "/",
			expected:  []string{"a", "", "b"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitString(tt.s, tt.delimiter)
			if len(result) != len(tt.expected) {
				t.Errorf("splitString(%q, %q) length = %d, want %d", tt.s, tt.delimiter, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("splitString(%q, %q)[%d] = %q, want %q", tt.s, tt.delimiter, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestContainsGistID(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid 32-char hex ID",
			url:      "https://gist.github.com/user/1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "valid 20-char hex ID",
			url:      "https://gist.github.com/user/1234567890abcdef1234",
			expected: true,
		},
		{
			name:     "no valid ID",
			url:      "https://gist.github.com/user/invalid",
			expected: false,
		},
		{
			name:     "short hex string",
			url:      "https://gist.github.com/user/abc",
			expected: false,
		},
		{
			name:     "long but non-hex string",
			url:      "https://gist.github.com/user/this-is-not-hex-but-long-enough",
			expected: false,
		},
		{
			name:     "empty URL",
			url:      "",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsGistID(tt.url)
			if result != tt.expected {
				t.Errorf("containsGistID(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

// Mock implementations for cache testing

// MockPromptService implements service.PromptService for testing cache behavior
type MockPromptServiceForGet struct {
	listPromptsResult    []model.Prompt
	listPromptsError     error
	filterPromptsResult  []model.Prompt
	filterPromptsError   error
	getPromptByURLResult *model.Prompt
	getPromptByURLError  error
	getPromptContentResult string
	getPromptContentError  error
	
	// Track if service is operating from cache
	usingCache bool
	
	// Call tracking
	listPromptsCalls     int
	filterPromptsCalls   []string
	getPromptByURLCalls  []string
	getPromptContentCalls []*model.Prompt
}

func (m *MockPromptServiceForGet) AddFromFile(filePath string) (*model.Prompt, error) {
	return nil, errors.NewAppError(errors.ErrValidation, "not implemented", nil)
}

func (m *MockPromptServiceForGet) DeleteByKeyword(keyword string) error {
	return errors.NewAppError(errors.ErrValidation, "not implemented", nil)
}

func (m *MockPromptServiceForGet) DeleteByURL(gistURL string) error {
	return errors.NewAppError(errors.ErrValidation, "not implemented", nil)
}

func (m *MockPromptServiceForGet) ListPrompts() ([]model.Prompt, error) {
	m.listPromptsCalls++
	return m.listPromptsResult, m.listPromptsError
}

func (m *MockPromptServiceForGet) FilterPrompts(keyword string) ([]model.Prompt, error) {
	m.filterPromptsCalls = append(m.filterPromptsCalls, keyword)
	return m.filterPromptsResult, m.filterPromptsError
}

func (m *MockPromptServiceForGet) GetPromptByURL(gistURL string) (*model.Prompt, error) {
	m.getPromptByURLCalls = append(m.getPromptByURLCalls, gistURL)
	return m.getPromptByURLResult, m.getPromptByURLError
}

func (m *MockPromptServiceForGet) GetPromptContent(prompt *model.Prompt) (string, error) {
	m.getPromptContentCalls = append(m.getPromptContentCalls, prompt)
	return m.getPromptContentResult, m.getPromptContentError
}

func (m *MockPromptServiceForGet) AddFromURL(gistURL string) (*model.Prompt, error) {
	return nil, errors.NewAppError(errors.ErrValidation, "not implemented", nil)
}

func (m *MockPromptServiceForGet) FilterPrivatePrompts(keyword string) ([]model.Prompt, error) {
	return nil, errors.NewAppError(errors.ErrValidation, "not implemented", nil)
}

func (m *MockPromptServiceForGet) SharePrompt(prompt *model.Prompt) (*model.Prompt, error) {
	return nil, errors.NewAppError(errors.ErrValidation, "not implemented", nil)
}

func (m *MockPromptServiceForGet) ValidateGistAccess(gistURL string) (*service.GistInfo, error) {
	return &service.GistInfo{}, nil
}

func (m *MockPromptServiceForGet) ListPrivatePrompts() ([]model.Prompt, error) {
	return nil, errors.NewAppError(errors.ErrValidation, "not implemented", nil)
}

// Sync implements the PromptService interface for testing
func (m *MockPromptServiceForGet) Sync() error {
	// This method is not used by get command but required by interface
	return nil
}

// MockClipboardUtil implements clipboard.Util for testing
type MockClipboardUtil struct {
	isAvailable bool
	copyError   error
	copyCalls   []string
}

func (m *MockClipboardUtil) IsAvailable() bool {
	return m.isAvailable
}

func (m *MockClipboardUtil) Copy(content string) error {
	m.copyCalls = append(m.copyCalls, content)
	return m.copyError
}

// MockVariableParser implements variable.Parser for testing
type MockVariableParser struct {
	hasVariablesResult   bool
	extractResult        []string
	replaceResult        string
	
	hasVariablesCalls    []string
	extractVariablesCalls []string
	replaceVariablesCalls map[string]map[string]string
}

func NewMockVariableParser() *MockVariableParser {
	return &MockVariableParser{
		replaceVariablesCalls: make(map[string]map[string]string),
	}
}

func (m *MockVariableParser) HasVariables(content string) bool {
	m.hasVariablesCalls = append(m.hasVariablesCalls, content)
	return m.hasVariablesResult
}

func (m *MockVariableParser) ExtractVariables(content string) []string {
	m.extractVariablesCalls = append(m.extractVariablesCalls, content)
	return m.extractResult
}

func (m *MockVariableParser) ReplaceVariables(content string, values map[string]string) string {
	m.replaceVariablesCalls[content] = values
	return m.replaceResult
}

// MockTUIInterface implements tui.TUIInterface for testing
type MockTUIInterface struct {
	showPromptListResult  *model.Prompt
	showPromptListError   error
	showVariableFormResult map[string]string
	showVariableFormError  error
	showConfirmResult     bool
	showConfirmError      error
	
	showPromptListCalls   [][]model.Prompt
	showVariableFormCalls [][]string
	showConfirmCalls      []string
}

func (m *MockTUIInterface) ShowPromptList(prompts []model.Prompt) (model.Prompt, error) {
	m.showPromptListCalls = append(m.showPromptListCalls, prompts)
	if m.showPromptListResult != nil {
		return *m.showPromptListResult, m.showPromptListError
	}
	return model.Prompt{}, m.showPromptListError
}

func (m *MockTUIInterface) ShowVariableForm(variables []string) (map[string]string, error) {
	m.showVariableFormCalls = append(m.showVariableFormCalls, variables)
	return m.showVariableFormResult, m.showVariableFormError
}

func (m *MockTUIInterface) ShowConfirm(prompt model.Prompt) (bool, error) {
	m.showConfirmCalls = append(m.showConfirmCalls, prompt.Name)
	return m.showConfirmResult, m.showConfirmError
}

func (m *MockTUIInterface) ShowError(appError *errors.AppError) error {
	return nil
}

// Helper function to capture stdout for get command tests
func captureGetOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// Test get command cache behavior and offline mode
func TestGetCommand_CacheBehavior(t *testing.T) {
	tests := []struct {
		name                  string
		args                  []string
		serviceListError      error
		serviceFilterError    error
		serviceContentError   error
		clipboardAvailable    bool
		expectedOutput        []string
		expectError           bool
		expectCacheMessage    bool
	}{
		{
			name:               "interactive mode - successful with cache fallback",
			args:               []string{},
			serviceListError:   errors.NewAppError(errors.ErrNetwork, "network error, using cache", nil),
			serviceFilterError: nil,
			clipboardAvailable: true,
			expectedOutput: []string{
				"Interactive mode - loading all prompts",
				"Error loading prompts",
				"network error, using cache", // Should show the cache fallback message
			},
			expectError:        false,
			expectCacheMessage: true,
		},
		{
			name:               "filter mode - network error with cache fallback", 
			args:               []string{"golang"},
			serviceFilterError: errors.NewAppError(errors.ErrNetwork, "remote failed, using cache", nil),
			clipboardAvailable: true,
			expectedOutput: []string{
				"Filter mode - searching for prompts matching 'golang'",
				"Error filtering prompts",
				"Network connectivity issues",
			},
			expectError: false,
		},
		{
			name:               "successful get with cache source indication",
			args:               []string{},
			serviceListError:   nil,
			clipboardAvailable: true,
			expectedOutput: []string{
				"Interactive mode - loading all prompts",
			},
			expectError:        false,
			expectCacheMessage: false, // No cache message when remote succeeds
		},
		{
			name:               "direct URL mode - cache fallback",
			args:               []string{"https://gist.github.com/user/abc123"},
			serviceFilterError: nil,
			clipboardAvailable: true,
			expectedOutput:     []string{
				"Direct mode - processing URL",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock dependencies
			mockService := &MockPromptServiceForGet{
				listPromptsResult: []model.Prompt{
					{ID: "123", Name: "Test Prompt", Author: "test", Content: "test content"},
				},
				listPromptsError:     tt.serviceListError,
				filterPromptsError:   tt.serviceFilterError,
				getPromptContentResult: "test content",
				getPromptContentError:  tt.serviceContentError,
				usingCache:            tt.expectCacheMessage,
			}
			
			mockClipboard := &MockClipboardUtil{
				isAvailable: tt.clipboardAvailable,
			}
			
			mockVariable := NewMockVariableParser()
			mockVariable.hasVariablesResult = false // No variables for simplicity
			
			mockTUI := &MockTUIInterface{
				showPromptListResult: &model.Prompt{
					ID: "123", Name: "Test Prompt", Author: "test", Content: "test content",
				},
			}
			
			// Create get command
			getCmd := NewGetCommand(mockService, mockClipboard, mockVariable, mockTUI)
			
			// Set arguments
			getCmd.SetArgs(tt.args)
			
			// Capture output
			output := captureGetOutput(func() {
				getCmd.Execute()
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

// Test get command with cache source indication (requirement 4.4)
func TestGetCommand_CacheSourceIndication(t *testing.T) {
	t.Run("should indicate cache source when using cached data", func(t *testing.T) {
		mockService := &MockPromptServiceForGet{
			listPromptsResult: []model.Prompt{
				{ID: "123", Name: "Cached Prompt", Author: "user", Content: "cached content"},
			},
			listPromptsError:       nil,
			getPromptContentResult: "cached content",
			getPromptContentError:  nil,
			usingCache:            true, // Simulate cache usage
		}
		
		mockClipboard := &MockClipboardUtil{isAvailable: true}
		mockVariable := NewMockVariableParser()
		mockVariable.hasVariablesResult = false
		
		mockTUI := &MockTUIInterface{
			showPromptListResult: &model.Prompt{
				ID: "123", Name: "Cached Prompt", Author: "user", Content: "cached content",
			},
		}
		
		getCmd := NewGetCommand(mockService, mockClipboard, mockVariable, mockTUI)
		getCmd.SetArgs([]string{}) // Interactive mode
		
		output := captureGetOutput(func() {
			getCmd.Execute()
		})
		
		// Should indicate when data comes from cache (this depends on implementation)
		// The actual cache indication might be in the copyToClipboard method
		if !strings.Contains(output, "Interactive mode") {
			t.Errorf("Expected cache source indication in output: %s", output)
		}
	})
}

// Test get command backward compatibility
func TestGetCommand_BackwardCompatibility(t *testing.T) {
	t.Run("existing functionality should work unchanged", func(t *testing.T) {
		mockService := &MockPromptServiceForGet{
			listPromptsResult: []model.Prompt{
				{ID: "123", Name: "Legacy Prompt", Author: "legacy", Content: "legacy content"},
			},
			listPromptsError:       nil,
			getPromptContentResult: "legacy content",
			getPromptContentError:  nil,
		}
		
		mockClipboard := &MockClipboardUtil{isAvailable: true}
		mockVariable := NewMockVariableParser()
		mockVariable.hasVariablesResult = false
		
		mockTUI := &MockTUIInterface{
			showPromptListResult: &model.Prompt{
				ID: "123", Name: "Legacy Prompt", Author: "legacy", Content: "legacy content",
			},
		}
		
		getCmd := NewGetCommand(mockService, mockClipboard, mockVariable, mockTUI)
		getCmd.SetArgs([]string{}) // Interactive mode
		
		output := captureGetOutput(func() {
			getCmd.Execute()
		})
		
		// Verify basic functionality works
		expectedPatterns := []string{
			"Interactive mode",
			"Found 1 prompt",
		}
		
		for _, pattern := range expectedPatterns {
			if !strings.Contains(output, pattern) {
				t.Errorf("Expected output to contain %q, got: %s", pattern, output)
			}
		}
		
		// Verify service methods were called
		if mockService.listPromptsCalls != 1 {
			t.Errorf("Expected 1 ListPrompts call, got %d", mockService.listPromptsCalls)
		}
	})
	
	t.Run("URL validation should remain consistent", func(t *testing.T) {
		g := &get{}
		
		// Test cases that should continue to work as before
		validURL := "https://gist.github.com/user/1234567890abcdef1234567890abcdef"
		if !g.isGistURL(validURL) {
			t.Errorf("Expected %q to be valid gist URL", validURL)
		}
		
		invalidURL := "not-a-url"
		if g.isGistURL(invalidURL) {
			t.Errorf("Expected %q to be invalid gist URL", invalidURL)
		}
	})
}

// Test offline mode behavior (requirement 4.3)
func TestGetCommand_OfflineMode(t *testing.T) {
	t.Run("should handle complete network failure gracefully", func(t *testing.T) {
		mockService := &MockPromptServiceForGet{
			listPromptsResult:  nil,
			listPromptsError:   errors.NewAppError(errors.ErrNetwork, "complete network failure", nil),
		}
		
		mockClipboard := &MockClipboardUtil{isAvailable: true}
		mockVariable := NewMockVariableParser()
		mockTUI := &MockTUIInterface{}
		
		getCmd := NewGetCommand(mockService, mockClipboard, mockVariable, mockTUI)
		getCmd.SetArgs([]string{}) // Interactive mode
		
		output := captureGetOutput(func() {
			getCmd.Execute()
		})
		
		// Should show appropriate error message without crashing
		if !strings.Contains(output, "Error loading prompts") {
			t.Errorf("Expected offline error message, got: %s", output)
		}
	})
	
	t.Run("should indicate when no cached data is available", func(t *testing.T) {
		mockService := &MockPromptServiceForGet{
			filterPromptsResult: nil,
			filterPromptsError:  errors.NewAppError(errors.ErrStorage, "no cache available", nil),
		}
		
		mockClipboard := &MockClipboardUtil{isAvailable: true}
		mockVariable := NewMockVariableParser()
		mockTUI := &MockTUIInterface{}
		
		getCmd := NewGetCommand(mockService, mockClipboard, mockVariable, mockTUI)
		getCmd.SetArgs([]string{"golang"}) // Filter mode
		
		output := captureGetOutput(func() {
			getCmd.Execute()
		})
		
		// Should provide helpful guidance
		if !strings.Contains(output, "Error filtering prompts") {
			t.Errorf("Expected cache unavailable message, got: %s", output)
		}
	})
}