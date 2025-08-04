package tui

import (
	"testing"

	"github.com/grigri/pv/internal/model"
)

func TestTestPromptGenerator(t *testing.T) {
	generator := NewTestPromptGenerator()
	
	// Test GeneratePrompt
	prompt1 := generator.GeneratePrompt()
	if prompt1.ID == "" {
		t.Error("expected non-empty ID")
	}
	if prompt1.Name == "" {
		t.Error("expected non-empty name")
	}
	if prompt1.Author == "" {
		t.Error("expected non-empty author")
	}
	
	// Test GeneratePrompts
	prompts := generator.GeneratePrompts(3)
	if len(prompts) != 3 {
		t.Errorf("expected 3 prompts, got %d", len(prompts))
	}
	
	// Ensure each prompt is unique
	for i := 0; i < len(prompts); i++ {
		for j := i + 1; j < len(prompts); j++ {
			if prompts[i].ID == prompts[j].ID {
				t.Errorf("expected unique IDs, found duplicate: %s", prompts[i].ID)
			}
		}
	}
	
	// Test GeneratePromptsWithPattern
	keyword := "golang"
	patternPrompts := generator.GeneratePromptsWithPattern(3, keyword)
	if len(patternPrompts) != 3 {
		t.Errorf("expected 3 pattern prompts, got %d", len(patternPrompts))
	}
	
	// Check that keyword appears in at least one field of each prompt
	for _, prompt := range patternPrompts {
		found := false
		if contains(prompt.Name, keyword) || contains(prompt.Author, keyword) || contains(prompt.Description, keyword) {
			found = true
		}
		if !found {
			t.Errorf("expected keyword %s to appear in prompt %s", keyword, prompt.ID)
		}
	}
	
	// Test GenerateEmptyPromptList
	empty := generator.GenerateEmptyPromptList()
	if len(empty) != 0 {
		t.Errorf("expected empty list, got %d items", len(empty))
	}
	
	// Test GeneratePromptWithSpecificData
	specific := generator.GeneratePromptWithSpecificData("test-id", "test-name", "test-author", "test-url")
	if specific.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", specific.ID)
	}
	if specific.Name != "test-name" {
		t.Errorf("expected name 'test-name', got %s", specific.Name)
	}
	if specific.Author != "test-author" {
		t.Errorf("expected author 'test-author', got %s", specific.Author)
	}
	if specific.GistURL != "test-url" {
		t.Errorf("expected URL 'test-url', got %s", specific.GistURL)
	}
	
	// Test GenerateRandomPrompt
	random := generator.GenerateRandomPrompt()
	if random.ID == "" {
		t.Error("expected non-empty random ID")
	}
	if random.Name == "" {
		t.Error("expected non-empty random name")
	}
}

func TestTUIStateValidator(t *testing.T) {
	mockTUI := NewMockTUI()
	validator := NewTUIStateValidator(mockTUI)
	
	// Test ValidateNoMethodCalls on fresh mock
	if !validator.ValidateNoMethodCalls() {
		t.Error("expected no method calls on fresh mock")
	}
	
	// Test after adding some calls
	testPrompts := CreateStandardTestPrompts()
	mockTUI.ShowPromptList(testPrompts)
	
	if validator.ValidateNoMethodCalls() {
		t.Error("expected method calls after ShowPromptList")
	}
	
	// Test ValidatePromptListCalled
	if !validator.ValidatePromptListCalled(testPrompts) {
		t.Error("expected ValidatePromptListCalled to return true")
	}
	
	// Test ValidateMethodCallCount
	if !validator.ValidateMethodCallCount("ShowPromptList", 1) {
		t.Error("expected ValidateMethodCallCount to return true for 1 call")
	}
	
	// Test with confirmation
	testPrompt := CreateSingleTestPrompt()
	mockTUI.ShowConfirm(testPrompt)
	
	if !validator.ValidateConfirmCalled(testPrompt) {
		t.Error("expected ValidateConfirmCalled to return true")
	}
	
	// Test ValidateMethodCallSequence
	expectedSequence := []string{"ShowPromptList", "ShowConfirm"}
	if !validator.ValidateMethodCallSequence(expectedSequence) {
		t.Error("expected ValidateMethodCallSequence to return true for correct sequence")
	}
}

