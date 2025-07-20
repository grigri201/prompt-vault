package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewForm(t *testing.T) {
	title := "Test Form"
	variables := []string{"name", "age", "city"}

	form := NewForm(title, variables)

	if form.Title != title {
		t.Errorf("Title = %q, want %q", form.Title, title)
	}

	if len(form.Fields) != len(variables) {
		t.Errorf("Fields length = %d, want %d", len(form.Fields), len(variables))
	}

	for i, field := range form.Fields {
		if field.Name != variables[i] {
			t.Errorf("Field[%d].Name = %q, want %q", i, field.Name, variables[i])
		}
		if field.Value != "" {
			t.Errorf("Field[%d].Value = %q, want empty", i, field.Value)
		}
		if !field.Required {
			t.Errorf("Field[%d].Required = false, want true", i)
		}
	}

	if form.CurrentField != 0 {
		t.Errorf("CurrentField = %d, want 0", form.CurrentField)
	}

	if form.Submitted {
		t.Error("Submitted should be false initially")
	}
}

func TestFormNavigation(t *testing.T) {
	tests := []struct {
		name        string
		startField  int
		key         string
		wantField   int
		description string
	}{
		{
			name:        "down arrow moves to next field",
			startField:  0,
			key:         "down",
			wantField:   1,
			description: "should move from field 0 to field 1",
		},
		{
			name:        "tab moves to next field",
			startField:  0,
			key:         "tab",
			wantField:   1,
			description: "tab navigation",
		},
		{
			name:        "up arrow moves to previous field",
			startField:  1,
			key:         "up",
			wantField:   0,
			description: "should move from field 1 to field 0",
		},
		{
			name:        "shift+tab moves to previous field",
			startField:  2,
			key:         "shift+tab",
			wantField:   1,
			description: "shift+tab navigation",
		},
		{
			name:        "down on last field stays",
			startField:  2,
			key:         "down",
			wantField:   2,
			description: "should stay on last field",
		},
		{
			name:        "up on first field stays",
			startField:  0,
			key:         "up",
			wantField:   0,
			description: "should stay on first field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewForm("Test", []string{"field1", "field2", "field3"})
			f.CurrentField = tt.startField

			model, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			updatedForm := model.(FormModel)

			if updatedForm.CurrentField != tt.wantField {
				t.Errorf("After %s from field %d: CurrentField = %d, want %d",
					tt.key, tt.startField, updatedForm.CurrentField, tt.wantField)
			}
		})
	}
}

func TestFormInput(t *testing.T) {
	form := NewForm("Test", []string{"name", "email"})

	// Type some text
	inputs := []string{"J", "o", "h", "n"}
	model := tea.Model(form)

	for _, char := range inputs {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(char)})
	}

	f := model.(FormModel)
	if f.Fields[0].Value != "John" {
		t.Errorf("Field value = %q, want %q", f.Fields[0].Value, "John")
	}

	// Test backspace
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	f = model.(FormModel)

	if f.Fields[0].Value != "Joh" {
		t.Errorf("After backspace, field value = %q, want %q", f.Fields[0].Value, "Joh")
	}
}

func TestFormSubmission(t *testing.T) {
	tests := []struct {
		name         string
		fields       []string
		fillValues   map[int]string
		currentField int
		wantError    bool
		errorField   string
	}{
		{
			name:   "submit with all fields filled",
			fields: []string{"name", "age"},
			fillValues: map[int]string{
				0: "John",
				1: "25",
			},
			currentField: 1,
			wantError:    false,
		},
		{
			name:   "submit with empty required field",
			fields: []string{"name", "age"},
			fillValues: map[int]string{
				0: "John",
				1: "",
			},
			currentField: 1,
			wantError:    true,
			errorField:   "age",
		},
		{
			name:   "enter on non-last field moves to next",
			fields: []string{"name", "age"},
			fillValues: map[int]string{
				0: "John",
			},
			currentField: 0,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := NewForm("Test", tt.fields)

			// Fill values
			for i, value := range tt.fillValues {
				form.Fields[i].Value = value
			}
			form.CurrentField = tt.currentField

			model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
			f := model.(FormModel)

			if tt.wantError {
				if !f.ShowError {
					t.Error("Expected error to be shown")
				}
				if !strings.Contains(f.ErrorMessage, tt.errorField) {
					t.Errorf("Error message %q doesn't mention field %q",
						f.ErrorMessage, tt.errorField)
				}
			} else {
				if f.ShowError {
					t.Errorf("Unexpected error: %s", f.ErrorMessage)
				}

				// Check if submitted (when on last field with all valid)
				if tt.currentField == len(tt.fields)-1 && len(tt.fillValues) == len(tt.fields) {
					if !f.Submitted {
						t.Error("Form should be submitted")
					}
					if cmd == nil {
						t.Error("Should quit after submission")
					}
				}
			}
		})
	}
}

