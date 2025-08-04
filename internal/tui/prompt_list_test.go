package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/grigri/pv/internal/model"
)

func TestNewPromptListModel(t *testing.T) {
	tests := []struct {
		name     string
		prompts  []model.Prompt
		mode     ListMode
		filter   string
		expected PromptListModel
	}{
		{
			name:    "creates model with empty prompts",
			prompts: []model.Prompt{},
			mode:    ListAll,
			filter:  "",
			expected: PromptListModel{
				prompts:  []model.Prompt{},
				cursor:   0,
				selected: nil,
				err:      nil,
				mode:     ListAll,
				filter:   "",
				loading:  false,
				done:     false,
				quit:     false,
			},
		},
		{
			name:    "creates model with single prompt",
			prompts: []model.Prompt{CreateSingleTestPrompt()},
			mode:    ListAll,
			filter:  "",
			expected: PromptListModel{
				prompts:  []model.Prompt{CreateSingleTestPrompt()},
				cursor:   0,
				selected: nil,
				err:      nil,
				mode:     ListAll,
				filter:   "",
				loading:  false,
				done:     false,
				quit:     false,
			},
		},
		{
			name:    "creates model with filtered mode",
			prompts: CreateStandardTestPrompts(),
			mode:    ListFiltered,
			filter:  "golang",
			expected: PromptListModel{
				prompts:  CreateStandardTestPrompts(),
				cursor:   0,
				selected: nil,
				err:      nil,
				mode:     ListFiltered,
				filter:   "golang",
				loading:  false,
				done:     false,
				quit:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewPromptListModel(tt.prompts, tt.mode, tt.filter)
			
			if len(model.prompts) != len(tt.expected.prompts) {
				t.Errorf("expected %d prompts, got %d", len(tt.expected.prompts), len(model.prompts))
			}
			
			if model.cursor != tt.expected.cursor {
				t.Errorf("expected cursor %d, got %d", tt.expected.cursor, model.cursor)
			}
			
			if model.mode != tt.expected.mode {
				t.Errorf("expected mode %v, got %v", tt.expected.mode, model.mode)
			}
			
			if model.filter != tt.expected.filter {
				t.Errorf("expected filter %s, got %s", tt.expected.filter, model.filter)
			}
			
			if model.loading != tt.expected.loading {
				t.Errorf("expected loading %v, got %v", tt.expected.loading, model.loading)
			}
			
			if model.done != tt.expected.done {
				t.Errorf("expected done %v, got %v", tt.expected.done, model.done)
			}
			
			if model.quit != tt.expected.quit {
				t.Errorf("expected quit %v, got %v", tt.expected.quit, model.quit)
			}
		})
	}
}

func TestNewLoadingPromptListModel(t *testing.T) {
	model := NewLoadingPromptListModel()
	
	if model.prompts != nil {
		t.Errorf("expected nil prompts, got %v", model.prompts)
	}
	
	if model.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", model.cursor)
	}
	
	if model.mode != ListAll {
		t.Errorf("expected mode ListAll, got %v", model.mode)
	}
	
	if model.filter != "" {
		t.Errorf("expected empty filter, got %s", model.filter)
	}
	
	if !model.loading {
		t.Errorf("expected loading true, got %v", model.loading)
	}
	
	if model.done {
		t.Errorf("expected done false, got %v", model.done)
	}
	
	if model.quit {
		t.Errorf("expected quit false, got %v", model.quit)
	}
}

func TestPromptListModel_Init(t *testing.T) {
	model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
	cmd := model.Init()
	
	if cmd != nil {
		t.Errorf("expected nil command, got %v", cmd)
	}
}

