package managers

import "context"

// Manager defines common manager operations
type Manager interface {
	// Initialize prepares the manager
	Initialize(ctx context.Context) error

	// Cleanup performs cleanup operations
	Cleanup() error

	// IsInitialized checks if manager is ready
	IsInitialized() bool
}

// BaseManager provides common functionality for managers
type BaseManager struct {
	initialized bool
}

// IsInitialized returns whether the manager is initialized
func (bm *BaseManager) IsInitialized() bool {
	return bm.initialized
}

// SetInitialized sets the initialization state
func (bm *BaseManager) SetInitialized(v bool) {
	bm.initialized = v
}

// Initialize provides default initialization (can be overridden)
func (bm *BaseManager) Initialize(ctx context.Context) error {
	bm.SetInitialized(true)
	return nil
}

// Cleanup provides default cleanup (can be overridden)
func (bm *BaseManager) Cleanup() error {
	bm.SetInitialized(false)
	return nil
}
