package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewVariableFormModel(t *testing.T) {
	variables := []string{"name", "age", "location"}
	model := NewVariableFormModel(variables)

	// Test initialization
	if len(model.variables) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(model.variables))
	}

	if len(model.inputs) != 3 {
		t.Errorf("Expected 3 inputs, got %d", len(model.inputs))
	}

	if len(model.values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(model.values))
	}

	if model.currentField != 0 {
		t.Errorf("Expected currentField to be 0, got %d", model.currentField)
	}

	if model.done {
		t.Error("Expected done to be false")
	}

	if model.cancelled {
		t.Error("Expected cancelled to be false")
	}
}

func TestVariableFormModel_Navigation(t *testing.T) {
	variables := []string{"var1", "var2", "var3"}
	model := NewVariableFormModel(variables)

	// Test next field
	newModel, _ := model.nextField()
	if newModel.currentField != 1 {
		t.Errorf("Expected currentField to be 1, got %d", newModel.currentField)
	}

	// Test wrapping around
	model.currentField = 2
	newModel, _ = model.nextField()
	if newModel.currentField != 0 {
		t.Errorf("Expected currentField to wrap to 0, got %d", newModel.currentField)
	}

	// Test previous field
	model.currentField = 1
	newModel, _ = model.prevField()
	if newModel.currentField != 0 {
		t.Errorf("Expected currentField to be 0, got %d", newModel.currentField)
	}

	// Test wrapping around backwards
	model.currentField = 0
	newModel, _ = model.prevField()
	if newModel.currentField != 2 {
		t.Errorf("Expected currentField to wrap to 2, got %d", newModel.currentField)
	}
}

func TestVariableFormModel_KeyHandling(t *testing.T) {
	variables := []string{"test"}
	model := NewVariableFormModel(variables)

	// Test cancel key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := model.Update(msg)
	if !newModel.(VariableFormModel).cancelled {
		t.Error("Expected model to be cancelled after Esc key")
	}
	if cmd == nil || cmd() != tea.Quit() {
		t.Error("Expected Quit command after cancel")
	}

	// Test tab navigation
	model = NewVariableFormModel([]string{"var1", "var2"})
	msg = tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ = model.Update(msg)
	if newModel.(VariableFormModel).currentField != 1 {
		t.Errorf("Expected currentField to be 1 after Tab, got %d", newModel.(VariableFormModel).currentField)
	}
}

func TestVariableFormModel_EmptyValidation(t *testing.T) {
	variables := []string{"required_field"}
	model := NewVariableFormModel(variables)

	// Try to submit without filling the field
	newModel, _ := model.handleEnter()
	if newModel.err == nil {
		t.Error("Expected validation error for empty field")
	}

	if newModel.done {
		t.Error("Expected form not to be done with validation error")
	}
}

func TestVariableFormModel_SuccessfulSubmission(t *testing.T) {
	variables := []string{"name"}
	model := NewVariableFormModel(variables)

	// Fill the field
	model.inputs[0].SetValue("John")
	model.values["name"] = "John"

	// Submit
	newModel, cmd := model.handleEnter()
	if newModel.err != nil {
		t.Errorf("Unexpected error: %v", newModel.err)
	}

	if !newModel.done {
		t.Error("Expected form to be done after successful submission")
	}

	if cmd == nil || cmd() != tea.Quit() {
		t.Error("Expected Quit command after successful submission")
	}

	values := newModel.GetValues()
	if values["name"] != "John" {
		t.Errorf("Expected name to be 'John', got %s", values["name"])
	}
}

func TestVariableFormModel_StateQueries(t *testing.T) {
	variables := []string{"test"}
	model := NewVariableFormModel(variables)

	// Test initial state
	if model.IsDone() {
		t.Error("Expected IsDone to be false initially")
	}

	if model.IsCancelled() {
		t.Error("Expected IsCancelled to be false initially")
	}

	if model.GetValues() != nil {
		t.Error("Expected GetValues to return nil when not done")
	}

	if model.GetError() != nil {
		t.Error("Expected GetError to return nil initially")
	}

	// Test cancelled state
	model.cancelled = true
	if !model.IsCancelled() {
		t.Error("Expected IsCancelled to be true after setting cancelled")
	}

	// Test done state
	model.cancelled = false
	model.done = true
	if !model.IsDone() {
		t.Error("Expected IsDone to be true after setting done")
	}

	values := model.GetValues()
	if values == nil {
		t.Error("Expected GetValues to return values when done")
	}
}

func TestVariableFormModel_View(t *testing.T) {
	variables := []string{"name", "email"}
	model := NewVariableFormModel(variables)

	// Test normal view
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Test error view
	model.err = fmt.Errorf("test error")
	errorView := model.View()
	if errorView == "" {
		t.Error("Expected non-empty error view")
	}

	if !strings.Contains(errorView, "test error") {
		t.Error("Expected error view to contain error message")
	}
}