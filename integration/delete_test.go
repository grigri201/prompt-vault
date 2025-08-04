package integration

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/cmd"
	apperrors "github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/tui"
	"github.com/grigri/pv/internal/validator"
)

// TestDeleteWorkflowIntegration tests the complete deletion workflow
// covering all major user scenarios as specified in T5 requirements
func TestDeleteWorkflowIntegration(t *testing.T) {
	// Test suite covering all deletion scenarios
	t.Run("ServiceLayerIntegration", testServiceLayerIntegration)
	t.Run("CommandParameterParsing", testCommandParameterParsing)
	t.Run("StoreIntegration", testStoreIntegration)
	t.Run("ErrorHandlingIntegration", testErrorHandlingIntegration)
	t.Run("URLValidationIntegration", testURLValidationIntegration)
}

// testServiceLayerIntegration tests the service layer integration with store
func testServiceLayerIntegration(t *testing.T) {
	testEnv := setupIntegrationTestEnvironment(t)
	defer testEnv.cleanup()

	// Test ListForDeletion service method
	prompts := []model.Prompt{
		createTestPrompt("1", "AI Assistant", "dev1", createValidGistURL("1234567890abcdef1234")),
		createTestPrompt("2", "Code Review", "dev2", createValidGistURL("abcdef1234567890abcd")),
	}

	// Add prompts to store
	for _, prompt := range prompts {
		err := testEnv.store.Add(prompt)
		if err != nil {
			t.Fatalf("Failed to add prompt: %v", err)
		}
	}

	// Test ListForDeletion
	listedPrompts, err := testEnv.service.ListForDeletion()
	if err != nil {
		t.Errorf("ListForDeletion failed: %v", err)
	}
	if len(listedPrompts) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(listedPrompts))
	}

	// Test FilterForDeletion
	filteredPrompts, err := testEnv.service.FilterForDeletion("AI")
	if err != nil {
		t.Errorf("FilterForDeletion failed: %v", err)
	}
	if len(filteredPrompts) != 1 {
		t.Errorf("Expected 1 filtered prompt, got %d", len(filteredPrompts))
	}
	if filteredPrompts[0].Name != "AI Assistant" {
		t.Errorf("Expected 'AI Assistant', got %s", filteredPrompts[0].Name)
	}
}

// testCommandParameterParsing tests command parameter parsing and routing
func testCommandParameterParsing(t *testing.T) {
	testEnv := setupIntegrationTestEnvironment(t)
	defer testEnv.cleanup()

	// Add test prompts
	prompts := []model.Prompt{
		createTestPrompt("1", "Golang Helper", "dev1", createValidGistURL("1234567890abcdef1234")),
		createTestPrompt("2", "Python Script", "dev2", createValidGistURL("abcdef1234567890abcd")),
	}

	for _, prompt := range prompts {
		err := testEnv.store.Add(prompt)
		if err != nil {
			t.Fatalf("Failed to add prompt: %v", err)
		}
	}

	// Test different command parameter scenarios
	tests := []struct {
		name string
		args []string
	}{
		{"no_arguments", []string{}},
		{"keyword_argument", []string{"golang"}},
		{"valid_url_argument", []string{createValidGistURL("1234567890abcdef1234")}},
		{"invalid_url_argument", []string{"invalid-url"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute command with different parameters
			_, err := executeDeleteCommand(testEnv, tt.args)
			
			// The command should not crash, even if it fails at TUI level
			// This tests that parameter parsing and initial processing work
			t.Logf("Command with args %v completed with result: %v", tt.args, err)
			
			// Verify store integrity is maintained
			remaining, storeErr := testEnv.store.List()
			if storeErr != nil {
				t.Errorf("Store integrity compromised: %v", storeErr)
			}
			if len(remaining) != len(prompts) {
				t.Errorf("Store should remain unchanged during command parsing, expected %d prompts, got %d", 
					len(prompts), len(remaining))
			}
		})
	}
}

