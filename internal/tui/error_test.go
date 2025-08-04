package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/grigri/pv/internal/errors"
)

func TestNewErrorModel(t *testing.T) {
	testError := fmt.Errorf("test error")
	model := NewErrorModel(testError)
	
	if model.err != testError {
		t.Error("expected error to be set")
	}
	
	if model.message != "test error" {
		t.Errorf("expected message 'test error', got %s", model.message)
	}
	
	if model.terminated {
		t.Error("expected terminated to be false initially")
	}
	
	if model.width != MinScreenWidth {
		t.Errorf("expected width %d, got %d", MinScreenWidth, model.width)
	}
	
	if model.height != MinScreenHeight {
		t.Errorf("expected height %d, got %d", MinScreenHeight, model.height)
	}
}

func TestNewErrorModel_WithAppError(t *testing.T) {
	appError := errors.NewAppError(errors.ErrValidation, "validation failed", fmt.Errorf("root cause"))
	model := NewErrorModel(appError)
	
	if model.err != appError {
		t.Error("expected app error to be set")
	}
	
	if model.message != "validation failed" {
		t.Errorf("expected message 'validation failed', got %s", model.message)
	}
	
	if model.errorType != errors.ErrValidation {
		t.Errorf("expected error type %v, got %v", errors.ErrValidation, model.errorType)
	}
}

func TestErrorModel_Init(t *testing.T) {
	model := NewErrorModel(fmt.Errorf("test"))
	cmd := model.Init()
	
	if cmd != nil {
		t.Errorf("expected nil command, got %v", cmd)
	}
}

func TestErrorModel_Update_KeyHandling(t *testing.T) {
	// Test any key press terminates the model
	testKeys := []string{"enter", "space", "escape", "q", "y", "n", "a"}
	
	for _, key := range testKeys {
		t.Run(fmt.Sprintf("key_%s", key), func(t *testing.T) {
			freshModel := NewErrorModel(fmt.Errorf("test"))
			
			var keyMsg tea.KeyMsg
			switch key {
			case "enter":
				keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
			case "space":
				keyMsg = tea.KeyMsg{Type: tea.KeySpace}
			case "escape":
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			default:
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			}
			
			updatedModel, cmd := freshModel.Update(keyMsg)
			newModel := updatedModel.(ErrorModel)
			
			if !newModel.terminated {
				t.Errorf("expected model to be terminated after key %s", key)
			}
			
			if cmd == nil {
				t.Errorf("expected quit command after key %s", key)
			}
		})
	}
}

func TestErrorModel_Update_WindowSizeMsg(t *testing.T) {
	model := NewErrorModel(fmt.Errorf("test"))
	
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(msg)
	newModel := updatedModel.(ErrorModel)
	
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

func TestErrorModel_Update_UnknownMessage(t *testing.T) {
	model := NewErrorModel(fmt.Errorf("test"))
	
	unknownMsg := struct{ data string }{data: "unknown"}
	updatedModel, cmd := model.Update(unknownMsg)
	newModel := updatedModel.(ErrorModel)
	
	if newModel.terminated != model.terminated {
		t.Error("expected model state to remain unchanged for unknown message")
	}
	
	if cmd != nil {
		t.Errorf("expected no command for unknown message, got %v", cmd)
	}
}

func TestErrorModel_View(t *testing.T) {
	testError := fmt.Errorf("this is a test error message")
	model := NewErrorModel(testError)
	view := model.View()
	
	if view == "" {
		t.Error("expected non-empty view")
	}
	
	// Should contain the error message
	if !contains(view, "this is a test error message") {
		t.Error("expected view to contain error message")
	}
	
	// Should contain continue instruction
	if !contains(view, "按任意键继续") {
		t.Error("expected view to contain continue instruction")
	}
}

func TestErrorModel_GetErrorIconAndColor(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expectIcon string
		expectColor string
	}{
		{
			name:        "validation error",
			err:         errors.NewAppError(errors.ErrValidation, "validation failed", fmt.Errorf("root")),
			expectIcon:  "⚠️",
			expectColor: ColorWarning,
		},
		{
			name:        "storage error", 
			err:         errors.NewAppError(errors.ErrStorage, "storage failed", fmt.Errorf("root")),
			expectIcon:  "💾",
			expectColor: ColorError,
		},
		{
			name:        "auth error",
			err:         errors.NewAppError(errors.ErrAuth, "auth failed", fmt.Errorf("root")),
			expectIcon:  "🔐",
			expectColor: ColorError,
		},
		{
			name:        "network error",
			err:         errors.NewAppError(errors.ErrNetwork, "network failed", fmt.Errorf("root")),
			expectIcon:  "🌐",
			expectColor: ColorWarning,
		},
		{
			name:        "generic error",
			err:         fmt.Errorf("generic error"),
			expectIcon:  "❌",
			expectColor: ColorError,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewErrorModel(tt.err)
			icon, color := model.getErrorIconAndColor()
			
			if icon != tt.expectIcon {
				t.Errorf("expected icon %s, got %s", tt.expectIcon, icon)
			}
			
			if color != tt.expectColor {
				t.Errorf("expected color %s, got %s", tt.expectColor, color)
			}
		})
	}
}

