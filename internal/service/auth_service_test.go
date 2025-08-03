package service

import (
	"testing"

	"github.com/grigri/pv/internal/auth"
	"github.com/grigri/pv/internal/errors"
)

// MockConfigStore is a mock implementation of config.Store for testing
type MockConfigStore struct {
	savedToken  string
	getError    error
	saveError   error
	deleteError error
}

func (m *MockConfigStore) SaveToken(token string) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.savedToken = token
	return nil
}

func (m *MockConfigStore) GetToken() (string, error) {
	if m.getError != nil {
		return "", m.getError
	}
	if m.savedToken == "" {
		return "", errors.ErrTokenNotFound
	}
	return m.savedToken, nil
}

func (m *MockConfigStore) DeleteToken() error {
	if m.deleteError != nil {
		return m.deleteError
	}
	m.savedToken = ""
	return nil
}

func (m *MockConfigStore) GetConfigPath() string {
	return "/mock/config/path"
}

func TestAuthService_Login(t *testing.T) {
	testCases := []struct {
		name             string
		token            string
		validationResult *auth.ValidationResult
		validationError  error
		saveError        error
		expectError      bool
		expectedError    error
	}{
		{
			name:  "successful login with valid token",
			token: "valid_token",
			validationResult: &auth.ValidationResult{
				IsValid:      true,
				HasGistScope: true,
				User: &auth.User{
					Login: "testuser",
					Email: "test@example.com",
				},
			},
			expectError: false,
		},
		{
			name:  "login with invalid token",
			token: "invalid_token",
			validationResult: &auth.ValidationResult{
				IsValid: false,
				Error:   "Invalid token",
			},
			expectError:   true,
			expectedError: errors.ErrInvalidToken,
		},
		{
			name:  "login with token missing gist scope",
			token: "no_gist_scope_token",
			validationResult: &auth.ValidationResult{
				IsValid:      true,
				HasGistScope: false,
				Error:        "Token lacks required 'gist' scope",
			},
			expectError:   true,
			expectedError: errors.ErrMissingScope,
		},
		{
			name:            "validation error",
			token:           "token",
			validationError: errors.ErrGitHubAPIUnavailable,
			expectError:     true,
			expectedError:   errors.ErrGitHubAPIUnavailable,
		},
		{
			name:  "save token error",
			token: "valid_token",
			validationResult: &auth.ValidationResult{
				IsValid:      true,
				HasGistScope: true,
				User: &auth.User{
					Login: "testuser",
				},
			},
			saveError:   errors.ErrTokenSaveFailed,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockConfigStore := &MockConfigStore{
				saveError: tc.saveError,
			}

			mockGitHubClient := &auth.MockGitHubClient{}

			mockTokenValidator := &MockTokenValidator{
				validateFunc: func(token string) (*auth.ValidationResult, error) {
					if tc.validationError != nil {
						return nil, tc.validationError
					}
					return tc.validationResult, nil
				},
			}

			// Create service
			service := NewAuthService(mockConfigStore, mockGitHubClient, mockTokenValidator)

			// Test login
			err := service.Login(tc.token)

			// Check error
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.expectedError != nil && err != tc.expectedError {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check token was saved
				if mockConfigStore.savedToken != tc.token {
					t.Errorf("Token not saved correctly: got %q, want %q", mockConfigStore.savedToken, tc.token)
				}
			}
		})
	}
}

