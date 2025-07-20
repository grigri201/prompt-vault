package gist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/models"
	goerrors "errors"
)

// Client wraps the GitHub API client for Gist operations
type Client struct {
	github *github.Client
}

// NewClient creates a new GitHub client with authentication
func NewClient(token string) (*Client, error) {
	// Validate token
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.NewValidationErrorMsg("NewClient", "token is required")
	}

	// Create authenticated client
	client := github.NewClient(nil).WithAuthToken(token)

	return &Client{
		github: client,
	}, nil
}

// ValidateToken validates the provided token and returns the username
func (c *Client) ValidateToken(ctx context.Context) (string, error) {
	// Get authenticated user
	user, resp, err := c.github.Users.Get(ctx, "")
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return "", errors.NewAuthError("ValidateToken", err)
		}
		if c.IsRateLimitError(err) {
			return "", errors.NewNetworkError("ValidateToken", err)
		}
		return "", errors.WrapWithMessage(err, "failed to validate token")
	}

	if user.Login == nil {
		return "", errors.NewValidationErrorMsg("ValidateToken", "unable to get username from API response")
	}

	return *user.Login, nil
}

// IsRateLimitError checks if an error is due to rate limiting
func (c *Client) IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a GitHub ErrorResponse
	var errResp *github.ErrorResponse
	if goerrors.As(err, &errResp) {
		// Check status code
		if errResp.Response != nil && errResp.Response.StatusCode == http.StatusForbidden {
			// Check for rate limit header
			remaining := errResp.Response.Header.Get("X-RateLimit-Remaining")
			if remaining == "0" {
				return true
			}

			// Check error message for rate limit text
			message := strings.ToLower(errResp.Message)
			if strings.Contains(message, "rate limit") {
				return true
			}
		}
	}

	// Also check the error string directly
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit")
}

// GetAPIError extracts the GitHub API error from an error
func (c *Client) GetAPIError(err error) *github.ErrorResponse {
	if err == nil {
		return nil
	}

	var errResp *github.ErrorResponse
	if goerrors.As(err, &errResp) {
		return errResp
	}

	return nil
}

// GetGitHubClient returns the underlying GitHub client for advanced operations
func (c *Client) GetGitHubClient() *github.Client {
	return c.github
}

// CreateGist creates a new private Gist with the given content
func (c *Client) CreateGist(ctx context.Context, gistName, description, content string) (string, string, error) {
	// Validate inputs
	if gistName == "" {
		return "", "", errors.NewValidationErrorMsg("CreateGist", "gist name is required")
	}
	if content == "" {
		return "", "", errors.NewValidationErrorMsg("CreateGist", "content is required")
	}

	// Prepare the gist request
	gistReq := &github.Gist{
		Description: github.String(description),
		Public:      github.Bool(false), // Always create private gists
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(gistName + ".yaml"): {
				Content: github.String(content),
			},
		},
	}

	// Create the gist
	gist, _, err := c.github.Gists.Create(ctx, gistReq)
	if err != nil {
		if c.IsRateLimitError(err) {
			return "", "", errors.NewNetworkError("CreateGist", err)
		}
		return "", "", errors.WrapWithMessage(err, "failed to create gist")
	}

	if gist.ID == nil || gist.HTMLURL == nil {
		return "", "", errors.NewNetworkErrorMsg("CreateGist", "unexpected response from GitHub API")
	}

	return *gist.ID, *gist.HTMLURL, nil
}

// UpdateGist updates an existing Gist with new content
func (c *Client) UpdateGist(ctx context.Context, gistID, gistName, description, content string) (string, error) {
	// Validate inputs
	if gistID == "" {
		return "", errors.NewValidationErrorMsg("UpdateGist", "gist ID is required")
	}
	if gistName == "" {
		return "", errors.NewValidationErrorMsg("UpdateGist", "gist name is required")
	}
	if content == "" {
		return "", errors.NewValidationErrorMsg("UpdateGist", "content is required")
	}

	// Prepare the update request
	gistReq := &github.Gist{
		Description: github.String(description),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(gistName + ".yaml"): {
				Content: github.String(content),
			},
		},
	}

	// Update the gist
	gist, resp, err := c.github.Gists.Edit(ctx, gistID, gistReq)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", errors.WrapWithMessage(err, "gist not found")
		}
		if resp != nil && resp.StatusCode == http.StatusForbidden && !c.IsRateLimitError(err) {
			return "", errors.NewAuthError("UpdateGist", err)
		}
		if c.IsRateLimitError(err) {
			return "", errors.NewNetworkError("UpdateGist", err)
		}
		return "", errors.WrapWithMessage(err, "failed to update gist")
	}

	if gist.HTMLURL == nil {
		return "", errors.NewNetworkErrorMsg("UpdateGist", "unexpected response from GitHub API")
	}

	return *gist.HTMLURL, nil
}

