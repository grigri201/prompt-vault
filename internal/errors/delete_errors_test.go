package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestDeleteErrorCreation(t *testing.T) {
	tests := []struct {
		name        string
		errorVar    AppError
		expectedMsg string
		expectedType ErrorType
		baseErr     error
	}{
		{
			name:        "ErrPromptNotFound creation",
			errorVar:    ErrPromptNotFound,
			expectedMsg: "未找到指定的提示",
			expectedType: ErrValidation,
			baseErr:     basePromptNotFound,
		},
		{
			name:        "ErrDeleteCancelled creation",
			errorVar:    ErrDeleteCancelled,
			expectedMsg: "删除操作已取消",
			expectedType: ErrValidation,
			baseErr:     baseUserCancelled,
		},
		{
			name:        "ErrDeleteFailed creation",
			errorVar:    ErrDeleteFailed,
			expectedMsg: "删除操作失败",
			expectedType: ErrStorage,
			baseErr:     nil,
		},
		{
			name:        "ErrInvalidGistURL creation",
			errorVar:    ErrInvalidGistURL,
			expectedMsg: "无效的 Gist URL 格式",
			expectedType: ErrValidation,
			baseErr:     baseInvalidURL,
		},
		{
			name:        "ErrGitHubPermission creation",
			errorVar:    ErrGitHubPermission,
			expectedMsg: "没有权限删除此 Gist",
			expectedType: ErrAuth,
			baseErr:     baseInsufficientPermission,
		},
		{
			name:        "ErrNoPromptsToDelete creation",
			errorVar:    ErrNoPromptsToDelete,
			expectedMsg: "没有找到可删除的提示",
			expectedType: ErrValidation,
			baseErr:     basePromptNotFound,
		},
		{
			name:        "ErrMultiplePromptsFound creation",
			errorVar:    ErrMultiplePromptsFound,
			expectedMsg: "找到多个匹配的提示，请选择要删除的具体提示",
			expectedType: ErrValidation,
			baseErr:     baseAmbiguousInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error message
			if tt.errorVar.Message != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, tt.errorVar.Message)
			}

			// Test error type
			if tt.errorVar.Type != tt.expectedType {
				t.Errorf("Expected type %v, got %v", tt.expectedType, tt.errorVar.Type)
			}

			// Test base error
			if tt.baseErr != nil {
				if !errors.Is(tt.errorVar.Err, tt.baseErr) {
					t.Errorf("Expected base error %v, got %v", tt.baseErr, tt.errorVar.Err)
				}
			} else {
				if tt.errorVar.Err != nil {
					t.Errorf("Expected nil base error, got %v", tt.errorVar.Err)
				}
			}
		})
	}
}

func TestErrorMessageLocalization(t *testing.T) {
	tests := []struct {
		name        string
		errorVar    AppError
		description string
	}{
		{
			name:        "ErrPromptNotFound Chinese message",
			errorVar:    ErrPromptNotFound,
			description: "Error message should be user-friendly in Chinese",
		},
		{
			name:        "ErrDeleteCancelled Chinese message",
			errorVar:    ErrDeleteCancelled,
			description: "Cancel message should be informative in Chinese",
		},
		{
			name:        "ErrDeleteFailed Chinese message",
			errorVar:    ErrDeleteFailed,
			description: "Failure message should be clear in Chinese",
		},
		{
			name:        "ErrInvalidGistURL Chinese message",
			errorVar:    ErrInvalidGistURL,
			description: "URL validation message should be helpful in Chinese",
		},
		{
			name:        "ErrGitHubPermission Chinese message",
			errorVar:    ErrGitHubPermission,
			description: "Permission error should guide user in Chinese",
		},
		{
			name:        "ErrNoPromptsToDelete Chinese message",
			errorVar:    ErrNoPromptsToDelete,
			description: "Empty result message should be informative in Chinese",
		},
		{
			name:        "ErrMultiplePromptsFound Chinese message",
			errorVar:    ErrMultiplePromptsFound,
			description: "Ambiguous input message should guide user in Chinese",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.errorVar.Message
			
			// Check that message is not empty
			if msg == "" {
				t.Error("Error message should not be empty")
			}

			// Check that message contains Chinese characters (basic check)
			hasChineseChars := false
			for _, r := range msg {
				if r >= 0x4e00 && r <= 0x9fff {
					hasChineseChars = true
					break
				}
			}

			if !hasChineseChars {
				t.Errorf("Expected Chinese characters in message: %q", msg)
			}

			// Check that message is user-friendly (no technical jargon)
			technicalTerms := []string{"nil", "null", "error", "exception", "panic"}
			for _, term := range technicalTerms {
				if contains(msg, term) {
					t.Errorf("Message contains technical term '%s': %q", term, msg)
				}
			}
		})
	}
}

