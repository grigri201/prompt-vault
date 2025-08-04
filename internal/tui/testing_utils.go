package tui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/grigri/pv/internal/model"
)

// TestPromptGenerator provides utilities for generating test prompt data
// with consistent and predictable data for testing scenarios.
type TestPromptGenerator struct {
	counter int
	baseURL string
}

// NewTestPromptGenerator creates a new test prompt data generator
func NewTestPromptGenerator() *TestPromptGenerator {
	return &TestPromptGenerator{
		counter: 0,
		baseURL: "https://gist.github.com/testuser",
	}
}

// GeneratePrompt creates a single test prompt with predictable data
func (g *TestPromptGenerator) GeneratePrompt() model.Prompt {
	g.counter++
	return model.Prompt{
		ID:          fmt.Sprintf("test-prompt-%d", g.counter),
		Name:        fmt.Sprintf("Test Prompt %d", g.counter),
		Author:      fmt.Sprintf("testuser%d", g.counter%3+1), // Cycle through 3 test users
		GistURL:     fmt.Sprintf("%s/%s", g.baseURL, fmt.Sprintf("abc123%d", g.counter)),
		Description: fmt.Sprintf("This is a test prompt number %d for testing purposes", g.counter),
		Tags:        []string{"test", "automation", fmt.Sprintf("category%d", g.counter%5+1)},
		Version:     "1.0.0",
		Content:     fmt.Sprintf("Test prompt content for prompt %d\nThis content is used for testing.", g.counter),
	}
}

// GeneratePrompts creates a slice of test prompts
func (g *TestPromptGenerator) GeneratePrompts(count int) []model.Prompt {
	prompts := make([]model.Prompt, count)
	for i := 0; i < count; i++ {
		prompts[i] = g.GeneratePrompt()
	}
	return prompts
}

// GeneratePromptsWithPattern creates prompts that match a specific pattern
// This is useful for testing keyword filtering functionality
func (g *TestPromptGenerator) GeneratePromptsWithPattern(count int, keyword string) []model.Prompt {
	prompts := make([]model.Prompt, count)
	for i := 0; i < count; i++ {
		prompt := g.GeneratePrompt()
		// Ensure the keyword appears in name, author, or description
		switch i % 3 {
		case 0:
			prompt.Name = fmt.Sprintf("%s %s", keyword, prompt.Name)
		case 1:
			prompt.Author = fmt.Sprintf("%s_%s", keyword, prompt.Author)
		case 2:
			prompt.Description = fmt.Sprintf("Contains %s keyword. %s", keyword, prompt.Description)
		}
		prompts[i] = prompt
	}
	return prompts
}

// GenerateEmptyPromptList returns an empty prompt slice for testing empty list scenarios
func (g *TestPromptGenerator) GenerateEmptyPromptList() []model.Prompt {
	return []model.Prompt{}
}

// GeneratePromptWithSpecificData creates a prompt with specific field values for targeted testing
func (g *TestPromptGenerator) GeneratePromptWithSpecificData(id, name, author, gistURL string) model.Prompt {
	return model.Prompt{
		ID:          id,
		Name:        name,
		Author:      author,
		GistURL:     gistURL,
		Description: fmt.Sprintf("Test prompt for %s by %s", name, author),
		Tags:        []string{"test", "specific"},
		Version:     "1.0.0",
		Content:     fmt.Sprintf("Specific test content for %s", name),
	}
}

// GenerateRandomPrompt creates a prompt with randomized data for stress testing
func (g *TestPromptGenerator) GenerateRandomPrompt() model.Prompt {
	rand.Seed(time.Now().UnixNano())
	
	names := []string{"AI Assistant", "Code Review", "SQL Helper", "Documentation", "Bug Fix", "Feature Plan"}
	authors := []string{"developer1", "coder2", "engineer3", "architect4", "specialist5"}
	
	g.counter++
	return model.Prompt{
		ID:          fmt.Sprintf("random-%d-%d", g.counter, rand.Int31()),
		Name:        names[rand.Intn(len(names))] + fmt.Sprintf(" %d", rand.Intn(100)),
		Author:      authors[rand.Intn(len(authors))],
		GistURL:     fmt.Sprintf("https://gist.github.com/%s/%x", authors[rand.Intn(len(authors))], rand.Int31()),
		Description: fmt.Sprintf("Random test prompt %d with generated content", g.counter),
		Tags:        []string{fmt.Sprintf("tag%d", rand.Intn(10)), "random", "test"},
		Version:     fmt.Sprintf("1.%d.%d", rand.Intn(10), rand.Intn(10)),
		Content:     fmt.Sprintf("Random content %d for testing purposes", rand.Int31()),
	}
}

