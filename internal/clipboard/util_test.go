package clipboard

import (
	"errors"
	"testing"
)

// MockClipboard is a mock implementation for testing clipboard operations
// without depending on system clipboard availability
type MockClipboard struct {
	content   string
	available bool
	copyError error
	readError error
}

func NewMockClipboard() *MockClipboard {
	return &MockClipboard{
		available: true, // Default to available
	}
}

func (m *MockClipboard) WriteAll(content string) error {
	if m.copyError != nil {
		return m.copyError
	}
	m.content = content
	return nil
}

func (m *MockClipboard) ReadAll() (string, error) {
	if m.readError != nil {
		return "", m.readError
	}
	return m.content, nil
}

func (m *MockClipboard) SetCopyError(err error) {
	m.copyError = err
}

func (m *MockClipboard) SetReadError(err error) {
	m.readError = err
}

func (m *MockClipboard) SetAvailable(available bool) {
	m.available = available
	if !available {
		m.SetReadError(errors.New("clipboard not available"))
	} else {
		m.SetReadError(nil)
	}
}

// mockUtil is a util implementation that uses MockClipboard for testing
type mockUtil struct {
	clipboard *MockClipboard
}

func newMockUtil(mockClipboard *MockClipboard) Util {
	return &mockUtil{clipboard: mockClipboard}
}

func (u *mockUtil) Copy(content string) error {
	return u.clipboard.WriteAll(content)
}

func (u *mockUtil) IsAvailable() bool {
	_, err := u.clipboard.ReadAll()
	return err == nil
}

func TestNewUtil(t *testing.T) {
	testCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "creates non-nil util",
			testFunc: func(t *testing.T) {
				util := NewUtil()
				if util == nil {
					t.Error("Expected non-nil util, got nil")
				}
			},
		},
		{
			name: "implements clipboard interface",
			testFunc: func(t *testing.T) {
				util := NewUtil()
				
				// Verify it implements the interface methods
				var _ Util = util
				
				// Test method existence by calling them (error handling tested separately)
				_ = util.Copy("")
				_ = util.IsAvailable()
			},
		},
		{
			name: "returns concrete util type",
			testFunc: func(t *testing.T) {
				util := NewUtil()
				// Just verify we got a non-nil implementation
				// The concrete type is internal and may change
				if util == nil {
					t.Error("Expected non-nil util implementation")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestUtil_Copy(t *testing.T) {
	testCases := []struct {
		name          string
		content       string
		setupMock     func(*MockClipboard)
		expectError   bool
		expectedError string
	}{
		{
			name:    "successful copy with normal content",
			content: "Hello, world!",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			expectError: false,
		},
		{
			name:    "successful copy with empty content",
			content: "",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			expectError: false,
		},
		{
			name:    "successful copy with unicode content",
			content: "你好世界 🌍 Hello émojis 🚀",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			expectError: false,
		},
		{
			name:    "successful copy with multiline content",
			content: "Line 1\nLine 2\r\nLine 3\rLine 4",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			expectError: false,
		},
		{
			name:    "successful copy with special characters",
			content: "Special chars: !@#$%^&*()[]{}|;':\",./<>?`~",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			expectError: false,
		},
		{
			name:    "successful copy with large content",
			content: generateLargeContent(10000), // 10KB content
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			expectError: false,
		},
		{
			name:    "copy fails when clipboard unavailable",
			content: "test content",
			setupMock: func(m *MockClipboard) {
				m.SetCopyError(errors.New("clipboard not available"))
			},
			expectError:   true,
			expectedError: "clipboard not available",
		},
		{
			name:    "copy fails with permission error",
			content: "test content",
			setupMock: func(m *MockClipboard) {
				m.SetCopyError(errors.New("permission denied"))
			},
			expectError:   true,
			expectedError: "permission denied",
		},
		{
			name:    "copy fails with system error",
			content: "test content",
			setupMock: func(m *MockClipboard) {
				m.SetCopyError(errors.New("system error: clipboard service unavailable"))
			},
			expectError:   true,
			expectedError: "system error: clipboard service unavailable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClipboard := NewMockClipboard()
			tc.setupMock(mockClipboard)
			
			util := newMockUtil(mockClipboard)
			err := util.Copy(tc.content)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tc.expectedError != "" && err.Error() != tc.expectedError {
					t.Errorf("Expected error %q, got %q", tc.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				
				// Verify content was actually copied
				if mockClipboard.content != tc.content {
					t.Errorf("Expected clipboard content %q, got %q", tc.content, mockClipboard.content)
				}
			}
		})
	}
}

func TestUtil_IsAvailable(t *testing.T) {
	testCases := []struct {
		name             string
		setupMock        func(*MockClipboard)
		expectedAvailable bool
	}{
		{
			name: "clipboard available - normal case",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			expectedAvailable: true,
		},
		{
			name: "clipboard available - with existing content",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
				m.content = "existing content"
			},
			expectedAvailable: true,
		},
		{
			name: "clipboard available - empty content",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
				m.content = ""
			},
			expectedAvailable: true,
		},
		{
			name: "clipboard unavailable - not accessible",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(false)
			},
			expectedAvailable: false,
		},
		{
			name: "clipboard unavailable - system error",
			setupMock: func(m *MockClipboard) {
				m.SetReadError(errors.New("system clipboard error"))
			},
			expectedAvailable: false,
		},
		{
			name: "clipboard unavailable - permission denied",
			setupMock: func(m *MockClipboard) {
				m.SetReadError(errors.New("permission denied"))
			},
			expectedAvailable: false,
		},
		{
			name: "clipboard unavailable - service not running",
			setupMock: func(m *MockClipboard) {
				m.SetReadError(errors.New("clipboard service not running"))
			},
			expectedAvailable: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClipboard := NewMockClipboard()
			tc.setupMock(mockClipboard)
			
			util := newMockUtil(mockClipboard)
			available := util.IsAvailable()

			if available != tc.expectedAvailable {
				t.Errorf("Expected IsAvailable() to return %v, got %v", tc.expectedAvailable, available)
			}
		})
	}
}

