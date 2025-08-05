package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grigri/pv/internal/model"
)

// BubbleTeaTUI implements the TUIInterface using the bubbletea framework.
// It serves as a factory and integration point for all TUI components,
// handling program startup, cleanup, and error management.
type BubbleTeaTUI struct {
	// altScreen determines whether to use alternative screen buffer
	// This is useful for full-screen TUI applications
	altScreen bool
	
	// mouseEnabled determines whether mouse input is enabled
	mouseEnabled bool
}

// NewBubbleTeaTUI creates a new instance of BubbleTeaTUI with default settings.
// By default, it uses alternative screen and disables mouse input for better
// compatibility with terminal environments.
func NewBubbleTeaTUI() *BubbleTeaTUI {
	return &BubbleTeaTUI{
		altScreen:    true,
		mouseEnabled: false,
	}
}

// NewBubbleTeaTUIWithOptions creates a new BubbleTeaTUI with custom options.
// This allows fine-tuning the TUI behavior for different use cases.
func NewBubbleTeaTUIWithOptions(altScreen, mouseEnabled bool) *BubbleTeaTUI {
	return &BubbleTeaTUI{
		altScreen:    altScreen,
		mouseEnabled: mouseEnabled,
	}
}

// ShowPromptList displays a list of prompts in an interactive interface
// and returns the user-selected prompt. This method implements the core
// list selection functionality for the delete command.
func (tui *BubbleTeaTUI) ShowPromptList(prompts []model.Prompt) (model.Prompt, error) {
	// Handle empty prompt list
	if len(prompts) == 0 {
		return model.Prompt{}, fmt.Errorf(ErrMsgNoPromptsFound)
	}

	// Create the prompt list model
	listModel := NewPromptListModel(prompts, ListAll, "")

	// Configure program options
	var options []tea.ProgramOption
	if tui.altScreen {
		options = append(options, tea.WithAltScreen())
	}
	if tui.mouseEnabled {
		options = append(options, tea.WithMouseCellMotion())
	}

	// Create and run the bubbletea program
	program := tea.NewProgram(listModel, options...)
	
	finalModel, err := program.Run()
	if err != nil {
		return model.Prompt{}, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
	}

	// Extract the final state from the model
	if promptListModel, ok := finalModel.(PromptListModel); ok {
		// Check if user quit without selection
		if promptListModel.HasQuit() {
			return model.Prompt{}, fmt.Errorf(ErrMsgUserCancelled)
		}

		// Check for errors during execution
		if err := promptListModel.GetError(); err != nil {
			return model.Prompt{}, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
		}

		// Check if a selection was made
		if selected := promptListModel.GetSelected(); selected != nil {
			return *selected, nil
		}

		// No selection made (should not happen if not quit)
		return model.Prompt{}, fmt.Errorf(ErrMsgInvalidSelection)
	}

	// Model type assertion failed
	return model.Prompt{}, fmt.Errorf("%s: invalid model type", ErrMsgTUIInitFailed)
}

// ShowPromptListFiltered displays a filtered list of prompts and returns
// the user-selected prompt. This method is used when showing search results.
func (tui *BubbleTeaTUI) ShowPromptListFiltered(prompts []model.Prompt, filter string) (model.Prompt, error) {
	// Handle empty filtered results
	if len(prompts) == 0 {
		return model.Prompt{}, fmt.Errorf("没有找到包含 \"%s\" 的提示", filter)
	}

	// Create the prompt list model with filtered mode
	listModel := NewPromptListModel(prompts, ListFiltered, filter)

	// Configure program options
	var options []tea.ProgramOption
	if tui.altScreen {
		options = append(options, tea.WithAltScreen())
	}
	if tui.mouseEnabled {
		options = append(options, tea.WithMouseCellMotion())
	}

	// Create and run the bubbletea program
	program := tea.NewProgram(listModel, options...)
	
	finalModel, err := program.Run()
	if err != nil {
		return model.Prompt{}, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
	}

	// Extract the final state from the model
	if promptListModel, ok := finalModel.(PromptListModel); ok {
		// Check if user quit without selection
		if promptListModel.HasQuit() {
			return model.Prompt{}, fmt.Errorf(ErrMsgUserCancelled)
		}

		// Check for errors during execution
		if err := promptListModel.GetError(); err != nil {
			return model.Prompt{}, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
		}

		// Check if a selection was made
		if selected := promptListModel.GetSelected(); selected != nil {
			return *selected, nil
		}

		// No selection made (should not happen if not quit)
		return model.Prompt{}, fmt.Errorf(ErrMsgInvalidSelection)
	}

	// Model type assertion failed
	return model.Prompt{}, fmt.Errorf("%s: invalid model type", ErrMsgTUIInitFailed)
}