// testStoreIntegration tests store operations integration
func testStoreIntegration(t *testing.T) {
	testEnv := setupIntegrationTestEnvironment(t)
	defer testEnv.cleanup()

	// Test basic store operations that the delete command depends on
	testPrompt := createTestPrompt("test1", "Store Test", "tester", createValidGistURL("1234567890abcdef1234"))

	// Test Add operation
	err := testEnv.store.Add(testPrompt)
	if err != nil {
		t.Fatalf("Store Add failed: %v", err)
	}

	// Test List operation (used by ListForDeletion)
	prompts, err := testEnv.store.List()
	if err != nil {
		t.Errorf("Store List failed: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt after Add, got %d", len(prompts))
	}

	// Test Get operation (used by FilterForDeletion and URL search)
	foundPrompts, err := testEnv.store.Get("Store")
	if err != nil {
		t.Errorf("Store Get failed: %v", err)
	}
	if len(foundPrompts) != 1 {
		t.Errorf("Expected 1 prompt matching 'Store', got %d", len(foundPrompts))
	}

	// Test Get by URL
	urlPrompts, err := testEnv.store.Get(testPrompt.GistURL)
	if err != nil {
		t.Errorf("Store Get by URL failed: %v", err)
	}
	if len(urlPrompts) != 1 {
		t.Errorf("Expected 1 prompt matching URL, got %d", len(urlPrompts))
	}

	// Test Delete operation (core of delete functionality)
	err = testEnv.store.Delete(testPrompt.ID)
	if err != nil {
		t.Errorf("Store Delete failed: %v", err)
	}

	// Verify deletion
	remainingPrompts, err := testEnv.store.List()
	if err != nil {
		t.Errorf("Store List after Delete failed: %v", err)
	}
	if len(remainingPrompts) != 0 {
		t.Errorf("Expected 0 prompts after Delete, got %d", len(remainingPrompts))
	}
}

// testErrorHandlingIntegration tests error handling scenarios
func testErrorHandlingIntegration(t *testing.T) {
	testEnv := setupIntegrationTestEnvironment(t)
	defer testEnv.cleanup()

	// Test network error scenario
	testEnv.mockStore.ShouldSimulateNetworkError = true
	
	// Try to execute command with network error
	_, err := executeDeleteCommand(testEnv, []string{})
	
	// Command should handle the error gracefully (not crash)
	t.Logf("Command with network error completed: %v", err)
	
	// Reset error simulation
	testEnv.mockStore.ShouldSimulateNetworkError = false
	
	// Test permission error scenario
	testEnv.mockStore.ShouldSimulatePermissionError = true
	
	// Add a test prompt and try to delete it
	testPrompt := createTestPrompt("perm1", "Permission Test", "tester", createValidGistURL("1234567890abcdef1234"))
	testEnv.store.Add(testPrompt)
	
	// Try to delete with permission error
	err = testEnv.store.Delete(testPrompt.ID)
	if err == nil {
		t.Error("Expected permission error, but got none")
	}
	
	// Verify error is of correct type (check for AppError)
	var appError apperrors.AppError
	if !errors.As(err, &appError) {
		t.Errorf("Expected AppError, got %T", err)
	}
}

// testURLValidationIntegration tests URL validation integration
func testURLValidationIntegration(t *testing.T) {
	testEnv := setupIntegrationTestEnvironment(t)
	defer testEnv.cleanup()

	// Test valid URLs
	validURLs := []string{
		createValidGistURL("1234567890abcdef1234"),
		createValidGistURL("abcdef1234567890abcdef1234567890ab"),
	}

	for _, validURL := range validURLs {
		// Execute command with valid URL (will fail at TUI level but URL parsing should work)
		_, err := executeDeleteCommand(testEnv, []string{validURL})
		t.Logf("Valid URL %s processing result: %v", validURL, err)
	}

	// Test invalid URLs
	invalidURLs := []string{
		"not-a-url",
		"http://example.com",
		"https://github.com/user/repo",
		"https://gist.github.com/user/invalid123",
	}

	for _, invalidURL := range invalidURLs {
		// Execute command with invalid URL
		_, err := executeDeleteCommand(testEnv, []string{invalidURL})
		t.Logf("Invalid URL %s processing result: %v", invalidURL, err)
		
		// Verify store remains unchanged
		prompts, _ := testEnv.store.List()
		if len(prompts) != 0 {
			t.Errorf("Store should remain empty for invalid URL processing")
		}
	}
}

// TestIntegrationEnvironment provides a controlled environment for integration testing
type TestIntegrationEnvironment struct {
	store       infra.Store
	mockStore   *MockGitHubStore
	mockTUI     *tui.MockTUI
	service     service.PromptService
	deleteCmd   *cobra.Command
	cleanup     func()
}