func TestPromptListModel_Update_KeyboardNavigation(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		key            string
		expectedCursor int
		expectDone     bool
		expectQuit     bool
	}{
		{
			name:           "down arrow moves cursor down",
			initialCursor:  0,
			key:            KeyDown,
			expectedCursor: 1,
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "up arrow moves cursor up",
			initialCursor:  1,
			key:            KeyUp,
			expectedCursor: 0,
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "up arrow at beginning stays at 0",
			initialCursor:  0,
			key:            KeyUp,
			expectedCursor: 0,
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "down arrow at end stays at last index",
			initialCursor:  1, // With 2 test prompts, last index is 1
			key:            KeyDown,
			expectedCursor: 1,
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "home key moves to beginning",
			initialCursor:  1,
			key:            KeyHome,
			expectedCursor: 0,
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "end key moves to end",
			initialCursor:  0,
			key:            KeyEnd,
			expectedCursor: 1, // With 2 test prompts, last index is 1
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "page up moves cursor up by 10 or to beginning",
			initialCursor:  1,
			key:            KeyPageUp,
			expectedCursor: 0,
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "page down moves cursor down by 10 or to end",
			initialCursor:  0,
			key:            KeyPageDown,
			expectedCursor: 1,
			expectDone:     false,
			expectQuit:     false,
		},
		{
			name:           "enter key selects current item",
			initialCursor:  0,
			key:            KeyEnter,
			expectedCursor: 0,
			expectDone:     true,
			expectQuit:     false,
		},
		{
			name:           "q key quits without selection",
			initialCursor:  0,
			key:            KeyQuit,
			expectedCursor: 0,
			expectDone:     true,
			expectQuit:     true,
		},
		{
			name:           "escape key quits without selection",
			initialCursor:  0,
			key:            KeyEscape,
			expectedCursor: 0,
			expectDone:     true,
			expectQuit:     true,
		},
		{
			name:           "ctrl+c quits without selection",
			initialCursor:  0,
			key:            KeyCtrlC,
			expectedCursor: 0,
			expectDone:     true,
			expectQuit:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
			model.cursor = tt.initialCursor
			
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == KeyUp {
				keyMsg = tea.KeyMsg{Type: tea.KeyUp}
			} else if tt.key == KeyDown {
				keyMsg = tea.KeyMsg{Type: tea.KeyDown}
			} else if tt.key == KeyHome {
				keyMsg = tea.KeyMsg{Type: tea.KeyHome}
			} else if tt.key == KeyEnd {
				keyMsg = tea.KeyMsg{Type: tea.KeyEnd}
			} else if tt.key == KeyPageUp {
				keyMsg = tea.KeyMsg{Type: tea.KeyPgUp}
			} else if tt.key == KeyPageDown {
				keyMsg = tea.KeyMsg{Type: tea.KeyPgDown}
			} else if tt.key == KeyEnter {
				keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
			} else if tt.key == KeyEscape {
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			} else if tt.key == KeyCtrlC {
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}
			
			updatedModel, cmd := model.Update(keyMsg)
			newModel := updatedModel.(PromptListModel)
			
			if newModel.cursor != tt.expectedCursor {
				t.Errorf("expected cursor %d, got %d", tt.expectedCursor, newModel.cursor)
			}
			
			if newModel.done != tt.expectDone {
				t.Errorf("expected done %v, got %v", tt.expectDone, newModel.done)
			}
			
			if newModel.quit != tt.expectQuit {
				t.Errorf("expected quit %v, got %v", tt.expectQuit, newModel.quit)
			}
			
			if tt.expectDone && (tt.expectQuit || cmd == nil) {
				// Should have quit command when done
				if tt.expectQuit && cmd == nil {
					t.Error("expected quit command when quitting")
				}
			}
		})
	}
}

func TestPromptListModel_Update_LoadingState(t *testing.T) {
	model := NewLoadingPromptListModel()
	
	// Keys should be ignored in loading state
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, cmd := model.Update(keyMsg)
	newModel := updatedModel.(PromptListModel)
	
	if newModel.cursor != 0 {
		t.Errorf("expected cursor to remain 0 in loading state, got %d", newModel.cursor)
	}
	
	if cmd != nil {
		t.Errorf("expected no command in loading state, got %v", cmd)
	}
}

func TestPromptListModel_Update_PromptsLoadedMsg(t *testing.T) {
	model := NewLoadingPromptListModel()
	testPrompts := CreateStandardTestPrompts()
	
	msg := promptsLoadedMsg{
		prompts: testPrompts,
		mode:    ListFiltered,
		filter:  "test",
	}
	
	updatedModel, cmd := model.Update(msg)
	newModel := updatedModel.(PromptListModel)
	
	if len(newModel.prompts) != len(testPrompts) {
		t.Errorf("expected %d prompts, got %d", len(testPrompts), len(newModel.prompts))
	}
	
	if newModel.mode != ListFiltered {
		t.Errorf("expected mode ListFiltered, got %v", newModel.mode)
	}
	
	if newModel.filter != "test" {
		t.Errorf("expected filter 'test', got %s", newModel.filter)
	}
	
	if newModel.loading {
		t.Errorf("expected loading false, got %v", newModel.loading)
	}
	
	if newModel.cursor != 0 {
		t.Errorf("expected cursor reset to 0, got %d", newModel.cursor)
	}
	
	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}
}

