package service

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Login authenticates with the provided GitHub token
	Login(token string) error

	// GetStatus returns the current authentication status
	GetStatus() (*AuthStatus, error)

	// Logout clears the stored authentication
	Logout() error
}

// AuthStatus represents the current authentication status
type AuthStatus struct {
	// IsAuthenticated indicates if the user is authenticated
	IsAuthenticated bool

	// Username is the GitHub username (empty if not authenticated)
	Username string

	// Email is the user's email address (empty if not authenticated)
	Email string
}