func TestUtil_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "copy and check availability repeatedly",
			testFunc: func(t *testing.T) {
				mockClipboard := NewMockClipboard()
				mockClipboard.SetAvailable(true)
				util := newMockUtil(mockClipboard)

				// Multiple operations should work consistently
				for i := 0; i < 10; i++ {
					if !util.IsAvailable() {
						t.Errorf("Expected clipboard to be available on iteration %d", i)
					}
					
					content := "test content " + string(rune('0'+i))
					if err := util.Copy(content); err != nil {
						t.Errorf("Unexpected error on iteration %d: %v", i, err)
					}
					
					if mockClipboard.content != content {
						t.Errorf("Expected content %q on iteration %d, got %q", content, i, mockClipboard.content)
					}
				}
			},
		},
		{
			name: "availability check after copy failure",
			testFunc: func(t *testing.T) {
				mockClipboard := NewMockClipboard()
				util := newMockUtil(mockClipboard)

				// Initially available
				mockClipboard.SetAvailable(true)
				if !util.IsAvailable() {
					t.Error("Expected clipboard to be initially available")
				}

				// Copy fails but availability might still be true (different operations)
				mockClipboard.SetCopyError(errors.New("copy failed"))
				err := util.Copy("test")
				if err == nil {
					t.Error("Expected copy to fail")
				}

				// IsAvailable should still work (it uses read operation)
				if !util.IsAvailable() {
					t.Error("Expected IsAvailable to still return true after copy failure")
				}
			},
		},
		{
			name: "copy with nil content handling",
			testFunc: func(t *testing.T) {
				mockClipboard := NewMockClipboard()
				mockClipboard.SetAvailable(true)
				util := newMockUtil(mockClipboard)

				// Empty string should be handled gracefully
				err := util.Copy("")
				if err != nil {
					t.Errorf("Unexpected error copying empty string: %v", err)
				}

				if mockClipboard.content != "" {
					t.Errorf("Expected empty content, got %q", mockClipboard.content)
				}
			},
		},
		{
			name: "very large content copy",
			testFunc: func(t *testing.T) {
				mockClipboard := NewMockClipboard()
				mockClipboard.SetAvailable(true)
				util := newMockUtil(mockClipboard)

				// Test with 1MB content
				largeContent := generateLargeContent(1024 * 1024)
				err := util.Copy(largeContent)
				if err != nil {
					t.Errorf("Unexpected error copying large content: %v", err)
				}

				if mockClipboard.content != largeContent {
					t.Error("Large content was not copied correctly")
				}

				if len(mockClipboard.content) != len(largeContent) {
					t.Errorf("Expected content length %d, got %d", len(largeContent), len(mockClipboard.content))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestUtil_RealClipboard_Integration(t *testing.T) {
	// These tests use the real clipboard implementation
	// They may fail in headless environments (CI/CD)
	// Skip if clipboard is not available
	
	realUtil := NewUtil()
	if !realUtil.IsAvailable() {
		t.Skip("Skipping real clipboard tests - clipboard not available in this environment")
	}

	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "real clipboard - simple text",
			content: "Hello, real clipboard!",
		},
		{
			name:    "real clipboard - empty content",
			content: "",
		},
		{
			name:    "real clipboard - unicode content",
			content: "测试内容 🚀 émojis",
		},
		{
			name:    "real clipboard - multiline content",
			content: "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test availability
			if !realUtil.IsAvailable() {
				t.Error("Expected real clipboard to be available")
			}

			// Test copy - we can't verify content without external dependencies
			// But we can verify the operation doesn't fail
			err := realUtil.Copy(tc.content)
			if err != nil {
				t.Errorf("Unexpected error copying to real clipboard: %v", err)
			}

			// Test availability again after copy
			if !realUtil.IsAvailable() {
				t.Error("Expected real clipboard to remain available after copy")
			}
		})
	}
}

