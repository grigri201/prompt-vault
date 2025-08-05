package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/service"
)

// Helper function to create AppError pointers for testing
func createAppError(errType errors.ErrorType, message string) error {
	appErr := errors.NewAppError(errType, message, nil)
	return &appErr
}

// MockSyncPromptService extends MockPromptService with sync-specific functionality
type MockSyncPromptService struct {
	*MockPromptService
	
	// Sync-specific behavior
	getPromptContentResult string
	getPromptContentError  error
	getPromptContentCalls  []string
	
	// Custom function overrides for sync testing
	getPromptContentFunc func(*model.Prompt) (string, error)
	listPromptsFunc      func() ([]model.Prompt, error)
	
	// Per-prompt error simulation for partial failure testing
	promptContentErrors map[string]error
}

func NewMockSyncPromptService() *MockSyncPromptService {
	return &MockSyncPromptService{
		MockPromptService:   NewMockPromptService(),
		getPromptContentCalls: make([]string, 0),
		promptContentErrors: make(map[string]error),
	}
}

func (m *MockSyncPromptService) GetPromptContent(prompt *model.Prompt) (string, error) {
	m.getPromptContentCalls = append(m.getPromptContentCalls, prompt.ID)
	
	if m.getPromptContentFunc != nil {
		return m.getPromptContentFunc(prompt)
	}
	
	// Check for per-prompt errors
	if err, exists := m.promptContentErrors[prompt.ID]; exists {
		return "", err
	}
	
	if m.getPromptContentError != nil {
		return "", m.getPromptContentError
	}
	
	return m.getPromptContentResult, nil
}

func (m *MockSyncPromptService) ListPrompts() ([]model.Prompt, error) {
	m.listPromptsCalls++
	if m.listPromptsFunc != nil {
		return m.listPromptsFunc()
	}
	if m.listPromptsError != nil {
		return nil, m.listPromptsError
	}
	return m.listPromptsResult, nil
}

func (m *MockSyncPromptService) Reset() {
	m.MockPromptService.Reset()
	m.getPromptContentResult = ""
	m.getPromptContentError = nil
	m.getPromptContentCalls = make([]string, 0)
	m.getPromptContentFunc = nil
	m.listPromptsFunc = nil
	m.promptContentErrors = make(map[string]error)
}

// Test data helpers for sync testing
func createSyncTestPrompts() []model.Prompt {
	return []model.Prompt{
		{
			ID:          "gist1234567890abcdef1234",
			Name:        "Go Best Practices",
			Author:      "Alice Smith",
			Description: "Comprehensive Go coding guidelines",
			Tags:        []string{"go", "best-practices"},
			Version:     "1.0",
			GistURL:     "https://gist.github.com/alice/gist1234567890abcdef1234",
		},
		{
			ID:          "gist4567890123abcdef4567",
			Name:        "Python Data Analysis",
			Author:      "Bob Johnson",
			Description: "Data analysis techniques in Python",
			Tags:        []string{"python", "data-science"},
			Version:     "2.1",
			GistURL:     "https://gist.github.com/bob/gist4567890123abcdef4567",
		},
		{
			ID:          "gist7890123456abcdef7890",
			Name:        "Docker Deployment Guide",
			Author:      "Carol Wilson",
			Description: "Container deployment best practices",
			Tags:        []string{"docker", "deployment"},
			Version:     "1.5",
			GistURL:     "https://gist.github.com/carol/gist7890123456abcdef7890",
		},
	}
}

