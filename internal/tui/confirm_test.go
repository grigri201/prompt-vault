package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/grigri/pv/internal/model"
)

func TestNewConfirmModel(t *testing.T) {
	testPrompt := CreateSingleTestPrompt()
	model := NewConfirmModel(testPrompt)
	
	if model.prompt.ID != testPrompt.ID {
		t.Errorf("expected prompt ID %s, got %s", testPrompt.ID, model.prompt.ID)
	}
	
	if model.confirmed {
		t.Error("expected confirmed to be false initially")
	}
	
	if model.cancelled {
		t.Error("expected cancelled to be false initially")
	}
	
	if model.err != nil {
		t.Errorf("expected no error initially, got %v", model.err)
	}
	
	if model.width != 60 {
		t.Errorf("expected width 60, got %d", model.width)
	}
	
	if model.height != 10 {
		t.Errorf("expected height 10, got %d", model.height)
	}
}

func TestConfirmModel_Init(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	cmd := model.Init()
	
	if cmd != nil {
		t.Errorf("expected nil command, got %v", cmd)
	}
}

func TestConfirmModel_Update_WindowSizeMsg(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(msg)
	newModel := updatedModel.(ConfirmModel)
	
	if newModel.width != 100 {
		t.Errorf("expected width 100, got %d", newModel.width)
	}
	
	if newModel.height != 50 {
		t.Errorf("expected height 50, got %d", newModel.height)
	}
	
	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}
}

func TestConfirmModel_Update_KeyboardInput(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		expectedConf bool
		expectedCanc bool
		expectQuit   bool
	}{
		{
			name:         "y key confirms",
			key:          "y",
			expectedConf: true,
			expectedCanc: false,
			expectQuit:   true,
		},
		{
			name:         "Y key confirms (uppercase)",
			key:          "Y",
			expectedConf: true,
			expectedCanc: false,
			expectQuit:   true,
		},
		{
			name:         "n key cancels",
			key:          "n",
			expectedConf: false,
			expectedCanc: true,
			expectQuit:   true,
		},
		{
			name:         "N key cancels (uppercase)",
			key:          "N",
			expectedConf: false,
			expectedCanc: true,
			expectQuit:   true,
		},
		{
			name:         "esc key cancels",
			key:          "esc",
			expectedConf: false,
			expectedCanc: true,
			expectQuit:   true,
		},
		{
			name:         "q key cancels",
			key:          "q",
			expectedConf: false,
			expectedCanc: true,
			expectQuit:   true,
		},
		{
			name:         "ctrl+c cancels",
			key:          "ctrl+c",
			expectedConf: false,
			expectedCanc: true,
			expectQuit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewConfirmModel(CreateSingleTestPrompt())
			
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "esc" {
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			} else if tt.key == "ctrl+c" {
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}
			
			updatedModel, cmd := model.Update(keyMsg)
			newModel := updatedModel.(ConfirmModel)
			
			if newModel.confirmed != tt.expectedConf {
				t.Errorf("expected confirmed %v, got %v", tt.expectedConf, newModel.confirmed)
			}
			
			if newModel.cancelled != tt.expectedCanc {
				t.Errorf("expected cancelled %v, got %v", tt.expectedCanc, newModel.cancelled)
			}
			
			if tt.expectQuit && cmd == nil {
				t.Error("expected quit command")
			}
			
			if !tt.expectQuit && cmd != nil {
				t.Errorf("expected no command, got %v", cmd)
			}
		})
	}
}

func TestConfirmModel_Update_IgnoredKeys(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	// Test that other keys don't change the model state
	ignoredKeys := []string{"a", "z", "1", "space", "tab", "enter"}
	
	for _, key := range ignoredKeys {
		t.Run(fmt.Sprintf("ignores_%s", key), func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			if key == "space" {
				keyMsg = tea.KeyMsg{Type: tea.KeySpace}
			} else if key == "tab" {
				keyMsg = tea.KeyMsg{Type: tea.KeyTab}
			} else if key == "enter" {
				keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
			}
			
			updatedModel, cmd := model.Update(keyMsg)
			newModel := updatedModel.(ConfirmModel)
			
			if newModel.confirmed {
				t.Error("expected confirmed to remain false")
			}
			
			if newModel.cancelled {
				t.Error("expected cancelled to remain false")
			}
			
			if cmd != nil {
				t.Errorf("expected no command, got %v", cmd)
			}
		})
	}
}

