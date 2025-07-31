package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/testhelpers"
)

func TestCompleteWorkflow_FileOperations(t *testing.T) {
	// Create test environment
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	// Create cache and config directories
	cacheDir := testhelpers.CreateTestCache(t, testDir)
	configDir := testhelpers.CreateTestConfig(t, testDir)

	// Verify directories were created
	testhelpers.ValidateDirectoryExists(t, cacheDir)
	testhelpers.ValidateDirectoryExists(t, configDir)

	t.Run("Create and validate test prompt file", func(t *testing.T) {
		promptFile := testhelpers.CreateTestPromptFile(t, testDir, "test-prompt")

		// Verify file was created
		testhelpers.ValidateFileExists(t, promptFile)

		// Verify file contains expected content
		testhelpers.ValidateFileContent(t, promptFile, "test-prompt")
		testhelpers.ValidateFileContent(t, promptFile, "testuser")
		testhelpers.ValidateFileContent(t, promptFile, "test")
	})

	t.Run("Create test index and entries", func(t *testing.T) {
		// Create test entries
		entries := []models.IndexEntry{
			testhelpers.CreateTestIndexEntry("API Docs", "user1", "gist1", []string{"api"}),
			testhelpers.CreateTestIndexEntry("Code Review", "user2", "gist2", []string{"code"}),
		}

		// Create index
		index := testhelpers.CreateTestIndex("testuser", entries)

		// Verify index structure
		assert.Equal(t, "testuser", index.Username)
		assert.Len(t, index.Entries, 2)
		assert.Equal(t, "API Docs", index.Entries[0].Name)
		assert.Equal(t, "Code Review", index.Entries[1].Name)
	})

	t.Run("Test prompt creation and conversion", func(t *testing.T) {
		// Create test prompt
		prompt := testhelpers.CreateTestPrompt("Test Workflow", "testuser", []string{"workflow", "test"})

		// Verify prompt structure
		assert.Equal(t, "Test Workflow", prompt.Name)
		assert.Equal(t, "testuser", prompt.Author)
		assert.Contains(t, prompt.Tags, "workflow")
		assert.Contains(t, prompt.Tags, "test")
		assert.NotEmpty(t, prompt.GistID)
		assert.NotEmpty(t, prompt.Content)

		// Test conversion to index entry
		entry := prompt.ToIndexEntry()
		assert.Equal(t, prompt.Name, entry.Name)
		assert.Equal(t, prompt.Author, entry.Author)
		assert.Equal(t, prompt.GistID, entry.GistID)
		assert.Equal(t, prompt.Tags, entry.Tags)
	})
}

