package gist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/models"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "creates client with valid token",
			token:   "ghp_validtoken123",
			wantErr: false,
		},
		{
			name:    "rejects empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "rejects whitespace token",
			token:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestClient_ValidateToken(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func(w http.ResponseWriter, r *http.Request)
		wantUsername   string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: "validates token successfully",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/user" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}
				if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
					t.Errorf("Unexpected authorization header: %s", auth)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"login": "testuser", "id": 12345}`))
			},
			wantUsername: "testuser",
			wantErr:      false,
		},
		{
			name: "handles unauthorized error",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message": "Bad credentials"}`))
			},
			wantErr:        true,
			wantErrMessage: "authentication failed",
		},
		{
			name: "handles rate limit error",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "API rate limit exceeded"}`))
			},
			wantErr:        true,
			wantErrMessage: "network error",
		},
		{
			name: "handles server error",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "Internal server error"}`))
			},
			wantErr: true,
		},
		{
			name: "handles network error",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				// Close connection immediately to simulate network error
				hj, ok := w.(http.Hijacker)
				if !ok {
					t.Error("ResponseWriter doesn't support hijacking")
					return
				}
				conn, _, err := hj.Hijack()
				if err != nil {
					t.Error("Failed to hijack connection")
					return
				}
				conn.Close()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.setupServer))
			defer server.Close()

			// Create client with test server URL
			client := &Client{
				github: github.NewClient(nil).WithAuthToken("test-token"),
			}
			// Override the base URL to use test server
			client.github.BaseURL, _ = client.github.BaseURL.Parse(server.URL + "/")

			username, err := client.ValidateToken(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("ValidateToken() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}

			if !tt.wantErr && username != tt.wantUsername {
				t.Errorf("ValidateToken() username = %v, want %v", username, tt.wantUsername)
			}
		})
	}
}