// TestNewSyncCommand tests the command creation and configuration
func TestNewSyncCommand(t *testing.T) {
	testCases := []struct {
		name          string
		promptService service.PromptService
		expectPanic   bool
	}{
		{
			name:          "successful command creation with valid service",
			promptService: NewMockSyncPromptService(),
			expectPanic:   false,
		},
		{
			name:          "command creation with nil service (should not panic)",
			promptService: nil,
			expectPanic:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.expectPanic {
					t.Errorf("Unexpected panic: %v", r)
				} else if r == nil && tc.expectPanic {
					t.Errorf("Expected panic but none occurred")
				}
			}()

			cmd := NewSyncCommand(tc.promptService)
			cobraCmd := (*cobra.Command)(cmd)

			if !tc.expectPanic {
				// Verify command structure
				if cobraCmd == nil {
					t.Fatal("Expected command but got nil")
				}

				// Check command configuration
				if cobraCmd.Use != "sync" {
					t.Errorf("Expected Use 'sync', got %q", cobraCmd.Use)
				}

				if cobraCmd.Short != "将远程 GitHub Gist 数据同步到本地缓存" {
					t.Errorf("Expected Short '将远程 GitHub Gist 数据同步到本地缓存', got %q", cobraCmd.Short)
				}

				if !strings.Contains(cobraCmd.Long, "同步命令实现完整的缓存同步流程") {
					t.Errorf("Expected Long description to contain sync flow description")
				}

				if !strings.Contains(cobraCmd.Example, "pv sync") {
					t.Errorf("Expected Example to contain 'pv sync'")
				}

				// Check Run function is set
				if cobraCmd.Run == nil {
					t.Errorf("Expected Run function to be set")
				}

				// Check verbose flag is available
				verboseFlag := cobraCmd.Flags().Lookup("verbose")
				if verboseFlag == nil {
					t.Errorf("Expected verbose flag to be available")
				}
			}
		})
	}
}

// TestSyncCommand_CommandConfiguration tests command parameters and flags
func TestSyncCommand_CommandConfiguration(t *testing.T) {
	mockService := NewMockSyncPromptService()
	cmd := NewSyncCommand(mockService)
	cobraCmd := (*cobra.Command)(cmd)

	// Test command properties
	expectedUse := "sync"
	if cobraCmd.Use != expectedUse {
		t.Errorf("Expected Use %q, got %q", expectedUse, cobraCmd.Use)
	}

	expectedShort := "将远程 GitHub Gist 数据同步到本地缓存"
	if cobraCmd.Short != expectedShort {
		t.Errorf("Expected Short %q, got %q", expectedShort, cobraCmd.Short)
	}

	// Test help text contains expected information
	longDescriptionKeywords := []string{
		"同步命令实现完整的缓存同步流程",
		"获取远程提示词索引列表",
		"串行下载所有提示词内容到本地缓存",
		"正在下载 X/Y",
		"单个提示词失败时继续处理其他提示词",
		"显示最终同步统计信息",
	}

	for _, keyword := range longDescriptionKeywords {
		if !strings.Contains(cobraCmd.Long, keyword) {
			t.Errorf("Expected Long description to contain %q", keyword)
		}
	}

	// Test example contains expected commands
	exampleCommands := []string{
		"pv sync",
		"pv sync --verbose",
	}

	for _, command := range exampleCommands {
		if !strings.Contains(cobraCmd.Example, command) {
			t.Errorf("Expected Example to contain %q", command)
		}
	}

	// Test verbose flag configuration
	verboseFlag := cobraCmd.Flags().Lookup("verbose")
	if verboseFlag == nil {
		t.Fatal("Expected verbose flag to be available")
	}

	if verboseFlag.Shorthand != "v" {
		t.Errorf("Expected verbose flag shorthand 'v', got %q", verboseFlag.Shorthand)
	}

	if verboseFlag.Usage != "显示详细的同步过程和错误信息" {
		t.Errorf("Expected verbose flag usage description, got %q", verboseFlag.Usage)
	}
}