// TUIStateValidator provides utilities for validating TUI component states
// and behaviors in test scenarios.
type TUIStateValidator struct {
	mockTUI *MockTUI
}

// NewTUIStateValidator creates a new TUI state validator
func NewTUIStateValidator(mockTUI *MockTUI) *TUIStateValidator {
	return &TUIStateValidator{
		mockTUI: mockTUI,
	}
}

// ValidatePromptListCalled verifies that ShowPromptList was called with expected prompts
func (v *TUIStateValidator) ValidatePromptListCalled(expectedPrompts []model.Prompt) bool {
	if len(v.mockTUI.ShowPromptListArgs) == 0 {
		return false
	}
	
	lastCall := v.mockTUI.ShowPromptListArgs[len(v.mockTUI.ShowPromptListArgs)-1]
	return comparePromptSlices(lastCall, expectedPrompts)
}

// ValidateConfirmCalled verifies that ShowConfirm was called with expected prompt
func (v *TUIStateValidator) ValidateConfirmCalled(expectedPrompt model.Prompt) bool {
	if len(v.mockTUI.ShowConfirmArgs) == 0 {
		return false
	}
	
	lastCall := v.mockTUI.ShowConfirmArgs[len(v.mockTUI.ShowConfirmArgs)-1]
	return lastCall.ID == expectedPrompt.ID
}

// ValidateMethodCallSequence verifies that methods were called in the expected order
func (v *TUIStateValidator) ValidateMethodCallSequence(expectedSequence []string) bool {
	if len(v.mockTUI.CallHistory) < len(expectedSequence) {
		return false
	}
	
	// Check the last N calls match the expected sequence
	startIndex := len(v.mockTUI.CallHistory) - len(expectedSequence)
	for i, expectedMethod := range expectedSequence {
		if v.mockTUI.CallHistory[startIndex+i].Method != expectedMethod {
			return false
		}
	}
	
	return true
}

// ValidateNoMethodCalls verifies that no TUI methods were called
func (v *TUIStateValidator) ValidateNoMethodCalls() bool {
	return len(v.mockTUI.CallHistory) == 0
}

// ValidateMethodCallCount verifies that a specific method was called exactly N times
func (v *TUIStateValidator) ValidateMethodCallCount(methodName string, expectedCount int) bool {
	return v.mockTUI.GetMethodCallCount(methodName) == expectedCount
}

// ErrorScenarioSimulator provides utilities for simulating various error conditions
// in TUI interactions to test error handling and recovery mechanisms.
type ErrorScenarioSimulator struct {
	mockTUI *MockTUI
}

// NewErrorScenarioSimulator creates a new error scenario simulator
func NewErrorScenarioSimulator(mockTUI *MockTUI) *ErrorScenarioSimulator {
	return &ErrorScenarioSimulator{
		mockTUI: mockTUI,
	}
}

// SimulateUserCancellation configures the mock to simulate user canceling the operation
func (s *ErrorScenarioSimulator) SimulateUserCancellation() {
	s.mockTUI.ShouldSimulateUserCancel = true
}

// SimulateSelectionError configures the mock to simulate an invalid selection error
func (s *ErrorScenarioSimulator) SimulateSelectionError() {
	s.mockTUI.ShouldSimulateSelectionErr = true
}

// SimulateConfirmationError configures the mock to simulate an error during confirmation
func (s *ErrorScenarioSimulator) SimulateConfirmationError() {
	s.mockTUI.ShouldSimulateConfirmErr = true
}

// SimulatePromptListError configures the mock to return an error when showing prompt list
func (s *ErrorScenarioSimulator) SimulatePromptListError(errorMsg string) {
	s.mockTUI.ShowPromptListErr = fmt.Errorf("%s", errorMsg)
}

// SimulateConfirmError configures the mock to return an error during confirmation
func (s *ErrorScenarioSimulator) SimulateConfirmError(errorMsg string) {
	s.mockTUI.ShowConfirmErr = fmt.Errorf("%s", errorMsg)
}

// SimulateEmptyPromptList tests behavior with empty prompt lists
func (s *ErrorScenarioSimulator) SimulateEmptyPromptList() []model.Prompt {
	return []model.Prompt{}
}

// SimulateNetworkTimeout simulates network-related delays and timeouts
func (s *ErrorScenarioSimulator) SimulateNetworkTimeout() {
	s.mockTUI.SimulateSlowResponse = true
	// In a real implementation, this might introduce delays
}

// TestScenarioBuilder provides a fluent interface for building complex test scenarios
type TestScenarioBuilder struct {
	mockTUI   *MockTUI
	generator *TestPromptGenerator
	validator *TUIStateValidator
	simulator *ErrorScenarioSimulator
}

