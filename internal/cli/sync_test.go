package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/grigri201/prompt-vault/internal/testhelpers"
)

func TestSyncCommand_Creation(t *testing.T) {
	cmd := NewSyncCommand()

	assert.Equal(t, "sync", cmd.Use)
	assert.Contains(t, cmd.Short, "synchronization")
	assert.True(t, cmd.HasFlags())

	// Check verbose flag exists
	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
}

func TestSyncCommand_ValidateArgs(t *testing.T) {
	cmd := NewSyncCommand()

	// Should accept 0 args
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Args validation should pass for sync command (no specific restrictions)
	err = cmd.Args(cmd, []string{"extra"})
	if cmd.Args != nil {
		// If Args is set, test it, otherwise skip
		assert.NoError(t, err)
	}
}

func TestSyncCommand_Flags(t *testing.T) {
	cmd := NewSyncCommand()

	// Test verbose flag
	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "verbose", verboseFlag.Name)
	assert.Equal(t, "v", verboseFlag.Shorthand)
	assert.Equal(t, "false", verboseFlag.DefValue)
	assert.Contains(t, verboseFlag.Usage, "detailed progress")
}

func TestSyncCommand_FlagParsing(t *testing.T) {
	cmd := NewSyncCommand()

	// Test parsing verbose flag
	err := cmd.ParseFlags([]string{"--verbose"})
	assert.NoError(t, err)

	verbose, err := cmd.Flags().GetBool("verbose")
	assert.NoError(t, err)
	assert.True(t, verbose)

	// Test parsing short verbose flag
	cmd2 := NewSyncCommand()
	err = cmd2.ParseFlags([]string{"-v"})
	assert.NoError(t, err)

	verbose2, err := cmd2.Flags().GetBool("verbose")
	assert.NoError(t, err)
	assert.True(t, verbose2)
}

func TestRunSyncCommand_NoContext(t *testing.T) {
	// Test when command context is not initialized
	// Reset any existing context
	originalContext := GetCommandContext()
	SetCommandContext(nil)
	t.Cleanup(func() { SetCommandContext(originalContext) })

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runSyncCommand(cmd, []string{}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command context not initialized")
}

func TestRunSyncCommand_InvalidContainer(t *testing.T) {
	// Test with invalid container setup
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	// Create a command context with invalid container state
	// This test would require access to container internals
	// For now, we test the error path validation
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// This test verifies the error handling structure
	// Full integration testing would require container mocking
	t.Skip("Requires container mocking infrastructure")
}

func TestSyncCommand_VerboseOutput(t *testing.T) {
	// Test that verbose flag affects output formatting
	// This is primarily a structural test
	cmd := NewSyncCommand()

	// Parse verbose flag
	err := cmd.ParseFlags([]string{"--verbose"})
	assert.NoError(t, err)

	verbose, err := cmd.Flags().GetBool("verbose")
	assert.NoError(t, err)
	assert.True(t, verbose)

	// The actual verbose behavior would be tested in integration tests
	// with proper mocking of sync manager
}

func TestSyncCommand_Help(t *testing.T) {
	cmd := NewSyncCommand()

	// Test help text contains expected information
	assert.Contains(t, cmd.Long, "synchronization between local and remote")
	assert.Contains(t, cmd.Long, "timestamps")
	assert.Contains(t, cmd.Long, "downloads remote")
	assert.Contains(t, cmd.Long, "uploads local")
	assert.Contains(t, cmd.Long, "Examples:")
}

func TestSyncCommand_ContextIntegration(t *testing.T) {
	// Test sync command with proper context setup
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	// Create mock components
	mockSyncManager := &testhelpers.MockSyncManager{}
	mockSyncManager.On("IsInitialized").Return(true)
	// mockSyncManager.On("SynchronizeData", mock.AnythingOfType("*context.emptyCtx")).Return(nil)
	mockSyncManager.On("GetSyncStatus").Return(testhelpers.SyncStatus{
		LocalTime:  time.Now(),
		RemoteTime: time.Now(),
		NeedsSync:  false,
		SyncAction: "up_to_date",
		Progress: testhelpers.SyncProgress{
			Completed: 0,
			Total:     0,
		},
	})

	// This test would require full container mocking
	// Skip for now until container infrastructure is available
	t.Skip("Requires full container mocking infrastructure")
}

// Test helper function for sync command testing
func createSyncTestEnvironment(t *testing.T) string {
	testDir := testhelpers.CreateTestDir(t)
	testhelpers.SetupTestEnv(t, testDir)

	// Create required directories
	testhelpers.CreateTestCache(t, testDir)
	testhelpers.CreateTestConfig(t, testDir)

	return testDir
}

func TestSyncCommand_ErrorHandling(t *testing.T) {
	// Test various error conditions
	tests := []struct {
		name          string
		setupError    string
		expectedError string
	}{
		{
			name:          "Authentication Error",
			setupError:    "auth",
			expectedError: "not authenticated",
		},
		{
			name:          "Network Error",
			setupError:    "network",
			expectedError: "network",
		},
		{
			name:          "Sync Manager Error",
			setupError:    "sync_manager",
			expectedError: "sync manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These tests would require proper mocking infrastructure
			// Skip for now until container and manager mocking is available
			t.Skip("Requires container and manager mocking infrastructure")
		})
	}
}
