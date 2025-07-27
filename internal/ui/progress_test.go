package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestProgressModel_Init(t *testing.T) {
	model := NewProgressModel("Loading...")
	cmd := model.Init()
	
	// Init should return a tick command
	if cmd == nil {
		t.Error("Expected Init to return a tick command, got nil")
	}
}

func TestProgressModel_Update(t *testing.T) {
	tests := []struct {
		name            string
		msg             tea.Msg
		expectQuit      bool
		expectTickCmd   bool
		expectStateChange bool
	}{
		{
			name:            "tick message",
			msg:             tickMsg(time.Now()),
			expectQuit:      false,
			expectTickCmd:   true,
			expectStateChange: true,
		},
		{
			name:            "quit key",
			msg:             tea.KeyMsg{Type: tea.KeyCtrlC},
			expectQuit:      true,
			expectTickCmd:   false,  // tea.Quit is returned, not tick
			expectStateChange: false,
		},
		{
			name:            "esc key",
			msg:             tea.KeyMsg{Type: tea.KeyEsc},
			expectQuit:      true,
			expectTickCmd:   false,  // tea.Quit is returned, not tick
			expectStateChange: false,
		},
		{
			name:            "other key",
			msg:             tea.KeyMsg{Type: tea.KeySpace},
			expectQuit:      false,
			expectTickCmd:   false,
			expectStateChange: false,
		},
		{
			name:            "progress complete message",
			msg:             ProgressCompleteMsg{},
			expectQuit:      true,
			expectTickCmd:   false,  // tea.Quit is returned, not tick
			expectStateChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewProgressModel("Testing...")
			initialFrame := model.(*ProgressModel).frame
			
			updatedModel, cmd := model.Update(tt.msg)
			progressModel := updatedModel.(*ProgressModel)
			
			// Check quit state
			if progressModel.quitting != tt.expectQuit {
				t.Errorf("Expected quitting = %v, got %v", tt.expectQuit, progressModel.quitting)
			}
			
			// Check if command is returned
			if tt.expectTickCmd && cmd == nil {
				t.Error("Expected tick command, got nil")
			} else if tt.expectQuit && cmd == nil {
				t.Error("Expected quit command, got nil")
			} else if !tt.expectTickCmd && !tt.expectQuit && cmd != nil {
				t.Errorf("Expected no command, got %v", cmd)
			}
			
			// Check state change
			if tt.expectStateChange && progressModel.frame == initialFrame {
				t.Error("Expected frame to change, but it didn't")
			} else if !tt.expectStateChange && progressModel.frame != initialFrame {
				t.Error("Expected frame to remain the same, but it changed")
			}
		})
	}
}

func TestProgressModel_View(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		quitting bool
		expected string
	}{
		{
			name:     "normal state",
			message:  "Loading data...",
			quitting: false,
			expected: "Loading data...",
		},
		{
			name:     "quitting state",
			message:  "Loading data...",
			quitting: true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &ProgressModel{
				message:  tt.message,
				quitting: tt.quitting,
				frame:    0,
			}
			
			view := model.View()
			
			if tt.quitting && view != "" {
				t.Errorf("Expected empty view when quitting, got %q", view)
			}
			
			if !tt.quitting && !contains(view, tt.expected) {
				t.Errorf("Expected view to contain %q, got %q", tt.expected, view)
			}
		})
	}
}

func TestProgressModel_SpinnerFrames(t *testing.T) {
	model := NewProgressModel("Testing...")
	progressModel := model.(*ProgressModel)
	
	// Test that spinner advances through frames
	frames := []string{}
	for i := 0; i < 10; i++ {
		view := progressModel.View()
		frames = append(frames, view)
		
		// Advance frame
		progressModel.frame++
	}
	
	// Check that we have different frames
	uniqueFrames := make(map[string]bool)
	for _, frame := range frames {
		uniqueFrames[frame] = true
	}
	
	if len(uniqueFrames) < 2 {
		t.Error("Expected spinner to have multiple unique frames")
	}
}

func TestProgressModel_SetMessage(t *testing.T) {
	model := NewProgressModel("Initial message")
	progressModel := model.(*ProgressModel)
	
	// Test SetMessage method
	newMessage := "Updated message"
	progressModel.SetMessage(newMessage)
	
	if progressModel.message != newMessage {
		t.Errorf("Expected message to be %q, got %q", newMessage, progressModel.message)
	}
	
	// Verify it appears in the view
	view := progressModel.View()
	if !contains(view, newMessage) {
		t.Errorf("Expected view to contain %q, got %q", newMessage, view)
	}
}

func TestProgress_Integration(t *testing.T) {
	// Test that progress can be started and stopped
	model := NewProgressModel("Processing...")
	
	// Simulate a few update cycles
	var cmd tea.Cmd
	cmd = model.Init()
	
	// First tick
	if cmd != nil {
		// This would normally be a tick command
		model, cmd = model.Update(tickMsg(time.Now()))
	}
	
	// Check that it's still running
	progressModel := model.(*ProgressModel)
	if progressModel.quitting {
		t.Error("Progress should not be quitting yet")
	}
	
	// Send complete message
	model, _ = model.Update(ProgressCompleteMsg{})
	progressModel = model.(*ProgressModel)
	
	if !progressModel.quitting {
		t.Error("Progress should be quitting after complete message")
	}
}

// Helper function for tests
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}