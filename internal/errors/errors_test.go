package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{
			name: "with custom message",
			err: &AppError{
				Type:    ErrTypeAuth,
				Op:      "login",
				Err:     errors.New("invalid token"),
				Message: "custom error message",
			},
			want: "custom error message",
		},
		{
			name: "without custom message",
			err: &AppError{
				Type: ErrTypeNetwork,
				Op:   "fetch",
				Err:  errors.New("timeout"),
			},
			want: "fetch: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	appErr := &AppError{
		Type: ErrTypeFileSystem,
		Op:   "read",
		Err:  baseErr,
	}

	if got := appErr.Unwrap(); got != baseErr {
		t.Errorf("AppError.Unwrap() = %v, want %v", got, baseErr)
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := &AppError{Type: ErrTypeAuth}
	err2 := &AppError{Type: ErrTypeAuth}
	err3 := &AppError{Type: ErrTypeNetwork}

	if !err1.Is(err2) {
		t.Error("Expected errors of same type to match")
	}

	if err1.Is(err3) {
		t.Error("Expected errors of different types not to match")
	}

	if err1.Is(errors.New("other error")) {
		t.Error("Expected AppError not to match non-AppError")
	}
}

func TestConstructors(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name     string
		fn       func(string, error) *AppError
		op       string
		err      error
		wantType ErrorType
		wantMsg  string
	}{
		{
			name:     "NewAuthError",
			fn:       NewAuthError,
			op:       "login",
			err:      baseErr,
			wantType: ErrTypeAuth,
			wantMsg:  "authentication failed during login",
		},
		{
			name:     "NewNetworkError",
			fn:       NewNetworkError,
			op:       "request",
			err:      baseErr,
			wantType: ErrTypeNetwork,
			wantMsg:  "network error during request",
		},
		{
			name:     "NewFileSystemError",
			fn:       NewFileSystemError,
			op:       "write",
			err:      baseErr,
			wantType: ErrTypeFileSystem,
			wantMsg:  "file system error during write",
		},
		{
			name:     "NewValidationError",
			fn:       NewValidationError,
			op:       "validate",
			err:      baseErr,
			wantType: ErrTypeValidation,
			wantMsg:  "validation failed during validate",
		},
		{
			name:     "NewParsingError",
			fn:       NewParsingError,
			op:       "parse",
			err:      baseErr,
			wantType: ErrTypeParsing,
			wantMsg:  "parsing failed during parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.op, tt.err)

			if got.Type != tt.wantType {
				t.Errorf("got Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Op != tt.op {
				t.Errorf("got Op = %v, want %v", got.Op, tt.op)
			}
			if got.Err != tt.err {
				t.Errorf("got Err = %v, want %v", got.Err, tt.err)
			}
			if got.Message != tt.wantMsg {
				t.Errorf("got Message = %v, want %v", got.Message, tt.wantMsg)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		err     error
		want    string
		wantNil bool
	}{
		{
			name:    "nil error",
			op:      "test",
			err:     nil,
			wantNil: true,
		},
		{
			name: "non-nil error",
			op:   "operation",
			err:  errors.New("failed"),
			want: "operation: failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.op, tt.err)

			if tt.wantNil {
				if got != nil {
					t.Errorf("Wrap() = %v, want nil", got)
				}
				return
			}

			if got.Error() != tt.want {
				t.Errorf("Wrap() = %v, want %v", got.Error(), tt.want)
			}
		})
	}
}

func TestIsType(t *testing.T) {
	authErr := NewAuthError("test", errors.New("auth failed"))
	netErr := NewNetworkError("test", errors.New("network failed"))
	wrappedAuthErr := fmt.Errorf("wrapped: %w", authErr)
	normalErr := errors.New("normal error")

	tests := []struct {
		name    string
		err     error
		errType ErrorType
		want    bool
	}{
		{
			name:    "direct auth error",
			err:     authErr,
			errType: ErrTypeAuth,
			want:    true,
		},
		{
			name:    "auth error checking for network type",
			err:     authErr,
			errType: ErrTypeNetwork,
			want:    false,
		},
		{
			name:    "network error",
			err:     netErr,
			errType: ErrTypeNetwork,
			want:    true,
		},
		{
			name:    "wrapped auth error",
			err:     wrappedAuthErr,
			errType: ErrTypeAuth,
			want:    true,
		},
		{
			name:    "normal error",
			err:     normalErr,
			errType: ErrTypeAuth,
			want:    false,
		},
		{
			name:    "nil error",
			err:     nil,
			errType: ErrTypeAuth,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsType(tt.err, tt.errType); got != tt.want {
				t.Errorf("IsType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAs(t *testing.T) {
	authErr := NewAuthError("test", errors.New("auth failed"))
	wrappedErr := fmt.Errorf("wrapped: %w", authErr)
	normalErr := errors.New("normal error")

	tests := []struct {
		name    string
		err     error
		want    bool
		wantErr *AppError
	}{
		{
			name:    "direct AppError",
			err:     authErr,
			want:    true,
			wantErr: authErr,
		},
		{
			name:    "wrapped AppError",
			err:     wrappedErr,
			want:    true,
			wantErr: authErr,
		},
		{
			name: "normal error",
			err:  normalErr,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target *AppError
			got := As(tt.err, &target)

			if got != tt.want {
				t.Errorf("As() = %v, want %v", got, tt.want)
			}

			if tt.want && target != tt.wantErr {
				t.Errorf("As() target = %v, want %v", target, tt.wantErr)
			}
		})
	}
}

func TestAs_NilTarget(t *testing.T) {
	err := NewAuthError("test", errors.New("failed"))

	if As(err, nil) {
		t.Error("As() with nil target should return false")
	}
}