func TestAuthService_GetStatus(t *testing.T) {
	testCases := []struct {
		name             string
		storedToken      string
		getTokenError    error
		validationResult *auth.ValidationResult
		validationError  error
		expectedStatus   *AuthStatus
		expectError      bool
	}{
		{
			name:        "authenticated user",
			storedToken: "valid_token",
			validationResult: &auth.ValidationResult{
				IsValid:      true,
				HasGistScope: true,
				User: &auth.User{
					Login: "testuser",
					Email: "test@example.com",
				},
			},
			expectedStatus: &AuthStatus{
				IsAuthenticated: true,
				Username:        "testuser",
				Email:           "test@example.com",
			},
		},
		{
			name:          "no stored token",
			getTokenError: errors.ErrTokenNotFound,
			expectedStatus: &AuthStatus{
				IsAuthenticated: false,
			},
		},
		{
			name:        "invalid stored token",
			storedToken: "invalid_token",
			validationResult: &auth.ValidationResult{
				IsValid: false,
			},
			expectedStatus: &AuthStatus{
				IsAuthenticated: false,
			},
		},
		{
			name:            "network error during validation",
			storedToken:     "some_token",
			validationError: errors.ErrGitHubAPIUnavailable,
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockConfigStore := &MockConfigStore{
				savedToken: tc.storedToken,
				getError:   tc.getTokenError,
			}

			mockGitHubClient := &auth.MockGitHubClient{}

			mockTokenValidator := &MockTokenValidator{
				validateFunc: func(token string) (*auth.ValidationResult, error) {
					if tc.validationError != nil {
						return nil, tc.validationError
					}
					return tc.validationResult, nil
				},
			}

			// Create service
			service := NewAuthService(mockConfigStore, mockGitHubClient, mockTokenValidator)

			// Test get status
			status, err := service.GetStatus()

			// Check error
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check status
				if status.IsAuthenticated != tc.expectedStatus.IsAuthenticated {
					t.Errorf("IsAuthenticated mismatch: got %v, want %v",
						status.IsAuthenticated, tc.expectedStatus.IsAuthenticated)
				}
				if status.Username != tc.expectedStatus.Username {
					t.Errorf("Username mismatch: got %q, want %q",
						status.Username, tc.expectedStatus.Username)
				}
				if status.Email != tc.expectedStatus.Email {
					t.Errorf("Email mismatch: got %q, want %q",
						status.Email, tc.expectedStatus.Email)
				}
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	testCases := []struct {
		name        string
		deleteError error
		expectError bool
	}{
		{
			name:        "successful logout",
			expectError: false,
		},
		{
			name:        "delete error",
			deleteError: errors.ErrTokenSaveFailed,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockConfigStore := &MockConfigStore{
				savedToken:  "some_token",
				deleteError: tc.deleteError,
			}

			mockGitHubClient := &auth.MockGitHubClient{}
			mockTokenValidator := &MockTokenValidator{}

			// Create service
			service := NewAuthService(mockConfigStore, mockGitHubClient, mockTokenValidator)

			// Test logout
			err := service.Logout()

			// Check error
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check token was deleted
				if mockConfigStore.savedToken != "" {
					t.Errorf("Token not deleted: still has value %q", mockConfigStore.savedToken)
				}
			}
		})
	}
}

func TestGetLoginMessage(t *testing.T) {
	username := "testuser"
	expected := "✓ Logged in as testuser"

	result := GetLoginMessage(username)
	if result != expected {
		t.Errorf("GetLoginMessage() = %q, want %q", result, expected)
	}
}

func TestGetStatusMessage(t *testing.T) {
	testCases := []struct {
		name     string
		status   *AuthStatus
		expected string
	}{
		{
			name: "authenticated with email",
			status: &AuthStatus{
				IsAuthenticated: true,
				Username:        "testuser",
				Email:           "test@example.com",
			},
			expected: "✓ Authenticated as testuser (test@example.com)",
		},
		{
			name: "authenticated without email",
			status: &AuthStatus{
				IsAuthenticated: true,
				Username:        "testuser",
			},
			expected: "✓ Authenticated as testuser",
		},
		{
			name: "not authenticated",
			status: &AuthStatus{
				IsAuthenticated: false,
			},
			expected: "✗ Not authenticated. Run 'pv auth login' to authenticate.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetStatusMessage(tc.status)
			if result != tc.expected {
				t.Errorf("GetStatusMessage() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestGetLogoutMessage(t *testing.T) {
	expected := "✓ Successfully logged out"

	result := GetLogoutMessage()
	if result != expected {
		t.Errorf("GetLogoutMessage() = %q, want %q", result, expected)
	}
}

// MockTokenValidator is a mock implementation of auth.TokenValidator
type MockTokenValidator struct {
	validateFunc func(token string) (*auth.ValidationResult, error)
}

func (m *MockTokenValidator) Validate(token string) (*auth.ValidationResult, error) {
	if m.validateFunc != nil {
		return m.validateFunc(token)
	}
	return &auth.ValidationResult{IsValid: false}, nil
}
