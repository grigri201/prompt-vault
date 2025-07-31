package errors

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorType_Constants(t *testing.T) {
	// Test that error type constants are defined
	assert.Equal(t, ErrorType(0), ErrTypeAuth)
	assert.Equal(t, ErrorType(1), ErrTypeNetwork)
	assert.Equal(t, ErrorType(2), ErrTypeFileSystem)
	assert.Equal(t, ErrorType(3), ErrTypeValidation)
	assert.Equal(t, ErrorType(4), ErrTypeParsing)
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError AppError
		expected string
	}{
		{
			name: "Error with custom message",
			appError: AppError{
				Type:    ErrTypeValidation,
				Op:      "TestOperation",
				Message: "Custom error message",
			},
			expected: "Custom error message",
		},
		{
			name: "Error without custom message",
			appError: AppError{
				Type: ErrTypeNetwork,
				Op:   "TestOperation",
				Err:  fmt.Errorf("underlying error"),
			},
			expected: "TestOperation: underlying error",
		},
		{
			name: "Error with both message and underlying error",
			appError: AppError{
				Type:    ErrTypeAuth,
				Op:      "TestOperation",
				Err:     fmt.Errorf("underlying error"),
				Message: "Priority message",
			},
			expected: "Priority message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := fmt.Errorf("underlying error")
	appErr := AppError{
		Type: ErrTypeNetwork,
		Op:   "TestOperation",
		Err:  underlyingErr,
	}

	unwrapped := appErr.Unwrap()
	assert.Equal(t, underlyingErr, unwrapped)

	// Test with nil underlying error
	appErrNil := AppError{
		Type: ErrTypeValidation,
		Op:   "TestOperation",
	}
	assert.Nil(t, appErrNil.Unwrap())
}

func TestAppError_Is(t *testing.T) {
	baseErr := &AppError{
		Type: ErrTypeValidation,
		Op:   "TestOperation",
	}

	tests := []struct {
		name     string
		target   error
		expected bool
	}{
		{
			name: "Same AppError type",
			target: &AppError{
				Type: ErrTypeValidation,
				Op:   "AnotherOperation",
			},
			expected: true,
		},
		{
			name: "Different AppError type",
			target: &AppError{
				Type: ErrTypeNetwork,
				Op:   "TestOperation",
			},
			expected: false,
		},
		{
			name:     "Non-AppError type",
			target:   fmt.Errorf("regular error"),
			expected: false,
		},
		{
			name:     "Nil target",
			target:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := baseErr.Is(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Context(t *testing.T) {
	context := map[string]interface{}{
		"user_id": "12345",
		"action":  "upload",
		"count":   42,
	}

	appErr := AppError{
		Type:    ErrTypeFileSystem,
		Op:      "TestOperation",
		Message: "Test error with context",
		Context: context,
	}

	assert.Equal(t, context, appErr.Context)
	assert.Equal(t, "12345", appErr.Context["user_id"])
	assert.Equal(t, "upload", appErr.Context["action"])
	assert.Equal(t, 42, appErr.Context["count"])
}

func TestSyncTimeoutError(t *testing.T) {
	duration := 30 * time.Second
	err := NewSyncTimeoutError("sync_operation", duration)

	assert.NotNil(t, err)
	assert.Equal(t, ErrTypeNetwork, err.Type)
	assert.Equal(t, "sync_operation", err.Op)
	assert.Equal(t, duration, err.Duration)
	assert.Contains(t, err.Error(), "timed out")
	assert.Contains(t, err.Error(), "30s")
}

func TestInvalidIDError(t *testing.T) {
	id := "invalid-id!"
	reason := "contains invalid characters"
	err := NewInvalidIDError("validate_id", id, reason)

	assert.NotNil(t, err)
	assert.Equal(t, ErrTypeValidation, err.Type)
	assert.Equal(t, "validate_id", err.Op)
	assert.Equal(t, id, err.ID)
	assert.Equal(t, reason, err.Reason)
	assert.Contains(t, err.Error(), "invalid ID")
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), reason)
}

func TestDuplicateConflictError(t *testing.T) {
	err := NewDuplicateConflictError(
		"upload_prompt",
		"Existing Prompt", "existing_author", "existing_id",
		"New Prompt", "new_author", "new_id",
	)

	assert.NotNil(t, err)
	assert.Equal(t, ErrTypeValidation, err.Type)
	assert.Equal(t, "upload_prompt", err.Op)
	assert.Equal(t, "Existing Prompt", err.ExistingName)
	assert.Equal(t, "existing_author", err.ExistingAuthor)
	assert.Equal(t, "existing_id", err.ExistingID)
	assert.Equal(t, "New Prompt", err.NewName)
	assert.Equal(t, "new_author", err.NewAuthor)
	assert.Equal(t, "new_id", err.NewID)
	assert.Contains(t, err.Error(), "duplicate prompt found")
	assert.Contains(t, err.Error(), "Existing Prompt")
	assert.Contains(t, err.Error(), "existing_author")
	assert.Contains(t, err.Error(), "existing_id")
}

func TestDuplicateConflictError_WithoutID(t *testing.T) {
	err := NewDuplicateConflictError(
		"upload_prompt",
		"Existing Prompt", "existing_author", "",
		"New Prompt", "new_author", "",
	)

	assert.NotNil(t, err)
	assert.Equal(t, "Existing Prompt", err.ExistingName)
	assert.Equal(t, "", err.ExistingID)
	assert.Contains(t, err.Error(), "duplicate prompt found")
	assert.NotContains(t, err.Error(), "(ID:")
}

func TestErrorTypeString(t *testing.T) {
	// Test error types can be used in switch statements
	var errType ErrorType = ErrTypeValidation

	var result string
	switch errType {
	case ErrTypeAuth:
		result = "auth"
	case ErrTypeNetwork:
		result = "network"
	case ErrTypeFileSystem:
		result = "filesystem"
	case ErrTypeValidation:
		result = "validation"
	case ErrTypeParsing:
		result = "parsing"
	default:
		result = "unknown"
	}

	assert.Equal(t, "validation", result)
}

func TestAppError_ChainCompatibility(t *testing.T) {
	// Test that AppError works with error wrapping/unwrapping
	baseErr := fmt.Errorf("base error")
	appErr := &AppError{
		Type: ErrTypeNetwork,
		Op:   "test_operation",
		Err:  baseErr,
	}

	// Test error interface
	var err error = appErr
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_operation")

	// Test unwrapping
	unwrapped := appErr.Unwrap()
	assert.Equal(t, baseErr, unwrapped)
}