func TestUtil_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(*MockClipboard)
		testFunc  func(t *testing.T, util Util)
	}{
		{
			name: "handle intermittent clipboard errors",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			testFunc: func(t *testing.T, util Util) {
				mockClipboard := util.(*mockUtil).clipboard

				// First operation succeeds
				if err := util.Copy("test1"); err != nil {
					t.Errorf("First copy should succeed: %v", err)
				}

				// Simulate intermittent failure
				mockClipboard.SetCopyError(errors.New("temporary failure"))
				if err := util.Copy("test2"); err == nil {
					t.Error("Expected copy to fail with temporary failure")
				}

				// Recovery - operation succeeds again
				mockClipboard.SetCopyError(nil)
				if err := util.Copy("test3"); err != nil {
					t.Errorf("Copy should succeed after recovery: %v", err)
				}
			},
		},
		{
			name: "handle availability check errors",
			setupMock: func(m *MockClipboard) {
				m.SetAvailable(true)
			},
			testFunc: func(t *testing.T, util Util) {
				mockClipboard := util.(*mockUtil).clipboard

				// Initially available
				if !util.IsAvailable() {
					t.Error("Expected clipboard to be initially available")
				}

				// Simulate read error
				mockClipboard.SetReadError(errors.New("read error"))
				if util.IsAvailable() {
					t.Error("Expected clipboard to be unavailable after read error")
				}

				// Recovery
				mockClipboard.SetReadError(nil)
				if !util.IsAvailable() {
					t.Error("Expected clipboard to be available after recovery")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClipboard := NewMockClipboard()
			tc.setupMock(mockClipboard)
			util := newMockUtil(mockClipboard)
			tc.testFunc(t, util)
		})
	}
}

func TestUtil_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to clipboard utility
	mockClipboard := NewMockClipboard()
	mockClipboard.SetAvailable(true)
	util := newMockUtil(mockClipboard)

	// Simple concurrency test - multiple goroutines using clipboard
	done := make(chan bool, 5)
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Test availability
			if !util.IsAvailable() {
				t.Errorf("Goroutine %d: Expected clipboard to be available", id)
			}
			
			// Test copy
			content := "test content from goroutine " + string(rune('0'+id))
			if err := util.Copy(content); err != nil {
				t.Errorf("Goroutine %d: Unexpected error: %v", id, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

// Helper function to generate large content for testing
func generateLargeContent(size int) string {
	content := make([]byte, size)
	for i := range content {
		content[i] = byte('A' + (i % 26))
	}
	return string(content)
}

// Benchmark tests
func BenchmarkUtil_Copy_SmallContent(b *testing.B) {
	mockClipboard := NewMockClipboard()
	mockClipboard.SetAvailable(true)
	util := newMockUtil(mockClipboard)
	content := "Small test content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		util.Copy(content)
	}
}

func BenchmarkUtil_Copy_LargeContent(b *testing.B) {
	mockClipboard := NewMockClipboard()
	mockClipboard.SetAvailable(true)
	util := newMockUtil(mockClipboard)
	content := generateLargeContent(10000) // 10KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		util.Copy(content)
	}
}

func BenchmarkUtil_IsAvailable(b *testing.B) {
	mockClipboard := NewMockClipboard()
	mockClipboard.SetAvailable(true)
	util := newMockUtil(mockClipboard)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		util.IsAvailable()
	}
}