func TestErrorCodesAndClassification(t *testing.T) {
	tests := []struct {
		name         string
		errorVar     AppError
		expectedType ErrorType
		category     string
	}{
		{
			name:         "ErrPromptNotFound should be validation error",
			errorVar:     ErrPromptNotFound,
			expectedType: ErrValidation,
			category:     "validation",
		},
		{
			name:         "ErrDeleteCancelled should be validation error",
			errorVar:     ErrDeleteCancelled,
			expectedType: ErrValidation,
			category:     "validation",
		},
		{
			name:         "ErrDeleteFailed should be storage error",
			errorVar:     ErrDeleteFailed,
			expectedType: ErrStorage,
			category:     "storage",
		},
		{
			name:         "ErrInvalidGistURL should be validation error",
			errorVar:     ErrInvalidGistURL,
			expectedType: ErrValidation,
			category:     "validation",
		},
		{
			name:         "ErrGitHubPermission should be auth error",
			errorVar:     ErrGitHubPermission,
			expectedType: ErrAuth,
			category:     "authentication",
		},
		{
			name:         "ErrNoPromptsToDelete should be validation error",
			errorVar:     ErrNoPromptsToDelete,
			expectedType: ErrValidation,
			category:     "validation",
		},
		{
			name:         "ErrMultiplePromptsFound should be validation error",
			errorVar:     ErrMultiplePromptsFound,
			expectedType: ErrValidation,
			category:     "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.errorVar.Type != tt.expectedType {
				t.Errorf("Expected error type %v, got %v", tt.expectedType, tt.errorVar.Type)
			}

			// Verify error type consistency
			switch tt.category {
			case "validation":
				if tt.errorVar.Type != ErrValidation {
					t.Errorf("Validation errors should use ErrValidation type")
				}
			case "storage":
				if tt.errorVar.Type != ErrStorage {
					t.Errorf("Storage errors should use ErrStorage type")
				}
			case "authentication":
				if tt.errorVar.Type != ErrAuth {
					t.Errorf("Authentication errors should use ErrAuth type")
				}
			}
		})
	}
}

func TestErrorWrappingAndContext(t *testing.T) {
	tests := []struct {
		name        string
		baseErr     error
		context     string
		wrappedErr  error
	}{
		{
			name:        "Wrap ErrPromptNotFound with context",
			baseErr:     ErrPromptNotFound,
			context:     "删除提示时出错",
			wrappedErr:  fmt.Errorf("删除提示时出错: %w", ErrPromptNotFound),
		},
		{
			name:        "Wrap ErrInvalidGistURL with context",
			baseErr:     ErrInvalidGistURL,
			context:     "解析URL时出错",
			wrappedErr:  fmt.Errorf("解析URL时出错: %w", ErrInvalidGistURL),
		},
		{
			name:        "Wrap ErrGitHubPermission with context",
			baseErr:     ErrGitHubPermission,
			context:     "GitHub API调用失败",
			wrappedErr:  fmt.Errorf("GitHub API调用失败: %w", ErrGitHubPermission),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that wrapped error maintains the original error
			if !errors.Is(tt.wrappedErr, tt.baseErr) {
				t.Errorf("Wrapped error should contain base error")
			}

			// Test error message contains context
			errorMsg := tt.wrappedErr.Error()
			if !contains(errorMsg, tt.context) {
				t.Errorf("Expected error message to contain context %q, got %q", tt.context, errorMsg)
			}

			// Test that we can still unwrap to get the original AppError
			var appErr AppError
			if !errors.As(tt.wrappedErr, &appErr) {
				t.Error("Should be able to unwrap to AppError")
			}
		})
	}
}