func TestFormView(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		fields       []string
		currentField int
		values       map[int]string
		showError    bool
		errorMessage string
		wantContains []string
	}{
		{
			name:         "basic form view",
			title:        "Enter Variables",
			fields:       []string{"name", "email"},
			currentField: 0,
			wantContains: []string{
				"Enter Variables",
				"Fill in the variables",
				"name*:",
				"email*:",
				"↑/↓ or tab: navigate",
			},
		},
		{
			name:         "with values",
			title:        "Test Form",
			fields:       []string{"field1", "field2"},
			currentField: 1,
			values: map[int]string{
				0: "value1",
				1: "value2",
			},
			wantContains: []string{
				"field1*: value1",
				"field2*: value2",
			},
		},
		{
			name:         "empty field shows placeholder",
			title:        "Test",
			fields:       []string{"field1"},
			currentField: 1, // Not on this field
			values:       map[int]string{},
			wantContains: []string{
				"(empty)",
			},
		},
		{
			name:         "with error",
			title:        "Test",
			fields:       []string{"required_field"},
			showError:    true,
			errorMessage: "Field 'required_field' is required",
			wantContains: []string{
				"Error:",
				"Field 'required_field' is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := NewForm(tt.title, tt.fields)
			form.CurrentField = tt.currentField
			form.ShowError = tt.showError
			form.ErrorMessage = tt.errorMessage

			// Set values
			for i, value := range tt.values {
				if i < len(form.Fields) {
					form.Fields[i].Value = value
				}
			}

			view := form.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(view, want) {
					t.Errorf("View missing %q\nGot: %s", want, view)
				}
			}
		})
	}
}

func TestFormQuit(t *testing.T) {
	form := NewForm("Test", []string{"field"})

	// Test esc key
	_, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("Expected quit command for 'esc' key")
	}

	// Test ctrl+c
	_, cmd = form.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Expected quit command for ctrl+c")
	}
}

func TestGetValues(t *testing.T) {
	form := NewForm("Test", []string{"name", "age", "city"})
	form.Fields[0].Value = "John"
	form.Fields[1].Value = "25"
	form.Fields[2].Value = "New York"

	values := form.GetValues()

	expected := map[string]string{
		"name": "John",
		"age":  "25",
		"city": "New York",
	}

	for key, want := range expected {
		if got, ok := values[key]; !ok || got != want {
			t.Errorf("GetValues()[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestSetFieldValue(t *testing.T) {
	form := NewForm("Test", []string{"field1", "field2"})

	form.SetFieldValue("field1", "value1")
	form.SetFieldValue("field2", "value2")
	form.SetFieldValue("nonexistent", "value3") // Should not panic

	if form.Fields[0].Value != "value1" {
		t.Errorf("Field[0].Value = %q, want %q", form.Fields[0].Value, "value1")
	}
	if form.Fields[1].Value != "value2" {
		t.Errorf("Field[1].Value = %q, want %q", form.Fields[1].Value, "value2")
	}
}

func TestFormReset(t *testing.T) {
	form := NewForm("Test", []string{"field1", "field2"})

	// Set some state
	form.Fields[0].Value = "value1"
	form.Fields[1].Value = "value2"
	form.CurrentField = 1
	form.Submitted = true
	form.ShowError = true
	form.ErrorMessage = "Some error"

	// Reset
	form.Reset()

	// Check all fields are reset
	for i, field := range form.Fields {
		if field.Value != "" {
			t.Errorf("Field[%d].Value = %q after reset, want empty", i, field.Value)
		}
	}

	if form.CurrentField != 0 {
		t.Errorf("CurrentField = %d after reset, want 0", form.CurrentField)
	}
	if form.Submitted {
		t.Error("Submitted should be false after reset")
	}
	if form.ShowError {
		t.Error("ShowError should be false after reset")
	}
	if form.ErrorMessage != "" {
		t.Errorf("ErrorMessage = %q after reset, want empty", form.ErrorMessage)
	}
}
