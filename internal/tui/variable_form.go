package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VariableFormModel represents the bubbletea model for collecting variable values
type VariableFormModel struct {
	variables    []string
	inputs       []textinput.Model
	currentField int
	values       map[string]string
	done         bool
	cancelled    bool
	err          error

	// Styles
	focusedStyle   lipgloss.Style
	blurredStyle   lipgloss.Style
	helpStyle      lipgloss.Style
	titleStyle     lipgloss.Style
	errorStyle     lipgloss.Style
	containerStyle lipgloss.Style
}

// NewVariableFormModel creates a new variable form model
func NewVariableFormModel(variables []string) VariableFormModel {
	inputs := make([]textinput.Model, len(variables))
	values := make(map[string]string)

	// Initialize text inputs
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = fmt.Sprintf("输入 %s 的值", variables[i])
		inputs[i].CharLimit = 200
		inputs[i].Width = VariableInputWidth

		if i == 0 {
			inputs[i].Focus()
		}

		values[variables[i]] = ""
	}

	return VariableFormModel{
		variables:    variables,
		inputs:       inputs,
		currentField: 0,
		values:       values,
		done:         false,
		cancelled:    false,
		err:          nil,

		focusedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true),

		blurredStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSecondary)),

		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			Margin(1, 0),

		titleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true).
			Margin(0, 0, 1, 0),

		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true),

		containerStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(1, 2).
			Width(VariableFormWidth),
	}
}

// Init implements the tea.Model interface
func (m VariableFormModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements the tea.Model interface
func (m VariableFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done || m.cancelled {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyCtrlC, KeyEscape, KeyQuit:
			m.cancelled = true
			return m, tea.Quit

		case KeyEnter:
			return m.handleEnter()

		case KeyTab, KeyDown:
			return m.nextField()

		case KeyShiftTab, KeyUp:
			return m.prevField()
		}
	}

	// Update current input field
	var cmd tea.Cmd
	m.inputs[m.currentField], cmd = m.inputs[m.currentField].Update(msg)
	m.values[m.variables[m.currentField]] = m.inputs[m.currentField].Value()

	return m, cmd
}

// View implements the tea.Model interface
func (m VariableFormModel) View() string {
	if m.err != nil {
		return m.errorStyle.Render(fmt.Sprintf("错误: %v", m.err))
	}

	var content strings.Builder

	// Title
	content.WriteString(m.titleStyle.Render("请填写提示词中的变量"))
	content.WriteString("\n\n")

	// Variable input fields
	for i, variable := range m.variables {
		style := m.blurredStyle
		indicator := "  "
		if i == m.currentField {
			style = m.focusedStyle
			indicator = "▶ "
		}

		// Variable label
		label := fmt.Sprintf("%s%s:", indicator, variable)
		if len(variable) > MaxVariableNameLength {
			label = fmt.Sprintf("%s%s...:", indicator, variable[:MaxVariableNameLength-3])
		}

		content.WriteString(style.Render(label))
		content.WriteString("\n")
		content.WriteString(m.inputs[i].View())
		content.WriteString("\n\n")
	}

	// Help text
	content.WriteString(m.helpStyle.Render(HelpTextVariableForm))

	return m.containerStyle.Render(content.String())
}

// handleEnter processes the Enter key press
func (m VariableFormModel) handleEnter() (VariableFormModel, tea.Cmd) {
	// Validate that all fields have values
	for i, variable := range m.variables {
		value := strings.TrimSpace(m.inputs[i].Value())
		if value == "" {
			m.err = fmt.Errorf("变量 '%s' 不能为空", variable)
			return m, nil
		}
		m.values[variable] = value
	}

	// All validation passed
	m.done = true
	return m, tea.Quit
}

// nextField moves to the next input field
func (m VariableFormModel) nextField() (VariableFormModel, tea.Cmd) {
	m.inputs[m.currentField].Blur()
	m.currentField = (m.currentField + 1) % len(m.inputs)
	return m.focusCurrentField()
}

// prevField moves to the previous input field
func (m VariableFormModel) prevField() (VariableFormModel, tea.Cmd) {
	m.inputs[m.currentField].Blur()
	m.currentField = (m.currentField - 1 + len(m.inputs)) % len(m.inputs)
	return m.focusCurrentField()
}

// focusCurrentField focuses the current input field
func (m VariableFormModel) focusCurrentField() (VariableFormModel, tea.Cmd) {
	cmd := m.inputs[m.currentField].Focus()
	return m, cmd
}

// IsDone returns true if the form is completed successfully
func (m VariableFormModel) IsDone() bool {
	return m.done
}

// IsCancelled returns true if the form was cancelled
func (m VariableFormModel) IsCancelled() bool {
	return m.cancelled
}

// GetValues returns the collected variable values
func (m VariableFormModel) GetValues() map[string]string {
	if !m.done {
		return nil
	}
	return m.values
}

// GetError returns any error that occurred
func (m VariableFormModel) GetError() error {
	return m.err
}