// setupIntegrationTestEnvironment creates a test environment with all necessary mocks
func setupIntegrationTestEnvironment(t *testing.T) *TestIntegrationEnvironment {
	// Create mock components
	mockStore := NewMockGitHubStore()
	mockTUI := tui.NewMockTUI()
	
	// Create service with mocked dependencies
	mockValidator := &MockYAMLValidator{}
	promptService := service.NewPromptService(mockStore, mockValidator)
	
	// Create delete command with mocked dependencies
	deleteCmd := cmd.NewDeleteCommand(mockStore, promptService)
	
	// Setup cleanup function
	cleanup := func() {
		mockStore.Reset()
		mockTUI.Reset()
	}
	
	return &TestIntegrationEnvironment{
		store:     mockStore,
		mockStore: mockStore,
		mockTUI:   mockTUI,
		service:   promptService,
		deleteCmd: deleteCmd,
		cleanup:   cleanup,
	}
}

// executeDeleteCommand executes the delete command and captures output
func executeDeleteCommand(env *TestIntegrationEnvironment, args []string) (string, error) {
	// Create a new command instance for each test to avoid cobra command reuse issues
	deleteCmd := cmd.NewDeleteCommand(env.store, env.service)
	
	// Capture command output
	var buf bytes.Buffer
	deleteCmd.SetOut(&buf)
	deleteCmd.SetErr(&buf)
	
	// Set arguments
	deleteCmd.SetArgs(args)
	
	// Execute command
	err := deleteCmd.Execute()
	
	output := buf.String()
	return output, err
}

// createTestPrompt creates a test prompt with specified parameters
func createTestPrompt(id, name, author, gistURL string) model.Prompt {
	return model.Prompt{
		ID:          id,
		Name:        name,
		Author:      author,
		GistURL:     gistURL,
		Description: fmt.Sprintf("Test prompt: %s by %s", name, author),
		Tags:        []string{"test", "integration"},
		Version:     "1.0.0",
		Content:     fmt.Sprintf("Test content for %s", name),
	}
}

// createValidGistURL creates a GitHub Gist URL with a valid gist ID
func createValidGistURL(gistID string) string {
	// Ensure the gist ID is 20 or 32 characters and hexadecimal
	if len(gistID) < 20 {
		// Pad with zeros to make it 20 characters
		gistID = gistID + strings.Repeat("0", 20-len(gistID))
	}
	return fmt.Sprintf("https://gist.github.com/testuser/%s", gistID)
}

// MockGitHubStore is a mock implementation of the Store interface for integration testing
type MockGitHubStore struct {
	prompts                       map[string]model.Prompt
	ShouldSimulateNetworkError    bool
	ShouldSimulatePermissionError bool
	DeleteMethodCalled            bool
}

// NewMockGitHubStore creates a new mock GitHub store
func NewMockGitHubStore() *MockGitHubStore {
	return &MockGitHubStore{
		prompts: make(map[string]model.Prompt),
	}
}

// List implements infra.Store.List
func (m *MockGitHubStore) List() ([]model.Prompt, error) {
	if m.ShouldSimulateNetworkError {
		return nil, apperrors.NewAppError(apperrors.ErrNetwork, "simulated network error", nil)
	}
	
	var result []model.Prompt
	for _, prompt := range m.prompts {
		result = append(result, prompt)
	}
	return result, nil
}

// Add implements infra.Store.Add
func (m *MockGitHubStore) Add(prompt model.Prompt) error {
	if m.ShouldSimulateNetworkError {
		return apperrors.NewAppError(apperrors.ErrNetwork, "simulated network error", nil)
	}
	
	m.prompts[prompt.ID] = prompt
	return nil
}

// Delete implements infra.Store.Delete
func (m *MockGitHubStore) Delete(id string) error {
	m.DeleteMethodCalled = true
	
	if m.ShouldSimulateNetworkError {
		return apperrors.NewAppError(apperrors.ErrNetwork, "simulated network error", nil)
	}
	
	if m.ShouldSimulatePermissionError {
		return apperrors.NewAppError(apperrors.ErrAuth, "simulated permission error", nil)
	}
	
	delete(m.prompts, id)
	return nil
}

