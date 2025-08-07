package errors

import "fmt"

// ErrorType represents the type of error
type ErrorType int

const (
	// ErrUnknown is for unknown errors
	ErrUnknown ErrorType = iota
	// ErrAuth is for authentication related errors
	ErrAuth
	// ErrStorage is for storage related errors
	ErrStorage
	// ErrNetwork is for network related errors
	ErrNetwork
	// ErrValidation is for validation related errors
	ErrValidation
	// ErrPermission is for permission related errors
	ErrPermission
	// ErrNotFound is for not found errors
	ErrNotFound
)

// AppError represents an application error
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e AppError) Unwrap() error {
	return e.Err
}

// Is allows error comparison
func (e AppError) Is(target error) bool {
	t, ok := target.(AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// NewAppError creates a new AppError
func NewAppError(errType ErrorType, message string, err error) AppError {
	return AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}
