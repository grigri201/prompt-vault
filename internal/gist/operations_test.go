package gist

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v73/github"
)

// TestClient wraps a Client and allows overriding specific methods for testing
type TestClient struct {
	*Client

	// Override functions
	updateGistFunc   func(ctx context.Context, gistID, name, description, content string) (string, error)
	createGistFunc   func(ctx context.Context, name, description, content string) (string, string, error)
	createPublicFunc func(ctx context.Context, name, description, content string) (string, string, error)
	deleteGistFunc   func(ctx context.Context, gistID string) error
	getGistFunc      func(ctx context.Context, gistID string) (*github.Gist, error)
	getAPIErrorFunc  func(err error) *github.ErrorResponse
	isRateLimitFunc  func(err error) bool

	// Track calls
	updateCalls int
	createCalls int
	deleteCalls int
	getCalls    int
}

func (t *TestClient) UpdateGist(ctx context.Context, gistID, name, description, content string) (string, error) {
	t.updateCalls++
	if t.updateGistFunc != nil {
		return t.updateGistFunc(ctx, gistID, name, description, content)
	}
	return "", fmt.Errorf("not implemented")
}

func (t *TestClient) CreateGist(ctx context.Context, name, description, content string) (string, string, error) {
	t.createCalls++
	if t.createGistFunc != nil {
		return t.createGistFunc(ctx, name, description, content)
	}
	return "", "", fmt.Errorf("not implemented")
}

func (t *TestClient) CreatePublicGist(ctx context.Context, name, description, content string) (string, string, error) {
	t.createCalls++
	if t.createPublicFunc != nil {
		return t.createPublicFunc(ctx, name, description, content)
	}
	return "", "", fmt.Errorf("not implemented")
}

func (t *TestClient) DeleteGist(ctx context.Context, gistID string) error {
	t.deleteCalls++
	if t.deleteGistFunc != nil {
		return t.deleteGistFunc(ctx, gistID)
	}
	return fmt.Errorf("not implemented")
}

func (t *TestClient) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	t.getCalls++
	if t.getGistFunc != nil {
		return t.getGistFunc(ctx, gistID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (t *TestClient) GetAPIError(err error) *github.ErrorResponse {
	if t.getAPIErrorFunc != nil {
		return t.getAPIErrorFunc(err)
	}
	if errResp, ok := err.(*github.ErrorResponse); ok {
		return errResp
	}
	return nil
}

func (t *TestClient) IsRateLimitError(err error) bool {
	if t.isRateLimitFunc != nil {
		return t.isRateLimitFunc(err)
	}
	return false
}

// mockLogger is a mock implementation of Logger for testing
type mockLogger struct {
	debugMessages []string
	infoMessages  []string
	errorMessages []string
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(msg, args...))
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(msg, args...))
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(msg, args...))
}

func TestNewGistOperations(t *testing.T) {
	tests := []struct {
		name      string
		config    GistOperationsConfig
		wantRetry int
	}{
		{
			name: "default retry count",
			config: GistOperationsConfig{
				Client: &Client{},
			},
			wantRetry: 3,
		},
		{
			name: "custom retry count",
			config: GistOperationsConfig{
				Client:     &Client{},
				RetryCount: 5,
			},
			wantRetry: 5,
		},
		{
			name: "zero retry count uses default",
			config: GistOperationsConfig{
				Client:     &Client{},
				RetryCount: 0,
			},
			wantRetry: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewGistOperations(tt.config)
			if ops.retryCount != tt.wantRetry {
				t.Errorf("NewGistOperations() retryCount = %v, want %v", ops.retryCount, tt.wantRetry)
			}
		})
	}
}

func TestGistOperations_CreateOrUpdate(t *testing.T) {
	tests := []struct {
		name        string
		gistID      string
		data        *GistData
		setupClient func() *TestClient
		wantCreated bool
		wantErr     bool
		errContains string
	}{
		{
			name:   "create new gist when no ID provided",
			gistID: "",
			data: &GistData{
				Name:    "test",
				Content: "content",
			},
			setupClient: func() *TestClient {
				return &TestClient{
					createGistFunc: func(ctx context.Context, name, description, content string) (string, string, error) {
						return "new-gist-id", "https://gist.github.com/new-gist-id", nil
					},
				}
			},
			wantCreated: true,
			wantErr:     false,
		},
		{
			name:   "update existing gist",
			gistID: "existing-id",
			data: &GistData{
				Name:    "test",
				Content: "content",
			},
			setupClient: func() *TestClient {
				return &TestClient{
					updateGistFunc: func(ctx context.Context, gistID, name, description, content string) (string, error) {
						return "https://gist.github.com/" + gistID, nil
					},
				}
			},
			wantCreated: false,
			wantErr:     false,
		},
		{
			name:   "create when update returns 404",
			gistID: "non-existent-id",
			data: &GistData{
				Name:    "test",
				Content: "content",
			},
			setupClient: func() *TestClient {
				return &TestClient{
					updateGistFunc: func(ctx context.Context, gistID, name, description, content string) (string, error) {
						return "", &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusNotFound},
							Message:  "Not Found",
						}
					},
					createGistFunc: func(ctx context.Context, name, description, content string) (string, string, error) {
						return "new-gist-id", "https://gist.github.com/new-gist-id", nil
					},
					getAPIErrorFunc: func(err error) *github.ErrorResponse {
						if errResp, ok := err.(*github.ErrorResponse); ok {
							return errResp
						}
						return nil
					},
				}
			},
			wantCreated: true,
			wantErr:     false,
		},
		{
			name:   "nil data returns error",
			gistID: "id",
			data:   nil,
			setupClient: func() *TestClient {
				return &TestClient{}
			},
			wantErr:     true,
			errContains: "gist data is required",
		},
		{
			name:   "empty name returns error",
			gistID: "id",
			data: &GistData{
				Content: "content",
			},
			setupClient: func() *TestClient {
				return &TestClient{}
			},
			wantErr:     true,
			errContains: "name and content are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			logger := &mockLogger{}

			ops := &GistOperations{
				client:     client,
				retryCount: 3,
				logger:     logger,
			}

			result, err := ops.CreateOrUpdate(context.Background(), tt.gistID, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateOrUpdate() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("CreateOrUpdate() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateOrUpdate() unexpected error = %v", err)
				return
			}

			if result.Created != tt.wantCreated {
				t.Errorf("CreateOrUpdate() Created = %v, want %v", result.Created, tt.wantCreated)
			}
		})
	}
}

