package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/imports"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/spf13/cobra"
)

// MockImportManager is a mock implementation of the import manager
type MockImportManager struct {
	ImportPromptFunc  func(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error)
	ImportPromptCalls []struct {
		GistURL string
		Index   *models.Index
	}
}

func (m *MockImportManager) ImportPrompt(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error) {
	m.ImportPromptCalls = append(m.ImportPromptCalls, struct {
		GistURL string
		Index   *models.Index
	}{gistURL, index})
	if m.ImportPromptFunc != nil {
		return m.ImportPromptFunc(ctx, gistURL, index)
	}
	return nil, fmt.Errorf("ImportPrompt not implemented")
}

// MockGistClientForImport is a mock implementation for import tests
type MockGistClientForImport struct {
	GetIndexFunc func(ctx context.Context) (*models.Index, error)
	UpdateIndexFunc func(ctx context.Context, index *models.Index) error
	GetGistFunc func(ctx context.Context, gistID string) (*github.Gist, error)
	GetGistByURLFunc func(ctx context.Context, gistURL string) (*github.Gist, error)
	ExtractGistIDFunc func(gistURL string) (string, error)
}

func (m *MockGistClientForImport) GetIndex(ctx context.Context) (*models.Index, error) {
	if m.GetIndexFunc != nil {
		return m.GetIndexFunc(ctx)
	}
	return nil, fmt.Errorf("GetIndex not implemented")
}

func (m *MockGistClientForImport) UpdateIndex(ctx context.Context, index *models.Index) error {
	if m.UpdateIndexFunc != nil {
		return m.UpdateIndexFunc(ctx, index)
	}
	return fmt.Errorf("UpdateIndex not implemented")
}

func (m *MockGistClientForImport) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	if m.GetGistFunc != nil {
		return m.GetGistFunc(ctx, gistID)
	}
	return nil, fmt.Errorf("GetGist not implemented")
}

func (m *MockGistClientForImport) GetGistByURL(ctx context.Context, gistURL string) (*github.Gist, error) {
	if m.GetGistByURLFunc != nil {
		return m.GetGistByURLFunc(ctx, gistURL)
	}
	return nil, fmt.Errorf("GetGistByURL not implemented")
}

func (m *MockGistClientForImport) ExtractGistID(gistURL string) (string, error) {
	if m.ExtractGistIDFunc != nil {
		return m.ExtractGistIDFunc(gistURL)
	}
	return "", fmt.Errorf("ExtractGistID not implemented")
}

func TestImportCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		importFunc     func(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error)
		expectError    bool
		expectedOutput string
		expectedCalls  int
	}{
		{
			name: "successful new import",
			args: []string{"https://gist.github.com/testuser/abc123"},
			importFunc: func(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error) {
				return &imports.ImportResult{
					GistID:     "abc123",
					IsUpdate:   false,
					NewVersion: "1.0.0",
				}, nil
			},
			expectError:    false,
			expectedOutput: "Successfully imported prompt (gist: abc123)",
			expectedCalls:  1,
		},
		{
			name: "successful update",
			args: []string{"https://gist.github.com/testuser/def456"},
			importFunc: func(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error) {
				return &imports.ImportResult{
					GistID:     "def456",
					IsUpdate:   true,
					OldVersion: "1.0.0",
					NewVersion: "2.0.0",
				}, nil
			},
			expectError:    false,
			expectedOutput: "Successfully updated prompt (gist: def456) from version 1.0.0 to 2.0.0",
			expectedCalls:  1,
		},
		{
			name:           "missing URL argument",
			args:           []string{},
			expectError:    true,
			expectedOutput: "accepts 1 arg(s), received 0",
			expectedCalls:  0,
		},
		{
			name: "import error",
			args: []string{"https://gist.github.com/testuser/error123"},
			importFunc: func(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error) {
				return nil, fmt.Errorf("invalid URL format")
			},
			expectError:    true,
			expectedOutput: "The provided URL is not valid. Please provide a valid GitHub gist URL",
			expectedCalls:  1,
		},
		{
			name: "private gist error",
			args: []string{"https://gist.github.com/testuser/private123"},
			importFunc: func(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error) {
				return nil, fmt.Errorf("cannot import private gist")
			},
			expectError:    true,
			expectedOutput: "private gist",
			expectedCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock manager
			mockManager := &MockImportManager{
				ImportPromptFunc: tt.importFunc,
			}

			// Create mock gist client
			mockGistClient := &MockGistClientForImport{
				GetIndexFunc: func(ctx context.Context) (*models.Index, error) {
					return &models.Index{
						Username:        "testuser",
						Entries:         []models.IndexEntry{},
						ImportedEntries: []models.IndexEntry{},
						UpdatedAt:       time.Now(),
					}, nil
				},
				UpdateIndexFunc: func(ctx context.Context, index *models.Index) error {
					return nil
				},
			}

			// Create command with mock
			cmd := newImportCmd(mockManager, mockGistClient)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Set args
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			// Check output
			output := buf.String()
			if !contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
			}

			// Check calls
			if len(mockManager.ImportPromptCalls) != tt.expectedCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectedCalls, len(mockManager.ImportPromptCalls))
			}
		})
	}
}

func TestImportCommand_Integration(t *testing.T) {
	// Test that the command is properly integrated with root command
	root := &cobra.Command{Use: "pv"}
	
	// Mock dependencies
	mockGistClient := &MockGistClientForImport{
		GetIndexFunc: func(ctx context.Context) (*models.Index, error) {
			return &models.Index{
				Username:        "testuser",
				ImportedEntries: []models.IndexEntry{},
			}, nil
		},
		UpdateIndexFunc: func(ctx context.Context, index *models.Index) error {
			return nil
		},
		GetGistByURLFunc: func(ctx context.Context, gistURL string) (*github.Gist, error) {
			return &github.Gist{
				ID:     github.String("abc123"),
				Public: github.Bool(true),
				Files: map[github.GistFilename]github.GistFile{
					"test.yaml": {
						Content: github.String(`---
name: Test Import
author: otheruser
category: test
tags: [test]
version: "1.0.0"
---
Content`),
					},
				},
			}, nil
		},
		ExtractGistIDFunc: func(gistURL string) (string, error) {
			return "abc123", nil
		},
	}

	mockUI := &MockUIForImport{
		ConfirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	// Create import manager
	importManager := imports.NewManager(mockGistClient, mockUI)

	// Add import command to root
	importCmd := newImportCmd(importManager, mockGistClient)
	root.AddCommand(importCmd)

	// Test command can be found
	cmd, _, err := root.Find([]string{"import", "https://gist.github.com/testuser/abc123"})
	if err != nil {
		t.Fatalf("Failed to find import command: %v", err)
	}

	if cmd.Use != "import <gist-url>" {
		t.Errorf("Expected command use to be 'import <gist-url>', got %q", cmd.Use)
	}
}

// MockUIForImport is a mock implementation of the UI interface for import
type MockUIForImport struct {
	ConfirmFunc func(message string) (bool, error)
}

func (m *MockUIForImport) Confirm(message string) (bool, error) {
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(message)
	}
	return false, nil
}