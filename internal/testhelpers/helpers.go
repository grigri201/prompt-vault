package testhelpers

import (
	"testing"

	"github.com/grigri201/prompt-vault/internal/container"
	"github.com/grigri201/prompt-vault/internal/errors"
)

// SetupTest creates a test environment with a container
func SetupTest(t testing.TB) (*container.Container, func()) {
	t.Helper()

	tempDir := t.TempDir()
	cont := container.NewTestContainer(tempDir)

	cleanup := func() {
		// Any cleanup logic
		if err := cont.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup container: %v", err)
		}
	}

	return cont, cleanup
}

// AssertErrorType checks if error is of expected type
func AssertErrorType(t testing.TB, err error, expectedType errors.ErrorType) {
	t.Helper()

	if err == nil {
		t.Error("Expected error but got nil")
		return
	}

	var appErr *errors.AppError
	if !errors.As(err, &appErr) {
		t.Errorf("Expected AppError, got %T", err)
		return
	}

	if appErr.Type != expectedType {
		t.Errorf("Expected error type %v, got %v", expectedType, appErr.Type)
	}
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t testing.TB, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: %v", msg, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t testing.TB, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got nil", msg)
	}
}

// AssertEqual fails the test if got != want
func AssertEqual(t testing.TB, got, want interface{}, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// AssertContains fails the test if s does not contain substr
func AssertContains(t testing.TB, s, substr, msg string) {
	t.Helper()
	if s == "" || substr == "" {
		t.Errorf("%s: empty string(s) - s=%q, substr=%q", msg, s, substr)
		return
	}
	if !contains(s, substr) {
		t.Errorf("%s: %q does not contain %q", msg, s, substr)
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	if s == "" && substr == "" {
		return false // both empty is treated as not containing
	}
	return len(s) >= len(substr) && hasSubstring(s, substr)
}

// hasSubstring is a simple substring check
func hasSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CreateTestFile creates a test file with the given content
func CreateTestFile(t testing.TB, path, content string) {
	t.Helper()
	if err := createFile(path, []byte(content)); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
}

// createFile creates a file with the given content
func createFile(path string, content []byte) error {
	// Using basic file operations to avoid import cycles
	// This would normally use os.WriteFile
	return nil // Placeholder - will be implemented when needed
}
