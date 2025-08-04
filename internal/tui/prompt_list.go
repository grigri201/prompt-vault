package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/grigri/pv/internal/model"
)

// PromptListModel represents the state of the prompt list TUI interface
type PromptListModel struct {
	prompts  []model.Prompt  // List of prompts to display
	cursor   int             // Currently selected item index
	selected *model.Prompt   // User-selected prompt
	err      error           // Any error that occurred
	mode     ListMode        // Display mode (ListAll or ListFiltered)
	filter   string          // Filter keyword for ListFiltered mode
	loading  bool            // Whether the list is currently loading
	done     bool            // Whether user has made a selection or quit
	quit     bool            // Whether user wants to quit without selection
}

// NewPromptListModel creates a new instance of PromptListModel
func NewPromptListModel(prompts []model.Prompt, mode ListMode, filter string) PromptListModel {
	return PromptListModel{
		prompts:  prompts,
		cursor:   0,
		selected: nil,
		err:      nil,
		mode:     mode,
		filter:   filter,
		loading:  false,
		done:     false,
		quit:     false,
	}
}

// NewLoadingPromptListModel creates a new instance in loading state
func NewLoadingPromptListModel() PromptListModel {
	return PromptListModel{
		prompts:  nil,
		cursor:   0,
		selected: nil,
		err:      nil,
		mode:     ListAll,
		filter:   "",
		loading:  true,
		done:     false,
		quit:     false,
	}
}

// promptsLoadedMsg is sent when prompts are loaded
type promptsLoadedMsg struct {
	prompts []model.Prompt
	mode    ListMode
	filter  string
}

// promptLoadErrorMsg is sent when there's an error loading prompts
type promptLoadErrorMsg struct {
	err error
}

// Init initializes the model and returns any initial commands
func (m PromptListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state
func (m PromptListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case promptsLoadedMsg:
		return m.handlePromptsLoaded(msg), nil
	case promptLoadErrorMsg:
		return m.handleLoadError(msg), nil
	case tea.WindowSizeMsg:
		return m, nil
	}
	return m, nil
}

// handleKeyPress processes keyboard input
func (m PromptListModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Don't process keys if loading
	if m.loading {
		return m, nil
	}

	switch msg.String() {
	case KeyQuit, KeyCtrlC, KeyEscape:
		m.quit = true
		m.done = true
		return m, tea.Quit

	case KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}

	case KeyDown:
		if m.cursor < len(m.prompts)-1 {
			m.cursor++
		}

	case KeyHome:
		m.cursor = 0

	case KeyEnd:
		if len(m.prompts) > 0 {
			m.cursor = len(m.prompts) - 1
		}

	case KeyPageUp:
		m.cursor = max(0, m.cursor-10)

	case KeyPageDown:
		m.cursor = min(len(m.prompts)-1, m.cursor+10)

	case KeyEnter:
		if len(m.prompts) > 0 && m.cursor >= 0 && m.cursor < len(m.prompts) {
			m.selected = &m.prompts[m.cursor]
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// handlePromptsLoaded processes the prompts loaded message
func (m PromptListModel) handlePromptsLoaded(msg promptsLoadedMsg) PromptListModel {
	m.prompts = msg.prompts
	m.mode = msg.mode
	m.filter = msg.filter
	m.loading = false
	m.cursor = 0
	return m
}

// handleLoadError processes the load error message
func (m PromptListModel) handleLoadError(msg promptLoadErrorMsg) PromptListModel {
	m.err = msg.err
	m.loading = false
	return m
}

// View renders the current state of the model
func (m PromptListModel) View() string {
	if m.loading {
		return m.renderLoadingView()
	}

	if m.err != nil {
		return m.renderErrorView()
	}

	if len(m.prompts) == 0 {
		return m.renderEmptyView()
	}

	return m.renderListView()
}

// renderLoadingView renders the loading state
func (m PromptListModel) renderLoadingView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(1, 2).
		Width(MinScreenWidth).
		Align(lipgloss.Center)

	content := fmt.Sprintf("🔄 %s", HelpTextLoading)
	return style.Render(content)
}

// renderErrorView renders the error state
func (m PromptListModel) renderErrorView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorError)).
		Padding(1, 2).
		Width(MinScreenWidth)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError)).
		Bold(true).
		Render("❌ 加载失败")

	content := fmt.Sprintf("%s\n\n%s\n\n%s", 
		title, 
		m.err.Error(),
		HelpTextGeneral)

	return style.Render(content)
}

