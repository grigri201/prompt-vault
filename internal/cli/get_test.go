package cli

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/testhelpers"
)

func TestGetCommand_Creation(t *testing.T) {
	cmd := NewGetCommand()

	assert.Equal(t, "get", cmd.Use[:3])
	assert.Contains(t, cmd.Short, "Search and retrieve")
	assert.True(t, cmd.HasFlags())

	// Check output flag exists
	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "o", outputFlag.Shorthand)
}

func TestGetCommand_ValidateArgs(t *testing.T) {
	cmd := NewGetCommand()

	// Should accept 0 args (show all)
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Should accept 1 arg (keyword search)
	err = cmd.Args(cmd, []string{"test"})
	assert.NoError(t, err)

	// Should reject 2+ args
	err = cmd.Args(cmd, []string{"test", "extra"})
	assert.Error(t, err)
}

func TestGetCommand_EmptyIndex(t *testing.T) {
	// Setup test environment
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	// Create empty cache directory
	cacheDir := testhelpers.CreateTestCache(t, testDir)

	// Create command
	cmd := NewGetCommand()

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock cache path function
	originalGetCachePathFunc := getCachePathFunc
	getCachePathFunc = func() string { return cacheDir }
	t.Cleanup(func() { getCachePathFunc = originalGetCachePathFunc })

	// Run command
	err := cmd.RunE(cmd, []string{})

	// Should handle empty index gracefully
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "No prompts found")
	assert.Contains(t, output, "pv sync")
}

func TestGetCommand_SearchFunctionality(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)
	cacheDir := testhelpers.CreateTestCache(t, testDir)

	// Create test index with sample data
	entries := []models.IndexEntry{
		testhelpers.CreateTestIndexEntry("API Documentation", "user1", "gist1", []string{"api", "docs"}),
		testhelpers.CreateTestIndexEntry("Code Review", "user2", "gist2", []string{"code", "review"}),
		testhelpers.CreateTestIndexEntry("Deployment Guide", "user1", "gist3", []string{"deploy", "ops"}),
	}

	testIndex := testhelpers.CreateTestIndex("testuser", entries)

	// Save test prompts to cache
	for _, entry := range entries {
		prompt := models.Prompt{
			PromptMeta: models.PromptMeta{
				Name:        entry.Name,
				Author:      entry.Author,
				Tags:        entry.Tags,
				Version:     entry.Version,
				Description: entry.Description,
			},
			GistID:    entry.GistID,
			GistURL:   entry.GistURL,
			Content:   "Test content for " + entry.Name,
			UpdatedAt: entry.UpdatedAt,
		}

		// Create cache manager and save prompt
		cachePath := getCachePathFunc()
		_, cacheManager := createManagersWithPath(cachePath)
		err := cacheManager.SavePrompt(&prompt)
		require.NoError(t, err)
	}

	// Save index
	cachePath := getCachePathFunc()
	_, cacheManager := createManagersWithPath(cachePath)
	err := cacheManager.SaveIndex(&testIndex)
	require.NoError(t, err)

	// Test search with keyword
	cmd := NewGetCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock cache path function
	originalGetCachePathFunc := getCachePathFunc
	getCachePathFunc = func() string { return cacheDir }
	t.Cleanup(func() { getCachePathFunc = originalGetCachePathFunc })

	// This test validates search results are displayed
	// Note: Interactive selection cannot be easily tested without mocking the UI
	err = cmd.RunE(cmd, []string{"api"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "API Documentation")
	assert.Contains(t, output, "user1")
}

func TestGetCommand_OutputToFile(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)
	outputFile := filepath.Join(testDir, "output.md")

	// Create command with output flag
	cmd := NewGetCommand()
	cmd.SetArgs([]string{"--output", outputFile, "test"})

	// Parse flags
	err := cmd.ParseFlags([]string{"--output", outputFile})
	require.NoError(t, err)

	// Verify flag was set correctly
	outputFlag, err := cmd.Flags().GetString("output")
	assert.NoError(t, err)
	assert.Equal(t, outputFile, outputFlag)
}