func TestPromptListModel_Update_PromptLoadErrorMsg(t *testing.T) {
	model := NewLoadingPromptListModel()
	testError := fmt.Errorf("test error")
	
	msg := promptLoadErrorMsg{err: testError}
	
	updatedModel, cmd := model.Update(msg)
	newModel := updatedModel.(PromptListModel)
	
	if newModel.err == nil {
		t.Error("expected error to be set")
	}
	
	if newModel.err.Error() != "test error" {
		t.Errorf("expected error message 'test error', got %s", newModel.err.Error())
	}
	
	if newModel.loading {
		t.Errorf("expected loading false, got %v", newModel.loading)
	}
	
	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}
}

func TestPromptListModel_Update_WindowSizeMsg(t *testing.T) {
	model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
	
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(msg)
	
	// Should return model and no command
	newModel := updatedModel.(PromptListModel)
	if len(newModel.prompts) != len(model.prompts) {
		t.Error("expected prompts to remain unchanged for window size message")
	}
	
	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}
}

func TestPromptListModel_EnterKeyWithEmptyList(t *testing.T) {
	model := NewPromptListModel([]model.Prompt{}, ListAll, "")
	
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(keyMsg)
	newModel := updatedModel.(PromptListModel)
	
	// Should not be done or have selection with empty list
	if newModel.done {
		t.Error("expected not done with empty list")
	}
	
	if newModel.selected != nil {
		t.Error("expected no selection with empty list")
	}
	
	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}
}

func TestPromptListModel_View_LoadingState(t *testing.T) {
	model := NewLoadingPromptListModel()
	view := model.View()
	
	if view == "" {
		t.Error("expected non-empty view")
	}
	
	// Should contain loading indicator
	if !contains(view, "🔄") {
		t.Error("expected loading indicator in view")
	}
}

func TestPromptListModel_View_ErrorState(t *testing.T) {
	model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
	model.err = fmt.Errorf("test error")
	
	view := model.View()
	
	if view == "" {
		t.Error("expected non-empty view")
	}
	
	// Should contain error indicator
	if !contains(view, "❌") {
		t.Error("expected error indicator in view")
	}
	
	// Should contain error message
	if !contains(view, "test error") {
		t.Error("expected error message in view")
	}
}

func TestPromptListModel_View_EmptyList(t *testing.T) {
	tests := []struct {
		name           string
		mode           ListMode
		filter         string
		expectedText   string
	}{
		{
			name:         "empty list all mode",
			mode:         ListAll,
			filter:       "",
			expectedText: "📝",
		},
		{
			name:         "empty list filtered mode",
			mode:         ListFiltered,
			filter:       "golang",
			expectedText: "🔍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewPromptListModel([]model.Prompt{}, tt.mode, tt.filter)
			view := model.View()
			
			if view == "" {
				t.Error("expected non-empty view")
			}
			
			if !contains(view, tt.expectedText) {
				t.Errorf("expected view to contain %s", tt.expectedText)
			}
		})
	}
}

func TestPromptListModel_View_ListState(t *testing.T) {
	model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
	view := model.View()
	
	if view == "" {
		t.Error("expected non-empty view")
	}
	
	// Should contain title
	if !contains(view, "🗑️") {
		t.Error("expected delete icon in view")
	}
	
	// Should contain prompt names
	if !contains(view, TestPromptName1) {
		t.Errorf("expected view to contain %s", TestPromptName1)
	}
	
	if !contains(view, TestPromptName2) {
		t.Errorf("expected view to contain %s", TestPromptName2)
	}
	
	// Should contain authors
	if !contains(view, TestAuthor1) {
		t.Errorf("expected view to contain %s", TestAuthor1)
	}
	
	// Should contain help text
	if !contains(view, HelpTextListNavigation) {
		t.Error("expected help text in view")
	}
}

