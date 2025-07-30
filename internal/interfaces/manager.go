package interfaces

import (
	"context"
)

// Manager defines the base interface for all managers
type Manager interface {
	// Initialize prepares the manager for use
	Initialize(ctx context.Context) error

	// Cleanup performs any necessary cleanup
	Cleanup() error

	// IsInitialized checks if the manager has been initialized
	IsInitialized() bool
}