// ShowConfirm displays a confirmation dialog for the given prompt
// and returns true if the user confirms the deletion, false if cancelled.
// This method implements the confirmation step required before deletion.
func (tui *BubbleTeaTUI) ShowConfirm(prompt model.Prompt) (bool, error) {
	// Create the confirmation model
	confirmModel := NewConfirmModel(prompt)

	// Configure program options
	var options []tea.ProgramOption
	if tui.altScreen {
		options = append(options, tea.WithAltScreen())
	}
	if tui.mouseEnabled {
		options = append(options, tea.WithMouseCellMotion())
	}

	// Create and run the bubbletea program
	program := tea.NewProgram(confirmModel, options...)
	
	finalModel, err := program.Run()
	if err != nil {
		return false, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
	}

	// Extract the final state from the model
	if confirmModelFinal, ok := finalModel.(ConfirmModel); ok {
		// Check for errors during execution
		if confirmModelFinal.HasError() {
			return false, fmt.Errorf("%s: confirmation dialog error", ErrMsgTUIRenderFailed)
		}

		// Check if user cancelled
		if confirmModelFinal.IsCancelled() {
			return false, fmt.Errorf(ErrMsgUserCancelled)
		}

		// Return confirmation result
		return confirmModelFinal.IsConfirmed(), nil
	}

	// Model type assertion failed
	return false, fmt.Errorf("%s: invalid model type", ErrMsgTUIInitFailed)
}

// ShowVariableForm displays a form for collecting variable values from the user.
// This method implements the variable input functionality for the get command.
func (tui *BubbleTeaTUI) ShowVariableForm(variables []string) (map[string]string, error) {
	// Handle empty variable list
	if len(variables) == 0 {
		return make(map[string]string), nil
	}

	// Create the variable form model
	formModel := NewVariableFormModel(variables)

	// Configure program options
	var options []tea.ProgramOption
	if tui.altScreen {
		options = append(options, tea.WithAltScreen())
	}
	if tui.mouseEnabled {
		options = append(options, tea.WithMouseCellMotion())
	}

	// Create and run the bubbletea program
	program := tea.NewProgram(formModel, options...)
	
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
	}

	// Extract the final state from the model
	if variableFormModel, ok := finalModel.(VariableFormModel); ok {
		// Check if user cancelled
		if variableFormModel.IsCancelled() {
			return nil, fmt.Errorf(ErrMsgUserCancelled)
		}

		// Check for errors during execution
		if err := variableFormModel.GetError(); err != nil {
			return nil, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
		}

		// Check if form was completed
		if variableFormModel.IsDone() {
			return variableFormModel.GetValues(), nil
		}

		// Form was not completed (should not happen if not cancelled)
		return nil, fmt.Errorf(ErrMsgInvalidSelection)
	}

	// Model type assertion failed
	return nil, fmt.Errorf("%s: invalid model type", ErrMsgTUIInitFailed)
}

// ShowError displays an error message to the user using the ErrorModel.
// This is a helper method for displaying errors in a consistent TUI format.
func (tui *BubbleTeaTUI) ShowError(err error) error {
	// Create the error model
	errorModel := NewErrorModel(err)

	// Configure program options
	var options []tea.ProgramOption
	if tui.altScreen {
		options = append(options, tea.WithAltScreen())
	}
	if tui.mouseEnabled {
		options = append(options, tea.WithMouseCellMotion())
	}

	// Create and run the bubbletea program
	program := tea.NewProgram(errorModel, options...)
	
	finalModel, runErr := program.Run()
	if runErr != nil {
		// If we can't even show the error, return the original error
		return fmt.Errorf("%s: %w (original error: %v)", ErrMsgTUIRenderFailed, runErr, err)
	}

	// Check if the error display was properly terminated
	if errorModelFinal, ok := finalModel.(ErrorModel); ok {
		if errorModelFinal.IsTerminated() {
			return nil // Error was displayed successfully
		}
	}

	// If we reach here, something went wrong with the error display
	return fmt.Errorf("%s: error display was not properly terminated", ErrMsgTUIRenderFailed)
}

// StartProgram is a low-level method that starts a bubbletea program
// with the given model and returns the final model state.
// This method provides direct access to bubbletea program execution
// for advanced use cases or testing.
func (tui *BubbleTeaTUI) StartProgram(model tea.Model) (tea.Model, error) {
	// Configure program options
	var options []tea.ProgramOption
	if tui.altScreen {
		options = append(options, tea.WithAltScreen())
	}
	if tui.mouseEnabled {
		options = append(options, tea.WithMouseCellMotion())
	}

	// Create and run the bubbletea program
	program := tea.NewProgram(model, options...)
	
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrMsgTUIRenderFailed, err)
	}

	return finalModel, nil
}

// SetAltScreen enables or disables the alternative screen buffer.
// When enabled, the TUI will take over the entire terminal screen.
// When disabled, the TUI will render inline with existing terminal content.
func (tui *BubbleTeaTUI) SetAltScreen(enabled bool) {
	tui.altScreen = enabled
}

// SetMouseEnabled enables or disables mouse input for the TUI.
// Mouse input can be useful for clicking on items but may not work
// in all terminal environments.
func (tui *BubbleTeaTUI) SetMouseEnabled(enabled bool) {
	tui.mouseEnabled = enabled
}

// IsAltScreenEnabled returns whether alternative screen is enabled.
func (tui *BubbleTeaTUI) IsAltScreenEnabled() bool {
	return tui.altScreen
}

// IsMouseEnabled returns whether mouse input is enabled.
func (tui *BubbleTeaTUI) IsMouseEnabled() bool {
	return tui.mouseEnabled
}

// Cleanup performs any necessary cleanup operations.
// Currently this is a no-op but is provided for future extensibility
// and to maintain a consistent interface.
func (tui *BubbleTeaTUI) Cleanup() {
	// Currently no cleanup is needed for bubbletea programs
	// as they handle their own cleanup automatically.
	// This method is provided for future extensibility.
}