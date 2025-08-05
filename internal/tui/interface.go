package tui

import "github.com/grigri/pv/internal/model"

// TUIInterface defines the contract for terminal user interface operations
// related to prompt management functionality. This interface enables both
// production bubbletea implementations and test mock implementations.
type TUIInterface interface {
	// ShowPromptList displays a list of prompts in an interactive interface
	// and returns the user-selected prompt. Returns an error if the user
	// cancels the operation or if there's an interface error.
	ShowPromptList(prompts []model.Prompt) (model.Prompt, error)

	// ShowConfirm displays a confirmation dialog for the given prompt
	// and returns true if the user confirms the deletion, false if cancelled.
	// Returns an error if there's an interface error.
	ShowConfirm(prompt model.Prompt) (bool, error)

	// ShowVariableForm displays a form for collecting variable values from the user.
	// Takes a list of variable names and returns a map of variable names to values.
	// Returns an error if the user cancels the operation or if there's an interface error.
	ShowVariableForm(variables []string) (map[string]string, error)
}

// ListMode represents different modes for displaying prompt lists
type ListMode int

const (
	// ListAll displays all available prompts without filtering
	ListAll ListMode = iota
	// ListFiltered displays prompts filtered by a search keyword
	ListFiltered
)

// String returns the string representation of ListMode
func (lm ListMode) String() string {
	switch lm {
	case ListAll:
		return "all"
	case ListFiltered:
		return "filtered"
	default:
		return "unknown"
	}
}

// Keyboard shortcut constants for TUI navigation and actions
const (
	// Navigation keys
	KeyUp       = "up"
	KeyDown     = "down"
	KeyPageUp   = "pgup"
	KeyPageDown = "pgdown"
	KeyHome     = "home"
	KeyEnd      = "end"

	// Action keys
	KeyEnter  = "enter"
	KeyEscape = "esc"
	KeyQuit   = "q"

	// Confirmation keys
	KeyYes = "y"
	KeyNo  = "n"

	// Special keys
	KeyTab       = "tab"
	KeyShiftTab  = "shift+tab"
	KeyBackspace = "backspace"
	KeyDelete    = "delete"
	KeySpace     = " "

	// Control keys
	KeyCtrlC = "ctrl+c"
	KeyCtrlD = "ctrl+d"
	KeyCtrlZ = "ctrl+z"
)

// TUI display constants for consistent formatting
const (
	// Screen dimensions
	MinScreenWidth  = 80
	MinScreenHeight = 24

	// List display settings
	MaxListItems         = 20
	ListItemPadding      = 2
	ListNumberWidth      = 4
	MaxPromptNameLength  = 40
	MaxAuthorNameLength  = 20

	// Confirmation dialog settings
	ConfirmDialogWidth  = 60
	ConfirmDialogHeight = 10

	// Variable form settings
	VariableFormWidth      = 70
	VariableFormMinHeight  = 15
	VariableInputWidth     = 50
	VariableLabelWidth     = 15
	MaxVariableNameLength  = 30

	// Colors and styling (compatible with lipgloss)
	ColorPrimary   = "#00D4AA"
	ColorSecondary = "#7C7C7C"
	ColorError     = "#FF0000"
	ColorWarning   = "#FFA500"
	ColorSuccess   = "#00FF00"
	ColorBorder    = "#444444"
	ColorText      = "#FFFFFF"
	ColorMuted     = "#888888"
)

// Error messages for TUI operations
const (
	ErrMsgNoPromptsFound     = "没有找到任何提示"
	ErrMsgListEmpty          = "提示列表为空"
	ErrMsgUserCancelled      = "用户取消了操作"
	ErrMsgInvalidSelection   = "无效的选择"
	ErrMsgTUIInitFailed      = "初始化用户界面失败"
	ErrMsgTUIRenderFailed    = "渲染界面失败"
	ErrMsgKeyboardInputError = "键盘输入处理错误"
)

// Help text constants for user guidance
const (
	HelpTextListNavigation = "↑/↓: 导航  Enter: 选择  q: 退出"
	HelpTextConfirmation   = "Y: 确认  N: 取消  Esc: 取消"
	HelpTextVariableForm   = "Tab/↓: 下一字段  Shift+Tab/↑: 上一字段  Enter: 确认  Esc: 取消"
	HelpTextGeneral        = "按 q 退出"
	HelpTextLoading        = "正在加载..."
)