func TestGetCommand_NoSearchResults(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)
	cacheDir := testhelpers.CreateTestCache(t, testDir)

	// Create test index with sample data
	entries := []models.IndexEntry{
		testhelpers.CreateTestIndexEntry("API Documentation", "user1", "gist1", []string{"api", "docs"}),
	}

	testIndex := testhelpers.CreateTestIndex("testuser", entries)

	// Save index
	cachePath := getCachePathFunc()
	_, cacheManager := createManagersWithPath(cachePath)
	err := cacheManager.SaveIndex(&testIndex)
	require.NoError(t, err)

	// Test search with keyword that has no matches
	cmd := NewGetCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock cache path function
	originalGetCachePathFunc := getCachePathFunc
	getCachePathFunc = func() string { return cacheDir }
	t.Cleanup(func() { getCachePathFunc = originalGetCachePathFunc })

	err = cmd.RunE(cmd, []string{"nonexistent"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No prompts found matching 'nonexistent'")
}

func TestGetCommand_Flags(t *testing.T) {
	cmd := NewGetCommand()

	// Test output flag
	outputFlag := cmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "output", outputFlag.Name)
	assert.Equal(t, "o", outputFlag.Shorthand)
	assert.Equal(t, "", outputFlag.DefValue)
	assert.Contains(t, outputFlag.Usage, "Output to file")
}

// Helper function to create a realistic test scenario
func createTestScenario(t *testing.T) (string, *models.Index) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	// Create realistic test entries
	entries := []models.IndexEntry{
		{
			GistID:      "api-docs-123",
			GistURL:     "https://gist.github.com/user/api-docs-123",
			Name:        "API Documentation Template",
			Author:      "developer1",
			Tags:        []string{"api", "documentation", "template"},
			Version:     "1.2",
			Description: "Generate comprehensive API documentation",
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			GistID:      "code-review-456",
			GistURL:     "https://gist.github.com/user/code-review-456",
			Name:        "Code Review Checklist",
			Author:      "reviewer",
			Tags:        []string{"code", "review", "checklist"},
			Version:     "2.0",
			Description: "Structured code review guidelines",
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
	}

	testIndex := testhelpers.CreateTestIndex("testuser", entries)

	return testDir, &testIndex
}

func TestGetCommand_Integration(t *testing.T) {
	testDir, testIndex := createTestScenario(t)
	cacheDir := testhelpers.CreateTestCache(t, testDir)

	// Save test index and prompts
	cachePath := getCachePathFunc()
	_, cacheManager := createManagersWithPath(cachePath)

	for _, entry := range testIndex.Entries {
		prompt := models.Prompt{
			PromptMeta: models.PromptMeta{
				Name:        entry.Name,
				Author:      entry.Author,
				Tags:        entry.Tags,
				Version:     entry.Version,
				Description: entry.Description,
			},
			GistID:    entry.GistID,
			GistURL:   entry.GistURL,
			Content:   fmt.Sprintf("Content for %s with {placeholder}", entry.Name),
			UpdatedAt: entry.UpdatedAt,
		}

		err := cacheManager.SavePrompt(&prompt)
		require.NoError(t, err)
	}

	err := cacheManager.SaveIndex(testIndex)
	require.NoError(t, err)

	// Test the get command displays search results
	cmd := NewGetCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock cache path function
	originalGetCachePathFunc := getCachePathFunc
	getCachePathFunc = func() string { return cacheDir }
	t.Cleanup(func() { getCachePathFunc = originalGetCachePathFunc })

	err = cmd.RunE(cmd, []string{"api"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Found 1 prompt(s)")
	assert.Contains(t, output, "API Documentation Template")
	assert.Contains(t, output, "developer1")
	assert.Contains(t, output, "api, documentation, template")
}
