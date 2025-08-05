package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestMockTUI_ShowVariableForm(t *testing.T) {
	mock := NewMockTUI()

	variables := []string{"name", "age", "city"}
	result, err := mock.ShowVariableForm(variables)

	// Test successful scenario
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Verify default values are created
	for _, variable := range variables {
		if value, exists := result[variable]; !exists {
			t.Errorf("Expected variable %s in result", variable)
		} else if !strings.Contains(value, variable) {
			t.Errorf("Expected value to contain variable name, got %s", value)
		}
	}

	// Verify method call was recorded
	if mock.GetMethodCallCount("ShowVariableForm") != 1 {
		t.Errorf("Expected 1 ShowVariableForm call, got %d", mock.GetMethodCallCount("ShowVariableForm"))
	}

	// Verify arguments were recorded
	if len(mock.ShowVariableFormArgs) != 1 {
		t.Errorf("Expected 1 ShowVariableFormArgs entry, got %d", len(mock.ShowVariableFormArgs))
	}

	if !compareStringSlices(mock.ShowVariableFormArgs[0], variables) {
		t.Error("Recorded arguments don't match expected")
	}
}

func TestMockTUI_ShowVariableForm_PreConfiguredValues(t *testing.T) {
	mock := NewMockTUI()

	// Pre-configure variable values
	expectedValues := map[string]string{
		"name": "John Doe",
		"age":  "30",
		"city": "New York",
	}
	mock.SetupVariableFormScenario(expectedValues)

	variables := []string{"name", "age", "city"}
	result, err := mock.ShowVariableForm(variables)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify pre-configured values are returned
	for key, expectedValue := range expectedValues {
		if actualValue, exists := result[key]; !exists {
			t.Errorf("Expected variable %s in result", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected %s=%s, got %s", key, expectedValue, actualValue)
		}
	}
}

func TestMockTUI_ShowVariableForm_UserCancel(t *testing.T) {
	mock := NewMockTUI()
	mock.ShouldSimulateUserCancel = true

	variables := []string{"test"}
	result, err := mock.ShowVariableForm(variables)

	if err == nil {
		t.Error("Expected error for user cancel")
	}

	if result != nil {
		t.Error("Expected nil result for user cancel")
	}

	if !strings.Contains(err.Error(), ErrMsgUserCancelled) {
		t.Errorf("Expected user cancelled error, got %v", err)
	}
}

func TestMockTUI_ShowVariableForm_Error(t *testing.T) {
	mock := NewMockTUI()
	err := mock.SetupErrorScenario("variable_form_error")
	if err != nil {
		t.Fatalf("Failed to setup error scenario: %v", err)
	}

	variables := []string{"test"}
	result, err := mock.ShowVariableForm(variables)

	if err == nil {
		t.Error("Expected error for variable form error scenario")
	}

	if result != nil {
		t.Error("Expected nil result for error scenario")
	}
}

func TestMockTUI_ShowVariableForm_EmptyVariables(t *testing.T) {
	mock := NewMockTUI()

	variables := []string{}
	result, err := mock.ShowVariableForm(variables)

	if err != nil {
		t.Errorf("Expected no error for empty variables, got %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result for empty variables")
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result map, got %d entries", len(result))
	}
}

func TestMockTUI_VerifyMethodCalled_ShowVariableForm(t *testing.T) {
	mock := NewMockTUI()

	variables := []string{"test1", "test2"}
	_, _ = mock.ShowVariableForm(variables)

	// Test verification with correct arguments
	if !mock.VerifyMethodCalled("ShowVariableForm", variables) {
		t.Error("Expected VerifyMethodCalled to return true for correct arguments")
	}

	// Test verification with incorrect arguments
	wrongVariables := []string{"wrong1", "wrong2"}
	if mock.VerifyMethodCalled("ShowVariableForm", wrongVariables) {
		t.Error("Expected VerifyMethodCalled to return false for incorrect arguments")
	}
}

func TestMockTUI_Reset_ClearsVariableFormData(t *testing.T) {
	mock := NewMockTUI()

	// Setup data
	mock.VariableFormResult = map[string]string{"test": "value"}
	mock.ShowVariableFormErr = fmt.Errorf("test error")
	mock.ShouldSimulateVariableFormErr = true
	mock.ShowVariableFormArgs = [][]string{{"test"}}

	// Call a method to add to history
	_, _ = mock.ShowVariableForm([]string{"test"})

	// Reset
	mock.Reset()

	// Verify reset
	if len(mock.VariableFormResult) != 0 {
		t.Error("Expected VariableFormResult to be empty after reset")
	}

	if mock.ShowVariableFormErr != nil {
		t.Error("Expected ShowVariableFormErr to be nil after reset")
	}

	if mock.ShouldSimulateVariableFormErr {
		t.Error("Expected ShouldSimulateVariableFormErr to be false after reset")
	}

	if len(mock.ShowVariableFormArgs) != 0 {
		t.Error("Expected ShowVariableFormArgs to be empty after reset")
	}

	if len(mock.CallHistory) != 0 {
		t.Error("Expected CallHistory to be empty after reset")
	}
}

func TestMockTUI_GetInteractionSummary_IncludesVariableForm(t *testing.T) {
	mock := NewMockTUI()

	// Call ShowVariableForm
	_, _ = mock.ShowVariableForm([]string{"test1", "test2"})

	summary := mock.GetInteractionSummary()

	if !strings.Contains(summary, "ShowVariableForm calls: 1") {
		t.Error("Expected summary to include ShowVariableForm call count")
	}

	if !strings.Contains(summary, "Last ShowVariableForm args: 2 variables") {
		t.Error("Expected summary to include ShowVariableForm arguments info")
	}
}