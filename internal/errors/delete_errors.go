package errors

import (
	"errors"
	"fmt"
)

// 删除操作相关的基础错误
var (
	// basePromptNotFound 基础提示未找到错误
	basePromptNotFound = fmt.Errorf("prompt not found")
	
	// baseUserCancelled 基础用户取消错误
	baseUserCancelled = fmt.Errorf("operation cancelled by user")
	
	// baseInvalidURL 基础无效URL错误
	baseInvalidURL = fmt.Errorf("invalid URL format")
	
	// baseInsufficientPermission 基础权限不足错误
	baseInsufficientPermission = fmt.Errorf("insufficient permission")
	
	// baseAmbiguousInput 基础输入模糊错误
	baseAmbiguousInput = fmt.Errorf("ambiguous input")
)

// Delete operation specific errors
var (
	// ErrPromptNotFound 表示未找到指定的提示
	ErrPromptNotFound = NewAppError(
		ErrValidation,
		"未找到指定的提示",
		basePromptNotFound,
	)

	// ErrDeleteCancelled 表示用户取消了删除操作
	ErrDeleteCancelled = NewAppError(
		ErrValidation,
		"删除操作已取消",
		baseUserCancelled,
	)

	// ErrDeleteFailed 表示删除操作失败
	ErrDeleteFailed = NewAppError(
		ErrStorage,
		"删除操作失败",
		nil,
	)

	// ErrInvalidGistURL 表示提供的 Gist URL 格式无效
	ErrInvalidGistURL = NewAppError(
		ErrValidation,
		"无效的 Gist URL 格式",
		baseInvalidURL,
	)

	// ErrGitHubPermission 表示没有权限删除指定的 Gist
	ErrGitHubPermission = NewAppError(
		ErrAuth,
		"没有权限删除此 Gist",
		baseInsufficientPermission,
	)

	// ErrNoPromptsToDelete 表示没有可删除的提示
	ErrNoPromptsToDelete = NewAppError(
		ErrValidation,
		"没有找到可删除的提示",
		basePromptNotFound,
	)

	// ErrMultiplePromptsFound 表示找到多个匹配的提示，需要用户选择
	ErrMultiplePromptsFound = NewAppError(
		ErrValidation,
		"找到多个匹配的提示，请选择要删除的具体提示",
		baseAmbiguousInput,
	)
)

// NewDeleteError 创建一个删除操作相关的错误
func NewDeleteError(errType ErrorType, message string, baseErr error) AppError {
	return NewAppError(errType, message, baseErr)
}

// IsDeleteCancelled 检查错误是否为用户取消删除操作
func IsDeleteCancelled(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseUserCancelled)
	}
	return false
}

// IsPromptNotFound 检查错误是否为提示未找到
func IsPromptNotFound(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, basePromptNotFound)
	}
	return false
}

// IsInvalidGistURL 检查错误是否为无效的 Gist URL
func IsInvalidGistURL(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseInvalidURL)
	}
	return false
}

// IsPermissionError 检查错误是否为权限相关错误
func IsPermissionError(err error) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, baseInsufficientPermission)
	}
	return false
}