package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/validator"
)

// MockStore is a mock implementation of infra.Store for testing
type MockStore struct {
	addError    error
	addFunc     func(model.Prompt) error
	deleteError error
	deleteFunc  func(string) error
	getError    error
	getFunc     func(string) ([]model.Prompt, error)
	listError   error
	listFunc    func() ([]model.Prompt, error)
	prompts     []model.Prompt
}

func (m *MockStore) List() ([]model.Prompt, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	if m.listFunc != nil {
		return m.listFunc()
	}
	return m.prompts, nil
}

func (m *MockStore) Add(prompt model.Prompt) error {
	if m.addFunc != nil {
		return m.addFunc(prompt)
	}
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
	if m.deleteFunc != nil {
		return m.deleteFunc(keyword)
	}
	// Remove from prompts slice for realistic behavior
	for i, prompt := range m.prompts {
		if prompt.ID == keyword {
			m.prompts = append(m.prompts[:i], m.prompts[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockStore) Update(prompt model.Prompt) error {
	return nil
}

func (m *MockStore) Get(keyword string) ([]model.Prompt, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if m.getFunc != nil {
		return m.getFunc(keyword)
	}
	// Return matching prompts for realistic behavior
	var matches []model.Prompt
	for _, prompt := range m.prompts {
		if strings.Contains(strings.ToLower(prompt.Name), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(prompt.Author), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(prompt.Description), strings.ToLower(keyword)) ||
			prompt.ID == keyword {
			matches = append(matches, prompt)
		}
	}
	return matches, nil
}

// MockYAMLValidator is a mock implementation of validator.YAMLValidator for testing
type MockYAMLValidator struct {
	validatePromptFileFunc func(content []byte) (*validator.PromptFileContent, error)
	validateRequiredFunc   func(prompt *validator.PromptFileContent) error
}

func (m *MockYAMLValidator) ValidatePromptFile(content []byte) (*validator.PromptFileContent, error) {
	if m.validatePromptFileFunc != nil {
		return m.validatePromptFileFunc(content)
	}
	// Default successful validation
	return &validator.PromptFileContent{
		Metadata: validator.PromptMetadata{
			Name:        "Test Prompt",
			Author:      "Test Author",
			Description: "Test Description",
			Tags:        []string{"test"},
			Version:     "1.0",
		},
		Content: "This is test prompt content",
	}, nil
}

func (m *MockYAMLValidator) ValidateRequired(prompt *validator.PromptFileContent) error {
	if m.validateRequiredFunc != nil {
		return m.validateRequiredFunc(prompt)
	}
	return nil
}

func TestPromptService_AddFromFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	testContent := `name: "Test Prompt"
author: "Test Author"
description: "Test Description"
tags:
  - "test"
version: "1.0"
---
This is test prompt content`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testCases := []struct {
		name                  string
		filePath              string
		mockStore             *MockStore
		mockValidator         *MockYAMLValidator
		expectError           bool
		expectedErrorType     errors.ErrorType
		expectedPromptName    string
		expectedPromptAuthor  string
		expectedPromptContent string
		expectedPromptTags    []string
		expectedPromptVersion string
		expectedPromptDesc    string
	}{
		{
			name:     "successful add from file",
			filePath: testFile,
			mockStore: &MockStore{
				addFunc: func(prompt model.Prompt) error {
					// Verify prompt fields are populated correctly
					if prompt.Name != "Test Prompt" {
						t.Errorf("Expected prompt name 'Test Prompt', got %q", prompt.Name)
					}
					if prompt.Author != "Test Author" {
						t.Errorf("Expected prompt author 'Test Author', got %q", prompt.Author)
					}
					expectedFullContent := `name: "Test Prompt"
author: "Test Author"
description: "Test Description"
tags:
  - "test"
version: "1.0"
---
This is test prompt content`
						if prompt.Content != expectedFullContent {
							t.Errorf("Expected prompt content %q, got %q", expectedFullContent, prompt.Content)
					}
					return nil
				},
			},
			mockValidator:         &MockYAMLValidator{},
			expectError:           false,
			expectedPromptName:    "Test Prompt",
			expectedPromptAuthor:  "Test Author",
			expectedPromptContent: `name: "Test Prompt"
author: "Test Author"
description: "Test Description"
tags:
  - "test"
version: "1.0"
---
This is test prompt content`,
			expectedPromptTags:    []string{"test"},
			expectedPromptVersion: "1.0",
			expectedPromptDesc:    "Test Description",
		},
		{
			name:              "empty file path",
			filePath:          "",
			mockStore:         &MockStore{},
			mockValidator:     &MockYAMLValidator{},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:              "whitespace only file path",
			filePath:          "   ",
			mockStore:         &MockStore{},
			mockValidator:     &MockYAMLValidator{},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:              "file not found",
			filePath:          "/non/existent/file.yaml",
			mockStore:         &MockStore{},
			mockValidator:     &MockYAMLValidator{},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:      "YAML validation failure",
			filePath:  testFile,
			mockStore: &MockStore{},
			mockValidator: &MockYAMLValidator{
				validatePromptFileFunc: func(content []byte) (*validator.PromptFileContent, error) {
					return nil, errors.NewAppError(
						errors.ErrValidation,
						"invalid YAML format",
						errors.ErrInvalidYAML,
					)
				},
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:      "required field validation failure",
			filePath:  testFile,
			mockStore: &MockStore{},
			mockValidator: &MockYAMLValidator{
				validateRequiredFunc: func(prompt *validator.PromptFileContent) error {
					return errors.ValidationError{
						Field:   "name",
						Message: "name is required and cannot be empty",
					}
				},
			},
			expectError: true,
		},
		{
			name:     "store add failure",
			filePath: testFile,
			mockStore: &MockStore{
				addError: errors.NewAppError(
					errors.ErrStorage,
					"failed to create gist",
					nil,
				),
			},
			mockValidator:     &MockYAMLValidator{},
			expectError:       true,
			expectedErrorType: errors.ErrStorage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with mocks
			service := NewPromptService(tc.mockStore, tc.mockValidator)

			// Test AddFromFile
			prompt, err := service.AddFromFile(tc.filePath)

			// Check error expectations
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Check error type if specified
				if tc.expectedErrorType != errors.ErrUnknown {
					appErr, ok := err.(errors.AppError)
					if !ok {
						// Could be ValidationError, check that separately
						if _, isValidationErr := err.(errors.ValidationError); !isValidationErr {
							t.Errorf("Expected AppError or ValidationError, got %T: %v", err, err)
						}
					} else if appErr.Type != tc.expectedErrorType {
						t.Errorf("Expected error type %v, got %v", tc.expectedErrorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify prompt fields
				if prompt == nil {
					t.Errorf("Expected prompt but got nil")
					return
				}

				if tc.expectedPromptName != "" && prompt.Name != tc.expectedPromptName {
					t.Errorf("Expected prompt name %q, got %q", tc.expectedPromptName, prompt.Name)
				}

				if tc.expectedPromptAuthor != "" && prompt.Author != tc.expectedPromptAuthor {
					t.Errorf("Expected prompt author %q, got %q", tc.expectedPromptAuthor, prompt.Author)
				}

				if tc.expectedPromptContent != "" && prompt.Content != tc.expectedPromptContent {
					t.Errorf("Expected prompt content %q, got %q", tc.expectedPromptContent, prompt.Content)
				}

				if tc.expectedPromptDesc != "" && prompt.Description != tc.expectedPromptDesc {
					t.Errorf("Expected prompt description %q, got %q", tc.expectedPromptDesc, prompt.Description)
				}

				if tc.expectedPromptVersion != "" && prompt.Version != tc.expectedPromptVersion {
					t.Errorf("Expected prompt version %q, got %q", tc.expectedPromptVersion, prompt.Version)
				}

				if len(tc.expectedPromptTags) > 0 {
					if len(prompt.Tags) != len(tc.expectedPromptTags) {
						t.Errorf("Expected %d tags, got %d", len(tc.expectedPromptTags), len(prompt.Tags))
					} else {
						for i, expectedTag := range tc.expectedPromptTags {
							if i < len(prompt.Tags) && prompt.Tags[i] != expectedTag {
								t.Errorf("Expected tag[%d] %q, got %q", i, expectedTag, prompt.Tags[i])
							}
						}
					}
				}

				// Verify ID and GistURL are initially empty (will be set by store)
				if prompt.ID != "" {
					t.Errorf("Expected empty ID, got %q", prompt.ID)
				}
				if prompt.GistURL != "" {
					t.Errorf("Expected empty GistURL, got %q", prompt.GistURL)
				}
			}
		})
	}
}

func TestPromptService_AddFromFile_EdgeCases(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	testCases := []struct {
		name          string
		setupFile     func(t *testing.T) string
		mockValidator *MockYAMLValidator
		expectError   bool
	}{
		{
			name: "file with read permissions issue",
			setupFile: func(t *testing.T) string {
				file := filepath.Join(tmpDir, "no_read.yaml")
				content := `name: "Test"
author: "Author"
---
Content`
				err := os.WriteFile(file, []byte(content), 0000) // No read permission
				if err != nil {
					t.Skip("Cannot create file with no read permissions on this system")
				}
				return file
			},
			mockValidator: &MockYAMLValidator{},
			expectError:   true,
		},
		{
			name: "empty file causes validation error",
			setupFile: func(t *testing.T) string {
				file := filepath.Join(tmpDir, "empty.yaml")
				err := os.WriteFile(file, []byte(""), 0644)
				if err != nil {
					t.Fatalf("Failed to create empty test file: %v", err)
				}
				return file
			},
			mockValidator: &MockYAMLValidator{
				validatePromptFileFunc: func(content []byte) (*validator.PromptFileContent, error) {
					if len(content) == 0 {
						return nil, errors.NewAppError(
							errors.ErrValidation,
							"empty file",
							errors.ErrInvalidYAML,
						)
					}
					return &validator.PromptFileContent{
						Metadata: validator.PromptMetadata{Name: "Test", Author: "Author"},
						Content:  "Content",
					}, nil
				},
			},
			expectError: true,
		},
		{
			name: "prompt file with invalid format (missing required fields)",
			setupFile: func(t *testing.T) string {
				file := filepath.Join(tmpDir, "invalid_format.yaml")
				content := `title: "Invalid Prompt"
description: "Missing name and author fields"
---
This prompt is missing required fields`
				err := os.WriteFile(file, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create invalid format test file: %v", err)
				}
				return file
			},
			mockValidator: &MockYAMLValidator{
				validatePromptFileFunc: func(content []byte) (*validator.PromptFileContent, error) {
					return nil, errors.NewAppError(
						errors.ErrValidation,
						"missing required fields: name, author",
						errors.ErrMissingRequired,
					)
				},
			},
			expectError: true,
		},
		{
			name: "md file cannot pass validation (invalid markdown structure)",
			setupFile: func(t *testing.T) string {
				file := filepath.Join(tmpDir, "invalid.md")
				content := `name: "Test Prompt"
author: "Test Author"
---
# Invalid Markdown
This markdown file has invalid structure that cannot be parsed properly.
[Broken link](
Missing closing bracket and other syntax errors.`
				err := os.WriteFile(file, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create invalid markdown test file: %v", err)
				}
				return file
			},
			mockValidator: &MockYAMLValidator{
				validatePromptFileFunc: func(content []byte) (*validator.PromptFileContent, error) {
					return nil, errors.NewAppError(
						errors.ErrValidation,
						"invalid markdown structure: broken syntax elements",
						errors.ErrInvalidMetadata,
					)
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.setupFile(t)
			defer func() {
				// Clean up by changing permissions back if needed
				os.Chmod(filePath, 0644)
				os.Remove(filePath)
			}()

			mockStore := &MockStore{}
			service := NewPromptService(mockStore, tc.mockValidator)

			_, err := service.AddFromFile(filePath)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPromptService_AddFromFile_ValidatorIntegration(t *testing.T) {
	// Create temporary test file with various content types
	tmpDir := t.TempDir()

	testCases := []struct {
		name        string
		fileContent string
		validator   *MockYAMLValidator
		expectError bool
	}{
		{
			name: "validator returns complex validation error",
			fileContent: `name: "Test"
author: "Author"
---
Content`,
			validator: &MockYAMLValidator{
				validatePromptFileFunc: func(content []byte) (*validator.PromptFileContent, error) {
					// Simulate complex parsing scenario
					return &validator.PromptFileContent{
						Metadata: validator.PromptMetadata{
							Name:   "Test",
							Author: "Author",
						},
						Content: "Content",
					}, nil
				},
				validateRequiredFunc: func(prompt *validator.PromptFileContent) error {
					// Simulate validation that checks content beyond just required fields
					if len(prompt.Content) < 10 {
						return errors.ValidationError{
							Field:   "content",
							Message: "content must be at least 10 characters",
						}
					}
					return nil
				},
			},
			expectError: true,
		},
		{
			name: "validator handles content with special characters",
			fileContent: `name: "Test with 特殊字符"
author: "Author with émojis 🚀"
description: "Description with\nnewlines and\ttabs"
tags:
  - "tag with spaces"
  - "标签"
version: "1.0.0-beta.1"
---
Content with special characters: àáâãäå æçèéêë ìíîï ñòóôõö øùúûü ý`,
			validator: &MockYAMLValidator{
				validatePromptFileFunc: func(content []byte) (*validator.PromptFileContent, error) {
					return &validator.PromptFileContent{
						Metadata: validator.PromptMetadata{
							Name:        "Test with 特殊字符",
							Author:      "Author with émojis 🚀",
							Description: "Description with\nnewlines and\ttabs",
							Tags:        []string{"tag with spaces", "标签"},
							Version:     "1.0.0-beta.1",
						},
						Content: "Content with special characters: àáâãäå æçèéêë ìíîï ñòóôõö øùúûü ý",
					}, nil
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test_"+tc.name+".yaml")
			err := os.WriteFile(testFile, []byte(tc.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(testFile)

			mockStore := &MockStore{}
			service := NewPromptService(mockStore, tc.validator)

			_, err = service.AddFromFile(testFile)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test data helpers for delete method tests
func createTestPrompts() []model.Prompt {
	return []model.Prompt{
		{
			ID:          "gist1234567890",
			Name:        "Go Code Review",
			Author:      "John Doe",
			Description: "A comprehensive guide for Go code review",
			Tags:        []string{"go", "review", "best-practice"},
			Version:     "1.0",
			Content:     "Go code review guidelines...",
			GistURL:     "https://gist.github.com/johndoe/gist1234567890",
		},
		{
			ID:          "gist4567890123",
			Name:        "SQL Query Optimizer",
			Author:      "Jane Smith",
			Description: "Tips for optimizing SQL queries",
			Tags:        []string{"sql", "performance"},
			Version:     "2.1",
			Content:     "SQL optimization techniques...",
			GistURL:     "https://gist.github.com/janesmith/gist4567890123",
		},
		{
			ID:          "gist7890123456",
			Name:        "Docker Best Practices",
			Author:      "John Doe",
			Description: "Best practices for Docker containerization",
			Tags:        []string{"docker", "containers"},
			Version:     "1.5",
			Content:     "Docker containerization guidelines...",
			GistURL:     "https://gist.github.com/johndoe/gist7890123456",
		},
	}
}

func TestPromptService_DeleteByKeyword(t *testing.T) {
	testPrompts := createTestPrompts()

	testCases := []struct {
		name              string
		keyword           string
		mockStore         *MockStore
		expectError       bool
		expectedErrorType errors.ErrorType
		expectedMessage   string
	}{
		{
			name:    "successful delete with single match",
			keyword: "SQL Query Optimizer",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError: false,
		},
		{
			name:    "successful delete with multiple matches",
			keyword: "John Doe",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError: false,
		},
		{
			name:    "empty keyword",
			keyword: "",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "whitespace only keyword",
			keyword: "   ",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "no matching prompts",
			keyword: "nonexistent",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError: true,
		},
		{
			name:    "store Get failure",
			keyword: "test",
			mockStore: &MockStore{
				prompts:  testPrompts,
				getError: errors.NewAppError(errors.ErrStorage, "database connection failed", nil),
			},
			expectError:       true,
			expectedErrorType: errors.ErrStorage,
		},
		{
			name:    "store Delete failure",
			keyword: "Go Code Review",
			mockStore: &MockStore{
				getFunc: func(keyword string) ([]model.Prompt, error) {
					// Return the matching prompt for Get, then fail on Delete
					prompts := createTestPrompts()
					for _, prompt := range prompts {
						if strings.Contains(strings.ToLower(prompt.Name), strings.ToLower(keyword)) {
							return []model.Prompt{prompt}, nil
						}
					}
					return []model.Prompt{}, nil
				},
				deleteFunc: func(id string) error {
					// Simulate permission denied when trying to delete
					return errors.NewAppError(errors.ErrAuth, "permission denied", nil)
				},
			},
			expectError:       true,
			expectedErrorType: errors.ErrStorage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockValidator := &MockYAMLValidator{}
			service := NewPromptService(tc.mockStore, mockValidator)

			err := service.DeleteByKeyword(tc.keyword)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if tc.expectedErrorType != errors.ErrUnknown {
					appErr, ok := err.(errors.AppError)
					if !ok {
						t.Errorf("Expected AppError, got %T: %v", err, err)
					} else if appErr.Type != tc.expectedErrorType {
						t.Errorf("Expected error type %v, got %v", tc.expectedErrorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}
		})
	}
}

func TestPromptService_DeleteByURL(t *testing.T) {
	testPrompts := createTestPrompts()

	testCases := []struct {
		name              string
		gistURL           string
		mockStore         *MockStore
		expectError       bool
		expectedErrorType errors.ErrorType
	}{
		{
			name:    "successful delete with valid URL",
			gistURL: "https://gist.github.com/johndoe/gist1234567890",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError: false,
		},
		{
			name:    "empty URL",
			gistURL: "",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "invalid URL format - not a GitHub URL",
			gistURL: "https://gitlab.com/user/repo",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "invalid URL format - malformed URL",
			gistURL: "not-a-url",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "invalid Gist ID format - too short",
			gistURL: "https://gist.github.com/user/abc123",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "prompt not found for valid URL",
			gistURL: "https://gist.github.com/user/nonexistent123456",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError: true,
		},
		{
			name:    "store Get failure",
			gistURL: "https://gist.github.com/johndoe/gist1234567890",
			mockStore: &MockStore{
				prompts:  testPrompts,
				getError: errors.NewAppError(errors.ErrStorage, "network error", nil),
			},
			expectError:       true,
			expectedErrorType: errors.ErrStorage,
		},
		{
			name:    "store Delete failure",
			gistURL: "https://gist.github.com/johndoe/gist1234567890",
			mockStore: &MockStore{
				getFunc: func(keyword string) ([]model.Prompt, error) {
					// Return the matching prompt for Get, then fail on Delete
					prompts := createTestPrompts()
					for _, prompt := range prompts {
						if prompt.ID == keyword {
							return []model.Prompt{prompt}, nil
						}
					}
					return []model.Prompt{}, nil
				},
				deleteFunc: func(id string) error {
					return errors.NewAppError(errors.ErrAuth, "github permission error", nil)
				},
			},
			expectError:       true,
			expectedErrorType: errors.ErrStorage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockValidator := &MockYAMLValidator{}
			service := NewPromptService(tc.mockStore, mockValidator)

			err := service.DeleteByURL(tc.gistURL)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if tc.expectedErrorType != errors.ErrUnknown {
					appErr, ok := err.(errors.AppError)
					if !ok {
						t.Errorf("Expected AppError, got %T: %v", err, err)
					} else if appErr.Type != tc.expectedErrorType {
						t.Errorf("Expected error type %v, got %v", tc.expectedErrorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}
		})
	}
}

func TestPromptService_ListForDeletion(t *testing.T) {
	testPrompts := createTestPrompts()

	testCases := []struct {
		name              string
		mockStore         *MockStore
		expectError       bool
		expectedErrorType errors.ErrorType
		expectedCount     int
	}{
		{
			name: "successful list with prompts",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:   false,
			expectedCount: 3,
		},
		{
			name: "empty prompt list",
			mockStore: &MockStore{
				prompts: []model.Prompt{},
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name: "no index error",
			mockStore: &MockStore{
				listError: infra.ErrNoIndex,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name: "empty index error",
			mockStore: &MockStore{
				listError: infra.ErrEmptyIndex,
			},
			expectError: true,
		},
		{
			name: "store list failure",
			mockStore: &MockStore{
				listError: errors.NewAppError(errors.ErrStorage, "database error", nil),
			},
			expectError:       true,
			expectedErrorType: errors.ErrStorage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockValidator := &MockYAMLValidator{}
			service := NewPromptService(tc.mockStore, mockValidator)

			prompts, err := service.ListForDeletion()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if tc.expectedErrorType != errors.ErrUnknown {
					appErr, ok := err.(errors.AppError)
					if !ok {
						t.Errorf("Expected AppError, got %T: %v", err, err)
					} else if appErr.Type != tc.expectedErrorType {
						t.Errorf("Expected error type %v, got %v", tc.expectedErrorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if len(prompts) != tc.expectedCount {
					t.Errorf("Expected %d prompts, got %d", tc.expectedCount, len(prompts))
				}
			}
		})
	}
}

func TestPromptService_FilterForDeletion(t *testing.T) {
	testPrompts := createTestPrompts()

	testCases := []struct {
		name              string
		keyword           string
		mockStore         *MockStore
		expectError       bool
		expectedErrorType errors.ErrorType
		expectedCount     int
	}{
		{
			name:    "successful filter with matches",
			keyword: "go",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:   false,
			expectedCount: 1, // Only "Go Code Review" matches
		},
		{
			name:    "successful filter with multiple matches",
			keyword: "John Doe",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:   false,
			expectedCount: 2, // Two prompts by John Doe
		},
		{
			name:    "no matches found",
			keyword: "nonexistent",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:   false,
			expectedCount: 0, // Empty slice, not an error
		},
		{
			name:    "empty keyword",
			keyword: "",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "whitespace only keyword",
			keyword: "   ",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "case insensitive matching",
			keyword: "SQL",
			mockStore: &MockStore{
				prompts: testPrompts,
			},
			expectError:   false,
			expectedCount: 1, // "SQL Query Optimizer" should match
		},
		{
			name:    "no index error",
			keyword: "test",
			mockStore: &MockStore{
				getError: infra.ErrNoIndex,
			},
			expectError:       true,
			expectedErrorType: errors.ErrValidation,
		},
		{
			name:    "empty index error",
			keyword: "test",
			mockStore: &MockStore{
				getError: infra.ErrEmptyIndex,
			},
			expectError: true,
		},
		{
			name:    "store get failure",
			keyword: "test",
			mockStore: &MockStore{
				getError: errors.NewAppError(errors.ErrStorage, "network error", nil),
			},
			expectError:       true,
			expectedErrorType: errors.ErrStorage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockValidator := &MockYAMLValidator{}
			service := NewPromptService(tc.mockStore, mockValidator)

			prompts, err := service.FilterForDeletion(tc.keyword)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if tc.expectedErrorType != errors.ErrUnknown {
					appErr, ok := err.(errors.AppError)
					if !ok {
						t.Errorf("Expected AppError, got %T: %v", err, err)
					} else if appErr.Type != tc.expectedErrorType {
						t.Errorf("Expected error type %v, got %v", tc.expectedErrorType, appErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if len(prompts) != tc.expectedCount {
					t.Errorf("Expected %d prompts, got %d", tc.expectedCount, len(prompts))
				}
			}
		})
	}
}

func TestPromptService_DeleteByKeyword_EdgeCases(t *testing.T) {
	testCases := []struct {
		name      string
		keyword   string
		mockStore *MockStore
		expectError bool
	}{
		{
			name:    "delete with special characters in keyword",
			keyword: "特殊字符 with émojis 🚀",
			mockStore: &MockStore{
				prompts: []model.Prompt{
					{
						ID:          "special123",
						Name:        "特殊字符 with émojis 🚀",
						Author:      "Test Author",
						Description: "Test with special characters",
						Content:     "Test content",
						GistURL:     "https://gist.github.com/test/special123",
					},
				},
			},
			expectError: false,
		},
		{
			name:    "delete with very long keyword",
			keyword: strings.Repeat("long", 100),
			mockStore: &MockStore{
				prompts: []model.Prompt{},
			},
			expectError: true, // No matching prompts
		},
		{
			name:    "delete when store returns empty results",
			keyword: "test",
			mockStore: &MockStore{
				getFunc: func(keyword string) ([]model.Prompt, error) {
					return []model.Prompt{}, nil
				},
			},
			expectError: true, // Should return ErrNoPromptsToDelete
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockValidator := &MockYAMLValidator{}
			service := NewPromptService(tc.mockStore, mockValidator)

			err := service.DeleteByKeyword(tc.keyword)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPromptService_DeleteByURL_EdgeCases(t *testing.T) {
	testCases := []struct {
		name      string
		gistURL   string
		mockStore *MockStore
		expectError bool
	}{
		{
			name:    "GitHub.com URL format",
			gistURL: "https://github.com/user/gist123456",
			mockStore: &MockStore{
				prompts: []model.Prompt{
					{
						ID:      "gist123456",
						Name:    "Test Prompt",
						Author:  "Test User",
						GistURL: "https://github.com/user/gist123456",
					},
				},
			},
			expectError: false,
		},
		{
			name:    "URL with query parameters",
			gistURL: "https://gist.github.com/user/gist123456?tab=readme",
			mockStore: &MockStore{
				prompts: []model.Prompt{
					{
						ID:      "gist123456",
						Name:    "Test Prompt",
						Author:  "Test User",
						GistURL: "https://gist.github.com/user/gist123456",
					},
				},
			},
			expectError: false,
		},
		{
			name:    "URL with fragment",
			gistURL: "https://gist.github.com/user/gist123456#file-test-md",
			mockStore: &MockStore{
				prompts: []model.Prompt{
					{
						ID:      "gist123456",
						Name:    "Test Prompt",
						Author:  "Test User",
						GistURL: "https://gist.github.com/user/gist123456",
					},
				},
			},
			expectError: false,
		},
		{
			name:        "invalid characters in gist ID",
			gistURL:     "https://gist.github.com/user/invalid@id",
			mockStore:   &MockStore{},
			expectError: true,
		},
		{
			name:        "whitespace only URL",
			gistURL:     "   ",
			mockStore:   &MockStore{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockValidator := &MockYAMLValidator{}
			service := NewPromptService(tc.mockStore, mockValidator)

			err := service.DeleteByURL(tc.gistURL)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