// TestSyncCommand_SuccessfulSync tests complete successful sync scenario (需求 3.1)
func TestSyncCommand_SuccessfulSync(t *testing.T) {
	testCases := []struct {
		name    string
		prompts []model.Prompt
		verbose bool
	}{
		{
			name:    "successful sync with multiple prompts - normal mode",
			prompts: createSyncTestPrompts(),
			verbose: false,
		},
		{
			name:    "successful sync with multiple prompts - verbose mode",
			prompts: createSyncTestPrompts(),
			verbose: true,
		},
		{
			name:    "successful sync with single prompt",
			prompts: createSyncTestPrompts()[:1],
			verbose: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := NewMockSyncPromptService()
			mockService.listPromptsResult = tc.prompts
			mockService.getPromptContentResult = "test content"

			cmd := NewSyncCommand(mockService)
			cobraCmd := (*cobra.Command)(cmd)
			if tc.verbose {
				cobraCmd.Flags().Set("verbose", "true")
			}

			// Capture output
			output := captureOutput(func() {
				cobraCmd.Run(cobraCmd, []string{})
			})

			// Verify service calls
			if mockService.listPromptsCalls != 1 {
				t.Errorf("Expected 1 ListPrompts call, got %d", mockService.listPromptsCalls)
			}

			expectedContentCalls := len(tc.prompts)
			if len(mockService.getPromptContentCalls) != expectedContentCalls {
				t.Errorf("Expected %d GetPromptContent calls, got %d", 
					expectedContentCalls, len(mockService.getPromptContentCalls))
			}

			// Verify output contains sync start message
			if !strings.Contains(output, "🔄 开始同步提示词...") {
				t.Errorf("Expected sync start message, got:\n%s", output)
			}

			// Verify progress display format (需求 3.2)
			for i, prompt := range tc.prompts {
				expectedProgress := fmt.Sprintf("⬇️  正在下载 %d/%d: %s", i+1, len(tc.prompts), prompt.Name)
				if !strings.Contains(output, expectedProgress) {
					t.Errorf("Expected progress message %q, got:\n%s", expectedProgress, output)
				}

				// In verbose mode, should show prompt ID
				if tc.verbose {
					if !strings.Contains(output, prompt.ID) {
						t.Errorf("Expected verbose mode to show prompt ID %s, got:\n%s", prompt.ID, output)
					}
				}
			}

			// Verify statistics display (需求 3.3)
			expectedStats := []string{
				"📊 同步完成统计:",
				fmt.Sprintf("总计: %d", len(tc.prompts)),
				fmt.Sprintf("成功: %d", len(tc.prompts)),
				"失败: 0",
				"跳过: 0",
				"🎉 所有提示词同步并缓存成功!",
			}

			for _, stat := range expectedStats {
				if !strings.Contains(output, stat) {
					t.Errorf("Expected statistic %q, got:\n%s", stat, output)
				}
			}

			// Verify success message
			successMessages := []string{
				"提示词已缓存到本地，现在可以离线使用。",
				"运行 'pv list' 查看同步的提示词。",
			}

			for _, msg := range successMessages {
				if !strings.Contains(output, msg) {
					t.Errorf("Expected success message %q, got:\n%s", msg, output)
				}
			}
		})
	}
}

// TestSyncCommand_EmptyPromptList tests scenario with no prompts to sync
func TestSyncCommand_EmptyPromptList(t *testing.T) {
	mockService := NewMockSyncPromptService()
	mockService.listPromptsResult = []model.Prompt{}

	cmd := NewSyncCommand(mockService)
	cobraCmd := (*cobra.Command)(cmd)

	output := captureOutput(func() {
		cobraCmd.Run(cobraCmd, []string{})
	})

	// Verify service calls
	if mockService.listPromptsCalls != 1 {
		t.Errorf("Expected 1 ListPrompts call, got %d", mockService.listPromptsCalls)
	}

	if len(mockService.getPromptContentCalls) != 0 {
		t.Errorf("Expected 0 GetPromptContent calls, got %d", len(mockService.getPromptContentCalls))
	}

	// Verify output messages
	expectedMessages := []string{
		"📝 没有找到需要同步的提示词。",
		"要添加提示词:",
		"在 GitHub Gists 中直接创建提示词",
		"使用 'pv add <file>' 添加新提示词",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected message %q, got:\n%s", msg, output)
		}
	}
}

