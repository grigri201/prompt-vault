package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grigri201/prompt-vault/internal/models"
)

// DuplicateChoice represents the user's choice when a duplicate is found
type DuplicateChoice int

const (
	// UpdateExisting indicates the user wants to update the existing prompt
	UpdateExisting DuplicateChoice = iota
	// CreateNew indicates the user wants to create a new prompt with a different ID
	CreateNew
	// Cancel indicates the user wants to cancel the operation
	Cancel
)

// DuplicateHandlerModel handles duplicate prompt interactions
type DuplicateHandlerModel struct {
	existing      *models.IndexEntry
	newPrompt     *models.Prompt
	choice        DuplicateChoice
	cursor        int
	showIDInput   bool
	newID         string
	confirmed     bool
	cancelled     bool
	errorMessage  string
}

// NewDuplicateHandler creates a new duplicate handler model
func NewDuplicateHandler(existing *models.IndexEntry, newPrompt *models.Prompt) DuplicateHandlerModel {
	return DuplicateHandlerModel{
		existing:  existing,
		newPrompt: newPrompt,
		cursor:    0,
	}
}

// Init initializes the model
func (m DuplicateHandlerModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m DuplicateHandlerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showIDInput {
		return m.handleIDInput(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 2 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0: // Update existing
				m.choice = UpdateExisting
				m.confirmed = true
				return m, tea.Quit
			case 1: // Create new
				m.choice = CreateNew
				m.showIDInput = true
				m.errorMessage = ""
			case 2: // Cancel
				m.choice = Cancel
				m.cancelled = true
				return m, tea.Quit
			}
		case "ctrl+c", "q":
			m.choice = Cancel
			m.cancelled = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// handleIDInput handles input when entering a new ID
func (m DuplicateHandlerModel) handleIDInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Validate the ID
			if m.newID == "" {
				m.errorMessage = "ID cannot be empty"
				return m, nil
			}
			// Additional validation will be done by the caller
			m.confirmed = true
			return m, tea.Quit
		case "esc":
			// Go back to menu
			m.showIDInput = false
			m.newID = ""
			m.errorMessage = ""
		case "ctrl+c":
			m.choice = Cancel
			m.cancelled = true
			return m, tea.Quit
		case "backspace":
			if len(m.newID) > 0 {
				m.newID = m.newID[:len(m.newID)-1]
			}
		default:
			// Add character if it's printable
			if len(msg.String()) == 1 {
				m.newID += msg.String()
			}
		}
	}

	return m, nil
}

// View renders the duplicate handler
func (m DuplicateHandlerModel) View() string {
	if m.showIDInput {
		return m.viewIDInput()
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	
	b.WriteString(headerStyle.Render("Duplicate Prompt Found"))
	b.WriteString("\n\n")

	// Show existing prompt details
	detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	b.WriteString("Existing prompt:\n")
	b.WriteString(detailStyle.Render(fmt.Sprintf("  Name: %s\n", m.existing.Name)))
	b.WriteString(detailStyle.Render(fmt.Sprintf("  Author: %s\n", m.existing.Author)))
	b.WriteString(detailStyle.Render(fmt.Sprintf("  Version: %s\n", m.existing.Version)))
	if m.existing.ID != "" {
		b.WriteString(detailStyle.Render(fmt.Sprintf("  ID: %s\n", m.existing.ID)))
	}
	b.WriteString(detailStyle.Render(fmt.Sprintf("  Last Updated: %s\n", m.existing.UpdatedAt.Format("2006-01-02 15:04:05"))))
	b.WriteString("\n")

	// Menu options
	options := []string{
		"Update existing prompt",
		"Create new prompt with different ID",
		"Cancel operation",
	}

	for i, option := range options {
		if i == m.cursor {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				Render("→ " + option))
		} else {
			b.WriteString("  " + option)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(detailStyle.Render("Use ↑/↓ to navigate, Enter to select, q to quit"))

	return b.String()
}

// viewIDInput renders the ID input view
func (m DuplicateHandlerModel) viewIDInput() string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	
	b.WriteString(headerStyle.Render("Enter New ID"))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString("Enter a unique ID for the new prompt:\n")
	b.WriteString("(Use only letters, numbers, hyphens, and underscores)\n\n")

	// Input field
	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(0, 1)

	b.WriteString("ID: ")
	b.WriteString(inputStyle.Render(m.newID))
	b.WriteString("\n\n")

	// Error message
	if m.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
		b.WriteString("\n\n")
	}

	// Instructions
	detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	b.WriteString(detailStyle.Render("Press Enter to confirm, Esc to go back"))

	return b.String()
}

// GetChoice returns the user's choice
func (m DuplicateHandlerModel) GetChoice() DuplicateChoice {
	return m.choice
}

// GetNewID returns the new ID entered by the user
func (m DuplicateHandlerModel) GetNewID() string {
	return m.newID
}

// IsConfirmed returns whether the user confirmed their choice
func (m DuplicateHandlerModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns whether the user cancelled the operation
func (m DuplicateHandlerModel) IsCancelled() bool {
	return m.cancelled
}