func TestPromptListModel_GettersAndSetters(t *testing.T) {
	model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
	
	// Test initial state
	if model.GetSelected() != nil {
		t.Error("expected no initial selection")
	}
	
	if model.IsDone() {
		t.Error("expected not done initially")
	}
	
	if model.HasQuit() {
		t.Error("expected not quit initially")
	}
	
	if model.GetError() != nil {
		t.Error("expected no initial error")
	}
	
	// Test after selection
	model.selected = &CreateStandardTestPrompts()[0]
	model.done = true
	
	if model.GetSelected() == nil {
		t.Error("expected selection after setting")
	}
	
	if !model.IsDone() {
		t.Error("expected done after setting")
	}
	
	// Test after quit
	model.quit = true
	if !model.HasQuit() {
		t.Error("expected quit after setting")
	}
	
	// Test after error
	testError := fmt.Errorf("test error")
	model.err = testError
	if model.GetError() != testError {
		t.Error("expected error after setting")
	}
}

func TestPromptListModel_SetPrompts(t *testing.T) {
	model := NewPromptListModel([]model.Prompt{}, ListAll, "")
	testPrompts := CreateStandardTestPrompts()
	
	cmd := model.SetPrompts(testPrompts, ListFiltered, "test")
	
	if cmd == nil {
		t.Error("expected command from SetPrompts")
	}
	
	// Execute the returned command to get the message
	msg := cmd()
	loadMsg, ok := msg.(promptsLoadedMsg)
	if !ok {
		t.Error("expected promptsLoadedMsg from command")
	}
	
	if len(loadMsg.prompts) != len(testPrompts) {
		t.Errorf("expected %d prompts, got %d", len(testPrompts), len(loadMsg.prompts))
	}
	
	if loadMsg.mode != ListFiltered {
		t.Errorf("expected mode ListFiltered, got %v", loadMsg.mode)
	}
	
	if loadMsg.filter != "test" {
		t.Errorf("expected filter 'test', got %s", loadMsg.filter)
	}
}

func TestPromptListModel_SetError(t *testing.T) {
	model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
	testError := fmt.Errorf("test error")
	
	cmd := model.SetError(testError)
	
	if cmd == nil {
		t.Error("expected command from SetError")
	}
	
	// Execute the returned command to get the message
	msg := cmd()
	errorMsg, ok := msg.(promptLoadErrorMsg)
	if !ok {
		t.Error("expected promptLoadErrorMsg from command")
	}
	
	if errorMsg.err != testError {
		t.Error("expected same error from command")
	}
}

func TestPromptListModel_RenderListItem(t *testing.T) {
	model := NewPromptListModel(CreateStandardTestPrompts(), ListAll, "")
	testPrompt := CreateSingleTestPrompt()
	
	// Test selected item
	selectedItem := model.renderListItem(0, testPrompt, true)
	if !contains(selectedItem, ">") {
		t.Error("expected selected item to contain selection indicator")
	}
	
	if !contains(selectedItem, testPrompt.Name) {
		t.Error("expected item to contain prompt name")
	}
	
	if !contains(selectedItem, testPrompt.Author) {
		t.Error("expected item to contain author")
	}
	
	// Test non-selected item
	normalItem := model.renderListItem(0, testPrompt, false)
	if contains(normalItem, ">") {
		t.Error("expected normal item not to contain selection indicator")
	}
	
	if !contains(normalItem, testPrompt.Name) {
		t.Error("expected item to contain prompt name")
	}
}

func TestPromptListModel_TruncateDisplay(t *testing.T) {
	// Create prompt with very long name and author
	longPrompt := model.Prompt{
		ID:      "test",
		Name:    "This is a very long prompt name that should be truncated when displayed in the list",
		Author:  "VeryLongAuthorNameThatShouldBeTruncated",
		GistURL: "https://gist.github.com/test/123",
	}
	
	listModel := NewPromptListModel([]model.Prompt{longPrompt}, ListAll, "")
	item := listModel.renderListItem(0, longPrompt, false)
	
	// Should contain truncation indicator
	if !contains(item, "...") {
		t.Error("expected truncated text to contain '...'")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}