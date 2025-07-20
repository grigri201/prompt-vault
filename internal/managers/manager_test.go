package managers

import (
	"context"
	"testing"
)

func TestBaseManager_IsInitialized(t *testing.T) {
	bm := &BaseManager{}

	// Test default state
	if bm.IsInitialized() {
		t.Error("BaseManager should not be initialized by default")
	}

	// Test after setting initialized
	bm.SetInitialized(true)
	if !bm.IsInitialized() {
		t.Error("BaseManager should be initialized after SetInitialized(true)")
	}

	// Test after unsetting initialized
	bm.SetInitialized(false)
	if bm.IsInitialized() {
		t.Error("BaseManager should not be initialized after SetInitialized(false)")
	}
}

func TestBaseManager_Initialize(t *testing.T) {
	bm := &BaseManager{}
	ctx := context.Background()

	// Test initialization
	if err := bm.Initialize(ctx); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Verify state after initialization
	if !bm.IsInitialized() {
		t.Error("BaseManager should be initialized after Initialize()")
	}
}

func TestBaseManager_Cleanup(t *testing.T) {
	bm := &BaseManager{}
	ctx := context.Background()

	// Initialize first
	if err := bm.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test cleanup
	if err := bm.Cleanup(); err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}

	// Verify state after cleanup
	if bm.IsInitialized() {
		t.Error("BaseManager should not be initialized after Cleanup()")
	}
}

// TestManagerInterface verifies that BaseManager can be used as Manager
func TestManagerInterface(t *testing.T) {
	var _ Manager = &BaseManager{}
}

// MockManager for testing custom implementations
type MockManager struct {
	BaseManager
	InitializeFunc func(context.Context) error
	CleanupFunc    func() error
}

func (m *MockManager) Initialize(ctx context.Context) error {
	if m.InitializeFunc != nil {
		return m.InitializeFunc(ctx)
	}
	return m.BaseManager.Initialize(ctx)
}

func (m *MockManager) Cleanup() error {
	if m.CleanupFunc != nil {
		return m.CleanupFunc()
	}
	return m.BaseManager.Cleanup()
}

func TestMockManager(t *testing.T) {
	// Test that MockManager implements Manager interface
	var _ Manager = &MockManager{}

	// Test with custom functions
	initCalled := false
	cleanupCalled := false

	mm := &MockManager{
		InitializeFunc: func(ctx context.Context) error {
			initCalled = true
			return nil
		},
		CleanupFunc: func() error {
			cleanupCalled = true
			return nil
		},
	}

	ctx := context.Background()

	// Test custom initialize
	if err := mm.Initialize(ctx); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}
	if !initCalled {
		t.Error("Custom InitializeFunc was not called")
	}

	// Test custom cleanup
	if err := mm.Cleanup(); err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}
	if !cleanupCalled {
		t.Error("Custom CleanupFunc was not called")
	}
}