func TestErrorScenarioSimulator(t *testing.T) {
	mockTUI := NewMockTUI()
	simulator := NewErrorScenarioSimulator(mockTUI)
	
	// Test SimulateUserCancellation
	simulator.SimulateUserCancellation()
	if !mockTUI.ShouldSimulateUserCancel {
		t.Error("expected ShouldSimulateUserCancel to be true")
	}
	
	// Test SimulateSelectionError
	mockTUI.Reset()
	simulator.SimulateSelectionError()
	if !mockTUI.ShouldSimulateSelectionErr {
		t.Error("expected ShouldSimulateSelectionErr to be true")
	}
	
	// Test SimulateConfirmationError
	mockTUI.Reset()
	simulator.SimulateConfirmationError()
	if !mockTUI.ShouldSimulateConfirmErr {
		t.Error("expected ShouldSimulateConfirmErr to be true")
	}
	
	// Test SimulatePromptListError
	mockTUI.Reset()
	errorMsg := "test prompt list error"
	simulator.SimulatePromptListError(errorMsg)
	if mockTUI.ShowPromptListErr == nil {
		t.Error("expected ShowPromptListErr to be set")
	}
	if mockTUI.ShowPromptListErr.Error() != errorMsg {
		t.Errorf("expected error message %s, got %s", errorMsg, mockTUI.ShowPromptListErr.Error())
	}
	
	// Test SimulateConfirmError
	mockTUI.Reset()
	confirmErrorMsg := "test confirm error"
	simulator.SimulateConfirmError(confirmErrorMsg)
	if mockTUI.ShowConfirmErr == nil {
		t.Error("expected ShowConfirmErr to be set")
	}
	if mockTUI.ShowConfirmErr.Error() != confirmErrorMsg {
		t.Errorf("expected error message %s, got %s", confirmErrorMsg, mockTUI.ShowConfirmErr.Error())
	}
	
	// Test SimulateEmptyPromptList
	emptyList := simulator.SimulateEmptyPromptList()
	if len(emptyList) != 0 {
		t.Errorf("expected empty list, got %d items", len(emptyList))
	}
	
	// Test SimulateNetworkTimeout
	simulator.SimulateNetworkTimeout()
	if !mockTUI.SimulateSlowResponse {
		t.Error("expected SimulateSlowResponse to be true")
	}
}

func TestTestScenarioBuilder(t *testing.T) {
	builder := NewTestScenarioBuilder()
	
	// Test GetMockTUI
	mockTUI := builder.GetMockTUI()
	if mockTUI == nil {
		t.Error("expected non-nil mock TUI")
	}
	
	// Test GetGenerator
	generator := builder.GetGenerator()
	if generator == nil {
		t.Error("expected non-nil generator")
	}
	
	// Test GetValidator
	validator := builder.GetValidator()
	if validator == nil {
		t.Error("expected non-nil validator")
	}
	
	// Test GetSimulator
	simulator := builder.GetSimulator()
	if simulator == nil {
		t.Error("expected non-nil simulator")
	}
	
	// Test fluent interface
	testPrompt := CreateSingleTestPrompt()
	builder = builder.WithPromptSelection(testPrompt).
		WithConfirmResult(true).
		WithUserCancellation()
	
	finalMock, finalValidator := builder.Build()
	
	if finalMock.SelectedPrompt.ID != testPrompt.ID {
		t.Error("expected selected prompt to be set")
	}
	
	if !finalMock.ConfirmResult {
		t.Error("expected confirm result to be true")
	}
	
	if !finalMock.ShouldSimulateUserCancel {
		t.Error("expected user cancellation to be set")
	}
	
	if finalValidator == nil {
		t.Error("expected non-nil validator from build")
	}
	
	// Test Reset
	builder.Reset()
	resetMock := builder.GetMockTUI()
	if resetMock.SelectedPrompt.ID != "" {
		t.Error("expected selected prompt to be reset")
	}
}

