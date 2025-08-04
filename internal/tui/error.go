package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grigri/pv/internal/errors"
)

// ErrorModel represents the TUI model for displaying error messages
// with appropriate styling and user interaction capabilities.
type ErrorModel struct {
	err         error
	message     string
	errorType   errors.ErrorType
	terminated  bool
	width       int
	height      int
}

// NewErrorModel creates a new ErrorModel with the given error.
// It automatically determines the error type and formats appropriate messages.
func NewErrorModel(err error) ErrorModel {
	var errorType errors.ErrorType
	var message string

	if appErr, ok := err.(*errors.AppError); ok {
		errorType = appErr.Type
		message = appErr.Message
	} else {
		errorType = errors.ErrUnknown
		message = err.Error()
	}

	return ErrorModel{
		err:       err,
		message:   message,
		errorType: errorType,
		width:     MinScreenWidth,
		height:    MinScreenHeight,
	}
}

// Init implements the bubbletea Model interface
func (m ErrorModel) Init() tea.Cmd {
	return nil
}

// Update implements the bubbletea Model interface
// Handles keyboard input for continuing after error display
func (m ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Any key press will terminate the error display
		switch msg.String() {
		case KeyCtrlC:
			m.terminated = true
			return m, tea.Quit
		default:
			// Any other key continues (exits the error display)
			m.terminated = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View implements the bubbletea Model interface
// Renders the error display with appropriate styling and icons
func (m ErrorModel) View() string {
	if m.width < MinScreenWidth {
		m.width = MinScreenWidth
	}

	// Get error icon and color based on error type
	icon, color := m.getErrorIconAndColor()

	// Create the main content box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(1, 2).
		Width(m.width - 4).
		Align(lipgloss.Center)

	// Create the title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Align(lipgloss.Center)

	// Create the message style
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Align(lipgloss.Center).
		Margin(1, 0)

	// Create the help text style
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Align(lipgloss.Center).
		Margin(1, 0)

	// Build the content
	title := titleStyle.Render(fmt.Sprintf("%s %s", icon, m.getErrorTitle()))
	message := messageStyle.Render(m.wrapMessage(m.message, m.width-8))
	help := helpStyle.Render("按任意键继续...")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		message,
		help,
	)

	box := boxStyle.Render(content)

	// Center the box vertically on screen
	if m.height > 0 {
		verticalPadding := (m.height - lipgloss.Height(box)) / 2
		if verticalPadding > 0 {
			box = lipgloss.NewStyle().
				Margin(verticalPadding, 0).
				Render(box)
		}
	}

	return box
}

// getErrorIconAndColor returns the appropriate icon and color for the error type
func (m ErrorModel) getErrorIconAndColor() (string, string) {
	switch m.errorType {
	case errors.ErrValidation:
		return "⚠️", ColorWarning
	case errors.ErrAuth:
		return "🔒", ColorError
	case errors.ErrStorage:
		return "💾", ColorError
	case errors.ErrNetwork:
		return "🌐", ColorError
	case errors.ErrPermission:
		return "🔐", ColorError
	default:
		return "❌", ColorError
	}
}

// getErrorTitle returns the appropriate title for the error type
func (m ErrorModel) getErrorTitle() string {
	switch m.errorType {
	case errors.ErrValidation:
		return "输入验证错误"
	case errors.ErrAuth:
		return "身份验证错误"
	case errors.ErrStorage:
		return "存储错误"
	case errors.ErrNetwork:
		return "网络错误"
	case errors.ErrPermission:
		return "权限错误"
	default:
		return "操作失败"
	}
}

// wrapMessage wraps the error message to fit within the specified width
func (m ErrorModel) wrapMessage(message string, maxWidth int) string {
	if len(message) <= maxWidth {
		return message
	}

	var lines []string
	words := strings.Fields(message)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxWidth {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine = currentLine + " " + word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// IsTerminated returns true if the error display has been terminated by user input
func (m ErrorModel) IsTerminated() bool {
	return m.terminated
}

// ShowError is a convenience function to display an error using bubbletea
// It blocks until the user presses any key to continue
func ShowError(err error) error {
	model := NewErrorModel(err)
	
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, runErr := program.Run()
	
	if runErr != nil {
		return fmt.Errorf("failed to run error display: %w", runErr)
	}

	// Check if the user terminated properly
	if errorModel, ok := finalModel.(ErrorModel); ok && errorModel.IsTerminated() {
		return nil
	}

	return fmt.Errorf("error display was not properly terminated")
}