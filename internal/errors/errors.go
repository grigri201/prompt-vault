package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	// ErrTypeAuth represents authentication errors
	ErrTypeAuth ErrorType = iota
	// ErrTypeNetwork represents network-related errors
	ErrTypeNetwork
	// ErrTypeFileSystem represents file system errors
	ErrTypeFileSystem
	// ErrTypeValidation represents validation errors
	ErrTypeValidation
	// ErrTypeParsing represents parsing errors
	ErrTypeParsing
)

// AppError represents application errors with context
type AppError struct {
	Type    ErrorType
	Op      string // Operation that failed
	Err     error  // Underlying error
	Message string // User-friendly message
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the target error is of the same type
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// Constructor functions

// NewAuthError creates a new authentication error
func NewAuthError(op string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeAuth,
		Op:      op,
		Err:     err,
		Message: fmt.Sprintf("authentication failed during %s", op),
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(op string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeNetwork,
		Op:      op,
		Err:     err,
		Message: fmt.Sprintf("network error during %s", op),
	}
}

// NewFileSystemError creates a new file system error
func NewFileSystemError(op string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeFileSystem,
		Op:      op,
		Err:     err,
		Message: fmt.Sprintf("file system error during %s", op),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(op string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeValidation,
		Op:      op,
		Err:     err,
		Message: fmt.Sprintf("validation failed during %s", op),
	}
}

// NewParsingError creates a new parsing error
func NewParsingError(op string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeParsing,
		Op:      op,
		Err:     err,
		Message: fmt.Sprintf("parsing failed during %s", op),
	}
}

// Helper functions

// Wrap wraps an error with operation context
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// IsType checks if an error is of a specific AppError type
func IsType(err error, errType ErrorType) bool {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// As is a wrapper around errors.As for convenience
func As(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	return asError(err, target)
}

// asError is a helper function that checks if err can be assigned to target
func asError(err error, target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **AppError:
		if appErr, ok := err.(*AppError); ok {
			*t = appErr
			return true
		}
		// Check if wrapped error is AppError
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			return asError(unwrapper.Unwrap(), target)
		}
	}
	return false
}

// NewError creates a new AppError with automatic type detection
func NewError(op string, err error) *AppError {
	errType := determineErrorType(err)
	return &AppError{
		Type:    errType,
		Op:      op,
		Err:     err,
		Message: fmt.Sprintf("%s: %v", op, err),
	}
}

// WrapError wraps an error with consistent AppError handling
func WrapError(op string, err error) error {
	if err == nil {
		return nil
	}

	// If it's already an AppError, preserve the type
	var appErr *AppError
	if As(err, &appErr) {
		return &AppError{
			Type:    appErr.Type,
			Op:      op,
			Err:     err,
			Message: fmt.Sprintf("%s: %v", op, err),
		}
	}

	// Create new AppError with automatic type detection
	return NewError(op, err)
}

// determineErrorType automatically determines the error type based on error content
func determineErrorType(err error) ErrorType {
	if err == nil {
		return ErrTypeValidation
	}

	errMsg := strings.ToLower(err.Error())

	// Check for authentication errors
	if strings.Contains(errMsg, "auth") || strings.Contains(errMsg, "token") ||
		strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "unauthorized") {
		return ErrTypeAuth
	}

	// Check for network errors
	if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "dial") ||
		strings.Contains(errMsg, "http") || strings.Contains(errMsg, "tls") ||
		strings.Contains(errMsg, "rate limit") {
		return ErrTypeNetwork
	}

	// Check for file system errors
	if strings.Contains(errMsg, "file") || strings.Contains(errMsg, "directory") ||
		strings.Contains(errMsg, "path") || strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "no such") || strings.Contains(errMsg, "exist") {
		return ErrTypeFileSystem
	}

	// Check for parsing errors
	if strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "unmarshal") ||
		strings.Contains(errMsg, "yaml") || strings.Contains(errMsg, "json") ||
		strings.Contains(errMsg, "syntax") || strings.Contains(errMsg, "invalid format") {
		return ErrTypeParsing
	}

	// Default to validation error
	return ErrTypeValidation
}
