package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressModel represents a progress indicator with spinner
type ProgressModel struct {
	message  string
	quitting bool
	frame    int
}

// tickMsg is sent to advance the spinner animation
type tickMsg time.Time

// ProgressCompleteMsg signals that the progress operation is complete
type ProgressCompleteMsg struct{}

// spinner frames for animation
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewProgressModel creates a new progress indicator model
func NewProgressModel(message string) tea.Model {
	return &ProgressModel{
		message: message,
		frame:   0,
	}
}

// Init starts the spinner animation
func (m *ProgressModel) Init() tea.Cmd {
	return tick()
}

// Update handles messages
func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		// Advance spinner frame
		m.frame = (m.frame + 1) % len(spinnerFrames)
		return m, tick()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}

	case ProgressCompleteMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the progress indicator
func (m *ProgressModel) View() string {
	if m.quitting {
		return ""
	}

	spinner := spinnerFrames[m.frame]
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return fmt.Sprintf("%s %s", style.Render(spinner), m.message)
}

// SetMessage updates the progress message
func (m *ProgressModel) SetMessage(message string) {
	m.message = message
}

// tick returns a command that sends a tick message after a delay
func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Progress is a higher-level progress indicator that can be used without tea.Program
type Progress struct {
	program *tea.Program
	model   *ProgressModel
}

// NewProgress creates a new progress indicator
func NewProgress(message string) *Progress {
	model := &ProgressModel{
		message: message,
		frame:   0,
	}

	program := tea.NewProgram(model)

	return &Progress{
		program: program,
		model:   model,
	}
}

// Start begins showing the progress indicator
func (p *Progress) Start() error {
	go func() {
		if _, err := p.program.Run(); err != nil {
			// Log error but don't crash
			fmt.Printf("Progress indicator error: %v\n", err)
		}
	}()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)
	return nil
}

// Stop stops the progress indicator
func (p *Progress) Stop() {
	p.program.Send(ProgressCompleteMsg{})
	// Give it a moment to clean up
	time.Sleep(50 * time.Millisecond)
}

// UpdateMessage changes the progress message
func (p *Progress) UpdateMessage(message string) {
	p.model.SetMessage(message)
}
