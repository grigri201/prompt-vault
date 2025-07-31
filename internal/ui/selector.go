package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// SelectorModel represents an interactive prompt selector
type SelectorModel struct {
	Items        []string
	Selected     int
	Input        string
	Confirmed    bool
	ShowError    bool
	ErrorMessage string
	MaxSelection int
}

// NewSelector creates a new selector model
func NewSelector(items []string) SelectorModel {
	return SelectorModel{
		Items:        items,
		Selected:     -1,
		Input:        "",
		Confirmed:    false,
		ShowError:    false,
		ErrorMessage: "",
		MaxSelection: len(items),
	}
}

// Init initializes the model
func (m SelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Try to parse the input as a number
			if m.Input != "" {
				num, err := strconv.Atoi(m.Input)
				if err != nil {
					m.ShowError = true
					m.ErrorMessage = "Please enter a valid number"
					return m, nil
				}

				// Check if number is in valid range
				if num < 1 || num > m.MaxSelection {
					m.ShowError = true
					m.ErrorMessage = fmt.Sprintf("Please enter a number between 1 and %d", m.MaxSelection)
					return m, nil
				}

				// Valid selection
				m.Selected = num - 1 // Convert to 0-based index
				m.Confirmed = true
				m.ShowError = false
				return m, tea.Quit
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		case "backspace":
			if len(m.Input) > 0 {
				m.Input = m.Input[:len(m.Input)-1]
				m.ShowError = false
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.Input += msg.String()
			m.ShowError = false
		}
	}

	return m, nil
}

// View renders the selector
func (m SelectorModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Select a prompt"))
	b.WriteString("\n\n")

	// Show items (up to 10 for visibility)
	displayItems := m.Items
	if len(displayItems) > 10 {
		displayItems = displayItems[:10]
		b.WriteString(fmt.Sprintf("Showing first 10 of %d items:\n", len(m.Items)))
	}

	for i, item := range displayItems {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}

	if len(m.Items) > 10 {
		b.WriteString(fmt.Sprintf("... and %d more\n", len(m.Items)-10))
	}

	b.WriteString("\n")

	// Input prompt
	b.WriteString("Enter selection (1-")
	b.WriteString(fmt.Sprintf("%d): ", m.MaxSelection))

	// Show current input
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	b.WriteString(inputStyle.Render(m.Input))

	// Cursor
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Blink(true).
		Render("█"))

	// Error message
	if m.ShowError {
		b.WriteString("\n\n")
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString(errorStyle.Render("Error: "))
		b.WriteString(errorStyle.Render(m.ErrorMessage))
	}

	// Help
	b.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(helpStyle.Render("enter: confirm • ctrl+c: cancel"))

	return b.String()
}

// GetSelection returns the selected item or empty string if none selected
func (m *SelectorModel) GetSelection() string {
	if m.Selected >= 0 && m.Selected < len(m.Items) {
		return m.Items[m.Selected]
	}
	return ""
}

// IsConfirmed returns whether a selection was confirmed
func (m *SelectorModel) IsConfirmed() bool {
	return m.Confirmed
}

// Reset clears the selection and input
func (m *SelectorModel) Reset() {
	m.Selected = -1
	m.Input = ""
	m.Confirmed = false
	m.ShowError = false
	m.ErrorMessage = ""
}
