package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSelector(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	s := NewSelector(items)

	if len(s.Items) != len(items) {
		t.Errorf("Items length = %d, want %d", len(s.Items), len(items))
	}

	if s.Selected != -1 {
		t.Errorf("Selected = %d, want -1", s.Selected)
	}

	if s.Input != "" {
		t.Errorf("Input = %q, want empty", s.Input)
	}

	if s.Confirmed {
		t.Error("Confirmed should be false initially")
	}

	if s.MaxSelection != 3 {
		t.Errorf("MaxSelection = %d, want 3", s.MaxSelection)
	}
}

func TestSelectorNumberInput(t *testing.T) {
	items := []string{"A", "B", "C", "D", "E"}

	tests := []struct {
		name      string
		keys      []string
		wantInput string
	}{
		{
			name:      "single digit",
			keys:      []string{"1"},
			wantInput: "1",
		},
		{
			name:      "multiple digits",
			keys:      []string{"1", "2", "3"},
			wantInput: "123",
		},
		{
			name:      "backspace removes digit",
			keys:      []string{"1", "2", "backspace"},
			wantInput: "1",
		},
		{
			name:      "backspace on empty does nothing",
			keys:      []string{"backspace"},
			wantInput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewSelector(items)
			model := tea.Model(selector)

			for _, key := range tt.keys {
				var msg tea.Msg
				if key == "backspace" {
					msg = tea.KeyMsg{Type: tea.KeyBackspace}
				} else {
					msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
				}
				model, _ = model.Update(msg)
			}

			s := model.(SelectorModel)
			if s.Input != tt.wantInput {
				t.Errorf("Input = %q, want %q", s.Input, tt.wantInput)
			}
		})
	}
}

func TestSelectorValidation(t *testing.T) {
	items := []string{"A", "B", "C"}
	
	tests := []struct {
		name         string
		input        string
		wantError    bool
		wantSelected int
		errorContains string
	}{
		{
			name:         "valid selection",
			input:        "2",
			wantError:    false,
			wantSelected: 1, // 0-based index
		},
		{
			name:          "non-numeric input",
			input:         "abc",
			wantError:     true,
			wantSelected:  -1,
			errorContains: "valid number",
		},
		{
			name:          "number too high",
			input:         "5",
			wantError:     true,
			wantSelected:  -1,
			errorContains: "between 1 and 3",
		},
		{
			name:          "number too low",
			input:         "0",
			wantError:     true,
			wantSelected:  -1,
			errorContains: "between 1 and 3",
		},
		{
			name:          "negative number",
			input:         "-1",
			wantError:     true,
			wantSelected:  -1,
			errorContains: "between 1 and 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSelector(items)
			s.Input = tt.input

			model, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
			s = model.(SelectorModel)

			if tt.wantError {
				if !s.ShowError {
					t.Error("Expected error to be shown")
				}
				if !strings.Contains(s.ErrorMessage, tt.errorContains) {
					t.Errorf("Error message %q doesn't contain %q", s.ErrorMessage, tt.errorContains)
				}
				if cmd != nil {
					t.Error("Should not quit on error")
				}
			} else {
				if s.ShowError {
					t.Errorf("Unexpected error: %s", s.ErrorMessage)
				}
				if s.Selected != tt.wantSelected {
					t.Errorf("Selected = %d, want %d", s.Selected, tt.wantSelected)
				}
				if !s.Confirmed {
					t.Error("Should be confirmed on valid selection")
				}
				if cmd == nil {
					t.Error("Should quit on valid selection")
				}
			}
		})
	}
}

func TestSelectorView(t *testing.T) {
	tests := []struct {
		name         string
		items        []string
		input        string
		showError    bool
		errorMessage string
		wantContains []string
	}{
		{
			name:  "basic view",
			items: []string{"Option A", "Option B", "Option C"},
			wantContains: []string{
				"Select a prompt",
				"1. Option A",
				"2. Option B", 
				"3. Option C",
				"Enter selection (1-3):",
				"enter: confirm",
			},
		},
		{
			name:  "with input",
			items: []string{"A", "B"},
			input: "12",
			wantContains: []string{
				"Enter selection (1-2):",
				"12",
			},
		},
		{
			name:         "with error",
			items:        []string{"A"},
			showError:    true,
			errorMessage: "Invalid input",
			wantContains: []string{
				"Error:",
				"Invalid input",
			},
		},
		{
			name:  "many items truncated",
			items: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
			wantContains: []string{
				"Showing first 10 of 12 items",
				"10. 10",
				"... and 2 more",
				"Enter selection (1-12):",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSelector(tt.items)
			s.Input = tt.input
			s.ShowError = tt.showError
			s.ErrorMessage = tt.errorMessage

			view := s.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(view, want) {
					t.Errorf("View missing %q\nGot: %s", want, view)
				}
			}
		})
	}
}

func TestSelectorQuit(t *testing.T) {
	s := NewSelector([]string{"A", "B"})

	// Test q key
	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("Expected quit command for 'q' key")
	}

	// Test ctrl+c
	_, cmd = s.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Expected quit command for ctrl+c")
	}
}

func TestGetSelection(t *testing.T) {
	items := []string{"First", "Second", "Third"}
	s := NewSelector(items)

	// No selection
	if s.GetSelection() != "" {
		t.Errorf("GetSelection() = %q, want empty", s.GetSelection())
	}

	// Valid selection
	s.Selected = 1
	if s.GetSelection() != "Second" {
		t.Errorf("GetSelection() = %q, want %q", s.GetSelection(), "Second")
	}

	// Out of bounds
	s.Selected = 10
	if s.GetSelection() != "" {
		t.Errorf("GetSelection() = %q, want empty for out of bounds", s.GetSelection())
	}
}

func TestIsConfirmed(t *testing.T) {
	s := NewSelector([]string{"A"})

	if s.IsConfirmed() {
		t.Error("IsConfirmed() should be false initially")
	}

	s.Confirmed = true
	if !s.IsConfirmed() {
		t.Error("IsConfirmed() should be true after confirmation")
	}
}

func TestReset(t *testing.T) {
	s := NewSelector([]string{"A", "B"})
	
	// Set some state
	s.Selected = 1
	s.Input = "123"
	s.Confirmed = true
	s.ShowError = true
	s.ErrorMessage = "Some error"

	// Reset
	s.Reset()

	if s.Selected != -1 {
		t.Errorf("Selected = %d after reset, want -1", s.Selected)
	}
	if s.Input != "" {
		t.Errorf("Input = %q after reset, want empty", s.Input)
	}
	if s.Confirmed {
		t.Error("Confirmed should be false after reset")
	}
	if s.ShowError {
		t.Error("ShowError should be false after reset")
	}
	if s.ErrorMessage != "" {
		t.Errorf("ErrorMessage = %q after reset, want empty", s.ErrorMessage)
	}
}