func TestClient_IsRateLimitError(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func(w http.ResponseWriter, r *http.Request)
		wantRateLimit bool
	}{
		{
			name: "detects rate limit error with header",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "API rate limit exceeded"}`))
			},
			wantRateLimit: true,
		},
		{
			name: "detects rate limit error from message",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "You have exceeded a secondary rate limit"}`))
			},
			wantRateLimit: true,
		},
		{
			name: "does not detect non-rate-limit 403",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "Resource not accessible by integration"}`))
			},
			wantRateLimit: false,
		},
		{
			name: "does not detect other errors",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			},
			wantRateLimit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.setupServer))
			defer server.Close()

			// Create client with test server URL
			client := &Client{
				github: github.NewClient(nil).WithAuthToken("test-token"),
			}
			client.github.BaseURL, _ = client.github.BaseURL.Parse(server.URL + "/")

			// Make a request that will fail
			_, _, err := client.github.Users.Get(context.Background(), "")

			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			isRateLimit := client.IsRateLimitError(err)
			if isRateLimit != tt.wantRateLimit {
				t.Errorf("IsRateLimitError() = %v, want %v", isRateLimit, tt.wantRateLimit)
			}
		})
	}
}

func TestClient_GetAPIError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantMessage string
		wantNil     bool
	}{
		{
			name: "extracts GitHub API error",
			err: &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
				Message: "Not Found",
			},
			wantMessage: "Not Found",
			wantNil:     false,
		},
		{
			name:    "returns nil for non-API error",
			err:     http.ErrServerClosed,
			wantNil: true,
		},
		{
			name:    "returns nil for nil error",
			err:     nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			apiErr := client.GetAPIError(tt.err)

			if tt.wantNil {
				if apiErr != nil {
					t.Errorf("GetAPIError() = %v, want nil", apiErr)
				}
			} else {
				if apiErr == nil {
					t.Error("GetAPIError() = nil, want non-nil")
				} else if apiErr.Message != tt.wantMessage {
					t.Errorf("GetAPIError().Message = %v, want %v", apiErr.Message, tt.wantMessage)
				}
			}
		})
	}
}

func TestClient_CreateGist(t *testing.T) {
	tests := []struct {
		name           string
		gistName       string
		description    string
		content        string
		setupServer    func(w http.ResponseWriter, r *http.Request)
		wantGistID     string
		wantURL        string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:        "creates gist successfully",
			gistName:    "testuser-example",
			description: "Example prompt template",
			content:     "Hello {name}!",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/gists" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}
				if r.Method != "POST" {
					t.Errorf("Unexpected method: %s", r.Method)
				}

				// Verify request body
				var reqBody map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}

				// Check that gist is private
				if public, ok := reqBody["public"].(bool); !ok || public {
					t.Error("Gist should be private")
				}

				// Check description
				if desc, ok := reqBody["description"].(string); !ok || desc != "Example prompt template" {
					t.Errorf("Unexpected description: %v", desc)
				}

				// Check files
				files, ok := reqBody["files"].(map[string]interface{})
				if !ok {
					t.Error("Missing files in request")
				}

				file, ok := files["testuser-example.yaml"].(map[string]interface{})
				if !ok {
					t.Error("Missing expected file in request")
				}

				if content, ok := file["content"].(string); !ok || content != "Hello {name}!" {
					t.Errorf("Unexpected file content: %v", content)
				}

				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{
					"id": "abc123def456",
					"html_url": "https://gist.github.com/testuser/abc123def456",
					"files": {
						"testuser-example.yaml": {
							"filename": "testuser-example.yaml",
							"content": "Hello {name}!"
						}
					}
				}`))
			},
			wantGistID: "abc123def456",
			wantURL:    "https://gist.github.com/testuser/abc123def456",
			wantErr:    false,
		},
		{
			name:           "handles empty gist name",
			gistName:       "",
			description:    "Test",
			content:        "Test",
			wantErr:        true,
			wantErrMessage: "gist name is required",
		},
		{
			name:           "handles empty content",
			gistName:       "test",
			description:    "Test",
			content:        "",
			wantErr:        true,
			wantErrMessage: "content is required",
		},
		{
			name:        "handles API error",
			gistName:    "test",
			description: "Test",
			content:     "Test",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message": "Validation Failed"}`))
			},
			wantErr:        true,
			wantErrMessage: "failed to create gist",
		},
		{
			name:        "handles rate limit",
			gistName:    "test",
			description: "Test",
			content:     "Test",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "API rate limit exceeded"}`))
			},
			wantErr:        true,
			wantErrMessage: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.setupServer))
				defer server.Close()
			}

			client := &Client{
				github: github.NewClient(nil).WithAuthToken("test-token"),
			}

			if server != nil {
				client.github.BaseURL, _ = client.github.BaseURL.Parse(server.URL + "/")
			}

			gistID, url, err := client.CreateGist(context.Background(), tt.gistName, tt.description, tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGist() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("CreateGist() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}

			if !tt.wantErr {
				if gistID != tt.wantGistID {
					t.Errorf("CreateGist() gistID = %v, want %v", gistID, tt.wantGistID)
				}
				if url != tt.wantURL {
					t.Errorf("CreateGist() url = %v, want %v", url, tt.wantURL)
				}
			}
		})
	}
}

func TestClient_UpdateGist(t *testing.T) {
	tests := []struct {
		name           string
		gistID         string
		gistName       string
		description    string
		content        string
		setupServer    func(w http.ResponseWriter, r *http.Request)
		wantURL        string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:        "updates gist successfully",
			gistID:      "abc123def456",
			gistName:    "testuser-example",
			description: "Updated prompt template",
			content:     "Hello {name}, welcome!",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/gists/abc123def456" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}
				if r.Method != "PATCH" {
					t.Errorf("Unexpected method: %s", r.Method)
				}

				// Verify request body
				var reqBody map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}

				// Check description
				if desc, ok := reqBody["description"].(string); !ok || desc != "Updated prompt template" {
					t.Errorf("Unexpected description: %v", desc)
				}

				// Check files
				files, ok := reqBody["files"].(map[string]interface{})
				if !ok {
					t.Error("Missing files in request")
				}

				file, ok := files["testuser-example.yaml"].(map[string]interface{})
				if !ok {
					t.Error("Missing expected file in request")
				}

				if content, ok := file["content"].(string); !ok || content != "Hello {name}, welcome!" {
					t.Errorf("Unexpected file content: %v", content)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "abc123def456",
					"html_url": "https://gist.github.com/testuser/abc123def456",
					"files": {
						"testuser-example.yaml": {
							"filename": "testuser-example.yaml",
							"content": "Hello {name}, welcome!"
						}
					}
				}`))
			},
			wantURL: "https://gist.github.com/testuser/abc123def456",
			wantErr: false,
		},
		{
			name:           "handles empty gist ID",
			gistID:         "",
			gistName:       "test",
			description:    "Test",
			content:        "Test",
			wantErr:        true,
			wantErrMessage: "gist ID is required",
		},
		{
			name:        "handles not found error",
			gistID:      "nonexistent",
			gistName:    "test",
			description: "Test",
			content:     "Test",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			},
			wantErr:        true,
			wantErrMessage: "gist not found",
		},
		{
			name:        "handles permission error",
			gistID:      "forbidden",
			gistName:    "test",
			description: "Test",
			content:     "Test",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "Must have admin rights to Repository"}`))
			},
			wantErr:        true,
			wantErrMessage: "authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.setupServer))
				defer server.Close()
			}

			client := &Client{
				github: github.NewClient(nil).WithAuthToken("test-token"),
			}

			if server != nil {
				client.github.BaseURL, _ = client.github.BaseURL.Parse(server.URL + "/")
			}

			url, err := client.UpdateGist(context.Background(), tt.gistID, tt.gistName, tt.description, tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateGist() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("UpdateGist() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}

			if !tt.wantErr {
				if url != tt.wantURL {
					t.Errorf("UpdateGist() url = %v, want %v", url, tt.wantURL)
				}
			}
		})
	}
}

func TestClient_GetGist(t *testing.T) {
	tests := []struct {
		name           string
		gistID         string
		setupServer    func(w http.ResponseWriter, r *http.Request)
		wantGist       *github.Gist
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:   "retrieves gist successfully",
			gistID: "abc123def456",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/gists/abc123def456" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}
				if r.Method != "GET" {
					t.Errorf("Unexpected method: %s", r.Method)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "abc123def456",
					"html_url": "https://gist.github.com/testuser/abc123def456",
					"description": "Test prompt",
					"public": false,
					"files": {
						"testuser-example.yaml": {
							"filename": "testuser-example.yaml",
							"content": "Hello {name}!",
							"size": 13
						}
					},
					"owner": {
						"login": "testuser"
					}
				}`))
			},
			wantGist: &github.Gist{
				ID:          github.String("abc123def456"),
				HTMLURL:     github.String("https://gist.github.com/testuser/abc123def456"),
				Description: github.String("Test prompt"),
				Public:      github.Bool(false),
			},
			wantErr: false,
		},
		{
			name:           "handles empty gist ID",
			gistID:         "",
			wantErr:        true,
			wantErrMessage: "gist ID is required",
		},
		{
			name:   "handles not found error",
			gistID: "nonexistent",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			},
			wantErr:        true,
			wantErrMessage: "gist not found",
		},
		{
			name:   "handles rate limit",
			gistID: "test123",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "API rate limit exceeded"}`))
			},
			wantErr:        true,
			wantErrMessage: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.setupServer))
				defer server.Close()
			}

			client := &Client{
				github: github.NewClient(nil).WithAuthToken("test-token"),
			}

			if server != nil {
				client.github.BaseURL, _ = client.github.BaseURL.Parse(server.URL + "/")
			}

			gist, err := client.GetGist(context.Background(), tt.gistID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetGist() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("GetGist() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}

			if !tt.wantErr && gist != nil && tt.wantGist != nil {
				if gist.GetID() != tt.wantGist.GetID() {
					t.Errorf("GetGist() ID = %v, want %v", gist.GetID(), tt.wantGist.GetID())
				}
				if gist.GetHTMLURL() != tt.wantGist.GetHTMLURL() {
					t.Errorf("GetGist() HTMLURL = %v, want %v", gist.GetHTMLURL(), tt.wantGist.GetHTMLURL())
				}
			}
		})
	}
}

func TestClient_DeleteGist(t *testing.T) {
	tests := []struct {
		name           string
		gistID         string
		setupServer    func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:   "deletes gist successfully",
			gistID: "abc123def456",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/gists/abc123def456" {
					t.Errorf("Unexpected path: %s", r.URL.Path)
				}
				if r.Method != "DELETE" {
					t.Errorf("Unexpected method: %s", r.Method)
				}

				w.WriteHeader(http.StatusNoContent)
			},
			wantErr: false,
		},
		{
			name:           "handles empty gist ID",
			gistID:         "",
			wantErr:        true,
			wantErrMessage: "gist ID is required",
		},
		{
			name:   "handles not found error",
			gistID: "nonexistent",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			},
			wantErr:        true,
			wantErrMessage: "gist not found",
		},
		{
			name:   "handles permission error",
			gistID: "forbidden",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "Must have admin rights to Repository"}`))
			},
			wantErr:        true,
			wantErrMessage: "authentication failed",
		},
		{
			name:   "handles rate limit",
			gistID: "test123",
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "API rate limit exceeded"}`))
			},
			wantErr:        true,
			wantErrMessage: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.setupServer))
				defer server.Close()
			}

			client := &Client{
				github: github.NewClient(nil).WithAuthToken("test-token"),
			}

			if server != nil {
				client.github.BaseURL, _ = client.github.BaseURL.Parse(server.URL + "/")
			}

			err := client.DeleteGist(context.Background(), tt.gistID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteGist() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("DeleteGist() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}
		})
	}
}

func TestClient_UpdateIndexGist(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		index          *models.Index
		existingGistID string
		setupServer    func(t *testing.T) *httptest.Server
		wantGistID     string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:     "creates new index gist when none exists",
			username: "testuser",
			index: &models.Index{
				Username: "testuser",
				Entries: []models.IndexEntry{
					{
						GistID:      "gist1",
						Name:        "Test Prompt",
						Author:      "testuser",
						Category:    "testing",
						Tags:        []string{"test"},
						Description: "Test description",
						UpdatedAt:   time.Now(),
					},
				},
				UpdatedAt: time.Now(),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/gists":
						// List gists to find index
						if r.Method == "GET" {
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(`[]`)) // No existing gists
						} else if r.Method == "POST" {
							// Create new index gist
							var reqBody map[string]interface{}
							if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
								t.Errorf("Failed to decode request body: %v", err)
							}

							// Verify gist properties
							if desc, ok := reqBody["description"].(string); !ok || desc != "Prompt Vault Index" {
								t.Errorf("Unexpected description: %v", desc)
							}

							files, ok := reqBody["files"].(map[string]interface{})
							if !ok {
								t.Error("Missing files in request")
							}

							file, ok := files["testuser-promptvault-index.json"].(map[string]interface{})
							if !ok {
								t.Error("Missing index file in request")
							}

							// Verify JSON content
							content, ok := file["content"].(string)
							if !ok {
								t.Error("Missing content in file")
							}

							var parsedIndex models.Index
							if err := json.Unmarshal([]byte(content), &parsedIndex); err != nil {
								t.Errorf("Invalid JSON content: %v", err)
							}

							if parsedIndex.Username != "testuser" {
								t.Errorf("Unexpected username in index: %v", parsedIndex.Username)
							}

							w.WriteHeader(http.StatusCreated)
							w.Write([]byte(`{
								"id": "newindex123",
								"html_url": "https://gist.github.com/testuser/newindex123"
							}`))
						}
					default:
						t.Errorf("Unexpected path: %s", r.URL.Path)
					}
				}))
			},
			wantGistID: "newindex123",
			wantErr:    false,
		},
		{
			name:           "updates existing index gist",
			username:       "testuser",
			existingGistID: "existingindex456",
			index: &models.Index{
				Username: "testuser",
				Entries: []models.IndexEntry{
					{
						GistID:    "gist1",
						Name:      "Updated Prompt",
						Author:    "testuser",
						Category:  "testing",
						Tags:      []string{"test", "updated"},
						UpdatedAt: time.Now(),
					},
				},
				UpdatedAt: time.Now(),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/gists":
						// List gists to find index
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`[{
							"id": "existingindex456",
							"description": "Prompt Vault Index",
							"files": {
								"testuser-promptvault-index.json": {
									"filename": "testuser-promptvault-index.json"
								}
							}
						}]`))
					case "/gists/existingindex456":
						if r.Method == "PATCH" {
							// Update existing index
							var reqBody map[string]interface{}
							if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
								t.Errorf("Failed to decode request body: %v", err)
							}

							files, ok := reqBody["files"].(map[string]interface{})
							if !ok {
								t.Error("Missing files in request")
							}

							file, ok := files["testuser-promptvault-index.json"].(map[string]interface{})
							if !ok {
								t.Error("Missing index file in request")
							}

							// Verify JSON content
							content, ok := file["content"].(string)
							if !ok {
								t.Error("Missing content in file")
							}

							var parsedIndex models.Index
							if err := json.Unmarshal([]byte(content), &parsedIndex); err != nil {
								t.Errorf("Invalid JSON content: %v", err)
							}

							w.WriteHeader(http.StatusOK)
							w.Write([]byte(`{
								"id": "existingindex456",
								"html_url": "https://gist.github.com/testuser/existingindex456"
							}`))
						}
					default:
						t.Errorf("Unexpected path: %s", r.URL.Path)
					}
				}))
			},
			wantGistID: "existingindex456",
			wantErr:    false,
		},
		{
			name:     "handles empty username",
			username: "",
			index: &models.Index{
				Username: "",
				Entries:  []models.IndexEntry{},
			},
			wantErr:        true,
			wantErrMessage: "username is required",
		},
		{
			name:           "handles nil index",
			username:       "testuser",
			index:          nil,
			wantErr:        true,
			wantErrMessage: "index is required",
		},
		{
			name:     "handles API error",
			username: "testuser",
			index: &models.Index{
				Username:  "testuser",
				Entries:   []models.IndexEntry{},
				UpdatedAt: time.Now(),
			},
			setupServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message": "Internal Server Error"}`))
				}))
			},
			wantErr:        true,
			wantErrMessage: "failed to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = tt.setupServer(t)
				defer server.Close()
			}

			client := &Client{
				github: github.NewClient(nil).WithAuthToken("test-token"),
			}

			if server != nil {
				client.github.BaseURL, _ = client.github.BaseURL.Parse(server.URL + "/")
			}

			gistID, err := client.UpdateIndexGist(context.Background(), tt.username, tt.index)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateIndexGist() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.wantErrMessage != "" && err != nil {
				if !contains(err.Error(), tt.wantErrMessage) {
					t.Errorf("UpdateIndexGist() error = %v, want error containing %v", err, tt.wantErrMessage)
				}
			}

			if !tt.wantErr && gistID != tt.wantGistID {
				t.Errorf("UpdateIndexGist() gistID = %v, want %v", gistID, tt.wantGistID)
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr) != -1))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
