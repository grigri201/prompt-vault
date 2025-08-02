package auth

import (
	"github.com/grigri/pv/internal/errors"
)

// TokenValidator defines the interface for validating GitHub tokens
type TokenValidator interface {
	// Validate checks if the token is valid and has required permissions
	Validate(token string) (*ValidationResult, error)
}

// ValidationResult contains the result of token validation
type ValidationResult struct {
	// IsValid indicates if the token is valid
	IsValid bool

	// HasGistScope indicates if the token has the required 'gist' scope
	HasGistScope bool

	// User contains the authenticated user information
	User *User

	// Error contains any error message
	Error string
}

// tokenValidator implements the TokenValidator interface
type tokenValidator struct {
	githubClient GitHubClient
}

// NewTokenValidator creates a new token validator
func NewTokenValidator(githubClient GitHubClient) TokenValidator {
	return &tokenValidator{
		githubClient: githubClient,
	}
}

// Validate checks if the token is valid and has required permissions
func (v *tokenValidator) Validate(token string) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:      false,
		HasGistScope: false,
	}

	// Validate token by checking scopes
	scopes, err := v.githubClient.ValidateScopes(token)
	if err != nil {
		if err == errors.ErrInvalidToken {
			result.Error = "Invalid token"
			return result, nil
		}
		return nil, err
	}

	// Check if token has gist scope
	for _, scope := range scopes {
		if scope == "gist" {
			result.HasGistScope = true
			break
		}
	}

	if !result.HasGistScope {
		result.Error = "Token lacks required 'gist' scope"
		return result, nil
	}

	// Get user information
	user, err := v.githubClient.GetAuthenticatedUser(token)
	if err != nil {
		if err == errors.ErrInvalidToken {
			result.Error = "Invalid token"
			return result, nil
		}
		return nil, err
	}

	result.IsValid = true
	result.User = user

	return result, nil
}
