// +build integration

package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Netflix/go-expect"
	"github.com/hinshun/vt10x"
)

// IntegrationTestSetup provides infrastructure for running TTY integration tests
// with bubbletea components using expect and vt10x libraries.
// This is used for testing the actual terminal interactions.
type IntegrationTestSetup struct {
	console    *expect.Console
	terminal   vt10x.Terminal
	timeout    time.Duration
	screenSize [2]int
}

// NewIntegrationTestSetup creates a new integration test environment
func NewIntegrationTestSetup() (*IntegrationTestSetup, error) {
	// Create a virtual terminal with standard dimensions
	terminal := vt10x.New(vt10x.WithSize(80, 24))
	
	// Create an expect console for interaction
	console, err := expect.NewConsole(expect.WithStdout(terminal))
	if err != nil {
		return nil, fmt.Errorf("failed to create console: %w", err)
	}

	return &IntegrationTestSetup{
		console:    console,
		terminal:   terminal,
		timeout:    10 * time.Second,
		screenSize: [2]int{80, 24},
	}, nil
}

// Close cleans up the integration test environment
func (setup *IntegrationTestSetup) Close() error {
	if setup.console != nil {
		return setup.console.Close()
	}
	return nil
}

// SetTimeout sets the timeout for expect operations
func (setup *IntegrationTestSetup) SetTimeout(timeout time.Duration) {
	setup.timeout = timeout
}

// SendKeys simulates keyboard input to the terminal
func (setup *IntegrationTestSetup) SendKeys(keys string) error {
	_, err := setup.console.Send(keys)
	return err
}

// ExpectString waits for a specific string to appear in the terminal output
func (setup *IntegrationTestSetup) ExpectString(expected string) error {
	_, err := setup.console.Expect(expect.String(expected), expect.WithTimeout(setup.timeout))
	return err
}

// ExpectPrompt waits for a prompt-like pattern to appear
func (setup *IntegrationTestSetup) ExpectPrompt() error {
	// Look for common prompt indicators
	_, err := setup.console.Expect(expect.Regexp(`[>$#]\s*$`), expect.WithTimeout(setup.timeout))
	return err
}

// GetScreenContent returns the current terminal screen content
func (setup *IntegrationTestSetup) GetScreenContent() string {
	return setup.terminal.String()
}

// GetScreenLines returns the terminal content as separate lines
func (setup *IntegrationTestSetup) GetScreenLines() []string {
	content := setup.terminal.String()
	return strings.Split(content, "\n")
}

// WaitForText waits for specific text to appear anywhere on screen
func (setup *IntegrationTestSetup) WaitForText(text string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for text: %s", text)
		case <-ticker.C:
			if strings.Contains(setup.GetScreenContent(), text) {
				return nil
			}
		}
	}
}

// SimulateUserInteraction simulates a sequence of user inputs with delays
func (setup *IntegrationTestSetup) SimulateUserInteraction(interactions []UserInteraction) error {
	for _, interaction := range interactions {
		// Wait for the expected screen state if specified
		if interaction.WaitFor != "" {
			if err := setup.WaitForText(interaction.WaitFor, setup.timeout); err != nil {
				return fmt.Errorf("failed to wait for '%s': %w", interaction.WaitFor, err)
			}
		}

		// Add delay if specified
		if interaction.Delay > 0 {
			time.Sleep(interaction.Delay)
		}

		// Send the input
		if err := setup.SendKeys(interaction.Input); err != nil {
			return fmt.Errorf("failed to send input '%s': %w", interaction.Input, err)
		}
	}
	return nil
}

// UserInteraction represents a single user input action in a test scenario
type UserInteraction struct {
	Input   string        // The keys to send
	WaitFor string        // Text to wait for before sending input
	Delay   time.Duration // Delay before sending input
}

// CreateDeleteInteractionScenario creates a standard delete command interaction scenario
func CreateDeleteInteractionScenario() []UserInteraction {
	return []UserInteraction{
		{
			WaitFor: "选择要删除的提示", // Wait for prompt list to appear
			Input:   "2",             // Select second item
			Delay:   500 * time.Millisecond,
		},
		{
			WaitFor: "确认删除", // Wait for confirmation dialog
			Input:   "y",       // Confirm deletion
			Delay:   200 * time.Millisecond,
		},
	}
}

// CreateCancelInteractionScenario creates a scenario where user cancels the operation
func CreateCancelInteractionScenario() []UserInteraction {
	return []UserInteraction{
		{
			WaitFor: "选择要删除的提示", // Wait for prompt list
			Input:   "q",             // Quit without selection
			Delay:   100 * time.Millisecond,
		},
	}
}

