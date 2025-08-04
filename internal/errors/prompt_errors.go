package errors

import "fmt"

// Prompt operation related errors
var (
	// ErrFileNotFound is returned when the prompt file cannot be found
	ErrFileNotFound = AppError{
		Type:    ErrValidation,
		Message: "Prompt file not found",
	}

	// ErrInvalidYAML is returned when the file is not valid YAML
	ErrInvalidYAML = AppError{
		Type:    ErrValidation,
		Message: "Invalid YAML format",
	}

	// ErrMissingRequired is returned when required fields are missing
	ErrMissingRequired = AppError{
		Type:    ErrValidation,
		Message: "Missing required fields",
	}

	// ErrInvalidMetadata is returned when metadata is invalid
	ErrInvalidMetadata = AppError{
		Type:    ErrValidation,
		Message: "Invalid metadata",
	}
)

// ValidationError represents a validation error with specific field information
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}