func TestConfirmModel_View_NormalState(t *testing.T) {
	testPrompt := CreateSingleTestPrompt()
	model := NewConfirmModel(testPrompt)
	view := model.View()
	
	if view == "" {
		t.Error("expected non-empty view")
	}
	
	// Should contain warning icon
	if !contains(view, "⚠️") {
		t.Error("expected warning icon in view")
	}
	
	// Should contain confirmation title
	if !contains(view, "确认删除提示") {
		t.Error("expected confirmation title in view")
	}
	
	// Should contain prompt name
	if !contains(view, testPrompt.Name) {
		t.Errorf("expected view to contain prompt name %s", testPrompt.Name)
	}
	
	// Should contain author
	if !contains(view, testPrompt.Author) {
		t.Errorf("expected view to contain author %s", testPrompt.Author)
	}
	
	// Should contain gist URL
	if !contains(view, testPrompt.GistURL) {
		t.Errorf("expected view to contain gist URL %s", testPrompt.GistURL)
	}
	
	// Should contain warning message
	if !contains(view, "此操作不可撤销") {
		t.Error("expected warning message in view")
	}
	
	// Should contain confirmation buttons
	if !contains(view, "[Y] 确认删除") {
		t.Error("expected confirm button in view")
	}
	
	if !contains(view, "[N] 取消") {
		t.Error("expected cancel button in view")
	}
	
	// Should contain help text
	if !contains(view, "Y: 确认") {
		t.Error("expected help text in view")
	}
}

func TestConfirmModel_View_WithDescription(t *testing.T) {
	testPrompt := CreateSingleTestPrompt()
	testPrompt.Description = "This is a test description"
	model := NewConfirmModel(testPrompt)
	view := model.View()
	
	// Should contain description
	if !contains(view, testPrompt.Description) {
		t.Errorf("expected view to contain description %s", testPrompt.Description)
	}
}

func TestConfirmModel_View_WithoutDescription(t *testing.T) {
	testPrompt := CreateSingleTestPrompt()
	testPrompt.Description = ""
	model := NewConfirmModel(testPrompt)
	view := model.View()
	
	// Should not contain description label when empty
	if contains(view, "描述:") {
		t.Error("expected view not to contain description label when empty")
	}
}

func TestConfirmModel_View_ErrorState(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	model.err = fmt.Errorf("test error")
	view := model.View()
	
	if view == "" {
		t.Error("expected non-empty view")
	}
	
	// Should contain error icon
	if !contains(view, "❌") {
		t.Error("expected error icon in view")
	}
	
	// Should contain error message
	if !contains(view, "test error") {
		t.Error("expected error message in view")
	}
	
	// Should contain continue instruction
	if !contains(view, "按任意键继续") {
		t.Error("expected continue instruction in view")
	}
}

