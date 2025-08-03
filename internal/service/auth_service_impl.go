package service

import (
	"fmt"

	"github.com/grigri/pv/internal/auth"
	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/errors"
)

// authService implements the AuthService interface
type authService struct {
	configStore    config.Store
	githubClient   auth.GitHubClient
	tokenValidator auth.TokenValidator
}

// NewAuthService creates a new authentication service
func NewAuthService(
	configStore config.Store,
	githubClient auth.GitHubClient,
	tokenValidator auth.TokenValidator,
) AuthService {
	return &authService{
		configStore:    configStore,
		githubClient:   githubClient,
		tokenValidator: tokenValidator,
	}
}

// Login authenticates with the provided GitHub token
func (s *authService) Login(token string) error {
	// Validate the token
	result, err := s.tokenValidator.Validate(token)
	if err != nil {
		return err
	}

	if !result.IsValid {
		if result.Error == "Token lacks required 'gist' scope" {
			return errors.ErrMissingScope
		}
		return errors.ErrInvalidToken
	}

	if !result.HasGistScope {
		return errors.ErrMissingScope
	}

	// Save the token
	if err := s.configStore.SaveToken(token); err != nil {
		return errors.NewAppError(
			errors.ErrStorage,
			"failed to save authentication token",
			err,
		)
	}

	return nil
}

// GetStatus returns the current authentication status
func (s *authService) GetStatus() (*AuthStatus, error) {
	status := &AuthStatus{
		IsAuthenticated: false,
	}

	// Try to get the stored token
	token, err := s.configStore.GetToken()
	if err != nil {
		if err == errors.ErrTokenNotFound {
			// Not authenticated
			return status, nil
		}
		return nil, err
	}

	// Validate the stored token
	result, err := s.tokenValidator.Validate(token)
	if err != nil {
		// If validation fails due to network issues, return error
		if err == errors.ErrGitHubAPIUnavailable {
			return nil, err
		}
		// For other errors, assume not authenticated
		return status, nil
	}

	if result.IsValid && result.User != nil {
		status.IsAuthenticated = true
		status.Username = result.User.Login
		status.Email = result.User.Email
	}

	return status, nil
}

// Logout clears the stored authentication
func (s *authService) Logout() error {
	if err := s.configStore.DeleteToken(); err != nil {
		return errors.NewAppError(
			errors.ErrStorage,
			"failed to clear authentication",
			err,
		)
	}

	return nil
}

// GetLoginMessage returns a success message for login
func GetLoginMessage(username string) string {
	return fmt.Sprintf("✓ Logged in as %s", username)
}

// GetStatusMessage returns a formatted status message
func GetStatusMessage(status *AuthStatus) string {
	if !status.IsAuthenticated {
		return "✗ Not authenticated. Run 'pv auth login' to authenticate."
	}

	if status.Email != "" {
		return fmt.Sprintf("✓ Authenticated as %s (%s)", status.Username, status.Email)
	}
	return fmt.Sprintf("✓ Authenticated as %s", status.Username)
}

// GetLogoutMessage returns a success message for logout
func GetLogoutMessage() string {
	return "✓ Successfully logged out"
}
