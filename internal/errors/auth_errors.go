package errors

// Authentication related errors
var (
	// ErrInvalidToken is returned when the provided token is invalid
	ErrInvalidToken = AppError{
		Type:    ErrAuth,
		Message: "Invalid GitHub token",
	}

	// ErrMissingScope is returned when the token lacks required 'gist' scope
	ErrMissingScope = AppError{
		Type:    ErrAuth,
		Message: "Token lacks required 'gist' scope",
	}

	// ErrTokenNotFound is returned when no authentication token is found
	ErrTokenNotFound = AppError{
		Type:    ErrAuth,
		Message: "No authentication token found. Please run 'pv auth login' first",
	}

	// ErrGitHubAPIUnavailable is returned when GitHub API is unreachable
	ErrGitHubAPIUnavailable = AppError{
		Type:    ErrNetwork,
		Message: "Unable to connect to GitHub API. Please check your internet connection",
	}

	// ErrTokenSaveFailed is returned when token cannot be saved
	ErrTokenSaveFailed = AppError{
		Type:    ErrStorage,
		Message: "Failed to save authentication token",
	}

	// ErrTokenLoadFailed is returned when token cannot be loaded
	ErrTokenLoadFailed = AppError{
		Type:    ErrStorage,
		Message: "Failed to load authentication token",
	}
)
