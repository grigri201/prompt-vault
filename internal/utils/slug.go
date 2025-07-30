package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// GenerateSlug creates a URL-friendly identifier from name and author
func GenerateSlug(name, author string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	author = strings.ToLower(author)
	
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	author = strings.ReplaceAll(author, " ", "-")
	
	// Remove special characters, keep only alphanumeric and hyphens
	validChars := regexp.MustCompile(`[^a-z0-9-]+`)
	name = validChars.ReplaceAllString(name, "")
	author = validChars.ReplaceAllString(author, "")
	
	// Remove consecutive hyphens
	multiHyphen := regexp.MustCompile(`-+`)
	name = multiHyphen.ReplaceAllString(name, "-")
	author = multiHyphen.ReplaceAllString(author, "-")
	
	// Trim hyphens from start and end
	name = strings.Trim(name, "-")
	author = strings.Trim(author, "-")
	
	// Prepend author for uniqueness
	return fmt.Sprintf("%s-%s", author, name)
}

// ValidateID checks if an ID contains only valid characters
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	
	// ID can only contain alphanumeric characters, hyphens, and underscores
	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validID.MatchString(id) {
		return fmt.Errorf("ID can only contain letters, numbers, hyphens, and underscores")
	}
	
	// Check minimum and maximum length
	if len(id) < 3 {
		return fmt.Errorf("ID must be at least 3 characters long")
	}
	if len(id) > 100 {
		return fmt.Errorf("ID must not exceed 100 characters")
	}
	
	return nil
}