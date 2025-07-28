package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/share"
	"github.com/spf13/cobra"
)

// MockShareManager is a mock implementation of the share manager
type MockShareManager struct {
	SharePromptFunc func(ctx context.Context, privateGistID string) (*share.ShareResult, error)
	SharePromptCalls []string
}

func (m *MockShareManager) SharePrompt(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
	m.SharePromptCalls = append(m.SharePromptCalls, privateGistID)
	if m.SharePromptFunc != nil {
		return m.SharePromptFunc(ctx, privateGistID)
	}
	return nil, fmt.Errorf("SharePrompt not implemented")
}

func TestShareCommand_Execute(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		shareFunc        func(ctx context.Context, privateGistID string) (*share.ShareResult, error)
		expectError      bool
		expectedOutput   string
		expectedCalls    []string
	}{
		{
			name: "successful new share",
			args: []string{"1234567890abcdef1234567890abcdef"},
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "public456",
					PublicGistURL: "https://gist.github.com/testuser/public456",
					IsUpdate:      false,
				}, nil
			},
			expectError:    false,
			expectedOutput: "Successfully created public gist: https://gist.github.com/testuser/public456",
			expectedCalls:  []string{"1234567890abcdef1234567890abcdef"},
		},
		{
			name: "successful update",
			args: []string{"abcdef1234567890abcdef1234567890"},
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return &share.ShareResult{
					PublicGistID:  "public999",
					PublicGistURL: "https://gist.github.com/testuser/public999",
					IsUpdate:      true,
				}, nil
			},
			expectError:    false,
			expectedOutput: "Successfully updated public gist: https://gist.github.com/testuser/public999",
			expectedCalls:  []string{"abcdef1234567890abcdef1234567890"},
		},
		{
			name:           "missing gist ID argument shows list",
			args:           []string{},
			expectError:    true,  // Will get TTY error in test environment
			expectedOutput: "Found 1 prompt(s) to share",  // Default test data has 1 prompt
			expectedCalls:  []string{},
		},
		{
			name: "share error",
			args: []string{"fedcba0987654321fedcba0987654321"},
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return nil, fmt.Errorf("gist not found")
			},
			expectError:    true,
			expectedOutput: "The gist was not found. Please check the gist ID and try again",
			expectedCalls:  []string{"fedcba0987654321fedcba0987654321"},
		},
		{
			name: "already public error",
			args: []string{"1111111111111111111111111111111a"},
			shareFunc: func(ctx context.Context, privateGistID string) (*share.ShareResult, error) {
				return nil, fmt.Errorf("cannot share: gist 1111111111111111111111111111111a is already public")
			},
			expectError:    true,
			expectedOutput: "already public",
			expectedCalls:  []string{"1111111111111111111111111111111a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock manager
			mockManager := &MockShareManager{
				SharePromptFunc: tt.shareFunc,
			}

			// Create command with mock
			cmd := newShareCmd(mockManager)

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
			if len(mockManager.SharePromptCalls) != len(tt.expectedCalls) {
				t.Errorf("Expected %d calls, got %d", len(tt.expectedCalls), len(mockManager.SharePromptCalls))
			}
			for i, expectedCall := range tt.expectedCalls {
				if i < len(mockManager.SharePromptCalls) && mockManager.SharePromptCalls[i] != expectedCall {
					t.Errorf("Expected call %d to be %q, got %q", i, expectedCall, mockManager.SharePromptCalls[i])
				}
			}
		})
	}
}

func TestShareCommand_Integration(t *testing.T) {
	// Test that the command is properly integrated with root command
	root := &cobra.Command{Use: "pv"}
	
	// Mock dependencies
	mockGistClient := &MockGistClient{
		GetGistFunc: func(ctx context.Context, gistID string) (*github.Gist, error) {
			return &github.Gist{
				ID:     github.String("1234567890abcdef1234567890abcdef"),
				Public: github.Bool(false),
				Files: map[github.GistFilename]github.GistFile{
					"test.yaml": {
						Content: github.String(`---
name: Test
author: testuser
category: test
tags: [test]
---
Content`),
					},
				},
			}, nil
		},
		CreatePublicGistFunc: func(ctx context.Context, gistName, description, content string) (string, string, error) {
			return "public456", "https://gist.github.com/testuser/public456", nil
		},
		ListUserGistsFunc: func(ctx context.Context, username string) ([]*github.Gist, error) {
			return []*github.Gist{}, nil
		},
	}

	mockUI := &MockUI{
		ConfirmFunc: func(message string) (bool, error) {
			return true, nil
		},
	}

	// Create share manager
	shareManager := share.NewManager(mockGistClient, mockUI, "testuser")

	// Add share command to root
	shareCmd := newShareCmd(shareManager)
	root.AddCommand(shareCmd)

	// Test command can be found
	cmd, _, err := root.Find([]string{"share", "1234567890abcdef1234567890abcdef"})
	if err != nil {
		t.Fatalf("Failed to find share command: %v", err)
	}

	if cmd.Use != "share [<gist-id>|<keyword>]" {
		t.Errorf("Expected command use to be 'share [<gist-id>|<keyword>]', got %q", cmd.Use)
	}
}

// MockGistClient is a mock implementation of the gist client
type MockGistClient struct {
	GetGistFunc          func(ctx context.Context, gistID string) (*github.Gist, error)
	CreatePublicGistFunc func(ctx context.Context, gistName, description, content string) (string, string, error)
	UpdateGistFunc       func(ctx context.Context, gistID, gistName, description, content string) (string, error)
	ListUserGistsFunc    func(ctx context.Context, username string) ([]*github.Gist, error)
}

func (m *MockGistClient) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	if m.GetGistFunc != nil {
		return m.GetGistFunc(ctx, gistID)
	}
	return nil, fmt.Errorf("GetGist not implemented")
}

func (m *MockGistClient) CreatePublicGist(ctx context.Context, gistName, description, content string) (string, string, error) {
	if m.CreatePublicGistFunc != nil {
		return m.CreatePublicGistFunc(ctx, gistName, description, content)
	}
	return "", "", fmt.Errorf("CreatePublicGist not implemented")
}

func (m *MockGistClient) UpdateGist(ctx context.Context, gistID, gistName, description, content string) (string, error) {
	if m.UpdateGistFunc != nil {
		return m.UpdateGistFunc(ctx, gistID, gistName, description, content)
	}
	return "", fmt.Errorf("UpdateGist not implemented")
}

func (m *MockGistClient) ListUserGists(ctx context.Context, username string) ([]*github.Gist, error) {
	if m.ListUserGistsFunc != nil {
		return m.ListUserGistsFunc(ctx, username)
	}
	return nil, fmt.Errorf("ListUserGists not implemented")
}

// MockUI is a mock implementation of the UI interface
type MockUI struct {
	ConfirmFunc func(message string) (bool, error)
}

func (m *MockUI) Confirm(message string) (bool, error) {
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(message)
	}
	return false, nil
}