func TestErrorModel_GetErrorTitle(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectTitle string
	}{
		{
			name:        "validation error",
			err:         errors.NewAppError(errors.ErrValidation, "validation failed", fmt.Errorf("root")),
			expectTitle: "输入验证错误",
		},
		{
			name:        "storage error",
			err:         errors.NewAppError(errors.ErrStorage, "storage failed", fmt.Errorf("root")),
			expectTitle: "存储操作失败",
		},
		{
			name:        "auth error",
			err:         errors.NewAppError(errors.ErrAuth, "auth failed", fmt.Errorf("root")),
			expectTitle: "权限验证失败",
		},
		{
			name:        "network error",
			err:         errors.NewAppError(errors.ErrNetwork, "network failed", fmt.Errorf("root")),
			expectTitle: "网络连接错误",
		},
		{
			name:        "permission error",
			err:         errors.NewAppError(errors.ErrPermission, "permission failed", fmt.Errorf("root")),
			expectTitle: "权限验证失败",
		},
		{
			name:        "generic error",
			err:         fmt.Errorf("generic error"),
			expectTitle: "系统错误",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewErrorModel(tt.err)
			title := model.getErrorTitle()
			
			if title != tt.expectTitle {
				t.Errorf("expected title %s, got %s", tt.expectTitle, title)
			}
		})
	}
}

func TestErrorModel_WrapMessage(t *testing.T) {
	model := NewErrorModel(fmt.Errorf("test"))
	
	tests := []struct {
		name     string
		message  string
		width    int
		expected int // expected minimum number of lines
	}{
		{
			name:     "short message",
			message:  "short",
			width:    50,
			expected: 1,
		},
		{
			name:     "long message",
			message:  "This is a very long error message that should be wrapped across multiple lines when displayed in the error dialog",
			width:    30,
			expected: 2, // Should be wrapped to multiple lines
		},
		{
			name:     "empty message",
			message:  "",
			width:    50,
			expected: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := model.wrapMessage(tt.message, tt.width)
			lines := strings.Split(wrapped, "\n")
			
			if len(lines) < tt.expected {
				t.Errorf("expected at least %d lines, got %d", tt.expected, len(lines))
			}
			
			// Check that each line is within reasonable bounds
			for i, line := range lines {
				if len(line) > tt.width+10 { // Allow some flexibility for wrapping
					t.Errorf("line %d significantly exceeds width limit: %d > %d", i, len(line), tt.width+10)
				}
			}
		})
	}
}

func TestErrorModel_IsTerminated(t *testing.T) {
	model := NewErrorModel(fmt.Errorf("test"))
	
	if model.IsTerminated() {
		t.Error("expected not terminated initially")
	}
	
	model.terminated = true
	if !model.IsTerminated() {
		t.Error("expected terminated after setting")
	}
}

func TestErrorModel_ViewWithDifferentSizes(t *testing.T) {
	testError := fmt.Errorf("test error with longer message")
	model := NewErrorModel(testError)
	
	// Test with different window sizes
	sizes := []struct {
		width  int
		height int
	}{
		{40, 10},
		{80, 20},
		{120, 30},
	}
	
	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%dx%d", size.width, size.height), func(t *testing.T) {
			model.width = size.width
			model.height = size.height
			
			view := model.View()
			if view == "" {
				t.Error("expected non-empty view")
			}
			
			// View should adapt to different sizes
			if !contains(view, "test error") {
				t.Error("expected view to contain error message regardless of size")
			}
		})
	}
}

func TestErrorModel_MultipleKeyPresses(t *testing.T) {
	model := NewErrorModel(fmt.Errorf("test"))
	
	// First key press should terminate
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(keyMsg)
	newModel := updatedModel.(ErrorModel)
	
	if !newModel.terminated {
		t.Error("expected model to be terminated after first key press")
	}
	
	if cmd == nil {
		t.Error("expected quit command after first key press")
	}
	
	// Subsequent key presses should not change state
	keyMsg2 := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel2, cmd2 := newModel.Update(keyMsg2)
	newModel2 := updatedModel2.(ErrorModel)
	
	if newModel2.terminated != newModel.terminated {
		t.Error("expected state to remain unchanged after termination")
	}
	
	// Should still return quit command for consistency
	if cmd2 == nil {
		t.Error("expected quit command even after termination")
	}
}