// TestSyncCommand_AuthenticationFailure tests authentication error scenarios (需求 3.4)
func TestSyncCommand_AuthenticationFailure(t *testing.T) {
	testCases := []struct {
		name         string
		errorType    errors.ErrorType
		errorMessage string
		expectedMsg  []string
	}{
		{
			name:         "authentication error",
			errorType:    errors.ErrAuth,
			errorMessage: "invalid token",
			expectedMsg: []string{
				"❌ 认证错误: invalid token",
				"请运行 'pv auth login' 重新登录 GitHub。",
			},
		},
		{
			name:         "network error",
			errorType:    errors.ErrNetwork,
			errorMessage: "connection timeout",
			expectedMsg: []string{
				"❌ 网络错误: connection timeout",
				"请检查网络连接后重试。",
			},
		},
		{
			name:         "generic app error",
			errorType:    errors.ErrStorage,
			errorMessage: "storage error",
			expectedMsg: []string{
				"❌ 错误: storage error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := NewMockSyncPromptService()
			appErr := errors.NewAppError(tc.errorType, tc.errorMessage, nil)
			mockService.listPromptsError = &appErr

			cmd := NewSyncCommand(mockService)
			cobraCmd := (*cobra.Command)(cmd)

			output := captureOutput(func() {
				cobraCmd.Run(cobraCmd, []string{})
			})

			// Verify service calls
			if mockService.listPromptsCalls != 1 {
				t.Errorf("Expected 1 ListPrompts call, got %d", mockService.listPromptsCalls)
			}

			if len(mockService.getPromptContentCalls) != 0 {
				t.Errorf("Expected 0 GetPromptContent calls, got %d", len(mockService.getPromptContentCalls))
			}

			// Verify error messages
			for _, msg := range tc.expectedMsg {
				if !strings.Contains(output, msg) {
					t.Errorf("Expected error message %q, got:\n%s", msg, output)
				}
			}
		})
	}
}

