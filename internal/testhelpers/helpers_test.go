package testhelpers

import (
	"fmt"
	"testing"

	pkgerrors "github.com/grigri201/prompt-vault/internal/errors"
)

func TestSetupTest(t *testing.T) {
	cont, cleanup := SetupTest(t)
	defer cleanup()

	if cont == nil {
		t.Error("SetupTest() returned nil container")
	}

	if cont.PathManager == nil {
		t.Error("Container PathManager is nil")
	}

	if cont.CacheManager == nil {
		t.Error("Container CacheManager is nil")
	}

	if cont.ConfigManager == nil {
		t.Error("Container ConfigManager is nil")
	}
}

func TestAssertErrorType(t *testing.T) {
	// Test with matching error type
	authErr := pkgerrors.NewAuthError("test", fmt.Errorf("base error"))

	// This should not fail
	AssertErrorType(t, authErr, pkgerrors.ErrTypeAuth)

	// For other tests, we'll use sub-tests that are expected to fail
	// and check that they do fail

	t.Run("nil error should fail", func(t *testing.T) {
		// We expect this to fail, so we'll check if it was reported
		innerT := &testRecorder{T: t}
		AssertErrorType(innerT, nil, pkgerrors.ErrTypeAuth)
		if !innerT.failed {
			t.Error("AssertErrorType should fail with nil error")
		}
	})

	t.Run("wrong error type should fail", func(t *testing.T) {
		innerT := &testRecorder{T: t}
		AssertErrorType(innerT, authErr, pkgerrors.ErrTypeNetwork)
		if !innerT.failed {
			t.Error("AssertErrorType should fail with wrong error type")
		}
	})

	t.Run("non-AppError should fail", func(t *testing.T) {
		innerT := &testRecorder{T: t}
		normalErr := fmt.Errorf("normal error")
		AssertErrorType(innerT, normalErr, pkgerrors.ErrTypeAuth)
		if !innerT.failed {
			t.Error("AssertErrorType should fail with non-AppError")
		}
	})
}

func TestAssertNoError(t *testing.T) {
	// Test with nil error (should pass)
	AssertNoError(t, nil, "test message")

	// Test with non-nil error (should fail)
	t.Run("non-nil error should fail", func(t *testing.T) {
		tr := &testRecorder{T: t}
		AssertNoError(tr, fmt.Errorf("error"), "test message")
		if !tr.failed {
			t.Error("AssertNoError should fail with non-nil error")
		}
	})
}

func TestAssertError(t *testing.T) {
	// Test with non-nil error (should pass)
	AssertError(t, fmt.Errorf("error"), "test message")

	// Test with nil error (should fail)
	t.Run("nil error should fail", func(t *testing.T) {
		tr := &testRecorder{T: t}
		AssertError(tr, nil, "test message")
		if !tr.failed {
			t.Error("AssertError should fail with nil error")
		}
	})
}

func TestAssertEqual(t *testing.T) {
	// Test with equal values
	AssertEqual(t, "test", "test", "test message")
	AssertEqual(t, 42, 42, "test message")

	// Test with different values
	t.Run("different values should fail", func(t *testing.T) {
		tr := &testRecorder{T: t}
		AssertEqual(tr, "test", "other", "test message")
		if !tr.failed {
			t.Error("AssertEqual should fail with different values")
		}
	})
}

func TestAssertContains(t *testing.T) {
	// Test when string contains substring
	AssertContains(t, "hello world", "world", "test message")

	// Test failure cases
	tests := []struct {
		name   string
		s      string
		substr string
	}{
		{"doesn't contain", "hello world", "foo"},
		{"empty string", "", "foo"},
		{"empty substring", "hello", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &testRecorder{T: t}
			AssertContains(tr, tt.s, tt.substr, "test message")
			if !tr.failed {
				t.Errorf("AssertContains should fail for %s", tt.name)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"contains at start", "hello world", "hello", true},
		{"contains in middle", "hello world", "lo wo", true},
		{"contains at end", "hello world", "world", true},
		{"doesn't contain", "hello world", "foo", false},
		{"empty substring", "hello", "", true},
		{"empty string", "", "foo", false},
		{"both empty", "", "", false},
		{"exact match", "hello", "hello", true},
		{"substring longer", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.expected)
			}
		})
	}
}

// testRecorder wraps a real testing.T and records if any failures occurred
type testRecorder struct {
	*testing.T
	failed bool
}

func (tr *testRecorder) Error(args ...interface{}) {
	tr.failed = true
	// Don't call the underlying T.Error to avoid test failure
}

func (tr *testRecorder) Errorf(format string, args ...interface{}) {
	tr.failed = true
	// Don't call the underlying T.Errorf to avoid test failure
}

func (tr *testRecorder) Fatal(args ...interface{}) {
	tr.failed = true
	// Don't call the underlying T.Fatal to avoid test termination
}

func (tr *testRecorder) Fatalf(format string, args ...interface{}) {
	tr.failed = true
	// Don't call the underlying T.Fatalf to avoid test termination
}
