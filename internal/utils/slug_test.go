package utils

import (
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name       string
		promptName string
		author     string
		want       string
	}{
		{
			name:       "simple name and author",
			promptName: "API Documentation",
			author:     "john",
			want:       "john-api-documentation",
		},
		{
			name:       "name with special characters",
			promptName: "API Documentation (v2.0)",
			author:     "john",
			want:       "john-api-documentation-v20",
		},
		{
			name:       "name with numbers",
			promptName: "Code Review 123",
			author:     "alice",
			want:       "alice-code-review-123",
		},
		{
			name:       "multiple spaces",
			promptName: "Code   Review   Template",
			author:     "bob",
			want:       "bob-code-review-template",
		},
		{
			name:       "author with special characters",
			promptName: "Test Prompt",
			author:     "john.doe",
			want:       "johndoe-test-prompt",
		},
		{
			name:       "uppercase to lowercase",
			promptName: "TEST PROMPT",
			author:     "ALICE",
			want:       "alice-test-prompt",
		},
		{
			name:       "leading and trailing spaces",
			promptName: "  Test Prompt  ",
			author:     "  john  ",
			want:       "john-test-prompt",
		},
		{
			name:       "consecutive special characters",
			promptName: "Test!!!Prompt",
			author:     "john",
			want:       "john-test-prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.promptName, tt.author)
			if got != tt.want {
				t.Errorf("GenerateSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid ID with letters",
			id:      "testprompt",
			wantErr: false,
		},
		{
			name:    "valid ID with numbers",
			id:      "test123",
			wantErr: false,
		},
		{
			name:    "valid ID with hyphens",
			id:      "test-prompt-123",
			wantErr: false,
		},
		{
			name:    "valid ID with underscores",
			id:      "test_prompt_123",
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "ID with spaces",
			id:      "test prompt",
			wantErr: true,
		},
		{
			name:    "ID with special characters",
			id:      "test@prompt",
			wantErr: true,
		},
		{
			name:    "ID too short",
			id:      "ab",
			wantErr: true,
		},
		{
			name:    "ID too long",
			id:      "a" + string(make([]byte, 100)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}