// TestSyncCommand_PartialSyncFailure tests partial failure scenarios (需求 3.5)
func TestSyncCommand_PartialSyncFailure(t *testing.T) {
	testCases := []struct {
		name            string
		prompts         []model.Prompt
		failingPrompts  map[string]error // promptID -> error
		verbose         bool
		expectedSuccess int
		expectedFailed  int
	}{
		{
			name:    "partial failure - one prompt fails",
			prompts: createSyncTestPrompts(),
			failingPrompts: map[string]error{
				"gist4567890123abcdef4567": createAppError(errors.ErrNetwork, "network timeout"),
			},
			verbose:         false,
			expectedSuccess: 2,
			expectedFailed:  1,
		},
		{
			name:    "partial failure - multiple prompts fail",
			prompts: createSyncTestPrompts(),
			failingPrompts: map[string]error{
				"gist1234567890abcdef1234": createAppError(errors.ErrStorage, "storage error"),
				"gist7890123456abcdef7890": createAppError(errors.ErrNetwork, "connection lost"),
			},
			verbose:         true,
			expectedSuccess: 1,
			expectedFailed:  2,
		},
		{
			name:    "all prompts fail",
			prompts: createSyncTestPrompts()[:2],
			failingPrompts: map[string]error{
				"gist1234567890abcdef1234": createAppError(errors.ErrAuth, "unauthorized"),
				"gist4567890123abcdef4567": createAppError(errors.ErrNetwork, "timeout"),
			},
			verbose:         false,
			expectedSuccess: 0,
			expectedFailed:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := NewMockSyncPromptService()
			mockService.listPromptsResult = tc.prompts
			mockService.getPromptContentResult = "test content"
			mockService.promptContentErrors = tc.failingPrompts

			cmd := NewSyncCommand(mockService)
			cobraCmd := (*cobra.Command)(cmd)
			if tc.verbose {
				cobraCmd.Flags().Set("verbose", "true")
			}

			output := captureOutput(func() {
				cobraCmd.Run(cobraCmd, []string{})
			})

			// Verify service calls
			if mockService.listPromptsCalls != 1 {
				t.Errorf("Expected 1 ListPrompts call, got %d", mockService.listPromptsCalls)
			}

			expectedContentCalls := len(tc.prompts)
			if len(mockService.getPromptContentCalls) != expectedContentCalls {
				t.Errorf("Expected %d GetPromptContent calls, got %d", 
					expectedContentCalls, len(mockService.getPromptContentCalls))
			}

			// Verify progress display for all prompts (including failed ones)
			for i, prompt := range tc.prompts {
				expectedProgress := fmt.Sprintf("⬇️  正在下载 %d/%d: %s", i+1, len(tc.prompts), prompt.Name)
				if !strings.Contains(output, expectedProgress) {
					t.Errorf("Expected progress message %q, got:\n%s", expectedProgress, output)
				}

				// Check for failure indicators
				if _, shouldFail := tc.failingPrompts[prompt.ID]; shouldFail {
					if tc.verbose {
						// In verbose mode, should show detailed error
						if !strings.Contains(output, fmt.Sprintf("下载 '%s' 失败:", prompt.Name)) {
							t.Errorf("Expected verbose failure message for %s, got:\n%s", prompt.Name, output)
						}
					} else {
						// In normal mode, should show simple failure indicator
						if !strings.Contains(output, "❌ 失败") {
							t.Errorf("Expected failure indicator, got:\n%s", output)
						}
					}
				}
			}

			// Verify statistics (需求 3.3)
			expectedStats := []string{
				"📊 同步完成统计:",
				fmt.Sprintf("总计: %d", len(tc.prompts)),
				fmt.Sprintf("成功: %d", tc.expectedSuccess),
				fmt.Sprintf("失败: %d", tc.expectedFailed),
				"跳过: 0",
			}

			for _, stat := range expectedStats {
				if !strings.Contains(output, stat) {
					t.Errorf("Expected statistic %q, got:\n%s", stat, output)
				}
			}

			// Verify result summary based on success/failure counts
			if tc.expectedSuccess == len(tc.prompts) {
				if !strings.Contains(output, "🎉 所有提示词同步并缓存成功!") {
					t.Errorf("Expected complete success message, got:\n%s", output)
				}
			} else if tc.expectedSuccess > 0 {
				expectedPartialMsg := fmt.Sprintf("✅ %d 个提示词同步并缓存成功", tc.expectedSuccess)
				if !strings.Contains(output, expectedPartialMsg) {
					t.Errorf("Expected partial success message %q, got:\n%s", expectedPartialMsg, output)
				}
				if tc.expectedFailed > 0 {
					expectedFailureMsg := fmt.Sprintf("，%d 个失败", tc.expectedFailed)
					if !strings.Contains(output, expectedFailureMsg) {
						t.Errorf("Expected partial failure message %q, got:\n%s", expectedFailureMsg, output)
					}
				}
			} else {
				if !strings.Contains(output, "❌ 同步失败，没有提示词被成功下载和缓存") {
					t.Errorf("Expected complete failure message, got:\n%s", output)
				}
			}

			// Verify error details based on verbose mode
			if tc.verbose && tc.expectedFailed > 0 {
				if !strings.Contains(output, "详细错误信息:") {
					t.Errorf("Expected detailed error section in verbose mode, got:\n%s", output)
				}
			} else if !tc.verbose && tc.expectedFailed > 0 {
				if !strings.Contains(output, "使用 --verbose 查看详细错误信息") {
					t.Errorf("Expected verbose suggestion message, got:\n%s", output)
				}
			}
		})
	}
}

