package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grigri/pv/internal/model"
)

// ConfirmModel represents the state of the confirmation dialog
type ConfirmModel struct {
	prompt      model.Prompt
	confirmed   bool
	cancelled   bool
	err         error
	width       int
	height      int
}

// NewConfirmModel creates a new confirmation model for the given prompt
func NewConfirmModel(prompt model.Prompt) ConfirmModel {
	return ConfirmModel{
		prompt:    prompt,
		confirmed: false,
		cancelled: false,
		width:     60, // ConfirmDialogWidth from interface.go
		height:    10, // ConfirmDialogHeight from interface.go
	}
}

// Init initializes the confirmation model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model state
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "y": // KeyYes from interface.go
			m.confirmed = true
			return m, tea.Quit

		case "n", "esc", "q": // KeyNo, KeyEscape, KeyQuit from interface.go
			m.cancelled = true
			return m, tea.Quit

		case "ctrl+c": // KeyCtrlC from interface.go
			m.cancelled = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the confirmation dialog
func (m ConfirmModel) View() string {
	if m.err != nil {
		return m.renderError()
	}

	return m.renderConfirmDialog()
}

// renderConfirmDialog renders the main confirmation interface
func (m ConfirmModel) renderConfirmDialog() string {
	// Define styles using constants from interface.go
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")). // ColorWarning
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width - 4)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")). // ColorText
		Width(m.width - 4).
		Align(lipgloss.Left).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")). // ColorPrimary
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")) // ColorText

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")). // ColorWarning
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width - 4).
		Padding(1, 0)

	confirmButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")). // ColorError
		Background(lipgloss.Color("#550000")).
		Bold(true).
		Padding(0, 2).
		Margin(0, 1)

	cancelButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C7C7C")). // ColorSecondary
		Background(lipgloss.Color("#333333")).
		Bold(true).
		Padding(0, 2).
		Margin(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")). // ColorMuted
		Align(lipgloss.Center).
		Width(m.width - 4)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")). // ColorBorder
		Width(m.width - 2).
		Align(lipgloss.Center)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("⚠️  确认删除提示"))
	content.WriteString("\n\n")

	// Prompt details
	content.WriteString(contentStyle.Render(
		fmt.Sprintf("%s %s\n", 
			labelStyle.Render("提示名称:"), 
			valueStyle.Render(m.truncateString(m.prompt.Name, 40))))) // MaxPromptNameLength

	content.WriteString(contentStyle.Render(
		fmt.Sprintf("%s %s\n", 
			labelStyle.Render("作者:"), 
			valueStyle.Render(m.truncateString(m.prompt.Author, 20))))) // MaxAuthorNameLength

	if m.prompt.Description != "" {
		content.WriteString(contentStyle.Render(
			fmt.Sprintf("%s %s\n", 
				labelStyle.Render("描述:"), 
				valueStyle.Render(m.truncateString(m.prompt.Description, 50)))))
	}

	content.WriteString(contentStyle.Render(
		fmt.Sprintf("%s %s\n", 
			labelStyle.Render("Gist URL:"), 
			valueStyle.Render(m.truncateString(m.prompt.GistURL, 60)))))

	content.WriteString("\n")

	// Warning message
	content.WriteString(warningStyle.Render("此操作不可撤销，确定要删除这个提示吗？"))
	content.WriteString("\n\n")

	// Buttons
	buttonsLine := confirmButtonStyle.Render("[Y] 确认删除") + 
		"    " + 
		cancelButtonStyle.Render("[N] 取消")
	
	content.WriteString(lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width - 4).
		Render(buttonsLine))
	content.WriteString("\n\n")

	// Help text
	content.WriteString(helpStyle.Render("Y: 确认  N: 取消  Esc: 取消")) // HelpTextConfirmation

	// Wrap in border
	return borderStyle.Render(content.String())
}

// renderError renders an error state
func (m ConfirmModel) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")). // ColorError
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width - 4).
		Padding(2)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF0000")). // ColorError
		Width(m.width - 2).
		Align(lipgloss.Center)

	content := fmt.Sprintf("❌ 确认对话框错误\n\n%s\n\n按任意键继续...", m.err.Error())
	
	return borderStyle.Render(errorStyle.Render(content))
}

// truncateString truncates a string to the specified length with ellipsis
func (m ConfirmModel) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	if maxLen <= 3 {
		return s[:maxLen]
	}
	
	return s[:maxLen-3] + "..."
}

// IsConfirmed returns true if the user confirmed the deletion
func (m ConfirmModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user cancelled the operation
func (m ConfirmModel) IsCancelled() bool {
	return m.cancelled
}

// HasError returns true if there's an error state
func (m ConfirmModel) HasError() bool {
	return m.err != nil
}

// SetError sets an error state for the model
func (m *ConfirmModel) SetError(err error) {
	m.err = err
}