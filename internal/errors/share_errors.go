package errors

import "fmt"

var (
	// Share 相关错误
	ErrGistAlreadyPublic = NewAppError(ErrValidation, "Gist 已经是公开的", nil)
	ErrGistAccessDenied  = NewAppError(ErrPermission, "没有访问 Gist 的权限", nil)
	ErrGistNotFound      = NewAppError(ErrNotFound, "找不到指定的 Gist", nil)
	
	// Add URL 相关错误
	ErrGistNotPublic     = NewAppError(ErrValidation, "只能导入公开的 Gist", nil)
	ErrInvalidPromptFormat = NewAppError(ErrValidation, "无效的 Prompt 格式", nil)
)

// NewShareError 创建分享操作相关的错误
func NewShareError(operation string, gistURL string, cause error) AppError {
	message := fmt.Sprintf("分享操作失败: %s (%s)", operation, gistURL)
	return NewAppError(ErrStorage, message, cause)
}

// NewAddFromURLError 创建从 URL 添加相关的错误
func NewAddFromURLError(reason string, gistURL string, cause error) AppError {
	message := fmt.Sprintf("从 URL 添加失败: %s (%s)", reason, gistURL)
	return NewAppError(ErrValidation, message, cause)
}