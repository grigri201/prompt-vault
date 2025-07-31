package errors

import (
	"fmt"
	"strings"
)

// DisplayError provides user-friendly error display
func DisplayError(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return formatAppError(appErr)
	}
	return err.Error()
}

// formatAppError formats an AppError for user-friendly display
func formatAppError(err *AppError) string {
	var builder strings.Builder

	// Add type-specific icon and prefix
	switch err.Type {
	case ErrTypeAuth:
		builder.WriteString("🔐 Authentication Error: ")
	case ErrTypeNetwork:
		builder.WriteString("🌐 Network Error: ")
	case ErrTypeFileSystem:
		builder.WriteString("📁 File System Error: ")
	case ErrTypeParsing:
		builder.WriteString("📝 Parsing Error: ")
	case ErrTypeValidation:
		builder.WriteString("✅ Validation Error: ")
	default:
		builder.WriteString("❌ Error: ")
	}

	// Add the main error message
	builder.WriteString(err.Error())

	// Add context information if available
	if len(err.Context) > 0 {
		builder.WriteString("\n")
		for key, value := range err.Context {
			builder.WriteString(fmt.Sprintf("   %s: %v\n", key, value))
		}
	}

	// Add solution suggestions
	if suggestion := getSuggestion(err.Type); suggestion != "" {
		builder.WriteString("\n💡 Suggestion: ")
		builder.WriteString(suggestion)
	}

	return builder.String()
}

// getSuggestion returns contextual suggestions for different error types
func getSuggestion(errorType ErrorType) string {
	switch errorType {
	case ErrTypeAuth:
		return "Please check your GitHub token validity, use 'pv login' to re-authenticate"
	case ErrTypeNetwork:
		return "Please check your internet connection and GitHub API status"
	case ErrTypeFileSystem:
		return "Please check file paths and permissions"
	case ErrTypeParsing:
		return "Please check file format and syntax"
	case ErrTypeValidation:
		return "Please check input data meets requirements"
	default:
		return ""
	}
}

// DisplayErrorCompact provides a more compact error display for CLI
func DisplayErrorCompact(err error) string {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ErrTypeAuth:
			return fmt.Sprintf("🔐 %s", appErr.Error())
		case ErrTypeNetwork:
			return fmt.Sprintf("🌐 %s", appErr.Error())
		case ErrTypeFileSystem:
			return fmt.Sprintf("📁 %s", appErr.Error())
		case ErrTypeParsing:
			return fmt.Sprintf("📝 %s", appErr.Error())
		case ErrTypeValidation:
			return fmt.Sprintf("✅ %s", appErr.Error())
		default:
			return fmt.Sprintf("❌ %s", appErr.Error())
		}
	}
	return err.Error()
}

// IsRetryable determines if an error might be resolved by retrying
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ErrTypeNetwork:
			// Network errors are often transient
			errMsg := strings.ToLower(appErr.Error())
			return strings.Contains(errMsg, "timeout") ||
				strings.Contains(errMsg, "connection") ||
				strings.Contains(errMsg, "temporary")
		case ErrTypeFileSystem:
			// Some file system errors might be retryable
			errMsg := strings.ToLower(appErr.Error())
			return strings.Contains(errMsg, "temporary") ||
				strings.Contains(errMsg, "busy")
		}
	}
	return false
}

// GetErrorCode returns a unique error code for programmatic handling
func GetErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ErrTypeAuth:
			return "AUTH_ERROR"
		case ErrTypeNetwork:
			return "NETWORK_ERROR"
		case ErrTypeFileSystem:
			return "FILESYSTEM_ERROR"
		case ErrTypeParsing:
			return "PARSING_ERROR"
		case ErrTypeValidation:
			return "VALIDATION_ERROR"
		}
	}
	return "UNKNOWN_ERROR"
}