func TestWorkflow_DataValidation(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	t.Run("Prompt validation workflow", func(t *testing.T) {
		// Test valid prompt
		validPrompt := testhelpers.CreateTestPrompt("Valid Prompt", "testuser", []string{"valid"})
		err := validPrompt.Validate()
		assert.NoError(t, err)

		// Test invalid prompt (missing required fields)
		invalidPrompt := models.Prompt{
			PromptMeta: models.PromptMeta{
				Name:   "", // Missing name
				Author: "testuser",
				Tags:   []string{"test"},
			},
		}
		err = invalidPrompt.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("Index operations workflow", func(t *testing.T) {
		index := &models.Index{
			Username:  "testuser",
			UpdatedAt: time.Now(),
		}

		// Add entries workflow
		entry1 := testhelpers.CreateTestIndexEntry("First", "user1", "gist1", []string{"first"})
		entry2 := testhelpers.CreateTestIndexEntry("Second", "user2", "gist2", []string{"second"})

		index.AddImportedEntry(entry1)
		index.AddImportedEntry(entry2)

		assert.Len(t, index.ImportedEntries, 2)

		// Find entry workflow
		found, exists := index.FindImportedEntry("gist1")
		assert.True(t, exists)
		assert.Equal(t, "First", found.Name)

		// Update entry workflow
		updatedEntry := testhelpers.CreateTestIndexEntry("First Updated", "user1", "gist1", []string{"first", "updated"})
		success := index.UpdateImportedEntry(updatedEntry)
		assert.True(t, success)

		// Verify update
		found, exists = index.FindImportedEntry("gist1")
		assert.True(t, exists)
		assert.Equal(t, "First Updated", found.Name)
		assert.Contains(t, found.Tags, "updated")
	})
}

func TestWorkflow_FileSystemOperations(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	t.Run("File creation and content validation", func(t *testing.T) {
		// Create various prompt files
		files := []string{"api-docs", "code-review", "deployment"}

		for _, filename := range files {
			promptFile := testhelpers.CreateTestPromptFile(t, testDir, filename)

			// Verify file exists and has correct content
			testhelpers.ValidateFileExists(t, promptFile)
			testhelpers.ValidateFileContent(t, promptFile, filename)

			// Verify YAML structure
			content, err := os.ReadFile(promptFile)
			require.NoError(t, err)

			contentStr := string(content)
			assert.Contains(t, contentStr, "---")
			assert.Contains(t, contentStr, "name:")
			assert.Contains(t, contentStr, "author:")
			assert.Contains(t, contentStr, "tags:")
		}
	})

	t.Run("Custom content workflow", func(t *testing.T) {
		customContent := `---
name: "Custom Prompt"
author: "custom_user"
tags: ["custom", "workflow"]
version: "2.0"
description: "A custom test prompt"
---
This is a custom prompt with {param1} and {param2}.
It has multiple lines and special formatting.`

		customFile := testhelpers.CreateTestPromptFileWithContent(t, testDir, "custom.yaml", customContent)

		testhelpers.ValidateFileExists(t, customFile)
		testhelpers.ValidateFileContent(t, customFile, "Custom Prompt")
		testhelpers.ValidateFileContent(t, customFile, "custom_user")
		testhelpers.ValidateFileContent(t, customFile, "{param1}")
		testhelpers.ValidateFileContent(t, customFile, "version: \"2.0\"")
	})
}

func TestWorkflow_ErrorHandling(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	t.Run("Invalid file operations", func(t *testing.T) {
		// Try to read non-existent file
		nonExistentFile := filepath.Join(testDir, "nonexistent.yaml")
		_, err := os.ReadFile(nonExistentFile)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Invalid prompt data", func(t *testing.T) {
		// Test various validation errors
		testCases := []struct {
			name   string
			prompt models.Prompt
			error  string
		}{
			{
				name: "Missing name",
				prompt: models.Prompt{
					PromptMeta: models.PromptMeta{
						Author: "testuser",
						Tags:   []string{"test"},
					},
				},
				error: "name is required",
			},
			{
				name: "Missing author",
				prompt: models.Prompt{
					PromptMeta: models.PromptMeta{
						Name: "Test Prompt",
						Tags: []string{"test"},
					},
				},
				error: "author is required",
			},
			{
				name: "Missing tags",
				prompt: models.Prompt{
					PromptMeta: models.PromptMeta{
						Name:   "Test Prompt",
						Author: "testuser",
						Tags:   []string{},
					},
				},
				error: "at least one tag is required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.prompt.Validate()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.error)
			})
		}
	})
}

func TestWorkflow_EnvironmentSetup(t *testing.T) {
	t.Run("Test environment variables", func(t *testing.T) {
		testDir := testhelpers.CreateTestDir(t)
		testhelpers.SetupTestEnv(t, testDir)

		// Verify environment variables are set
		assert.Equal(t, testDir, os.Getenv("HOME"))
		assert.Equal(t, filepath.Join(testDir, "cache"), os.Getenv("CACHE_DIR"))
		assert.Equal(t, filepath.Join(testDir, "config"), os.Getenv("CONFIG_DIR"))
	})

	t.Run("Directory structure", func(t *testing.T) {
		testDir := testhelpers.CreateTestDir(t)
		testhelpers.SetupTestEnv(t, testDir)

		cacheDir := testhelpers.CreateTestCache(t, testDir)
		configDir := testhelpers.CreateTestConfig(t, testDir)

		// Verify directory structure
		assert.DirExists(t, testDir)
		assert.DirExists(t, cacheDir)
		assert.DirExists(t, configDir)

		// Verify paths match environment variables
		assert.Equal(t, cacheDir, os.Getenv("CACHE_DIR"))
		assert.Equal(t, configDir, os.Getenv("CONFIG_DIR"))
	})
}

func TestWorkflow_TimeHandling(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	t.Run("Timestamp operations", func(t *testing.T) {
		now := time.Now()

		// Create prompt with timestamp
		prompt := testhelpers.CreateTestPrompt("Time Test", "testuser", []string{"time"})

		// Verify timestamp is recent
		assert.True(t, prompt.UpdatedAt.After(now.Add(-time.Minute)))
		assert.True(t, prompt.UpdatedAt.Before(now.Add(time.Minute)))

		// Test index entry timestamp
		entry := prompt.ToIndexEntry()
		assert.Equal(t, prompt.UpdatedAt, entry.UpdatedAt)
	})

	t.Run("Version timestamp generation", func(t *testing.T) {
		prompt := models.Prompt{
			PromptMeta: models.PromptMeta{
				Name:   "Version Test",
				Author: "testuser",
				Tags:   []string{"version"},
			},
		}

		// Set default version (timestamp-based)
		prompt.SetDefaultVersion()
		assert.NotEmpty(t, prompt.Version)
		assert.Regexp(t, `^\d+$`, prompt.Version) // Should be numeric timestamp
	})
}

func TestWorkflow_DataIntegrity(t *testing.T) {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	t.Run("Prompt to IndexEntry conversion integrity", func(t *testing.T) {
		originalPrompt := testhelpers.CreateTestPrompt("Integrity Test", "testuser", []string{"integrity", "test"})
		originalPrompt.Version = "2.1"
		originalPrompt.Description = "Test description for integrity"
		originalPrompt.ID = "integrity-test-id"

		// Convert to index entry
		entry := originalPrompt.ToIndexEntry()

		// Verify all fields are preserved
		assert.Equal(t, originalPrompt.GistID, entry.GistID)
		assert.Equal(t, originalPrompt.GistURL, entry.GistURL)
		assert.Equal(t, originalPrompt.Name, entry.Name)
		assert.Equal(t, originalPrompt.Author, entry.Author)
		assert.Equal(t, originalPrompt.Tags, entry.Tags)
		assert.Equal(t, originalPrompt.Version, entry.Version)
		assert.Equal(t, originalPrompt.Description, entry.Description)
		assert.Equal(t, originalPrompt.ID, entry.ID)
		assert.Equal(t, originalPrompt.UpdatedAt, entry.UpdatedAt)

		// Content should not be in index entry (per design)
		// This is verified by checking the JSON tag is "-"
	})

	t.Run("Index operations maintain consistency", func(t *testing.T) {
		index := testhelpers.CreateTestIndex("testuser", []models.IndexEntry{})

		// Add multiple entries
		entries := []models.IndexEntry{
			testhelpers.CreateTestIndexEntry("First", "user1", "gist1", []string{"first"}),
			testhelpers.CreateTestIndexEntry("Second", "user2", "gist2", []string{"second"}),
			testhelpers.CreateTestIndexEntry("Third", "user3", "gist3", []string{"third"}),
		}

		for _, entry := range entries {
			index.AddImportedEntry(entry)
		}

		assert.Len(t, index.ImportedEntries, 3)

		// Verify each entry is findable
		for _, entry := range entries {
			found, exists := index.FindImportedEntry(entry.GistID)
			assert.True(t, exists)
			assert.Equal(t, entry.Name, found.Name)
			assert.Equal(t, entry.Author, found.Author)
		}
	})
}