func TestConfirmModel_TruncateString(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string unchanged",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "exact length unchanged",
			input:    "exactly10c",
			maxLen:   10,
			expected: "exactly10c",
		},
		{
			name:     "long string truncated",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is...",
		},
		{
			name:     "very short maxLen",
			input:    "hello",
			maxLen:   3,
			expected: "hel",
		},
		{
			name:     "maxLen less than ellipsis",
			input:    "hello",
			maxLen:   2,
			expected: "he",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConfirmModel_IsConfirmed(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	if model.IsConfirmed() {
		t.Error("expected not confirmed initially")
	}
	
	model.confirmed = true
	if !model.IsConfirmed() {
		t.Error("expected confirmed after setting")
	}
}

func TestConfirmModel_IsCancelled(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	if model.IsCancelled() {
		t.Error("expected not cancelled initially")
	}
	
	model.cancelled = true
	if !model.IsCancelled() {
		t.Error("expected cancelled after setting")
	}
}

func TestConfirmModel_HasError(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	if model.HasError() {
		t.Error("expected no error initially")
	}
	
	model.err = fmt.Errorf("test error")
	if !model.HasError() {
		t.Error("expected error after setting")
	}
}

func TestConfirmModel_SetError(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	testError := fmt.Errorf("test error")
	
	model.SetError(testError)
	
	if model.err != testError {
		t.Error("expected error to be set")
	}
	
	if !model.HasError() {
		t.Error("expected HasError to return true after setting error")
	}
}

func TestConfirmModel_View_TruncatedFields(t *testing.T) {
	// Create prompt with very long fields
	longPrompt := model.Prompt{
		ID:          "test",
		Name:        "This is a very long prompt name that should be truncated when displayed in the confirmation dialog",
		Author:      "VeryLongAuthorNameThatShouldBeTruncatedInTheConfirmationDialog",
		GistURL:     "https://gist.github.com/verylongusername/verylonggistidthatshouldbetrunacted123456789",
		Description: "This is a very long description that should be truncated when displayed in the confirmation dialog for deletion",
	}
	
	confirmModel := NewConfirmModel(longPrompt)
	view := confirmModel.View()
	
	// Should contain truncation indicators
	if !contains(view, "...") {
		t.Error("expected truncated text to contain '...'")
	}
	
	// Should still be displayable
	if view == "" {
		t.Error("expected non-empty view even with long fields")
	}
}

func TestConfirmModel_KeyHandling_CaseInsensitive(t *testing.T) {
	// Test that both uppercase and lowercase work
	tests := []struct {
		key          string
		expectConfirm bool
		expectCancel  bool
	}{
		{"y", true, false},
		{"Y", true, false},
		{"n", false, true},
		{"N", false, true},
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("key_%s", tt.key), func(t *testing.T) {
			model := NewConfirmModel(CreateSingleTestPrompt())
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			
			updatedModel, cmd := model.Update(keyMsg)
			newModel := updatedModel.(ConfirmModel)
			
			if newModel.confirmed != tt.expectConfirm {
				t.Errorf("expected confirmed %v, got %v", tt.expectConfirm, newModel.confirmed)
			}
			
			if newModel.cancelled != tt.expectCancel {
				t.Errorf("expected cancelled %v, got %v", tt.expectCancel, newModel.cancelled)
			}
			
			if cmd == nil {
				t.Error("expected quit command")
			}
		})
	}
}

func TestConfirmModel_StateTransitions(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	// Initially both should be false
	if model.IsConfirmed() || model.IsCancelled() {
		t.Error("expected both confirmed and cancelled to be false initially")
	}
	
	// After confirmation, confirmed should be true but cancelled false
	model.confirmed = true
	if !model.IsConfirmed() || model.IsCancelled() {
		t.Error("expected confirmed true and cancelled false after confirmation")
	}
	
	// Reset and test cancellation
	model.confirmed = false
	model.cancelled = true
	if model.IsConfirmed() || !model.IsCancelled() {
		t.Error("expected confirmed false and cancelled true after cancellation")
	}
}

func TestConfirmModel_UnknownMessageTypes(t *testing.T) {
	model := NewConfirmModel(CreateSingleTestPrompt())
	
	// Test with unknown message type
	unknownMsg := struct{ data string }{data: "unknown"}
	updatedModel, cmd := model.Update(unknownMsg)
	
	// Should return model with same state
	newModel := updatedModel.(ConfirmModel)
	if newModel.confirmed != model.confirmed || newModel.cancelled != model.cancelled {
		t.Error("expected model state to remain unchanged for unknown message type")
	}
	
	if cmd != nil {
		t.Errorf("expected no command for unknown message type, got %v", cmd)
	}
}

func TestConfirmModel_View_FieldDisplayOrder(t *testing.T) {
	testPrompt := model.Prompt{
		ID:          "test-id",
		Name:        "Test Name",
		Author:      "Test Author", 
		GistURL:     "https://gist.github.com/test/123",
		Description: "Test Description",
	}
	
	confirmModel := NewConfirmModel(testPrompt)
	view := confirmModel.View()
	
	// Check that fields appear in expected order
	namePos := findStringPos(view, "提示名称:")
	authorPos := findStringPos(view, "作者:")
	descPos := findStringPos(view, "描述:")
	urlPos := findStringPos(view, "Gist URL:")
	
	if namePos == -1 {
		t.Error("expected to find name label")
	}
	
	if authorPos == -1 {
		t.Error("expected to find author label")
	}
	
	if descPos == -1 {
		t.Error("expected to find description label")
	}
	
	if urlPos == -1 {
		t.Error("expected to find URL label")
	}
	
	// Name should come before author
	if namePos > authorPos {
		t.Error("expected name to appear before author")
	}
	
	// Description should come after author but before URL
	if descPos < authorPos || descPos > urlPos {
		t.Error("expected description to appear between author and URL")
	}
}

// Helper function to find position of substring in string
func findStringPos(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}