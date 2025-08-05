package tui

import (
	"errors"
	"fmt"

	"github.com/grigri/pv/internal/model"
)

// MockTUI is a test implementation of TUIInterface that allows for
// pre-configured responses and method call history tracking.
// This enables comprehensive testing of command workflows
// without requiring actual terminal interaction.
type MockTUI struct {
	// Pre-configured return values for testing scenarios
	SelectedPrompt      model.Prompt
	ConfirmResult       bool
	VariableFormResult  map[string]string
	ShowPromptListErr   error
	ShowConfirmErr      error
	ShowVariableFormErr error

	// Method call history for verification in tests
	CallHistory           []MethodCall
	ShowPromptListArgs    [][]model.Prompt
	ShowConfirmArgs       []model.Prompt
	ShowVariableFormArgs  [][]string

	// Test scenario configurations
	ShouldSimulateUserCancel      bool
	ShouldSimulateSelectionErr    bool
	ShouldSimulateConfirmErr      bool
	ShouldSimulateVariableFormErr bool
	SimulateSlowResponse          bool
}

// MethodCall represents a recorded method invocation for test verification
type MethodCall struct {
	Method    string
	Args      interface{}
	Timestamp int64
}

// NewMockTUI creates a new MockTUI instance with default test configuration
func NewMockTUI() *MockTUI {
	return &MockTUI{
		CallHistory:           make([]MethodCall, 0),
		ShowPromptListArgs:    make([][]model.Prompt, 0),
		ShowConfirmArgs:       make([]model.Prompt, 0),
		ShowVariableFormArgs:  make([][]string, 0),
		VariableFormResult:    make(map[string]string),
	}
}

// ShowPromptList implements TUIInterface.ShowPromptList for testing
// It records the method call and returns pre-configured values or errors
// based on test scenarios.
func (m *MockTUI) ShowPromptList(prompts []model.Prompt) (model.Prompt, error) {
	// Record method call for verification
	m.CallHistory = append(m.CallHistory, MethodCall{
		Method: "ShowPromptList",
		Args:   prompts,
	})
	m.ShowPromptListArgs = append(m.ShowPromptListArgs, prompts)

	// Simulate user cancellation scenario
	if m.ShouldSimulateUserCancel {
		return model.Prompt{}, errors.New(ErrMsgUserCancelled)
	}

	// Simulate selection error scenario
	if m.ShouldSimulateSelectionErr {
		return model.Prompt{}, errors.New(ErrMsgInvalidSelection)
	}

	// Return pre-configured error if set
	if m.ShowPromptListErr != nil {
		return model.Prompt{}, m.ShowPromptListErr
	}

	// Handle empty prompts list
	if len(prompts) == 0 {
		return model.Prompt{}, errors.New(ErrMsgNoPromptsFound)
	}

	// Return pre-configured prompt or first prompt if not set
	if m.SelectedPrompt.ID != "" {
		return m.SelectedPrompt, nil
	}

	// Default: return first prompt from the list
	return prompts[0], nil
}

// ShowConfirm implements TUIInterface.ShowConfirm for testing
// It records the method call and returns pre-configured confirmation result
// based on test scenarios.
func (m *MockTUI) ShowConfirm(prompt model.Prompt) (bool, error) {
	// Record method call for verification
	m.CallHistory = append(m.CallHistory, MethodCall{
		Method: "ShowConfirm",
		Args:   prompt,
	})
	m.ShowConfirmArgs = append(m.ShowConfirmArgs, prompt)

	// Simulate confirmation error scenario
	if m.ShouldSimulateConfirmErr {
		return false, errors.New(ErrMsgTUIRenderFailed)
	}

	// Return pre-configured error if set
	if m.ShowConfirmErr != nil {
		return false, m.ShowConfirmErr
	}

	// Return pre-configured confirmation result
	return m.ConfirmResult, nil
}

// ShowVariableForm implements TUIInterface.ShowVariableForm for testing
// It records the method call and returns pre-configured variable values
// based on test scenarios.
func (m *MockTUI) ShowVariableForm(variables []string) (map[string]string, error) {
	// Record method call for verification
	m.CallHistory = append(m.CallHistory, MethodCall{
		Method: "ShowVariableForm",
		Args:   variables,
	})
	m.ShowVariableFormArgs = append(m.ShowVariableFormArgs, variables)

	// Simulate user cancellation scenario
	if m.ShouldSimulateUserCancel {
		return nil, errors.New(ErrMsgUserCancelled)
	}

	// Simulate variable form error scenario
	if m.ShouldSimulateVariableFormErr {
		return nil, errors.New(ErrMsgTUIRenderFailed)
	}

	// Return pre-configured error if set
	if m.ShowVariableFormErr != nil {
		return nil, m.ShowVariableFormErr
	}

	// Return pre-configured variable values if set
	if len(m.VariableFormResult) > 0 {
		return m.VariableFormResult, nil
	}

	// Default: create default values for all variables
	result := make(map[string]string)
	for _, variable := range variables {
		result[variable] = "test_value_" + variable
	}
	return result, nil
}

// Reset clears all recorded data and resets the mock to initial state
// This is useful for cleaning up between test cases.
func (m *MockTUI) Reset() {
	m.SelectedPrompt = model.Prompt{}
	m.ConfirmResult = false
	m.VariableFormResult = make(map[string]string)
	m.ShowPromptListErr = nil
	m.ShowConfirmErr = nil
	m.ShowVariableFormErr = nil
	m.CallHistory = make([]MethodCall, 0)
	m.ShowPromptListArgs = make([][]model.Prompt, 0)
	m.ShowConfirmArgs = make([]model.Prompt, 0)
	m.ShowVariableFormArgs = make([][]string, 0)
	m.ShouldSimulateUserCancel = false
	m.ShouldSimulateSelectionErr = false
	m.ShouldSimulateConfirmErr = false
	m.ShouldSimulateVariableFormErr = false
	m.SimulateSlowResponse = false
}