// CreateConfirmCancelScenario creates a scenario where user cancels at confirmation
func CreateConfirmCancelScenario() []UserInteraction {
	return []UserInteraction{
		{
			WaitFor: "选择要删除的提示", // Wait for prompt list
			Input:   "1",             // Select first item
			Delay:   300 * time.Millisecond,
		},
		{
			WaitFor: "确认删除", // Wait for confirmation
			Input:   "n",       // Cancel deletion
			Delay:   200 * time.Millisecond,
		},
	}
}

// TestHelper provides additional utilities for integration testing
type TestHelper struct {
	setup *IntegrationTestSetup
}

// NewTestHelper creates a new test helper with integration setup
func NewTestHelper() (*TestHelper, error) {
	setup, err := NewIntegrationTestSetup()
	if err != nil {
		return nil, err
	}

	return &TestHelper{
		setup: setup,
	}, nil
}

// Close cleans up the test helper
func (h *TestHelper) Close() error {
	return h.setup.Close()
}

// GetSetup returns the underlying integration test setup
func (h *TestHelper) GetSetup() *IntegrationTestSetup {
	return h.setup
}

// ValidateScreenContains checks if the screen contains specific text
func (h *TestHelper) ValidateScreenContains(expectedText string) bool {
	return strings.Contains(h.setup.GetScreenContent(), expectedText)
}

// ValidateScreenDoesNotContain checks if the screen does NOT contain specific text
func (h *TestHelper) ValidateScreenDoesNotContain(forbiddenText string) bool {
	return !strings.Contains(h.setup.GetScreenContent(), forbiddenText)
}

// ValidatePromptListDisplayed checks if the prompt list is properly displayed
func (h *TestHelper) ValidatePromptListDisplayed() bool {
	content := h.setup.GetScreenContent()
	// Check for list indicators and navigation help
	hasListIndicators := strings.Contains(content, "[1]") || strings.Contains(content, "[2]")
	hasNavigationHelp := strings.Contains(content, "↑/↓: 导航") || strings.Contains(content, "Enter: 选择")
	return hasListIndicators && hasNavigationHelp
}

// ValidateConfirmationDisplayed checks if the confirmation dialog is properly displayed
func (h *TestHelper) ValidateConfirmationDisplayed() bool {
	content := h.setup.GetScreenContent()
	hasConfirmText := strings.Contains(content, "确认删除")
	hasOptions := strings.Contains(content, "[Y]") && strings.Contains(content, "[N]")
	return hasConfirmText && hasOptions
}

// ValidateSuccessMessage checks if a success message is displayed
func (h *TestHelper) ValidateSuccessMessage() bool {
	content := h.setup.GetScreenContent()
	return strings.Contains(content, "删除成功") || strings.Contains(content, "已删除")
}

// ValidateErrorMessage checks if an error message is displayed
func (h *TestHelper) ValidateErrorMessage() bool {
	content := h.setup.GetScreenContent()
	return strings.Contains(content, "错误") || strings.Contains(content, "失败")
}

// RunDeleteCommandTest runs a complete delete command test scenario
func (h *TestHelper) RunDeleteCommandTest(scenario []UserInteraction) error {
	// Start the delete command (this would be called from the actual test)
	// For now, this is a placeholder that would be implemented in actual integration tests
	
	// Simulate the user interactions
	return h.setup.SimulateUserInteraction(scenario)
}

// Integration test examples and documentation

/* Example Integration Test Usage:

func TestDeleteCommandIntegration(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Test interactive delete scenario
	scenario := CreateDeleteInteractionScenario()
	err = helper.RunDeleteCommandTest(scenario)
	if err != nil {
		t.Errorf("Delete command test failed: %v", err)
	}

	// Validate the expected outcome
	if !helper.ValidateSuccessMessage() {
		t.Error("Expected success message not found")
	}
}

func TestDeleteCommandCancellation(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Test cancellation scenario
	scenario := CreateCancelInteractionScenario()
	err = helper.RunDeleteCommandTest(scenario)
	if err != nil {
		t.Errorf("Cancel test failed: %v", err)
	}

	// Validate that no deletion occurred
	if helper.ValidateSuccessMessage() {
		t.Error("Unexpected success message found after cancellation")
	}
}

*/

// Note: To use this integration testing infrastructure:
// 1. Build tests with integration tag: go test -tags=integration
// 2. These tests require a TTY environment and may not work in CI without special setup
// 3. For CI/CD, consider using the MockTUI for unit tests and run integration tests separately
// 4. The expect and vt10x libraries provide powerful terminal simulation capabilities
// 5. Always clean up resources by calling Close() methods in defer statements