// TestSyncCommand_ProgressDisplay tests progress display functionality (需求 3.2)
func TestSyncCommand_ProgressDisplay(t *testing.T) {
	prompts := createSyncTestPrompts()
	
	testCases := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "progress display in normal mode",
			verbose: false,
		},
		{
			name:    "progress display in verbose mode",
			verbose: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := NewMockSyncPromptService()
			mockService.listPromptsResult = prompts
			mockService.getPromptContentResult = "test content"

			cmd := NewSyncCommand(mockService)
			cobraCmd := (*cobra.Command)(cmd)
			if tc.verbose {
				cobraCmd.Flags().Set("verbose", "true")
			}

			output := captureOutput(func() {
				cobraCmd.Run(cobraCmd, []string{})
			})

			// Verify progress format matches specification: "正在下载 X/Y"
			for i, prompt := range prompts {
				expectedProgress := fmt.Sprintf("⬇️  正在下载 %d/%d: %s", i+1, len(prompts), prompt.Name)
				if !strings.Contains(output, expectedProgress) {
					t.Errorf("Expected exact progress format %q, got:\n%s", expectedProgress, output)
				}

				// In verbose mode, should also show prompt ID
				if tc.verbose {
					expectedVerboseProgress := fmt.Sprintf("(%s)", prompt.ID)
					if !strings.Contains(output, expectedVerboseProgress) {
						t.Errorf("Expected verbose progress with ID %q, got:\n%s", expectedVerboseProgress, output)
					}
				}
			}

			// Verify initial count display
			expectedCountMsg := fmt.Sprintf("📋 发现 %d 个提示词需要同步", len(prompts))
			if !strings.Contains(output, expectedCountMsg) {
				t.Errorf("Expected count message %q, got:\n%s", expectedCountMsg, output)
			}
		})
	}
}

// TestSyncCommand_ErrorStatistics tests error reporting and statistics (需求 3.3)
func TestSyncCommand_ErrorStatistics(t *testing.T) {
	prompts := createSyncTestPrompts()
	
	testCases := []struct {
		name           string
		failingPrompts map[string]error
		verbose        bool
		expectedErrors []string
	}{
		{
			name: "single error - verbose mode",
			failingPrompts: map[string]error{
				"gist1234567890abcdef1234": createAppError(errors.ErrNetwork, "timeout"),
			},
			verbose: true,
			expectedErrors: []string{
				"下载 'Go Best Practices' 失败:",
			},
		},
		{
			name: "multiple errors - verbose mode",
			failingPrompts: map[string]error{
				"gist1234567890abcdef1234": createAppError(errors.ErrNetwork, "timeout"),
				"gist7890123456abcdef7890": createAppError(errors.ErrAuth, "unauthorized"),
			},
			verbose: true,
			expectedErrors: []string{
				"下载 'Go Best Practices' 失败:",
				"下载 'Docker Deployment Guide' 失败:",
				"详细错误信息:",
			},
		},
		{
			name: "multiple errors - normal mode",
			failingPrompts: map[string]error{
				"gist1234567890abcdef1234": createAppError(errors.ErrNetwork, "timeout"),
				"gist4567890123abcdef4567": createAppError(errors.ErrStorage, "storage error"),
			},
			verbose: false,
			expectedErrors: []string{
				"使用 --verbose 查看详细错误信息",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := NewMockSyncPromptService()
			mockService.listPromptsResult = prompts
			mockService.getPromptContentResult = "test content"
			mockService.promptContentErrors = tc.failingPrompts

			cmd := NewSyncCommand(mockService)
			cobraCmd := (*cobra.Command)(cmd)
			if tc.verbose {
				cobraCmd.Flags().Set("verbose", "true")
			}

			output := captureOutput(func() {
				cobraCmd.Run(cobraCmd, []string{})
			})

			// Verify error count in statistics
			expectedFailCount := fmt.Sprintf("失败: %d", len(tc.failingPrompts))
			if !strings.Contains(output, expectedFailCount) {
				t.Errorf("Expected fail count %q, got:\n%s", expectedFailCount, output)
			}

			// Verify expected error messages
			for _, expectedError := range tc.expectedErrors {
				if !strings.Contains(output, expectedError) {
					t.Errorf("Expected error message %q, got:\n%s", expectedError, output)
				}
			}

			// Verify statistics section always appears
			if !strings.Contains(output, "📊 同步完成统计:") {
				t.Errorf("Expected statistics section, got:\n%s", output)
			}
		})
	}
}

