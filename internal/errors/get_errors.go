package errors

import (
	"errors"
	"fmt"
)

// 获取操作相关的基础错误
var (
	// baseNoPromptsFound 基础没有找到提示词错误
	baseNoPromptsFound = fmt.Errorf("no prompts found")
	
	// baseInvalidGistURL 基础无效URL错误
	baseInvalidGistURL = fmt.Errorf("invalid gist URL")
	
	// baseVariableRequired 基础变量必填错误
	baseVariableRequired = fmt.Errorf("variable value required")
	
	// baseClipboardUnavailable 基础剪贴板不可用错误
	baseClipboardUnavailable = fmt.Errorf("clipboard unavailable")
	
	// basePromptContentError 基础提示词内容错误
	basePromptContentError = fmt.Errorf("prompt content error")
)

// GetError represents a get operation error with structured information
type GetError struct {
	Operation string
	Cause     error
}

// Error implements the error interface
func (e *GetError) Error() string {
	return fmt.Sprintf("获取提示词失败: %s - %v", e.Operation, e.Cause)
}

// Unwrap returns the wrapped error
func (e *GetError) Unwrap() error {
	return e.Cause
}

// Is allows error comparison
func (e *GetError) Is(target error) bool {
	t, ok := target.(*GetError)
	if !ok {
		return false
	}
	return e.Operation == t.Operation
}

// Get operation specific errors
var (
	// ErrNoPromptsFound 表示未找到任何提示词
	ErrNoPromptsFound = NewAppError(
		ErrValidation,
		"未找到任何提示词",
		baseNoPromptsFound,
	)

	// ErrInvalidGetGistURL 表示获取时提供的 Gist URL 格式无效
	ErrInvalidGetGistURL = NewAppError(
		ErrValidation,
		"无效的 Gist URL 格式",
		baseInvalidGistURL,
	)

	// ErrVariableRequired 表示变量值不能为空
	ErrVariableRequired = NewAppError(
		ErrValidation,
		"变量值不能为空",
		baseVariableRequired,
	)

	// ErrClipboardUnavailable 表示剪贴板不可用
	ErrClipboardUnavailable = NewAppError(
		ErrStorage,
		"剪贴板不可用",
		baseClipboardUnavailable,
	)

	// ErrGetCancelled 表示用户取消了获取操作
	ErrGetCancelled = NewAppError(
		ErrValidation,
		"获取操作已取消",
		baseUserCancelled,
	)

	// ErrPromptContentEmpty 表示提示词内容为空
	ErrPromptContentEmpty = NewAppError(
		ErrValidation,
		"提示词内容为空",
		basePromptContentError,
	)

	// ErrVariableParsingFailed 表示变量解析失败
	ErrVariableParsingFailed = NewAppError(
		ErrValidation,
		"变量解析失败",
		basePromptContentError,
	)
)

// NewGetError 创建一个获取操作相关的错误
func NewGetError(operation string, cause error) *GetError {
	return &GetError{
		Operation: operation,
		Cause:     cause,
	}
}

// IsGetCancelled 检查错误是否为用户取消获取操作
func IsGetCancelled(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseUserCancelled)
	}
	return false
}

// IsNoPromptsFound 检查错误是否为未找到提示词
func IsNoPromptsFound(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseNoPromptsFound)
	}
	return false
}

// IsInvalidGistURL 检查错误是否为无效的 Gist URL (复用删除错误检查函数)
func IsInvalidGetGistURL(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseInvalidGistURL)
	}
	return false
}

// IsVariableRequired 检查错误是否为变量值必填
func IsVariableRequired(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseVariableRequired)
	}
	return false
}

// IsClipboardUnavailable 检查错误是否为剪贴板不可用
func IsClipboardUnavailable(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseClipboardUnavailable)
	}
	return false
}