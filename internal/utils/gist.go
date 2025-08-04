package utils

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/grigri/pv/internal/errors"
)

// ExtractGistID extracts and validates a GitHub Gist ID from a Gist URL
// This function provides comprehensive validation and should be used across the codebase
// to maintain consistency and follow DRY principles.
func ExtractGistID(gistURL string) (string, error) {
	// Validate that the input is not empty
	if strings.TrimSpace(gistURL) == "" {
		return "", errors.ErrInvalidGistURL
	}

	// Parse the URL to validate its format
	parsedURL, err := url.Parse(gistURL)
	if err != nil {
		return "", errors.ErrInvalidGistURL
	}

	// Check if it's a GitHub Gist URL
	if parsedURL.Host != "gist.github.com" && parsedURL.Host != "github.com" {
		return "", errors.ErrInvalidGistURL
	}

	// Extract the Gist ID from the path
	// GitHub Gist URLs are in the format: https://gist.github.com/{user}/{gist_id}
	// Or sometimes just: https://gist.github.com/{gist_id}
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	
	var gistID string
	if len(pathParts) >= 2 {
		// Format: /user/gist_id
		gistID = pathParts[len(pathParts)-1]
	} else if len(pathParts) == 1 {
		// Format: /gist_id
		gistID = pathParts[0]
	} else {
		return "", errors.ErrInvalidGistURL
	}

	// Remove .git suffix if present
	gistID = strings.TrimSuffix(gistID, ".git")

	// Validate Gist ID format (should be alphanumeric with possible dashes)
	gistIDPattern := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !gistIDPattern.MatchString(gistID) {
		return "", errors.ErrInvalidGistURL
	}

	// Minimum length check for Gist IDs (GitHub uses long IDs)
	if len(gistID) < 10 {
		return "", errors.ErrInvalidGistURL
	}

	return gistID, nil
}

// ExtractGistIDFromURL is a simple version that extracts Gist ID from URL
// without validation. This is kept for backward compatibility and performance
// in cases where validation is not needed.
func ExtractGistIDFromURL(gistURL string) string {
	// Simple extraction without validation for performance
	parts := strings.Split(gistURL, "/")
	if len(parts) >= 2 {
		gistID := parts[len(parts)-1]
		// Remove .git suffix if present
		return strings.TrimSuffix(gistID, ".git")
	}
	return ""
}