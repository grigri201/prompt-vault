package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestMockManager_Initialize(t *testing.T) {
	// Test with custom function
	called := false
	mock := &MockManager{
		InitializeFunc: func(ctx context.Context) error {
			called = true
			return nil
		},
	}

	err := mock.Initialize(context.Background())
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}
	if !called {
		t.Error("InitializeFunc was not called")
	}

	// Test default behavior
	mock2 := &MockManager{}
	err = mock2.Initialize(context.Background())
	if err != nil {
		t.Errorf("Initialize() with default behavior error = %v", err)
	}
	if !mock2.IsInitialized() {
		t.Error("Default Initialize should set initialized to true")
	}
}

func TestMockManager_SavePrompt(t *testing.T) {
	prompt := &models.Prompt{
		GistID: "test-id",
		PromptMeta: models.PromptMeta{
			Name:     "Test Prompt",
			Author:   "Test Author",
			Category: "Test",
			Tags:     []string{"test"},
		},
	}

	// Test with custom function
	var savedPrompt *models.Prompt
	mock := &MockManager{
		SavePromptFunc: func(p *models.Prompt) error {
			savedPrompt = p
			return nil
		},
	}

	err := mock.SavePrompt(prompt)
	if err != nil {
		t.Errorf("SavePrompt() error = %v", err)
	}
	if savedPrompt != prompt {
		t.Error("SavePromptFunc was not called with correct prompt")
	}
}

func TestMockManager_GetPrompt(t *testing.T) {
	// Test with custom function
	expectedPrompt := &models.Prompt{
		GistID: "custom-id",
		PromptMeta: models.PromptMeta{
			Name:     "Custom Prompt",
			Author:   "Custom Author",
			Category: "Custom",
			Tags:     []string{"custom"},
		},
	}
	mock := &MockManager{
		GetPromptFunc: func(id string) (*models.Prompt, error) {
			if id == "custom-id" {
				return expectedPrompt, nil
			}
			return nil, errors.New("not found")
		},
	}

	prompt, err := mock.GetPrompt("custom-id")
	if err != nil {
		t.Errorf("GetPrompt() error = %v", err)
	}
	if prompt != expectedPrompt {
		t.Error("GetPromptFunc did not return expected prompt")
	}

	// Test error case
	_, err = mock.GetPrompt("wrong-id")
	if err == nil {
		t.Error("GetPrompt() should return error for wrong ID")
	}

	// Test default behavior
	mock2 := &MockManager{}
	prompt2, err := mock2.GetPrompt("any-id")
	if err != nil {
		t.Errorf("GetPrompt() with default behavior error = %v", err)
	}
	if prompt2.GistID != "any-id" {
		t.Error("Default GetPrompt should return prompt with requested ID")
	}
}

func TestMockManager_GetIndex(t *testing.T) {
	// Test with custom function
	expectedIndex := &models.Index{
		Username: "test-user",
		Entries: []models.IndexEntry{
			{Name: "Prompt 1", Author: "Author 1"},
			{Name: "Prompt 2", Author: "Author 2"},
		},
	}
	mock := &MockManager{
		GetIndexFunc: func() (*models.Index, error) {
			return expectedIndex, nil
		},
	}

	index, err := mock.GetIndex()
	if err != nil {
		t.Errorf("GetIndex() error = %v", err)
	}
	if index.Username != expectedIndex.Username {
		t.Error("GetIndexFunc did not return expected index")
	}
	if len(index.Entries) != 2 {
		t.Error("GetIndexFunc did not return expected entries")
	}
}
