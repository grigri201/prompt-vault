package errors

import (
	"errors"
	"testing"
)

func TestGetShareErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "auth error",
			err:  NewAuthError("share", errors.New("no token")),
			want: ErrMsgAuthRequired,
		},
		{
			name: "network error",
			err:  NewNetworkError("share", errors.New("timeout")),
			want: ErrMsgShareNetworkError,
		},
		{
			name: "already public error",
			err:  errors.New("gist is already public"),
			want: ErrMsgSharePublicGist,
		},
		{
			name: "not found error",
			err:  errors.New("gist not found"),
			want: ErrMsgShareGistNotFound,
		},
		{
			name: "YAML error",
			err:  errors.New("no YAML file found"),
			want: ErrMsgShareNoYAMLFile,
		},
		{
			name: "cancelled error",
			err:  errors.New("operation cancelled by user"),
			want: ErrMsgShareCancelled,
		},
		{
			name: "generic error",
			err:  errors.New("something went wrong"),
			want: "Share failed: something went wrong",
		},
		{
			name: "app error with custom message",
			err: &AppError{
				Type:    ErrTypeValidation,
				Op:      "share",
				Err:     errors.New("validation failed"),
				Message: "Custom validation error",
			},
			want: "Custom validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetShareErrorMessage(tt.err)
			if got != tt.want {
				t.Errorf("GetShareErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetImportErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "auth error",
			err:  NewAuthError("import", errors.New("no token")),
			want: ErrMsgAuthRequired,
		},
		{
			name: "network error",
			err:  NewNetworkError("import", errors.New("timeout")),
			want: ErrMsgNetworkTimeout,
		},
		{
			name: "validation error",
			err:  NewValidationError("import", errors.New("invalid format")),
			want: ErrMsgImportNoValidPrompt,
		},
		{
			name: "invalid URL error",
			err:  errors.New("invalid URL format"),
			want: ErrMsgImportInvalidURL,
		},
		{
			name: "not GitHub URL error",
			err:  errors.New("not a GitHub gist URL"),
			want: ErrMsgImportNotGitHubURL,
		},
		{
			name: "not found error",
			err:  errors.New("gist not found"),
			want: ErrMsgImportGistNotFound,
		},
		{
			name: "private gist error",
			err:  errors.New("cannot import private gist"),
			want: ErrMsgImportPrivateGist,
		},
		{
			name: "valid prompt error",
			err:  errors.New("not a valid prompt file"),
			want: ErrMsgImportNoValidPrompt,
		},
		{
			name: "required fields error",
			err:  errors.New("missing required fields"),
			want: ErrMsgImportMissingFields,
		},
		{
			name: "cancelled error",
			err:  errors.New("import cancelled"),
			want: ErrMsgImportCancelled,
		},
		{
			name: "generic error",
			err:  errors.New("unknown error"),
			want: "Import failed: unknown error",
		},
		{
			name: "app error with custom message",
			err: &AppError{
				Type:    ErrTypeParsing,
				Op:      "import",
				Err:     errors.New("parse failed"),
				Message: "Custom parse error",
			},
			want: "Custom parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetImportErrorMessage(tt.err)
			if got != tt.want {
				t.Errorf("GetImportErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{
			name:   "exact match",
			s:      "hello",
			substr: "hello",
			want:   true,
		},
		{
			name:   "substring match",
			s:      "hello world",
			substr: "world",
			want:   true,
		},
		{
			name:   "case insensitive match",
			s:      "Hello World",
			substr: "hello",
			want:   true,
		},
		{
			name:   "no match",
			s:      "hello",
			substr: "goodbye",
			want:   false,
		},
		{
			name:   "empty substring",
			s:      "hello",
			substr: "",
			want:   true,
		},
		{
			name:   "empty string",
			s:      "",
			substr: "hello",
			want:   false,
		},
		{
			name:   "substring longer than string",
			s:      "hi",
			substr: "hello",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestWrapWithMessage(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		msg     string
		wantMsg string
		wantNil bool
	}{
		{
			name:    "nil error",
			err:     nil,
			msg:     "custom message",
			wantNil: true,
		},
		{
			name:    "regular error",
			err:     errors.New("original error"),
			msg:     "custom message",
			wantMsg: "custom message: original error",
		},
		{
			name: "app error",
			err: &AppError{
				Type:    ErrTypeAuth,
				Op:      "login",
				Err:     errors.New("auth failed"),
				Message: "original message",
			},
			msg:     "new message",
			wantMsg: "new message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapWithMessage(tt.err, tt.msg)
			if tt.wantNil {
				if got != nil {
					t.Errorf("WrapWithMessage() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("WrapWithMessage() = nil, want non-nil")
			}
			if got.Error() != tt.wantMsg {
				t.Errorf("WrapWithMessage() error = %v, want %v", got.Error(), tt.wantMsg)
			}
			
			// For AppError, verify type is preserved
			if appErr, ok := tt.err.(*AppError); ok {
				var gotAppErr *AppError
				if As(got, &gotAppErr) {
					if gotAppErr.Type != appErr.Type {
						t.Errorf("WrapWithMessage() preserved type = %v, want %v", gotAppErr.Type, appErr.Type)
					}
				} else {
					t.Error("WrapWithMessage() did not preserve AppError type")
				}
			}
		})
	}
}

func TestMessageFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(string, string) *AppError
		op       string
		msg      string
		wantType ErrorType
		wantMsg  string
	}{
		{
			name:     "NewAuthErrorMsg",
			function: NewAuthErrorMsg,
			op:       "login",
			msg:      "Invalid credentials",
			wantType: ErrTypeAuth,
			wantMsg:  "Invalid credentials",
		},
		{
			name:     "NewValidationErrorMsg",
			function: NewValidationErrorMsg,
			op:       "validate",
			msg:      "Missing required field",
			wantType: ErrTypeValidation,
			wantMsg:  "Missing required field",
		},
		{
			name:     "NewFileSystemErrorMsg",
			function: NewFileSystemErrorMsg,
			op:       "read",
			msg:      "File not found",
			wantType: ErrTypeFileSystem,
			wantMsg:  "File not found",
		},
		{
			name:     "NewParsingErrorMsg",
			function: NewParsingErrorMsg,
			op:       "parse",
			msg:      "Invalid syntax",
			wantType: ErrTypeParsing,
			wantMsg:  "Invalid syntax",
		},
		{
			name:     "NewNetworkErrorMsg",
			function: NewNetworkErrorMsg,
			op:       "fetch",
			msg:      "Connection timeout",
			wantType: ErrTypeNetwork,
			wantMsg:  "Connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.function(tt.op, tt.msg)
			if err.Type != tt.wantType {
				t.Errorf("%s() type = %v, want %v", tt.name, err.Type, tt.wantType)
			}
			if err.Op != tt.op {
				t.Errorf("%s() op = %v, want %v", tt.name, err.Op, tt.op)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("%s() message = %v, want %v", tt.name, err.Message, tt.wantMsg)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("%s() Error() = %v, want %v", tt.name, err.Error(), tt.wantMsg)
			}
		})
	}
}