package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a single form field
type FormField struct {
	Name     string
	Value    string
	Required bool
}

// FormModel represents a multi-field form
type FormModel struct {
	Title        string
	Fields       []FormField
	CurrentField int
	Submitted    bool
	ShowError    bool
	ErrorMessage string
}

// NewForm creates a new form model
func NewForm(title string, variables []string) FormModel {
	fields := make([]FormField, len(variables))
	for i, v := range variables {
		fields[i] = FormField{
			Name:     v,
			Value:    "",
			Required: true,
		}
	}

	return FormModel{
		Title:        title,
		Fields:       fields,
		CurrentField: 0,
		Submitted:    false,
		ShowError:    false,
		ErrorMessage: "",
	}
}

// Init initializes the model
func (m FormModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if m.CurrentField > 0 {
				m.CurrentField--
			}
			m.ShowError = false
		case "down", "tab":
			if m.CurrentField < len(m.Fields)-1 {
				m.CurrentField++
			}
			m.ShowError = false
		case "enter":
			// Check if on the last field
			if m.CurrentField == len(m.Fields)-1 {
				// Validate all required fields
				for i, field := range m.Fields {
					if field.Required && strings.TrimSpace(field.Value) == "" {
						m.ShowError = true
						m.ErrorMessage = fmt.Sprintf("Field '%s' is required", field.Name)
						m.CurrentField = i
						return m, nil
					}
				}
				// All valid, submit
				m.Submitted = true
				return m, tea.Quit
			} else {
				// Move to next field
				m.CurrentField++
			}
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "backspace":
			if len(m.Fields[m.CurrentField].Value) > 0 {
				// Convert to runes to handle multi-byte characters properly
				runes := []rune(m.Fields[m.CurrentField].Value)
				if len(runes) > 0 {
					m.Fields[m.CurrentField].Value = string(runes[:len(runes)-1])
				}
			}
		default:
			// Add character to current field
			// Handle both single-byte and multi-byte characters (including Chinese)
			if msg.Type == tea.KeyRunes {
				// This handles multi-byte characters properly
				m.Fields[m.CurrentField].Value += string(msg.Runes)
			} else if len(msg.String()) > 0 && !strings.HasPrefix(msg.String(), "ctrl+") && !strings.HasPrefix(msg.String(), "alt+") {
				// Fallback for other key types
				m.Fields[m.CurrentField].Value += msg.String()
			}
		}
	}

	return m, nil
}

// View renders the form
func (m FormModel) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render(m.Title))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString("Fill in the variables below:\n\n")

	// Fields
	for i, field := range m.Fields {
		// Field name
		fieldStyle := lipgloss.NewStyle()
		if i == m.CurrentField {
			fieldStyle = fieldStyle.
				Foreground(lipgloss.Color("86")).
				Bold(true)
		} else {
			fieldStyle = fieldStyle.Foreground(lipgloss.Color("246"))
		}

		b.WriteString(fieldStyle.Render(field.Name))
		if field.Required {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Render("*"))
		}
		b.WriteString(": ")

		// Field value
		valueStyle := lipgloss.NewStyle()
		if i == m.CurrentField {
			valueStyle = valueStyle.Foreground(lipgloss.Color("230"))
		} else {
			valueStyle = valueStyle.Foreground(lipgloss.Color("250"))
		}

		value := field.Value
		if value == "" && i != m.CurrentField {
			value = "(empty)"
			valueStyle = valueStyle.Foreground(lipgloss.Color("241"))
		}
		b.WriteString(valueStyle.Render(value))

		// Cursor for current field
		if i == m.CurrentField {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Blink(true).
				Render("█"))
		}

		b.WriteString("\n")
	}

	// Error message
	if m.ShowError {
		b.WriteString("\n")
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString(errorStyle.Render("Error: "))
		b.WriteString(errorStyle.Render(m.ErrorMessage))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	help := []string{
		"↑/↓ or tab: navigate",
		"enter: next/submit",
		"esc: cancel",
	}
	b.WriteString(helpStyle.Render(strings.Join(help, " • ")))

	return b.String()
}

// GetValues returns a map of field names to values
func (m *FormModel) GetValues() map[string]string {
	values := make(map[string]string)
	for _, field := range m.Fields {
		values[field.Name] = field.Value
	}
	return values
}

// IsSubmitted returns whether the form was submitted
func (m *FormModel) IsSubmitted() bool {
	return m.Submitted
}

// SetFieldValue sets the value of a specific field
func (m *FormModel) SetFieldValue(name, value string) {
	for i := range m.Fields {
		if m.Fields[i].Name == name {
			m.Fields[i].Value = value
			break
		}
	}
}

// Reset clears all field values and resets state
func (m *FormModel) Reset() {
	for i := range m.Fields {
		m.Fields[i].Value = ""
	}
	m.CurrentField = 0
	m.Submitted = false
	m.ShowError = false
	m.ErrorMessage = ""
}