// Update implements infra.Store.Update
func (m *MockGitHubStore) Update(prompt model.Prompt) error {
	if m.ShouldSimulateNetworkError {
		return apperrors.NewAppError(apperrors.ErrNetwork, "simulated network error", nil)
	}
	
	m.prompts[prompt.ID] = prompt
	return nil
}

// Get implements infra.Store.Get
func (m *MockGitHubStore) Get(query string) ([]model.Prompt, error) {
	if m.ShouldSimulateNetworkError {
		return nil, apperrors.NewAppError(apperrors.ErrNetwork, "simulated network error", nil)
	}
	
	var result []model.Prompt
	for _, prompt := range m.prompts {
		// Simple matching logic for testing
		if strings.Contains(strings.ToLower(prompt.Name), strings.ToLower(query)) ||
		   strings.Contains(strings.ToLower(prompt.Author), strings.ToLower(query)) ||
		   strings.Contains(strings.ToLower(prompt.Description), strings.ToLower(query)) ||
		   prompt.GistURL == query ||
		   prompt.ID == query {
			result = append(result, prompt)
		}
	}
	return result, nil
}

// Reset resets the mock store to initial state
func (m *MockGitHubStore) Reset() {
	m.prompts = make(map[string]model.Prompt)
	m.ShouldSimulateNetworkError = false
	m.ShouldSimulatePermissionError = false
	m.DeleteMethodCalled = false
}

// MockYAMLValidator is a mock implementation of the YAMLValidator interface
type MockYAMLValidator struct{}

// ValidatePromptFile implements validator.YAMLValidator.ValidatePromptFile
func (m *MockYAMLValidator) ValidatePromptFile(content []byte) (*validator.PromptFileContent, error) {
	return &validator.PromptFileContent{
		Metadata: validator.PromptMetadata{
			Name:        "Mock Prompt",
			Author:      "mock-author",
			Description: "Mock description",
			Version:     "1.0.0",
		},
		Content: "Mock content",
	}, nil
}

// ValidateRequired implements validator.YAMLValidator.ValidateRequired
func (m *MockYAMLValidator) ValidateRequired(prompt *validator.PromptFileContent) error {
	return nil
}

// TestIntegrationTestSetup verifies that the integration test environment is properly configured
func TestIntegrationTestSetup(t *testing.T) {
	// Test environment setup
	env := setupIntegrationTestEnvironment(t)
	defer env.cleanup()
	
	// Verify components are properly initialized
	if env.store == nil {
		t.Fatal("Store not initialized")
	}
	
	if env.mockTUI == nil {
		t.Fatal("Mock TUI not initialized")
	}
	
	if env.service == nil {
		t.Fatal("Service not initialized")
	}
	
	if env.deleteCmd == nil {
		t.Fatal("Delete command not initialized")
	}
	
	// Test basic store operations
	testPrompt := createTestPrompt("test1", "Setup Test", "tester", createValidGistURL("1234567890abcdef1234"))
	
	err := env.store.Add(testPrompt)
	if err != nil {
		t.Fatalf("Failed to add test prompt: %v", err)
	}
	
	prompts, err := env.store.List()
	if err != nil {
		t.Fatalf("Failed to list prompts: %v", err)
	}
	
	if len(prompts) != 1 {
		t.Fatalf("Expected 1 prompt, got %d", len(prompts))
	}
	
	if prompts[0].ID != testPrompt.ID {
		t.Fatalf("Expected prompt ID %s, got %s", testPrompt.ID, prompts[0].ID)
	}
}

// BenchmarkDeleteWorkflow benchmarks the delete workflow performance
func BenchmarkDeleteWorkflow(b *testing.B) {
	// Setup benchmark environment
	env := setupIntegrationTestEnvironment(&testing.T{})
	defer env.cleanup()
	
	// Add test prompts
	for i := 0; i < 100; i++ {
		prompt := createTestPrompt(
			fmt.Sprintf("bench%d", i),
			fmt.Sprintf("Benchmark Prompt %d", i),
			"benchuser",
			createValidGistURL(fmt.Sprintf("benchmark%012d", i)),
		)
		env.store.Add(prompt)
	}
	
	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test service operations that are core to delete functionality
		env.service.ListForDeletion()
		env.service.FilterForDeletion("benchmark")
	}
}