// NewTestScenarioBuilder creates a new test scenario builder
func NewTestScenarioBuilder() *TestScenarioBuilder {
	mockTUI := NewMockTUI()
	return &TestScenarioBuilder{
		mockTUI:   mockTUI,
		generator: NewTestPromptGenerator(),
		validator: NewTUIStateValidator(mockTUI),
		simulator: NewErrorScenarioSimulator(mockTUI),
	}
}

// GetMockTUI returns the mock TUI instance
func (b *TestScenarioBuilder) GetMockTUI() *MockTUI {
	return b.mockTUI
}

// GetGenerator returns the prompt generator
func (b *TestScenarioBuilder) GetGenerator() *TestPromptGenerator {
	return b.generator
}

// GetValidator returns the state validator
func (b *TestScenarioBuilder) GetValidator() *TUIStateValidator {
	return b.validator
}

// GetSimulator returns the error scenario simulator
func (b *TestScenarioBuilder) GetSimulator() *ErrorScenarioSimulator {
	return b.simulator
}

// WithPromptSelection configures the mock to return a specific prompt selection
func (b *TestScenarioBuilder) WithPromptSelection(prompt model.Prompt) *TestScenarioBuilder {
	b.mockTUI.SelectedPrompt = prompt
	return b
}

// WithConfirmResult configures the mock to return a specific confirmation result
func (b *TestScenarioBuilder) WithConfirmResult(confirmed bool) *TestScenarioBuilder {
	b.mockTUI.ConfirmResult = confirmed
	return b
}

// WithUserCancellation configures the mock to simulate user cancellation
func (b *TestScenarioBuilder) WithUserCancellation() *TestScenarioBuilder {
	b.simulator.SimulateUserCancellation()
	return b
}

// WithSelectionError configures the mock to simulate selection errors
func (b *TestScenarioBuilder) WithSelectionError() *TestScenarioBuilder {
	b.simulator.SimulateSelectionError()
	return b
}

// WithConfirmationError configures the mock to simulate confirmation errors
func (b *TestScenarioBuilder) WithConfirmationError() *TestScenarioBuilder {
	b.simulator.SimulateConfirmationError()
	return b
}

// Reset resets the builder to initial state
func (b *TestScenarioBuilder) Reset() *TestScenarioBuilder {
	b.mockTUI.Reset()
	return b
}

// Build finalizes the test scenario and returns the configured components
func (b *TestScenarioBuilder) Build() (*MockTUI, *TUIStateValidator) {
	return b.mockTUI, b.validator
}

// Common test data constants for consistent testing
const (
	TestPromptID1     = "test-prompt-1"
	TestPromptID2     = "test-prompt-2"
	TestPromptName1   = "Test AI Assistant"
	TestPromptName2   = "Test Code Review"
	TestAuthor1       = "testuser1"
	TestAuthor2       = "testuser2"
	TestGistURL1      = "https://gist.github.com/testuser1/abc123"
	TestGistURL2      = "https://gist.github.com/testuser2/def456"
	TestKeyword       = "golang"
	TestDescription1  = "Test prompt for AI assistance"
	TestDescription2  = "Test prompt for code review"
)

// CreateStandardTestPrompts creates a standard set of test prompts for consistent testing
func CreateStandardTestPrompts() []model.Prompt {
	return []model.Prompt{
		{
			ID:          TestPromptID1,
			Name:        TestPromptName1,
			Author:      TestAuthor1,
			GistURL:     TestGistURL1,
			Description: TestDescription1,
			Tags:        []string{"test", "ai", "assistant"},
			Version:     "1.0.0",
			Content:     "Test content for AI assistant prompt",
		},
		{
			ID:          TestPromptID2,
			Name:        TestPromptName2,
			Author:      TestAuthor2,
			GistURL:     TestGistURL2,
			Description: TestDescription2,
			Tags:        []string{"test", "code", "review"},
			Version:     "1.0.0",
			Content:     "Test content for code review prompt",
		},
	}
}

// CreateEmptyTestPrompts returns an empty prompt list for testing edge cases
func CreateEmptyTestPrompts() []model.Prompt {
	return []model.Prompt{}
}

// CreateSingleTestPrompt creates a single test prompt for simple test scenarios
func CreateSingleTestPrompt() model.Prompt {
	return model.Prompt{
		ID:          TestPromptID1,
		Name:        TestPromptName1,
		Author:      TestAuthor1,
		GistURL:     TestGistURL1,
		Description: TestDescription1,
		Tags:        []string{"test", "single"},
		Version:     "1.0.0",
		Content:     "Single test prompt content",
	}
}