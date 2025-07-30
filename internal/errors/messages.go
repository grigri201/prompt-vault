package errors

import (
	"errors"
	"fmt"
	"strings"
)

// NewAuthErrorMsg creates an authentication error with a custom message
func NewAuthErrorMsg(op string, msg string) *AppError {
	return &AppError{
		Type:    ErrTypeAuth,
		Op:      op,
		Err:     errors.New(msg),
		Message: msg,
	}
}

// NewValidationErrorMsg creates a validation error with a custom message
func NewValidationErrorMsg(op string, msg string) *AppError {
	return &AppError{
		Type:    ErrTypeValidation,
		Op:      op,
		Err:     errors.New(msg),
		Message: msg,
	}
}

// NewFileSystemErrorMsg creates a file system error with a custom message
func NewFileSystemErrorMsg(op string, msg string) *AppError {
	return &AppError{
		Type:    ErrTypeFileSystem,
		Op:      op,
		Err:     errors.New(msg),
		Message: msg,
	}
}

// NewParsingErrorMsg creates a parsing error with a custom message
func NewParsingErrorMsg(op string, msg string) *AppError {
	return &AppError{
		Type:    ErrTypeParsing,
		Op:      op,
		Err:     errors.New(msg),
		Message: msg,
	}
}

// NewNetworkErrorMsg creates a network error with a custom message
func NewNetworkErrorMsg(op string, msg string) *AppError {
	return &AppError{
		Type:    ErrTypeNetwork,
		Op:      op,
		Err:     errors.New(msg),
		Message: msg,
	}
}

// WrapWithMessage wraps an error with a custom message while preserving the type
func WrapWithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}

	// If it's already an AppError, preserve the type
	var appErr *AppError
	if As(err, &appErr) {
		return &AppError{
			Type:    appErr.Type,
			Op:      appErr.Op,
			Err:     appErr.Err,
			Message: msg,
		}
	}

	// Otherwise, wrap as generic error
	return fmt.Errorf("%s: %w", msg, err)
}

// Share command specific error messages
const (
	// Share errors
	ErrMsgSharePublicGist    = "Cannot share a gist that is already public"
	ErrMsgShareGistNotFound  = "The gist was not found. Please check the gist ID and try again"
	ErrMsgShareNoYAMLFile    = "The gist does not contain a valid YAML prompt file"
	ErrMsgShareInvalidPrompt = "The prompt file has an invalid format"
	ErrMsgShareCreateFailed  = "Failed to create public gist. Please check your permissions and try again"
	ErrMsgShareUpdateFailed  = "Failed to update existing public gist"
	ErrMsgShareCancelled     = "Share operation was cancelled by user"
	ErrMsgShareNetworkError  = "Network error occurred while sharing. Please check your internet connection"

	// Import errors
	ErrMsgImportInvalidURL     = "The provided URL is not valid. Please provide a valid GitHub gist URL"
	ErrMsgImportNotGitHubURL   = "The URL is not a GitHub gist URL. Only GitHub gists can be imported"
	ErrMsgImportGistNotFound   = "The gist was not found. Please check the URL and ensure it's public"
	ErrMsgImportPrivateGist    = "Cannot import private gists. Only public gists can be imported"
	ErrMsgImportNoValidPrompt  = "The gist does not contain a valid prompt file"
	ErrMsgImportMissingFields  = "The prompt is missing required fields (name, author, category, tags, or version)"
	ErrMsgImportInvalidVersion = "The prompt has an invalid version format"
	ErrMsgImportCancelled      = "Import operation was cancelled by user"
	ErrMsgImportIndexUpdate    = "Failed to update the index after import"

	// Common messages
	ErrMsgAuthRequired   = "Authentication required. Please run 'pv login' first"
	ErrMsgNetworkTimeout = "Request timed out. Please check your internet connection and try again"
)

// GetShareErrorMessage returns a user-friendly error message for share errors
func GetShareErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for specific error types
	var appErr *AppError
	if As(err, &appErr) {
		// If it has a custom message and it's not the default one, use it
		if appErr.Message != "" && !isDefaultMessage(appErr) {
			return appErr.Message
		}

		// Otherwise, use type-specific messages
		switch appErr.Type {
		case ErrTypeAuth:
			return ErrMsgAuthRequired
		case ErrTypeNetwork:
			return ErrMsgShareNetworkError
		}
	}

	// Check error message content
	errMsg := err.Error()
	switch {
	case contains(errMsg, "already public"):
		return ErrMsgSharePublicGist
	case contains(errMsg, "not found"):
		return ErrMsgShareGistNotFound
	case contains(errMsg, "YAML"):
		return ErrMsgShareNoYAMLFile
	case contains(errMsg, "cancelled"):
		return ErrMsgShareCancelled
	default:
		return fmt.Sprintf("Share failed: %v", err)
	}
}

// GetImportErrorMessage returns a user-friendly error message for import errors
func GetImportErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for specific error types
	var appErr *AppError
	if As(err, &appErr) {
		// If it has a custom message and it's not the default one, use it
		if appErr.Message != "" && !isDefaultMessage(appErr) {
			return appErr.Message
		}

		// Otherwise, use type-specific messages
		switch appErr.Type {
		case ErrTypeAuth:
			return ErrMsgAuthRequired
		case ErrTypeNetwork:
			return ErrMsgNetworkTimeout
		case ErrTypeValidation:
			return ErrMsgImportNoValidPrompt
		}
	}

	// Check error message content
	errMsg := err.Error()
	switch {
	case contains(errMsg, "invalid URL"):
		return ErrMsgImportInvalidURL
	case contains(errMsg, "not a GitHub gist URL"):
		return ErrMsgImportNotGitHubURL
	case contains(errMsg, "not found"):
		return ErrMsgImportGistNotFound
	case contains(errMsg, "private"):
		return ErrMsgImportPrivateGist
	case contains(errMsg, "valid prompt"):
		return ErrMsgImportNoValidPrompt
	case contains(errMsg, "required"):
		return ErrMsgImportMissingFields
	case contains(errMsg, "cancelled"):
		return ErrMsgImportCancelled
	default:
		return fmt.Sprintf("Import failed: %v", err)
	}
}

// isDefaultMessage checks if the AppError has a default generated message
func isDefaultMessage(appErr *AppError) bool {
	if appErr == nil || appErr.Message == "" {
		return true
	}

	// Check if it matches the default message pattern
	defaultPrefixes := []string{
		"authentication failed during ",
		"network error during ",
		"file system error during ",
		"validation failed during ",
		"parsing failed during ",
	}

	for _, prefix := range defaultPrefixes {
		if contains(appErr.Message, prefix) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
