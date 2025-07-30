package errors

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name         string
		op           string
		err          error
		expectedType ErrorType
	}{
		{
			name:         "auth error detection",
			op:           "login",
			err:          errors.New("authentication failed"),
			expectedType: ErrTypeAuth,
		},
		{
			name:         "token error detection",
			op:           "validate",
			err:          errors.New("invalid token"),
			expectedType: ErrTypeAuth,
		},
		{
			name:         "network error detection",
			op:           "fetch",
			err:          errors.New("connection timeout"),
			expectedType: ErrTypeNetwork,
		},
		{
			name:         "http error detection",
			op:           "request",
			err:          errors.New("HTTP 500 error"),
			expectedType: ErrTypeNetwork,
		},
		{
			name:         "file error detection",
			op:           "read",
			err:          errors.New("file not found"),
			expectedType: ErrTypeFileSystem,
		},
		{
			name:         "directory error detection",
			op:           "create",
			err:          errors.New("directory already exists"),
			expectedType: ErrTypeFileSystem,
		},
		{
			name:         "parsing error detection",
			op:           "parse",
			err:          errors.New("invalid YAML syntax"),
			expectedType: ErrTypeParsing,
		},
		{
			name:         "unmarshal error detection",
			op:           "decode",
			err:          errors.New("cannot unmarshal data"),
			expectedType: ErrTypeParsing,
		},
		{
			name:         "default to validation",
			op:           "validate",
			err:          errors.New("value is required"),
			expectedType: ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := NewError(tt.op, tt.err)

			if appErr.Type != tt.expectedType {
				t.Errorf("NewError() type = %v, want %v", appErr.Type, tt.expectedType)
			}

			if appErr.Op != tt.op {
				t.Errorf("NewError() op = %v, want %v", appErr.Op, tt.op)
			}

			if appErr.Err != tt.err {
				t.Errorf("NewError() err = %v, want %v", appErr.Err, tt.err)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name         string
		op           string
		err          error
		checkType    bool
		expectedType ErrorType
	}{
		{
			name: "nil error",
			op:   "test",
			err:  nil,
		},
		{
			name:         "wrap regular error",
			op:           "operation",
			err:          errors.New("some error"),
			checkType:    true,
			expectedType: ErrTypeValidation,
		},
		{
			name:         "wrap AppError preserves type",
			op:           "outer",
			err:          NewAuthError("inner", errors.New("auth failed")),
			checkType:    true,
			expectedType: ErrTypeAuth,
		},
		{
			name:         "wrap network error",
			op:           "http-call",
			err:          errors.New("dial tcp: connection refused"),
			checkType:    true,
			expectedType: ErrTypeNetwork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(tt.op, tt.err)

			if tt.err == nil {
				if wrapped != nil {
					t.Errorf("WrapError() = %v, want nil", wrapped)
				}
				return
			}

			var appErr *AppError
			if !As(wrapped, &appErr) {
				t.Fatal("WrapError() should return an AppError")
			}

			if appErr.Op != tt.op {
				t.Errorf("WrapError() op = %v, want %v", appErr.Op, tt.op)
			}

			if tt.checkType && appErr.Type != tt.expectedType {
				t.Errorf("WrapError() type = %v, want %v", appErr.Type, tt.expectedType)
			}
		})
	}
}

func TestDetermineErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		// Auth errors
		{"auth keyword", errors.New("authentication required"), ErrTypeAuth},
		{"token keyword", errors.New("invalid token provided"), ErrTypeAuth},
		{"permission keyword", errors.New("permission denied"), ErrTypeAuth},
		{"forbidden keyword", errors.New("access forbidden"), ErrTypeAuth},
		{"unauthorized keyword", errors.New("unauthorized access"), ErrTypeAuth},

		// Network errors
		{"network keyword", errors.New("network unreachable"), ErrTypeNetwork},
		{"connection keyword", errors.New("connection refused"), ErrTypeNetwork},
		{"timeout keyword", errors.New("request timeout"), ErrTypeNetwork},
		{"dial keyword", errors.New("dial tcp failed"), ErrTypeNetwork},
		{"http keyword", errors.New("HTTP 404"), ErrTypeNetwork},
		{"tls keyword", errors.New("TLS handshake failed"), ErrTypeNetwork},
		{"rate limit keyword", errors.New("rate limit exceeded"), ErrTypeNetwork},

		// File system errors
		{"file keyword", errors.New("file not found"), ErrTypeFileSystem},
		{"directory keyword", errors.New("directory does not exist"), ErrTypeFileSystem},
		{"path keyword", errors.New("invalid path"), ErrTypeFileSystem},
		{"no such keyword", errors.New("no such file"), ErrTypeFileSystem},
		{"exist keyword", errors.New("already exists"), ErrTypeFileSystem},

		// Parsing errors
		{"parse keyword", errors.New("failed to parse"), ErrTypeParsing},
		{"unmarshal keyword", errors.New("cannot unmarshal"), ErrTypeParsing},
		{"yaml keyword", errors.New("invalid YAML"), ErrTypeParsing},
		{"json keyword", errors.New("invalid JSON"), ErrTypeParsing},
		{"syntax keyword", errors.New("syntax error"), ErrTypeParsing},
		{"invalid format keyword", errors.New("invalid format provided"), ErrTypeParsing},

		// Default validation
		{"unknown error", errors.New("something went wrong"), ErrTypeValidation},
		{"nil error", nil, ErrTypeValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineErrorType(tt.err)
			if result != tt.expected {
				t.Errorf("determineErrorType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStandardizedErrorConsistency(t *testing.T) {
	// Test that NewError and constructor functions can work together
	err1 := NewError("test", errors.New("auth failed"))
	err2 := NewAuthError("test", errors.New("auth failed"))

	// Both should be auth errors
	if !IsType(err1, ErrTypeAuth) {
		t.Error("NewError with auth message should create auth error")
	}
	if !IsType(err2, ErrTypeAuth) {
		t.Error("NewAuthError should create auth error")
	}

	// Test wrapping preserves type
	wrapped1 := WrapError("outer", err1)
	wrapped2 := WrapError("outer", err2)

	if !IsType(wrapped1, ErrTypeAuth) {
		t.Error("WrapError should preserve auth type")
	}
	if !IsType(wrapped2, ErrTypeAuth) {
		t.Error("WrapError should preserve auth type")
	}
}
