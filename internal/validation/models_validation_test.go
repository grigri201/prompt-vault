package validation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestPromptMeta_DesignCompliance(t *testing.T) {
	// 验证PromptMeta结构符合设计文档
	meta := models.PromptMeta{
		Name:        "Test Prompt",
		Author:      "testuser",
		Tags:        []string{"test"},
		Version:     "1.0",
		Description: "Test description",
		Parent:      "parent-id",
		ID:          "test-id",
	}

	// 测试YAML序列化（只应包含YAML标签）
	yamlData, err := yaml.Marshal(meta)
	require.NoError(t, err)

	yamlStr := string(yamlData)
	assert.Contains(t, yamlStr, "name: Test Prompt")
	assert.Contains(t, yamlStr, "author: testuser")
	assert.Contains(t, yamlStr, "tags:")
	assert.NotContains(t, yamlStr, "category") // 确保category字段已移除

	// 测试反序列化
	var unmarshaled models.PromptMeta
	err = yaml.Unmarshal(yamlData, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, meta.Name, unmarshaled.Name)
	assert.Equal(t, meta.Author, unmarshaled.Author)
}

func TestIndexEntry_DesignCompliance(t *testing.T) {
	// 验证IndexEntry结构符合设计文档
	entry := models.IndexEntry{
		GistID:      "test123",
		GistURL:     "https://gist.github.com/user/test123",
		Name:        "Test Entry",
		Author:      "testuser",
		Tags:        []string{"test"},
		Version:     "1.0",
		Description: "Test description",
		Parent:      "parent-id",
		ID:          "test-id",
		UpdatedAt:   time.Now(),
	}

	// 测试JSON序列化
	jsonData, err := json.Marshal(entry)
	require.NoError(t, err)

	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"gist_id":"test123"`)
	assert.Contains(t, jsonStr, `"name":"Test Entry"`)
	assert.NotContains(t, jsonStr, `"category"`) // 确保category字段已移除

	// 验证omitempty标签工作
	emptyEntry := models.IndexEntry{
		GistID:    "test",
		GistURL:   "https://test.com",
		Name:      "Test",
		Author:    "user",
		Tags:      []string{"test"},
		UpdatedAt: time.Now(),
	}

	emptyJSON, err := json.Marshal(emptyEntry)
	require.NoError(t, err)

	// 空的可选字段不应出现在JSON中
	emptyStr := string(emptyJSON)
	assert.NotContains(t, emptyStr, `"version"`)
	assert.NotContains(t, emptyStr, `"description"`)
	assert.NotContains(t, emptyStr, `"parent"`)
}

func TestPromptMeta_Validation(t *testing.T) {
	tests := []struct {
		name      string
		meta      models.PromptMeta
		wantError bool
	}{
		{
			name: "valid meta",
			meta: models.PromptMeta{
				Name:   "Test",
				Author: "testuser",
				Tags:   []string{"test"},
			},
			wantError: false,
		},
		{
			name: "missing name",
			meta: models.PromptMeta{
				Author: "testuser",
				Tags:   []string{"test"},
			},
			wantError: true,
		},
		{
			name: "missing author",
			meta: models.PromptMeta{
				Name: "Test",
				Tags: []string{"test"},
			},
			wantError: true,
		},
		{
			name: "missing tags",
			meta: models.PromptMeta{
				Name:   "Test",
				Author: "testuser",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIndex_Operations(t *testing.T) {
	index := &models.Index{}
	entry := models.IndexEntry{
		GistID:    "test123",
		Name:      "Test Entry",
		Author:    "testuser",
		Tags:      []string{"test"},
		UpdatedAt: time.Now(),
	}

	// Test AddImportedEntry
	index.AddImportedEntry(entry)
	assert.Len(t, index.ImportedEntries, 1)
	assert.Equal(t, "test123", index.ImportedEntries[0].GistID)

	// Test FindImportedEntry
	found, exists := index.FindImportedEntry("test123")
	assert.True(t, exists)
	assert.Equal(t, "Test Entry", found.Name)

	// Test UpdateImportedEntry
	entry.Name = "Updated Entry"
	updated := index.UpdateImportedEntry(entry)
	assert.True(t, updated)
	assert.Equal(t, "Updated Entry", index.ImportedEntries[0].Name)

	// Test not found
	_, exists = index.FindImportedEntry("nonexistent")
	assert.False(t, exists)
}
