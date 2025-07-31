package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptMeta_Validate(t *testing.T) {
	tests := []struct {
		name    string
		meta    PromptMeta
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid prompt meta",
			meta: PromptMeta{
				Name:   "Test Prompt",
				Author: "testuser",
				Tags:   []string{"test", "example"},
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			meta: PromptMeta{
				Author: "testuser",
				Tags:   []string{"test"},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "Missing author",
			meta: PromptMeta{
				Name: "Test Prompt",
				Tags: []string{"test"},
			},
			wantErr: true,
			errMsg:  "author is required",
		},
		{
			name: "Missing tags",
			meta: PromptMeta{
				Name:   "Test Prompt",
				Author: "testuser",
				Tags:   []string{},
			},
			wantErr: true,
			errMsg:  "at least one tag is required",
		},
		{
			name: "Valid with all optional fields",
			meta: PromptMeta{
				Name:        "Test Prompt",
				Author:      "testuser",
				Tags:        []string{"test"},
				Version:     "1.0",
				Description: "Test description",
				Parent:      "parent-id",
				ID:          "test-id",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromptMeta_ValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"Valid ID", "test-prompt-123", false},
		{"Valid underscore", "test_prompt_123", false},
		{"Empty ID (optional)", "", false},
		{"Too short", "ab", true},
		{"Too long", strings.Repeat("a", 101), true},
		{"Invalid characters", "test prompt!", true},
		{"Invalid characters with spaces", "test prompt", true},
		{"Valid minimum length", "abc", false},
		{"Valid maximum length", strings.Repeat("a", 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := PromptMeta{ID: tt.id}
			err := meta.ValidateID()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromptMeta_SetDefaultVersion(t *testing.T) {
	t.Run("Sets version when empty", func(t *testing.T) {
		meta := PromptMeta{}
		meta.SetDefaultVersion()
		assert.NotEmpty(t, meta.Version)
		assert.Regexp(t, `^\d+$`, meta.Version)
	})

	t.Run("Preserves existing version", func(t *testing.T) {
		meta := PromptMeta{Version: "1.0"}
		meta.SetDefaultVersion()
		assert.Equal(t, "1.0", meta.Version)
	})
}

func TestPrompt_ToIndexEntry(t *testing.T) {
	now := time.Now()
	prompt := Prompt{
		PromptMeta: PromptMeta{
			Name:        "Test Prompt",
			Author:      "testuser",
			Tags:        []string{"test"},
			Version:     "1.0",
			Description: "Test description",
			Parent:      "parent-id",
			ID:          "test-id",
		},
		GistID:    "gist123",
		GistURL:   "https://gist.github.com/user/gist123",
		UpdatedAt: now,
		Content:   "Test content",
	}

	entry := prompt.ToIndexEntry()

	assert.Equal(t, "gist123", entry.GistID)
	assert.Equal(t, "https://gist.github.com/user/gist123", entry.GistURL)
	assert.Equal(t, "Test Prompt", entry.Name)
	assert.Equal(t, "testuser", entry.Author)
	assert.Equal(t, []string{"test"}, entry.Tags)
	assert.Equal(t, "1.0", entry.Version)
	assert.Equal(t, "Test description", entry.Description)
	assert.Equal(t, "parent-id", entry.Parent)
	assert.Equal(t, "test-id", entry.ID)
	assert.Equal(t, now, entry.UpdatedAt)
}

func TestPrompt_Validation_Integration(t *testing.T) {
	t.Run("Full prompt validation with invalid ID", func(t *testing.T) {
		prompt := Prompt{
			PromptMeta: PromptMeta{
				Name:   "Test Prompt",
				Author: "testuser",
				Tags:   []string{"test"},
				ID:     "a", // Too short
			},
		}

		err := prompt.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 3 characters")
	})

	t.Run("Full prompt validation success", func(t *testing.T) {
		prompt := Prompt{
			PromptMeta: PromptMeta{
				Name:   "Test Prompt",
				Author: "testuser",
				Tags:   []string{"test"},
				ID:     "valid-id",
			},
		}

		err := prompt.Validate()
		assert.NoError(t, err)
	})
}