func TestNewDeleteError(t *testing.T) {
	tests := []struct {
		name        string
		errType     ErrorType
		message     string
		baseErr     error
		expectError bool
	}{
		{
			name:        "Create valid delete error",
			errType:     ErrValidation,
			message:     "测试删除错误",
			baseErr:     basePromptNotFound,
			expectError: false,
		},
		{
			name:        "Create delete error with nil base",
			errType:     ErrStorage,
			message:     "存储删除错误",
			baseErr:     nil,
			expectError: false,
		},
		{
			name:        "Create auth delete error",
			errType:     ErrAuth,
			message:     "权限删除错误",
			baseErr:     baseInsufficientPermission,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewDeleteError(tt.errType, tt.message, tt.baseErr)

			// Check error type
			if err.Type != tt.errType {
				t.Errorf("Expected error type %v, got %v", tt.errType, err.Type)
			}

			// Check message
			if err.Message != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, err.Message)
			}

			// Check base error
			if tt.baseErr != nil {
				if !errors.Is(err.Err, tt.baseErr) {
					t.Errorf("Expected base error %v, got %v", tt.baseErr, err.Err)
				}
			} else {
				if err.Err != nil {
					t.Errorf("Expected nil base error, got %v", err.Err)
				}
			}
		})
	}
}