// TestSyncCommand_VerboseMode tests verbose mode specific functionality
func TestSyncCommand_VerboseMode(t *testing.T) {
	prompts := createSyncTestPrompts()[:2] // Use fewer prompts for clearer testing
	
	mockService := NewMockSyncPromptService()
	mockService.listPromptsResult = prompts
	mockService.getPromptContentResult = "test content"
	// Make one prompt fail
	mockService.promptContentErrors = map[string]error{
		"gist4567890123abcdef4567": createAppError(errors.ErrNetwork, "connection timeout"),
	}

	cmd := NewSyncCommand(mockService)
	cobraCmd := (*cobra.Command)(cmd)
	cobraCmd.Flags().Set("verbose", "true")

	output := captureOutput(func() {
		cobraCmd.Run(cobraCmd, []string{})
	})

	// Verify verbose-specific features
	verboseFeatures := []string{
		// Progress should show prompt IDs
		"(gist1234567890abcdef1234)",
		"(gist4567890123abcdef4567)",
		
		// Success should show detailed message
		"✅ 成功缓存",
		
		// Failures should show detailed error messages
		"❌ 下载 'Python Data Analysis' 失败:",
		"connection timeout",
		
		// Error details section
		"详细错误信息:",
	}

	for _, feature := range verboseFeatures {
		if !strings.Contains(output, feature) {
			t.Errorf("Expected verbose feature %q, got:\n%s", feature, output)
		}
	}
}

// TestSyncCommand_ContinuesOnIndividualFailures tests that sync continues processing other prompts when one fails
func TestSyncCommand_ContinuesOnIndividualFailures(t *testing.T) {
	prompts := createSyncTestPrompts()
	
	mockService := NewMockSyncPromptService()
	mockService.listPromptsResult = prompts
	mockService.getPromptContentResult = "test content"
	
	// Make the middle prompt fail
	mockService.promptContentErrors = map[string]error{
		"gist4567890123abcdef4567": errors.NewAppError(errors.ErrNetwork, "network error", nil),
	}

	cmd := NewSyncCommand(mockService)
	cobraCmd := (*cobra.Command)(cmd)

	output := captureOutput(func() {
		cobraCmd.Run(cobraCmd, []string{})
	})

	// Verify all prompts were attempted
	expectedContentCalls := len(prompts)
	if len(mockService.getPromptContentCalls) != expectedContentCalls {
		t.Errorf("Expected %d GetPromptContent calls, got %d", 
			expectedContentCalls, len(mockService.getPromptContentCalls))
	}

	// Verify all prompt IDs were called
	for _, prompt := range prompts {
		found := false
		for _, calledID := range mockService.getPromptContentCalls {
			if calledID == prompt.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected prompt %s to be processed, but it wasn't called", prompt.ID)
		}
	}

	// Verify statistics show correct success/failure counts
	expectedStats := []string{
		"总计: 3",
		"成功: 2",
		"失败: 1",
	}

	for _, stat := range expectedStats {
		if !strings.Contains(output, stat) {
			t.Errorf("Expected statistic %q, got:\n%s", stat, output)
		}
	}

	// Verify partial success message
	if !strings.Contains(output, "✅ 2 个提示词同步并缓存成功，1 个失败") {
		t.Errorf("Expected partial success message, got:\n%s", output)
	}
}