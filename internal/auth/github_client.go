package auth

// GitHubClient defines the interface for GitHub API operations
type GitHubClient interface {
	// GetAuthenticatedUser retrieves the authenticated user information
	GetAuthenticatedUser(token string) (*User, error)

	// ValidateScopes checks if the token has the required scopes
	ValidateScopes(token string) ([]string, error)
}

// User represents a GitHub user
type User struct {
	// Login is the GitHub username
	Login string `json:"login"`

	// Email is the user's primary email address
	Email string `json:"email"`

	// Name is the user's display name
	Name string `json:"name"`
}