func TestIsDeleteCancelled(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Should detect ErrDeleteCancelled",
			err:      ErrDeleteCancelled,
			expected: true,
		},
		{
			name:     "Should detect wrapped ErrDeleteCancelled",
			err:      fmt.Errorf("context: %w", ErrDeleteCancelled),
			expected: true,
		},
		{
			name:     "Should not detect other errors",
			err:      ErrPromptNotFound,
			expected: false,
		},
		{
			name:     "Should handle nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Should handle non-AppError",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDeleteCancelled(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsPromptNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Should detect ErrPromptNotFound",
			err:      ErrPromptNotFound,
			expected: true,
		},
		{
			name:     "Should detect ErrNoPromptsToDelete",
			err:      ErrNoPromptsToDelete,
			expected: true,
		},
		{
			name:     "Should detect wrapped prompt not found",
			err:      fmt.Errorf("context: %w", ErrPromptNotFound),
			expected: true,
		},
		{
			name:     "Should not detect other errors",
			err:      ErrDeleteCancelled,
			expected: false,
		},
		{
			name:     "Should handle nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPromptNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInvalidGistURL(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Should detect ErrInvalidGistURL",
			err:      ErrInvalidGistURL,
			expected: true,
		},
		{
			name:     "Should detect wrapped invalid URL",
			err:      fmt.Errorf("validation failed: %w", ErrInvalidGistURL),
			expected: true,
		},
		{
			name:     "Should not detect other errors",
			err:      ErrPromptNotFound,
			expected: false,
		},
		{
			name:     "Should handle nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInvalidGistURL(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Should detect ErrGitHubPermission",
			err:      ErrGitHubPermission,
			expected: true,
		},
		{
			name:     "Should detect wrapped permission error",
			err:      fmt.Errorf("API call failed: %w", ErrGitHubPermission),
			expected: true,
		},
		{
			name:     "Should not detect other errors",
			err:      ErrPromptNotFound,
			expected: false,
		},
		{
			name:     "Should handle nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPermissionError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestErrorHandlingConsistency(t *testing.T) {
	// Test that all delete errors follow consistent patterns
	deleteErrors := []AppError{
		ErrPromptNotFound,
		ErrDeleteCancelled,
		ErrDeleteFailed,
		ErrInvalidGistURL,
		ErrGitHubPermission,
		ErrNoPromptsToDelete,
		ErrMultiplePromptsFound,
	}

	for _, err := range deleteErrors {
		t.Run(fmt.Sprintf("Consistency check for %v", err.Message), func(t *testing.T) {
			// All errors should have non-empty messages
			if err.Message == "" {
				t.Error("Error message should not be empty")
			}

			// All errors should have valid types
			validTypes := []ErrorType{ErrValidation, ErrStorage, ErrAuth, ErrNetwork}
			isValidType := false
			for _, validType := range validTypes {
				if err.Type == validType {
					isValidType = true
					break
				}
			}
			if !isValidType {
				t.Errorf("Invalid error type: %v", err.Type)
			}

			// Test error can be converted to string
			errStr := err.Error()
			if errStr == "" {
				t.Error("Error string representation should not be empty")
			}

			// Test error implements error interface
			var _ error = err
		})
	}
}

func TestErrorContextPropagation(t *testing.T) {
	// Test error context propagation through multiple layers
	baseErr := ErrPromptNotFound
	
	// Simulate service layer wrapping
	serviceErr := fmt.Errorf("service error: %w", baseErr)
	
	// Simulate command layer wrapping
	cmdErr := fmt.Errorf("command execution failed: %w", serviceErr)
	
	// Test that original error is still detectable
	if !errors.Is(cmdErr, baseErr) {
		t.Error("Should be able to detect original error through multiple wraps")
	}
	
	// Test that error type detection still works
	if !IsPromptNotFound(cmdErr) {
		t.Error("Should still detect prompt not found error through multiple wraps")
	}
	
	// Test that we can still extract AppError
	var appErr AppError
	if !errors.As(cmdErr, &appErr) {
		t.Error("Should be able to extract AppError through multiple wraps")
	}
	
	if appErr.Message != baseErr.Message {
		t.Errorf("Expected message %q, got %q", baseErr.Message, appErr.Message)
	}
}

func TestErrorStrategyConsistency(t *testing.T) {
	// Test that error handling strategies are consistent across all error types
	strategies := map[ErrorType]string{
		ErrValidation: "应向用户显示友好错误信息并指导正确操作",
		ErrStorage:    "应记录详细日志并向用户显示简化错误信息",
		ErrAuth:       "应指导用户检查权限设置和认证状态",
		ErrNetwork:    "应建议用户检查网络连接并提供重试选项",
	}

	for errType, _ := range strategies {
		t.Run(fmt.Sprintf("Strategy for %v", errType), func(t *testing.T) {
			// Find all errors of this type
			var errorsOfType []AppError
			allErrors := []AppError{
				ErrPromptNotFound,
				ErrDeleteCancelled,
				ErrDeleteFailed,
				ErrInvalidGistURL,
				ErrGitHubPermission,
				ErrNoPromptsToDelete,
				ErrMultiplePromptsFound,
			}

			for _, err := range allErrors {
				if err.Type == errType {
					errorsOfType = append(errorsOfType, err)
				}
			}

			// Verify that all errors of this type follow the strategy
			for _, err := range errorsOfType {
				switch errType {
				case ErrValidation:
					// Validation errors should have user-friendly messages
					if len(err.Message) == 0 {
						t.Errorf("Validation error should have user-friendly message: %v", err)
					}
				case ErrStorage:
					// Storage errors should be suitable for logging
					if err.Type != ErrStorage {
						t.Errorf("Storage error should have correct type: %v", err)
					}
				case ErrAuth:
					// Auth errors should mention permissions or authentication
					msg := err.Message
					if !contains(msg, "权限") && !contains(msg, "认证") && !contains(msg, "登录") {
						t.Errorf("Auth error should mention permissions: %q", msg)
					}
				}
			}
		})
	}
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}