package gist

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
)

// Logger is an optional interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// GistData represents data for creating/updating a gist
type GistData struct {
	Name        string
	Description string
	Content     string
	Public      bool
}

// GistResult represents the result of a gist operation
type GistResult struct {
	ID      string
	URL     string
	Created bool // true if created, false if updated
}

// GistClient defines the interface for Gist operations
type GistClient interface {
	UpdateGist(ctx context.Context, gistID, name, description, content string) (string, error)
	CreateGist(ctx context.Context, name, description, content string) (string, string, error)
	CreatePublicGist(ctx context.Context, name, description, content string) (string, string, error)
	DeleteGist(ctx context.Context, gistID string) error
	GetGist(ctx context.Context, gistID string) (*github.Gist, error)
	GetAPIError(err error) *github.ErrorResponse
	IsRateLimitError(err error) bool
}

// GistOperations provides high-level operations for Gist management
type GistOperations struct {
	client     GistClient
	retryCount int
	logger     Logger // Optional logger interface
}

// GistOperationsConfig configures the GistOperations behavior
type GistOperationsConfig struct {
	Client     GistClient
	RetryCount int
	Logger     Logger
}

// NewGistOperations creates a new GistOperations instance
func NewGistOperations(config GistOperationsConfig) *GistOperations {
	retryCount := config.RetryCount
	if retryCount <= 0 {
		retryCount = 3 // Default retry count
	}

	return &GistOperations{
		client:     config.Client,
		retryCount: retryCount,
		logger:     config.Logger,
	}
}

// CreateOrUpdate creates a new gist or updates existing one
// Handles 404 errors by creating a new gist
func (g *GistOperations) CreateOrUpdate(ctx context.Context, gistID string, data *GistData) (*GistResult, error) {
	if data == nil {
		return nil, errors.NewValidationErrorMsg("CreateOrUpdate", "gist data is required")
	}

	if data.Name == "" || data.Content == "" {
		return nil, errors.NewValidationErrorMsg("CreateOrUpdate", "name and content are required")
	}

	// If no gistID provided, create a new gist
	if gistID == "" {
		return g.create(ctx, data)
	}

	// Try to update existing gist
	url, err := g.client.UpdateGist(ctx, gistID, data.Name, data.Description, data.Content)
	if err != nil {
		// Check if it's a 404 error
		if g.is404Error(err) {
			if g.logger != nil {
				g.logger.Debug("Gist not found, creating new one", "gistID", gistID)
			}
			// Gist doesn't exist, create a new one
			return g.create(ctx, data)
		}
		return nil, err
	}

	return &GistResult{
		ID:      gistID,
		URL:     url,
		Created: false,
	}, nil
}

// CreateOrUpdateWithRetry performs operation with automatic retry
func (g *GistOperations) CreateOrUpdateWithRetry(ctx context.Context, gistID string, data *GistData) (*GistResult, error) {
	var lastErr error

	for i := 0; i < g.retryCount; i++ {
		result, err := g.CreateOrUpdate(ctx, gistID, data)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry validation errors
		if errors.IsType(err, errors.ErrTypeValidation) {
			return nil, err
		}

		// Don't retry auth errors
		if errors.IsType(err, errors.ErrTypeAuth) {
			return nil, err
		}

		// Check if we should retry
		if !g.isRetryableError(err) {
			return nil, err
		}

		if g.logger != nil {
			g.logger.Info("Retrying gist operation", "attempt", i+1, "error", err.Error())
		}

		// Exponential backoff
		if i < g.retryCount-1 {
			backoff := time.Duration(1<<uint(i)) * time.Second
			select {
			case <-time.After(backoff):
				// Continue to next retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, lastErr
}

// DeleteSafely deletes a gist, ignoring 404 errors
func (g *GistOperations) DeleteSafely(ctx context.Context, gistID string) error {
	if gistID == "" {
		return errors.NewValidationErrorMsg("DeleteSafely", "gist ID is required")
	}

	err := g.client.DeleteGist(ctx, gistID)
	if err != nil {
		// Ignore 404 errors
		if g.is404Error(err) {
			if g.logger != nil {
				g.logger.Debug("Gist already deleted or doesn't exist", "gistID", gistID)
			}
			return nil
		}
		return err
	}

	return nil
}

// FetchPromptGist fetches and parses a prompt gist
func (g *GistOperations) FetchPromptGist(ctx context.Context, gistID string) (*models.Prompt, error) {
	if gistID == "" {
		return nil, errors.NewValidationErrorMsg("FetchPromptGist", "gist ID is required")
	}

	// Fetch the gist
	gist, err := g.client.GetGist(ctx, gistID)
	if err != nil {
		return nil, err
	}

	// Find the prompt file
	var content string
	var filename string
	for name, file := range gist.Files {
		if file.Content != nil {
			content = *file.Content
			filename = string(name)
			break
		}
	}

	if content == "" {
		return nil, errors.NewValidationErrorMsg("FetchPromptGist", "no content found in gist")
	}

	// Parse the prompt
	yamlParser := parser.NewYAMLParser(parser.YAMLParserConfig{
		Strict: false, // Be lenient when fetching
	})

	prompt, err := yamlParser.ParsePromptFile(content)
	if err != nil {
		return nil, errors.WrapWithMessage(err, fmt.Sprintf("failed to parse prompt from %s", filename))
	}

	// Set gist metadata
	prompt.GistID = gistID
	if gist.HTMLURL != nil {
		prompt.GistURL = *gist.HTMLURL
	}
	if gist.UpdatedAt != nil {
		prompt.UpdatedAt = gist.UpdatedAt.Time
	}

	return prompt, nil
}

// create creates a new gist
func (g *GistOperations) create(ctx context.Context, data *GistData) (*GistResult, error) {
	var id, url string
	var err error

	if data.Public {
		id, url, err = g.client.CreatePublicGist(ctx, data.Name, data.Description, data.Content)
	} else {
		id, url, err = g.client.CreateGist(ctx, data.Name, data.Description, data.Content)
	}

	if err != nil {
		return nil, err
	}

	return &GistResult{
		ID:      id,
		URL:     url,
		Created: true,
	}, nil
}

// is404Error checks if an error is a 404 Not Found error
func (g *GistOperations) is404Error(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a GitHub ErrorResponse with 404 status
	apiErr := g.client.GetAPIError(err)
	if apiErr != nil && apiErr.Response != nil {
		return apiErr.Response.StatusCode == http.StatusNotFound
	}

	// Also check the error message
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found")
}

// isRetryableError determines if an error should be retried
func (g *GistOperations) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Rate limit errors are retryable
	if g.client.IsRateLimitError(err) {
		return true
	}

	// Network errors are retryable
	if errors.IsType(err, errors.ErrTypeNetwork) {
		return true
	}

	// Check for specific HTTP status codes
	apiErr := g.client.GetAPIError(err)
	if apiErr != nil && apiErr.Response != nil {
		status := apiErr.Response.StatusCode
		// Retry on 5xx errors and certain 4xx errors
		return status >= 500 || status == 429 || status == 408
	}

	return false
}
