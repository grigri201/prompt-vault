package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grigri201/prompt-vault/internal/models"
)

func TestDuplicateHandler_Navigation(t *testing.T) {
	existing := &models.IndexEntry{
		Name:      "Test Prompt",
		Author:    "john",
		Version:   "1.0",
		ID:        "john-test",
		UpdatedAt: time.Now(),
	}
	
	newPrompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:   "Test Prompt Updated",
			Author: "john",
		},
	}
	
	m := NewDuplicateHandler(existing, newPrompt)
	
	// Test initial state
	if m.cursor != 0 {
		t.Errorf("Initial cursor position should be 0, got %d", m.cursor)
	}
	
	// Test down navigation
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updatedModel.(DuplicateHandlerModel)
	if m.cursor != 1 {
		t.Errorf("After pressing 'j', cursor should be 1, got %d", m.cursor)
	}
	
	// Test down navigation with arrow key
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(DuplicateHandlerModel)
	if m.cursor != 2 {
		t.Errorf("After pressing down arrow, cursor should be 2, got %d", m.cursor)
	}
	
	// Test boundary - can't go beyond last option
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(DuplicateHandlerModel)
	if m.cursor != 2 {
		t.Errorf("Cursor should stay at 2 (last option), got %d", m.cursor)
	}
	
	// Test up navigation
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(DuplicateHandlerModel)
	if m.cursor != 1 {
		t.Errorf("After pressing up arrow, cursor should be 1, got %d", m.cursor)
	}
}

func TestDuplicateHandler_Selection(t *testing.T) {
	existing := &models.IndexEntry{
		Name:      "Test Prompt",
		Author:    "john",
		Version:   "1.0",
		UpdatedAt: time.Now(),
	}
	
	newPrompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:   "Test Prompt Updated",
			Author: "john",
		},
	}
	
	tests := []struct {
		name           string
		cursorPosition int
		expectedChoice DuplicateChoice
		expectQuit     bool
		expectIDInput  bool
	}{
		{
			name:           "select update existing",
			cursorPosition: 0,
			expectedChoice: UpdateExisting,
			expectQuit:     true,
			expectIDInput:  false,
		},
		{
			name:           "select create new",
			cursorPosition: 1,
			expectedChoice: CreateNew,
			expectQuit:     false,
			expectIDInput:  true,
		},
		{
			name:           "select cancel",
			cursorPosition: 2,
			expectedChoice: Cancel,
			expectQuit:     true,
			expectIDInput:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDuplicateHandler(existing, newPrompt)
			m.cursor = tt.cursorPosition
			
			updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			m = updatedModel.(DuplicateHandlerModel)
			
			if m.choice != tt.expectedChoice {
				t.Errorf("Expected choice %v, got %v", tt.expectedChoice, m.choice)
			}
			
			if tt.expectQuit && cmd == nil {
				t.Error("Expected quit command, got nil")
			}
			
			if m.showIDInput != tt.expectIDInput {
				t.Errorf("Expected showIDInput to be %v, got %v", tt.expectIDInput, m.showIDInput)
			}
		})
	}
}

func TestDuplicateHandler_IDInput(t *testing.T) {
	existing := &models.IndexEntry{
		Name:      "Test Prompt",
		Author:    "john",
		Version:   "1.0",
		UpdatedAt: time.Now(),
	}
	
	newPrompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:   "Test Prompt Updated",
			Author: "john",
		},
	}
	
	m := NewDuplicateHandler(existing, newPrompt)
	m.showIDInput = true
	m.choice = CreateNew
	
	// Type some characters
	testID := "test-id-123"
	for _, ch := range testID {
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = updatedModel.(DuplicateHandlerModel)
	}
	
	if m.newID != testID {
		t.Errorf("Expected newID to be '%s', got '%s'", testID, m.newID)
	}
	
	// Test backspace
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updatedModel.(DuplicateHandlerModel)
	
	expectedID := testID[:len(testID)-1]
	if m.newID != expectedID {
		t.Errorf("After backspace, expected newID to be '%s', got '%s'", expectedID, m.newID)
	}
	
	// Test enter with non-empty ID
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(DuplicateHandlerModel)
	
	if !m.confirmed {
		t.Error("Expected confirmed to be true after pressing Enter")
	}
	
	if cmd == nil {
		t.Error("Expected quit command after confirming ID")
	}
}

func TestDuplicateHandler_EmptyIDValidation(t *testing.T) {
	existing := &models.IndexEntry{
		Name:      "Test Prompt",
		Author:    "john",
		Version:   "1.0",
		UpdatedAt: time.Now(),
	}
	
	newPrompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:   "Test Prompt Updated",
			Author: "john",
		},
	}
	
	m := NewDuplicateHandler(existing, newPrompt)
	m.showIDInput = true
	m.choice = CreateNew
	m.newID = "" // Empty ID
	
	// Try to submit empty ID
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(DuplicateHandlerModel)
	
	if m.confirmed {
		t.Error("Should not confirm with empty ID")
	}
	
	if cmd != nil {
		t.Error("Should not quit with empty ID")
	}
	
	if m.errorMessage == "" {
		t.Error("Expected error message for empty ID")
	}
}

func TestDuplicateHandler_Cancel(t *testing.T) {
	existing := &models.IndexEntry{
		Name:      "Test Prompt",
		Author:    "john",
		Version:   "1.0",
		UpdatedAt: time.Now(),
	}
	
	newPrompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:   "Test Prompt Updated",
			Author: "john",
		},
	}
	
	// Test cancel with 'q'
	m := NewDuplicateHandler(existing, newPrompt)
	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updatedModel.(DuplicateHandlerModel)
	
	if m.choice != Cancel {
		t.Errorf("Expected choice to be Cancel, got %v", m.choice)
	}
	
	if !m.cancelled {
		t.Error("Expected cancelled to be true")
	}
	
	if cmd == nil {
		t.Error("Expected quit command")
	}
	
	// Test cancel with Ctrl+C
	m2 := NewDuplicateHandler(existing, newPrompt)
	updatedModel2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m2 = updatedModel2.(DuplicateHandlerModel)
	
	if m2.choice != Cancel {
		t.Errorf("Expected choice to be Cancel with Ctrl+C, got %v", m2.choice)
	}
	
	if !m2.cancelled {
		t.Error("Expected cancelled to be true with Ctrl+C")
	}
}

func TestDuplicateHandler_EscapeFromIDInput(t *testing.T) {
	existing := &models.IndexEntry{
		Name:      "Test Prompt",
		Author:    "john",
		Version:   "1.0",
		UpdatedAt: time.Now(),
	}
	
	newPrompt := &models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:   "Test Prompt Updated",
			Author: "john",
		},
	}
	
	m := NewDuplicateHandler(existing, newPrompt)
	m.showIDInput = true
	m.choice = CreateNew
	m.newID = "some-text"
	m.errorMessage = "some error"
	
	// Press Escape
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updatedModel.(DuplicateHandlerModel)
	
	if m.showIDInput {
		t.Error("Should return to menu after pressing Escape")
	}
	
	if m.newID != "" {
		t.Error("Should clear newID after pressing Escape")
	}
	
	if m.errorMessage != "" {
		t.Error("Should clear error message after pressing Escape")
	}
}