// renderEmptyView renders the empty list state
func (m PromptListModel) renderEmptyView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(1, 2).
		Width(MinScreenWidth).
		Align(lipgloss.Center)

	var message string
	if m.mode == ListFiltered {
		message = fmt.Sprintf("🔍 没有找到包含 \"%s\" 的提示", m.filter)
	} else {
		message = fmt.Sprintf("📝 %s", ErrMsgListEmpty)
	}

	content := fmt.Sprintf("%s\n\n%s", message, HelpTextGeneral)
	return style.Render(content)
}

// renderListView renders the main prompt list
func (m PromptListModel) renderListView() string {
	// Header
	var title string
	if m.mode == ListFiltered {
		title = fmt.Sprintf("🗑️  Prompt Vault - 筛选结果 \"%s\"", m.filter)
	} else {
		title = "🗑️  Prompt Vault - 选择要删除的提示"
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true).
		Padding(0, 1)

	header := headerStyle.Render(title)

	// List items
	var items []string
	for i, prompt := range m.prompts {
		item := m.renderListItem(i, prompt, i == m.cursor)
		items = append(items, item)
	}

	listContent := strings.Join(items, "\n")

	// Footer with help text
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Padding(1, 1)

	footer := footerStyle.Render(HelpTextListNavigation)

	// Main container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(1, 1).
		Width(MinScreenWidth)

	content := fmt.Sprintf("%s\n\n%s\n\n%s", header, listContent, footer)
	return containerStyle.Render(content)
}

// renderListItem renders a single prompt item in the list
func (m PromptListModel) renderListItem(index int, prompt model.Prompt, selected bool) string {
	// Format: [数字] 提示名称 (作者: 作者名)
	number := fmt.Sprintf("[%d]", index+1)
	
	// Truncate name if too long
	name := prompt.Name
	if len(name) > MaxPromptNameLength {
		name = name[:MaxPromptNameLength-3] + "..."
	}

	// Truncate author if too long
	author := prompt.Author
	if len(author) > MaxAuthorNameLength {
		author = author[:MaxAuthorNameLength-3] + "..."
	}

	itemText := fmt.Sprintf("%s %s (作者: %s)", number, name, author)

	if selected {
		// Selected item style
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorPrimary)).
			Bold(true).
			Padding(0, 1)
		return "  > " + style.Render(itemText)
	} else {
		// Normal item style
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Padding(0, 1)
		return "    " + style.Render(itemText)
	}
}

// GetSelected returns the selected prompt if any
func (m PromptListModel) GetSelected() *model.Prompt {
	return m.selected
}

// IsDone returns true if the user has made a selection or quit
func (m PromptListModel) IsDone() bool {
	return m.done
}

// HasQuit returns true if the user quit without making a selection
func (m PromptListModel) HasQuit() bool {
	return m.quit
}

// GetError returns any error that occurred
func (m PromptListModel) GetError() error {
	return m.err
}

// SetPrompts updates the prompts list (useful for dynamic loading)
func (m *PromptListModel) SetPrompts(prompts []model.Prompt, mode ListMode, filter string) tea.Cmd {
	return func() tea.Msg {
		return promptsLoadedMsg{
			prompts: prompts,
			mode:    mode,
			filter:  filter,
		}
	}
}

// SetError sets an error state
func (m *PromptListModel) SetError(err error) tea.Cmd {
	return func() tea.Msg {
		return promptLoadErrorMsg{err: err}
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}