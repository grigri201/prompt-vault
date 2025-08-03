package auth

// MockGitHubClient is a mock implementation of GitHubClient for testing
type MockGitHubClient struct {
	// GetAuthenticatedUserFunc allows customizing the behavior of GetAuthenticatedUser
	GetAuthenticatedUserFunc func(token string) (*User, error)

	// ValidateScopesFunc allows customizing the behavior of ValidateScopes
	ValidateScopesFunc func(token string) ([]string, error)

	// Calls tracks the method calls for verification
	Calls struct {
		GetAuthenticatedUser []string // tokens passed
		ValidateScopes       []string // tokens passed
	}
}

// NewMockGitHubClient creates a new mock GitHub client
func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		GetAuthenticatedUserFunc: func(token string) (*User, error) {
			// Default implementation
			if token == "valid_token" {
				return &User{
					Login: "testuser",
					Email: "test@example.com",
					Name:  "Test User",
				}, nil
			}
			return nil, nil
		},
		ValidateScopesFunc: func(token string) ([]string, error) {
			// Default implementation
			if token == "valid_token" {
				return []string{"gist", "repo"}, nil
			}
			return []string{}, nil
		},
	}
}

// GetAuthenticatedUser implements GitHubClient
func (m *MockGitHubClient) GetAuthenticatedUser(token string) (*User, error) {
	m.Calls.GetAuthenticatedUser = append(m.Calls.GetAuthenticatedUser, token)
	if m.GetAuthenticatedUserFunc != nil {
		return m.GetAuthenticatedUserFunc(token)
	}
	return nil, nil
}

// ValidateScopes implements GitHubClient
func (m *MockGitHubClient) ValidateScopes(token string) ([]string, error) {
	m.Calls.ValidateScopes = append(m.Calls.ValidateScopes, token)
	if m.ValidateScopesFunc != nil {
		return m.ValidateScopesFunc(token)
	}
	return []string{}, nil
}

// Reset clears all recorded calls
func (m *MockGitHubClient) Reset() {
	m.Calls.GetAuthenticatedUser = nil
	m.Calls.ValidateScopes = nil
}

// Verify interface compliance
var _ GitHubClient = (*MockGitHubClient)(nil)
