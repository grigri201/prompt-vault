package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaginatorModel represents a paginated view
type PaginatorModel struct {
	Items       []string
	PageSize    int
	CurrentPage int
	TotalPages  int
	Width       int
	Height      int
}

// NewPaginator creates a new paginator model
func NewPaginator(items []string, pageSize int) PaginatorModel {
	totalPages := (len(items) + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginatorModel{
		Items:       items,
		PageSize:    pageSize,
		CurrentPage: 1,
		TotalPages:  totalPages,
		Width:       80,
		Height:      24,
	}
}

// Init initializes the model
func (m PaginatorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m PaginatorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.CurrentPage > 1 {
				m.CurrentPage--
			}
		case "right", "l":
			if m.CurrentPage < m.TotalPages {
				m.CurrentPage++
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

// View renders the paginator
func (m PaginatorModel) View() string {
	if len(m.Items) == 0 {
		return "No items to display."
	}

	// Calculate current page items
	startIdx := (m.CurrentPage - 1) * m.PageSize
	endIdx := startIdx + m.PageSize
	if endIdx > len(m.Items) {
		endIdx = len(m.Items)
	}

	// Build the view
	var b strings.Builder

	// Header
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Prompt List"))
	b.WriteString("\n\n")

	// Items
	for i := startIdx; i < endIdx; i++ {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, m.Items[i]))
	}

	// Fill empty space
	itemsShown := endIdx - startIdx
	for i := itemsShown; i < m.PageSize; i++ {
		b.WriteString("\n")
	}

	// Footer with pagination info
	b.WriteString("\n")
	pageInfo := fmt.Sprintf("Page %d of %d", m.CurrentPage, m.TotalPages)
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(pageInfo))

	// Navigation help
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.TotalPages > 1 {
		b.WriteString(helpStyle.Render("← →: navigate • q: quit"))
	} else {
		b.WriteString(helpStyle.Render("q: quit"))
	}

	return b.String()
}

// GetCurrentItems returns the items on the current page
func (m *PaginatorModel) GetCurrentItems() []string {
	if len(m.Items) == 0 {
		return []string{}
	}

	startIdx := (m.CurrentPage - 1) * m.PageSize
	endIdx := startIdx + m.PageSize
	if endIdx > len(m.Items) {
		endIdx = len(m.Items)
	}

	return m.Items[startIdx:endIdx]
}

// GetCurrentRange returns the start and end indices for the current page
func (m *PaginatorModel) GetCurrentRange() (int, int) {
	if len(m.Items) == 0 {
		return 0, 0
	}

	startIdx := (m.CurrentPage - 1) * m.PageSize
	endIdx := startIdx + m.PageSize
	if endIdx > len(m.Items) {
		endIdx = len(m.Items)
	}

	return startIdx, endIdx
}