func TestGistOperations_CreateOrUpdateWithRetry(t *testing.T) {
	tests := []struct {
		name        string
		gistID      string
		data        *GistData
		setupClient func() *TestClient
		wantErr     bool
		wantCalls   int
	}{
		{
			name:   "success on first try",
			gistID: "test-id",
			data: &GistData{
				Name:    "test",
				Content: "content",
			},
			setupClient: func() *TestClient {
				return &TestClient{
					updateGistFunc: func(ctx context.Context, gistID, name, description, content string) (string, error) {
						return "https://gist.github.com/" + gistID, nil
					},
				}
			},
			wantErr:   false,
			wantCalls: 1,
		},
		{
			name:   "retry on rate limit error",
			gistID: "test-id",
			data: &GistData{
				Name:    "test",
				Content: "content",
			},
			setupClient: func() *TestClient {
				callCount := 0
				return &TestClient{
					updateGistFunc: func(ctx context.Context, gistID, name, description, content string) (string, error) {
						callCount++
						if callCount < 3 {
							return "", &github.ErrorResponse{
								Response: &http.Response{
									StatusCode: http.StatusForbidden,
									Header:     http.Header{"X-RateLimit-Remaining": []string{"0"}},
								},
								Message: "rate limit exceeded",
							}
						}
						return "https://gist.github.com/" + gistID, nil
					},
					isRateLimitFunc: func(err error) bool {
						if errResp, ok := err.(*github.ErrorResponse); ok {
							if errResp.Response != nil && errResp.Response.StatusCode == http.StatusForbidden {
								remaining := errResp.Response.Header.Get("X-RateLimit-Remaining")
								return remaining == "0"
							}
						}
						return false
					},
				}
			},
			wantErr:   false,
			wantCalls: 3,
		},
		{
			name:   "don't retry validation errors",
			gistID: "test-id",
			data:   nil,
			setupClient: func() *TestClient {
				return &TestClient{}
			},
			wantErr:   true,
			wantCalls: 0, // CreateOrUpdate will be called 0 times because validation happens before
		},
		{
			name:   "retry on 500 error",
			gistID: "test-id",
			data: &GistData{
				Name:    "test",
				Content: "content",
			},
			setupClient: func() *TestClient {
				callCount := 0
				return &TestClient{
					updateGistFunc: func(ctx context.Context, gistID, name, description, content string) (string, error) {
						callCount++
						if callCount < 2 {
							return "", &github.ErrorResponse{
								Response: &http.Response{StatusCode: http.StatusInternalServerError},
								Message:  "Internal Server Error",
							}
						}
						return "https://gist.github.com/" + gistID, nil
					},
					getAPIErrorFunc: func(err error) *github.ErrorResponse {
						if errResp, ok := err.(*github.ErrorResponse); ok {
							return errResp
						}
						return nil
					},
				}
			},
			wantErr:   false,
			wantCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			logger := &mockLogger{}

			ops := &GistOperations{
				client:     client,
				retryCount: 3,
				logger:     logger,
			}

			_, err := ops.CreateOrUpdateWithRetry(context.Background(), tt.gistID, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateOrUpdateWithRetry() error = nil, wantErr = true")
				}
			} else if err != nil {
				t.Errorf("CreateOrUpdateWithRetry() unexpected error = %v", err)
			}

			if client.updateCalls != tt.wantCalls {
				t.Errorf("CreateOrUpdateWithRetry() calls = %v, want %v", client.updateCalls, tt.wantCalls)
			}
		})
	}
}

func TestGistOperations_DeleteSafely(t *testing.T) {
	tests := []struct {
		name        string
		gistID      string
		setupClient func() *TestClient
		wantErr     bool
		errContains string
	}{
		{
			name:   "successful delete",
			gistID: "test-id",
			setupClient: func() *TestClient {
				return &TestClient{
					deleteGistFunc: func(ctx context.Context, gistID string) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "ignore 404 error",
			gistID: "test-id",
			setupClient: func() *TestClient {
				return &TestClient{
					deleteGistFunc: func(ctx context.Context, gistID string) error {
						return &github.ErrorResponse{
							Response: &http.Response{StatusCode: http.StatusNotFound},
							Message:  "Not Found",
						}
					},
					getAPIErrorFunc: func(err error) *github.ErrorResponse {
						if errResp, ok := err.(*github.ErrorResponse); ok {
							return errResp
						}
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "empty gist ID returns error",
			gistID: "",
			setupClient: func() *TestClient {
				return &TestClient{}
			},
			wantErr:     true,
			errContains: "gist ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			logger := &mockLogger{}

			ops := &GistOperations{
				client:     client,
				retryCount: 3,
				logger:     logger,
			}

			err := ops.DeleteSafely(context.Background(), tt.gistID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteSafely() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("DeleteSafely() error = %v, want error containing %v", err, tt.errContains)
				}
			} else if err != nil {
				t.Errorf("DeleteSafely() unexpected error = %v", err)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
