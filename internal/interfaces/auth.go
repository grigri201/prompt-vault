package interfaces

import (
	"context"
)

// AuthReader defines read operations for authentication
type AuthReader interface {
	// GetToken retrieves the stored authentication token
	GetToken() (string, error)

	// GetUsername retrieves the authenticated username
	GetUsername() (string, error)

	// IsAuthenticated checks if the user is authenticated
	IsAuthenticated() bool
}

// AuthWriter defines write operations for authentication
type AuthWriter interface {
	// SaveToken saves the authentication token
	SaveToken(token string) error

	// SaveUsername saves the authenticated username
	SaveUsername(username string) error

	// ClearAuthentication removes stored authentication data
	ClearAuthentication() error
}

// AuthManager combines read and write operations with authentication logic
type AuthManager interface {
	AuthReader
	AuthWriter

	// AuthenticateWithToken validates a token and stores authentication data
	AuthenticateWithToken(ctx context.Context, token string) (string, error)

	// Initialize prepares the auth manager for use
	Initialize(ctx context.Context) error

	// Cleanup performs any necessary cleanup
	Cleanup() error
}
