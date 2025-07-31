package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/grigri201/prompt-vault/internal/models"
)

// CreateTestDir creates a temporary directory for testing
func CreateTestDir(t *testing.T) string {
	testDir := t.TempDir()
	return testDir
}

// SetupTestEnv sets up environment variables for testing
func SetupTestEnv(t *testing.T, testDir string) {
	// Set test environment variables
	os.Setenv("HOME", testDir)
	os.Setenv("CACHE_DIR", filepath.Join(testDir, "cache"))
	os.Setenv("CONFIG_DIR", filepath.Join(testDir, "config"))

	// Cleanup function
	t.Cleanup(func() {
		os.Unsetenv("HOME")
		os.Unsetenv("CACHE_DIR")
		os.Unsetenv("CONFIG_DIR")
	})
}

// CreateTestPrompt creates a test prompt with default values
func CreateTestPrompt(name, author string, tags []string) models.Prompt {
	return models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:        name,
			Author:      author,
			Tags:        tags,
			Version:     "1.0",
			Description: fmt.Sprintf("Test description for %s", name),
		},
		GistID:    fmt.Sprintf("gist-%s", name),
		GistURL:   fmt.Sprintf("https://gist.github.com/test/%s", name),
		Content:   fmt.Sprintf("Test content for %s with {variable}", name),
		UpdatedAt: time.Now(),
	}
}

// CreateTestPromptWithContent creates a test prompt with specific content
func CreateTestPromptWithContent(name, author, content string, tags []string) models.Prompt {
	return models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:        name,
			Author:      author,
			Tags:        tags,
			Version:     "1.0",
			Description: fmt.Sprintf("Test description for %s", name),
		},
		GistID:    fmt.Sprintf("gist-%s", name),
		GistURL:   fmt.Sprintf("https://gist.github.com/test/%s", name),
		Content:   content,
		UpdatedAt: time.Now(),
	}
}

// CreateTestPromptFile creates a YAML prompt file for testing
func CreateTestPromptFile(t *testing.T, dir, name string) string {
	content := fmt.Sprintf(`---
name: "%s"
author: "testuser"
tags: ["test", "example"]
version: "1.0"
description: "Test prompt for %s"
---
This is a test prompt for %s with {variable} and {another_var}.`, name, name, name)

	filename := filepath.Join(dir, name+".yaml")
	err := os.WriteFile(filename, []byte(content), 0644)
	require.NoError(t, err)

	return filename
}

// CreateTestPromptFileWithContent creates a YAML prompt file with specific content
func CreateTestPromptFileWithContent(t *testing.T, dir, filename, content string) string {
	fullPath := filepath.Join(dir, filename)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	require.NoError(t, err)
	return fullPath
}

// CreateTestIndex creates a test index with sample entries
func CreateTestIndex(username string, entries []models.IndexEntry) models.Index {
	return models.Index{
		Username:        username,
		Entries:         entries,
		ImportedEntries: []models.IndexEntry{},
		UpdatedAt:       time.Now(),
	}
}

// CreateTestIndexEntry creates a test index entry
func CreateTestIndexEntry(name, author, gistID string, tags []string) models.IndexEntry {
	return models.IndexEntry{
		GistID:      gistID,
		GistURL:     fmt.Sprintf("https://gist.github.com/test/%s", gistID),
		Name:        name,
		Author:      author,
		Tags:        tags,
		Version:     "1.0",
		Description: fmt.Sprintf("Test entry for %s", name),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestConfig creates a test configuration directory
func CreateTestConfig(t *testing.T, testDir string) string {
	configDir := filepath.Join(testDir, "config")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	return configDir
}

// CreateTestCache creates a test cache directory
func CreateTestCache(t *testing.T, testDir string) string {
	cacheDir := filepath.Join(testDir, "cache")
	err := os.MkdirAll(cacheDir, 0755)
	require.NoError(t, err)
	return cacheDir
}

// ValidateFileExists checks if a file exists and is readable
func ValidateFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	require.NoError(t, err, "File should exist: %s", path)
}

// ValidateFileContent checks if a file contains expected content
func ValidateFileContent(t *testing.T, path, expectedContent string) {
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), expectedContent)
}

// ValidateDirectoryExists checks if a directory exists
func ValidateDirectoryExists(t *testing.T, path string) {
	info, err := os.Stat(path)
	require.NoError(t, err, "Directory should exist: %s", path)
	require.True(t, info.IsDir(), "Path should be a directory: %s", path)
}

// CleanupTestFiles removes test files (though t.TempDir() handles this automatically)
func CleanupTestFiles(paths ...string) {
	for _, path := range paths {
		os.RemoveAll(path)
	}
}
