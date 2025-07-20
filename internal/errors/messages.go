package errors

import (
	"errors"
	"fmt"
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