func TestMockTUI(t *testing.T) {
	mockTUI := NewMockTUI()
	
	// Test initial state
	if len(mockTUI.CallHistory) != 0 {
		t.Error("expected empty call history initially")
	}
	
	// Test ShowPromptList with normal scenario
	testPrompts := CreateStandardTestPrompts()
	mockTUI.SelectedPrompt = testPrompts[0]
	
	result, err := mockTUI.ShowPromptList(testPrompts)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	if result.ID != testPrompts[0].ID {
		t.Errorf("expected prompt ID %s, got %s", testPrompts[0].ID, result.ID)
	}
	
	// Test ShowPromptList with empty list
	_, err = mockTUI.ShowPromptList([]model.Prompt{})
	if err == nil {
		t.Error("expected error for empty prompt list")
	}
	
	// Test ShowPromptList with user cancellation
	mockTUI.Reset()
	mockTUI.ShouldSimulateUserCancel = true
	_, err = mockTUI.ShowPromptList(testPrompts)
	if err == nil {
		t.Error("expected error for user cancellation")
	}
	
	// Test ShowConfirm
	mockTUI.Reset()
	mockTUI.ConfirmResult = true
	testPrompt := CreateSingleTestPrompt()
	
	confirmed, err := mockTUI.ShowConfirm(testPrompt)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	if !confirmed {
		t.Error("expected confirmation to be true")
	}
	
	// Test GetMethodCallCount
	count := mockTUI.GetMethodCallCount("ShowConfirm")
	if count != 1 {
		t.Errorf("expected 1 ShowConfirm call, got %d", count)
	}
	
	// Test GetLastMethodCall
	lastCall := mockTUI.GetLastMethodCall()
	if lastCall == nil {
		t.Error("expected non-nil last method call")
	}
	
	if lastCall.Method != "ShowConfirm" {
		t.Errorf("expected last method to be ShowConfirm, got %s", lastCall.Method)
	}
	
	// Test VerifyMethodCalled
	if !mockTUI.VerifyMethodCalled("ShowConfirm", testPrompt) {
		t.Error("expected VerifyMethodCalled to return true")
	}
	
	// Test setup methods
	mockTUI.SetupInteractiveDeleteScenario(testPrompt, true)
	if mockTUI.SelectedPrompt.ID != testPrompt.ID {
		t.Error("expected selected prompt to be set by setup")
	}
	
	mockTUI.SetupDirectDeleteScenario(false)
	if mockTUI.ConfirmResult {
		t.Error("expected confirm result to be false")
	}
	
	// Test SetupErrorScenario
	err = mockTUI.SetupErrorScenario("user_cancel")
	if err != nil {
		t.Errorf("expected no error setting up scenario, got %v", err)
	}
	
	err = mockTUI.SetupErrorScenario("unknown_error")
	if err == nil {
		t.Error("expected error for unknown error type")
	}
	
	// Test GetInteractionSummary
	summary := mockTUI.GetInteractionSummary()
	if summary == "" {
		t.Error("expected non-empty interaction summary")
	}
}

func TestCreateStandardTestPrompts(t *testing.T) {
	prompts := CreateStandardTestPrompts()
	
	if len(prompts) != 2 {
		t.Errorf("expected 2 standard prompts, got %d", len(prompts))
	}
	
	if prompts[0].ID != TestPromptID1 {
		t.Errorf("expected first prompt ID %s, got %s", TestPromptID1, prompts[0].ID)
	}
	
	if prompts[1].ID != TestPromptID2 {
		t.Errorf("expected second prompt ID %s, got %s", TestPromptID2, prompts[1].ID)
	}
}

func TestCreateEmptyTestPrompts(t *testing.T) {
	prompts := CreateEmptyTestPrompts()
	
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts, got %d", len(prompts))
	}
}

func TestCreateSingleTestPrompt(t *testing.T) {
	prompt := CreateSingleTestPrompt()
	
	if prompt.ID != TestPromptID1 {
		t.Errorf("expected prompt ID %s, got %s", TestPromptID1, prompt.ID)
	}
	
	if prompt.Name != TestPromptName1 {
		t.Errorf("expected prompt name %s, got %s", TestPromptName1, prompt.Name)
	}
}

func TestComparePromptSlices(t *testing.T) {
	prompts1 := CreateStandardTestPrompts()
	prompts2 := CreateStandardTestPrompts()
	
	// Same content should be equal
	if !comparePromptSlices(prompts1, prompts2) {
		t.Error("expected equal prompt slices to compare as equal")
	}
	
	// Different lengths should not be equal
	if comparePromptSlices(prompts1, prompts1[:1]) {
		t.Error("expected different length slices to compare as not equal")
	}
	
	// Different IDs should not be equal
	prompts2[0].ID = "different-id"
	if comparePromptSlices(prompts1, prompts2) {
		t.Error("expected slices with different IDs to compare as not equal")
	}
}

func TestListModeString(t *testing.T) {
	tests := []struct {
		mode     ListMode
		expected string
	}{
		{ListAll, "all"},
		{ListFiltered, "filtered"},
		{ListMode(999), "unknown"},
	}
	
	for _, tt := range tests {
		result := tt.mode.String()
		if result != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, result)
		}
	}
}