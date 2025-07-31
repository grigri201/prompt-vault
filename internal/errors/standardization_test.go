package errors

import (
	"fmt"
	"testing"
)

func TestErrorStandardization(t *testing.T) {
	tests := []struct {
		name         string
		createError  func() error
		expectedType ErrorType
	}{
		{
			name: "Validation Error",
			createError: func() error {
				return NewValidationErrorMsg("Test", "validation failed")
			},
			expectedType: ErrTypeValidation,
		},
		{
			name: "Network Error with Context",
			createError: func() error {
				return NewNetworkErrorWithURL("Test", "https://api.github.com",
					"request failed", fmt.Errorf("connection timeout"))
			},
			expectedType: ErrTypeNetwork,
		},
		{
			name: "File System Error",
			createError: func() error {
				return NewFileSystemError("Test", fmt.Errorf("permission denied"))
			},
			expectedType: ErrTypeFileSystem,
		},
		{
			name: "Auth Error",
			createError: func() error {
				return NewAuthErrorMsg("Test", "token expired")
			},
			expectedType: ErrTypeAuth,
		},
		{
			name: "Parsing Error",
			createError: func() error {
				return NewParsingError("Test", fmt.Errorf("invalid YAML"))
			},
			expectedType: ErrTypeParsing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()

			appErr, ok := err.(*AppError)
			if !ok {
				t.Errorf("error should implement AppError interface")
				return
			}

			if appErr.Type != tt.expectedType {
				t.Errorf("expected error type %v, got %v", tt.expectedType, appErr.Type)
			}

			if appErr.Error() == "" {
				t.Error("error message should not be empty")
			}

			// Test friendly display
			display := DisplayError(err)
			if display == "" {
				t.Error("display should not be empty")
			}

			// Test compact display
			compact := DisplayErrorCompact(err)
			if compact == "" {
				t.Error("compact display should not be empty")
			}

			// Test error code
			code := GetErrorCode(err)
			if code == "UNKNOWN_ERROR" {
				t.Error("should have a specific error code")
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	originalErr := fmt.Errorf("original error")

	wrappedErr := WrapError("TestOperation", originalErr)

	appErr, ok := wrappedErr.(*AppError)
	if !ok {
		t.Error("wrapped error should be AppError")
		return
	}

	if appErr.Op != "TestOperation" {
		t.Errorf("expected operation 'TestOperation', got '%s'", appErr.Op)
	}

	if appErr.Unwrap() != originalErr {
		t.Error("should preserve original error")
	}
}

func TestErrorWithContext(t *testing.T) {
	ctx := map[string]interface{}{
		"field": "username",
		"value": "invalid@name",
	}

	err := NewErrorWithContext(ErrTypeValidation, "TestOp",
		fmt.Errorf("validation failed"), ctx)

	if len(err.Context) != 2 {
		t.Errorf("expected 2 context items, got %d", len(err.Context))
	}

	if err.Context["field"] != "username" {
		t.Errorf("expected context field 'username', got '%v'", err.Context["field"])
	}

	// Test display includes context
	display := DisplayError(err)
	if display == "" {
		t.Error("display should not be empty")
	}
}

func TestConvenienceContextFunctions(t *testing.T) {
	// Test field validation error
	fieldErr := NewValidationErrorWithField("TestOp", "email", "invalid email format")
	if fieldErr.Type != ErrTypeValidation {
		t.Error("should be validation error")
	}
	if fieldErr.Context["field"] != "email" {
		t.Error("should have field context")
	}

	// Test network error with URL
	networkErr := NewNetworkErrorWithURL("TestOp", "https://api.github.com",
		"request failed", fmt.Errorf("timeout"))
	if networkErr.Type != ErrTypeNetwork {
		t.Error("should be network error")
	}
	if networkErr.Context["url"] != "https://api.github.com" {
		t.Error("should have URL context")
	}

	// Test file system error with path
	fsErr := NewFileSystemErrorWithPath("TestOp", "/tmp/test.txt",
		fmt.Errorf("permission denied"))
	if fsErr.Type != ErrTypeFileSystem {
		t.Error("should be file system error")
	}
	if fsErr.Context["path"] != "/tmp/test.txt" {
		t.Error("should have path context")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "Network timeout error",
			err:       NewNetworkErrorMsg("Test", "connection timeout"),
			retryable: true,
		},
		{
			name:      "Network connection error",
			err:       NewNetworkErrorMsg("Test", "connection refused"),
			retryable: true,
		},
		{
			name:      "Validation error",
			err:       NewValidationErrorMsg("Test", "invalid input"),
			retryable: false,
		},
		{
			name:      "File system busy error",
			err:       NewFileSystemErrorMsg("Test", "device busy"),
			retryable: true,
		},
		{
			name:      "Auth error",
			err:       NewAuthErrorMsg("Test", "unauthorized"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsRetryable(tt.err) != tt.retryable {
				t.Errorf("expected retryable=%v for %s", tt.retryable, tt.name)
			}
		})
	}
}

func TestErrorTypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ErrorType
	}{
		{
			name:     "Auth error detection",
			errMsg:   "authentication failed: invalid token",
			expected: ErrTypeAuth,
		},
		{
			name:     "Network error detection",
			errMsg:   "network timeout occurred",
			expected: ErrTypeNetwork,
		},
		{
			name:     "File error detection",
			errMsg:   "file not found: /tmp/test.txt",
			expected: ErrTypeFileSystem,
		},
		{
			name:     "Parse error detection",
			errMsg:   "failed to parse YAML content",
			expected: ErrTypeParsing,
		},
		{
			name:     "Default validation",
			errMsg:   "some unknown error",
			expected: ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errMsg)
			detected := determineErrorType(err)
			if detected != tt.expected {
				t.Errorf("expected %v, got %v for message: %s",
					tt.expected, detected, tt.errMsg)
			}
		})
	}
}

func TestDisplayFormatting(t *testing.T) {
	err := NewValidationErrorWithField("TestValidation", "email",
		"email format is invalid")

	display := DisplayError(err)

	// Should contain emoji and type
	if !containsString(display, "✅") {
		t.Error("should contain validation emoji")
	}

	// Should contain suggestion
	if !containsString(display, "💡 Suggestion:") {
		t.Error("should contain suggestion")
	}

	// Should contain context
	if !containsString(display, "field: email") {
		t.Error("should contain field context")
	}
}

// Helper function for string containment check
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