// GetGist retrieves a Gist by ID
func (c *Client) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	// Validate input
	if gistID == "" {
		return nil, errors.NewValidationErrorMsg("GetGist", "gist ID is required")
	}

	// Get the gist
	gist, resp, err := c.github.Gists.Get(ctx, gistID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, errors.WrapWithMessage(err, "gist not found")
		}
		if c.IsRateLimitError(err) {
			return nil, errors.NewNetworkError("GetGist", err)
		}
		return nil, errors.WrapWithMessage(err, "failed to get gist")
	}

	return gist, nil
}

// DeleteGist deletes a Gist by ID
func (c *Client) DeleteGist(ctx context.Context, gistID string) error {
	// Validate input
	if gistID == "" {
		return errors.NewValidationErrorMsg("DeleteGist", "gist ID is required")
	}

	// Delete the gist
	resp, err := c.github.Gists.Delete(ctx, gistID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return errors.WrapWithMessage(err, "gist not found")
		}
		if resp != nil && resp.StatusCode == http.StatusForbidden && !c.IsRateLimitError(err) {
			return errors.NewAuthError("DeleteGist", err)
		}
		if c.IsRateLimitError(err) {
			return errors.NewNetworkError("DeleteGist", err)
		}
		return errors.WrapWithMessage(err, "failed to delete gist")
	}

	return nil
}

// UpdateIndexGist updates or creates the index Gist for a user
func (c *Client) UpdateIndexGist(ctx context.Context, username string, index *models.Index) (string, error) {
	// Validate inputs
	if username == "" {
		return "", errors.NewValidationErrorMsg("UpdateIndexGist", "username is required")
	}
	if index == nil {
		return "", errors.NewValidationErrorMsg("UpdateIndexGist", "index is required")
	}

	// Set the username in the index
	index.Username = username

	// Marshal index to JSON
	indexJSON, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return "", errors.WrapWithMessage(err, "failed to marshal index")
	}

	// Gist filename for index
	indexFilename := fmt.Sprintf("%s-promptvault-index.json", username)

	// Try to find existing index gist
	gists, _, err := c.github.Gists.List(ctx, "", &github.GistListOptions{})
	if err != nil {
		return "", errors.WrapWithMessage(err, "failed to list gists")
	}

	// Look for existing index gist
	var existingGistID string
	for _, gist := range gists {
		if gist.Description != nil && *gist.Description == "Prompt Vault Index" {
			// Check if it has the index file
			for filename := range gist.Files {
				if string(filename) == indexFilename {
					existingGistID = *gist.ID
					break
				}
			}
			if existingGistID != "" {
				break
			}
		}
	}

	if existingGistID != "" {
		// Update existing gist
		gistUpdate := &github.Gist{
			Files: map[github.GistFilename]github.GistFile{
				github.GistFilename(indexFilename): {
					Content: github.String(string(indexJSON)),
				},
			},
		}

		_, _, err = c.github.Gists.Edit(ctx, existingGistID, gistUpdate)
		if err != nil {
			return "", errors.WrapWithMessage(err, "failed to update index gist")
		}

		return existingGistID, nil
	}

	// Create new index gist
	newGist := &github.Gist{
		Description: github.String("Prompt Vault Index"),
		Public:      github.Bool(false),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(indexFilename): {
				Content: github.String(string(indexJSON)),
			},
		},
	}

	createdGist, _, err := c.github.Gists.Create(ctx, newGist)
	if err != nil {
		return "", errors.WrapWithMessage(err, "failed to create index gist")
	}

	if createdGist.ID == nil {
		return "", errors.NewNetworkErrorMsg("UpdateIndexGist", "unexpected response from GitHub API")
	}

	return *createdGist.ID, nil
}