// GetMethodCallCount returns the number of times a specific method was called
func (m *MockTUI) GetMethodCallCount(methodName string) int {
	count := 0
	for _, call := range m.CallHistory {
		if call.Method == methodName {
			count++
		}
	}
	return count
}

// GetLastMethodCall returns the last recorded method call, or nil if no calls
func (m *MockTUI) GetLastMethodCall() *MethodCall {
	if len(m.CallHistory) == 0 {
		return nil
	}
	return &m.CallHistory[len(m.CallHistory)-1]
}

// VerifyMethodCalled verifies that a specific method was called with expected arguments
func (m *MockTUI) VerifyMethodCalled(methodName string, expectedArgs interface{}) bool {
	for _, call := range m.CallHistory {
		if call.Method == methodName {
			switch methodName {
			case "ShowPromptList":
				if prompts, ok := call.Args.([]model.Prompt); ok {
					if expected, ok := expectedArgs.([]model.Prompt); ok {
						return comparePromptSlices(prompts, expected)
					}
				}
			case "ShowConfirm":
				if prompt, ok := call.Args.(model.Prompt); ok {
					if expected, ok := expectedArgs.(model.Prompt); ok {
						return prompt.ID == expected.ID
					}
				}
			case "ShowVariableForm":
				if variables, ok := call.Args.([]string); ok {
					if expected, ok := expectedArgs.([]string); ok {
						return compareStringSlices(variables, expected)
					}
				}
			}
		}
	}
	return false
}

// SetupInteractiveDeleteScenario configures the mock for interactive delete testing
func (m *MockTUI) SetupInteractiveDeleteScenario(selectedPrompt model.Prompt, confirmResult bool) {
	m.SelectedPrompt = selectedPrompt
	m.ConfirmResult = confirmResult
	m.ShowPromptListErr = nil
	m.ShowConfirmErr = nil
}

// SetupFilterDeleteScenario configures the mock for filtered delete testing
func (m *MockTUI) SetupFilterDeleteScenario(selectedPrompt model.Prompt, confirmResult bool) {
	m.SetupInteractiveDeleteScenario(selectedPrompt, confirmResult)
}

// SetupDirectDeleteScenario configures the mock for direct URL delete testing
func (m *MockTUI) SetupDirectDeleteScenario(confirmResult bool) {
	m.ConfirmResult = confirmResult
	m.ShowConfirmErr = nil
}

// SetupVariableFormScenario configures the mock for variable form testing
func (m *MockTUI) SetupVariableFormScenario(variableValues map[string]string) {
	m.VariableFormResult = variableValues
	m.ShowVariableFormErr = nil
}

// SetupGetCommandScenario configures the mock for full get command testing
func (m *MockTUI) SetupGetCommandScenario(selectedPrompt model.Prompt, variableValues map[string]string) {
	m.SelectedPrompt = selectedPrompt
	m.VariableFormResult = variableValues
	m.ShowPromptListErr = nil
	m.ShowVariableFormErr = nil
}

// SetupErrorScenario configures the mock to simulate various error conditions
func (m *MockTUI) SetupErrorScenario(errorType string) error {
	switch errorType {
	case "user_cancel":
		m.ShouldSimulateUserCancel = true
	case "selection_error":
		m.ShouldSimulateSelectionErr = true
	case "confirm_error":
		m.ShouldSimulateConfirmErr = true
	case "variable_form_error":
		m.ShouldSimulateVariableFormErr = true
	case "show_prompt_list_error":
		m.ShowPromptListErr = errors.New("simulated prompt list error")
	case "show_confirm_error":
		m.ShowConfirmErr = errors.New("simulated confirm error")
	case "show_variable_form_error":
		m.ShowVariableFormErr = errors.New("simulated variable form error")
	default:
		return fmt.Errorf("unknown error type: %s", errorType)
	}
	return nil
}

// GetInteractionSummary returns a summary of all interactions for debugging
func (m *MockTUI) GetInteractionSummary() string {
	summary := fmt.Sprintf("MockTUI Interaction Summary:\n")
	summary += fmt.Sprintf("- Total method calls: %d\n", len(m.CallHistory))
	summary += fmt.Sprintf("- ShowPromptList calls: %d\n", m.GetMethodCallCount("ShowPromptList"))
	summary += fmt.Sprintf("- ShowConfirm calls: %d\n", m.GetMethodCallCount("ShowConfirm"))
	summary += fmt.Sprintf("- ShowVariableForm calls: %d\n", m.GetMethodCallCount("ShowVariableForm"))
	
	if len(m.ShowPromptListArgs) > 0 {
		summary += fmt.Sprintf("- Last ShowPromptList args: %d prompts\n", len(m.ShowPromptListArgs[len(m.ShowPromptListArgs)-1]))
	}
	
	if len(m.ShowConfirmArgs) > 0 {
		lastConfirm := m.ShowConfirmArgs[len(m.ShowConfirmArgs)-1]
		summary += fmt.Sprintf("- Last ShowConfirm args: prompt '%s'\n", lastConfirm.Name)
	}
	
	if len(m.ShowVariableFormArgs) > 0 {
		lastVariables := m.ShowVariableFormArgs[len(m.ShowVariableFormArgs)-1]
		summary += fmt.Sprintf("- Last ShowVariableForm args: %d variables\n", len(lastVariables))
	}
	
	return summary
}

// comparePromptSlices compares two prompt slices for equality
func comparePromptSlices(a, b []model.Prompt) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			return false
		}
	}
	return true
}

// compareStringSlices compares two string slices for equality
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}