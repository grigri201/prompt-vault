package errors

import (
	"fmt"
	"time"
)

// SyncTimeoutError represents a sync operation timeout
type SyncTimeoutError struct {
	AppError
	Duration time.Duration
}

// NewSyncTimeoutError creates a new sync timeout error
func NewSyncTimeoutError(operation string, duration time.Duration) *SyncTimeoutError {
	return &SyncTimeoutError{
		AppError: AppError{
			Op:      operation,
			Type:    ErrTypeNetwork,
			Message: fmt.Sprintf("sync operation timed out after %v", duration),
		},
		Duration: duration,
	}
}

// InvalidIDError represents an invalid ID error
type InvalidIDError struct {
	AppError
	ID     string
	Reason string
}

// NewInvalidIDError creates a new invalid ID error
func NewInvalidIDError(operation string, id string, reason string) *InvalidIDError {
	return &InvalidIDError{
		AppError: AppError{
			Op:      operation,
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("invalid ID '%s': %s", id, reason),
		},
		ID:     id,
		Reason: reason,
	}
}

// DuplicateConflictError represents a duplicate conflict
type DuplicateConflictError struct {
	AppError
	ExistingName   string
	ExistingAuthor string
	ExistingID     string
	NewName        string
	NewAuthor      string
	NewID          string
}

// NewDuplicateConflictError creates a new duplicate conflict error
func NewDuplicateConflictError(operation string, existingName, existingAuthor, existingID, newName, newAuthor, newID string) *DuplicateConflictError {
	message := fmt.Sprintf("duplicate prompt found: existing '%s' by %s", existingName, existingAuthor)
	if existingID != "" {
		message += fmt.Sprintf(" (ID: %s)", existingID)
	}
	
	return &DuplicateConflictError{
		AppError: AppError{
			Op:      operation,
			Type:    ErrTypeValidation,
			Message: message,
		},
		ExistingName:   existingName,
		ExistingAuthor: existingAuthor,
		ExistingID:     existingID,
		NewName:        newName,
		NewAuthor:      newAuthor,
		NewID:          newID,
	}
}