package clipboard

import (
	"github.com/atotto/clipboard"
)

// Util defines the interface for clipboard operations.
// It provides cross-platform clipboard functionality for copying content
// and checking clipboard availability in the system.
type Util interface {
	// Copy copies the provided content to the system clipboard.
	// Returns an error if the clipboard operation fails or if the clipboard
	// is not accessible. The content parameter should contain the text to be copied.
	Copy(content string) error

	// IsAvailable checks if the system clipboard is available and accessible.
	// Returns true if clipboard operations can be performed, false otherwise.
	// This method can be used to verify clipboard availability before attempting
	// copy operations, allowing for graceful fallback when clipboard is unavailable.
	IsAvailable() bool
}

// util is the concrete implementation of the Util interface.
// It provides cross-platform clipboard operations using the atotto/clipboard library.
type util struct{}

// NewUtil creates a new instance of clipboard utility.
// This factory function returns a concrete implementation that can perform
// clipboard operations across different operating systems (Windows, macOS, Linux).
func NewUtil() Util {
	return &util{}
}

// Copy implements the Util interface by copying content to the system clipboard.
// It uses the github.com/atotto/clipboard library to handle cross-platform
// clipboard access. The method returns an error if the clipboard operation fails,
// such as when the clipboard is not accessible or the system doesn't support
// clipboard operations.
//
// Parameters:
//   - content: The text content to be copied to clipboard
//
// Returns:
//   - error: nil if successful, or an error describing the failure
func (u *util) Copy(content string) error {
	return clipboard.WriteAll(content)
}

// IsAvailable implements the Util interface by checking clipboard availability.
// It attempts to perform a test clipboard operation to determine if the system
// clipboard is accessible and functional. This method can be used to verify
// clipboard availability before attempting copy operations.
//
// Returns:
//   - bool: true if clipboard is available and accessible, false otherwise
func (u *util) IsAvailable() bool {
	// Test clipboard availability by attempting to read from it
	// If this operation succeeds, the clipboard is available
	_, err := clipboard.ReadAll()
	return err == nil
}
