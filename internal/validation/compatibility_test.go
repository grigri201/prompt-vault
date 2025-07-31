package validation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestBackwardCompatibility(t *testing.T) {
	// 测试旧格式数据的兼容性
	t.Run("Legacy YAML with Category", func(t *testing.T) {
		legacyYAML := `---
name: "Legacy Prompt"
author: "testuser"
category: "development"
tags: ["test"]
---
Legacy content`

		// 应该能够解析包含category的旧格式
		var legacy models.LegacyPromptMeta
		err := yaml.Unmarshal([]byte(legacyYAML), &legacy)
		require.NoError(t, err)

		assert.Equal(t, "Legacy Prompt", legacy.Name)
		assert.Equal(t, "development", legacy.Category)

		// 测试迁移功能
		migrated := models.MigrateLegacyPrompt(legacy)
		assert.Equal(t, "Legacy Prompt", migrated.Name)
		assert.Contains(t, migrated.Tags, "development") // category转为tag
		assert.Contains(t, migrated.Tags, "test")        // 原有tag保留
	})

	t.Run("Legacy YAML without Category", func(t *testing.T) {
		legacyYAML := `---
name: "No Category Prompt"
author: "testuser"
tags: ["test", "example"]
---
Content without category`

		var legacy models.LegacyPromptMeta
		err := yaml.Unmarshal([]byte(legacyYAML), &legacy)
		require.NoError(t, err)

		assert.Equal(t, "No Category Prompt", legacy.Name)
		assert.Equal(t, "", legacy.Category) // 空category

		// 测试迁移功能
		migrated := models.MigrateLegacyPrompt(legacy)
		assert.Equal(t, "No Category Prompt", migrated.Name)
		assert.Equal(t, []string{"test", "example"}, migrated.Tags) // tags保持不变
	})

	t.Run("Legacy Index with Category", func(t *testing.T) {
		legacyIndexJSON := `{
  "username": "testuser",
  "entries": [
    {
      "gist_id": "test123",
      "name": "Legacy Entry",
      "author": "testuser",
      "category": "development",
      "tags": ["test"],
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "updated_at": "2024-01-01T00:00:00Z"
}`

		// 应该能够解析包含category的旧索引
		var legacyIndex struct {
			Username string `json:"username"`
			Entries  []struct {
				GistID   string   `json:"gist_id"`
				Name     string   `json:"name"`
				Author   string   `json:"author"`
				Category string   `json:"category"`
				Tags     []string `json:"tags"`
			} `json:"entries"`
		}

		err := json.Unmarshal([]byte(legacyIndexJSON), &legacyIndex)
		require.NoError(t, err)

		assert.Equal(t, "development", legacyIndex.Entries[0].Category)

		// 可以转换为新格式
		newEntry := models.IndexEntry{
			GistID: legacyIndex.Entries[0].GistID,
			Name:   legacyIndex.Entries[0].Name,
			Author: legacyIndex.Entries[0].Author,
			Tags:   append(legacyIndex.Entries[0].Tags, legacyIndex.Entries[0].Category),
		}

		assert.Contains(t, newEntry.Tags, "development")
		assert.Contains(t, newEntry.Tags, "test")
	})

	t.Run("Category to Tags Migration Edge Cases", func(t *testing.T) {
		// 测试category已存在于tags中的情况
		legacy := models.LegacyPromptMeta{
			Name:     "Edge Case",
			Author:   "testuser",
			Category: "test",
			Tags:     []string{"test", "example"}, // category已在tags中
		}

		migrated := models.MigrateLegacyPrompt(legacy)
		// category不应该重复添加
		tagCount := 0
		for _, tag := range migrated.Tags {
			if tag == "test" {
				tagCount++
			}
		}
		assert.Equal(t, 1, tagCount, "Category should not be duplicated in tags")
	})
}

func TestDataFormatEvolution(t *testing.T) {
	t.Run("Old to New Field Mapping", func(t *testing.T) {
		// 测试字段映射的完整性
		legacy := models.LegacyPromptMeta{
			Name:        "Complete Test",
			Author:      "testuser",
			Category:    "ai-prompts",
			Tags:        []string{"chatgpt", "assistant"},
			Version:     "2.1",
			Description: "A comprehensive test prompt",
			Parent:      "parent-gist-123",
			ID:          "unique-id-456",
		}

		migrated := models.MigrateLegacyPrompt(legacy)

		// 验证所有字段正确迁移
		assert.Equal(t, legacy.Name, migrated.Name)
		assert.Equal(t, legacy.Author, migrated.Author)
		assert.Equal(t, legacy.Version, migrated.Version)
		assert.Equal(t, legacy.Description, migrated.Description)
		assert.Equal(t, legacy.Parent, migrated.Parent)
		assert.Equal(t, legacy.ID, migrated.ID)

		// 验证tags包含原tags和category
		assert.Contains(t, migrated.Tags, "chatgpt")
		assert.Contains(t, migrated.Tags, "assistant")
		assert.Contains(t, migrated.Tags, "ai-prompts")
		assert.Len(t, migrated.Tags, 3) // 2个原tag + 1个category
	})

	t.Run("Empty Category Handling", func(t *testing.T) {
		legacy := models.LegacyPromptMeta{
			Name:   "No Category",
			Author: "testuser",
			Tags:   []string{"tag1", "tag2"},
			// Category is empty
		}

		migrated := models.MigrateLegacyPrompt(legacy)
		assert.Equal(t, []string{"tag1", "tag2"}, migrated.Tags)
	})
}

func TestMigrationValidation(t *testing.T) {
	t.Run("Migrated Data Validation", func(t *testing.T) {
		legacy := models.LegacyPromptMeta{
			Name:     "Valid Legacy",
			Author:   "testuser",
			Category: "development",
			Tags:     []string{"code"},
		}

		migrated := models.MigrateLegacyPrompt(legacy)

		// 迁移后的数据应该通过验证
		err := migrated.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid Legacy Data", func(t *testing.T) {
		legacy := models.LegacyPromptMeta{
			// Missing required fields
			Name: "Missing Author",
			Tags: []string{"test"},
		}

		migrated := models.MigrateLegacyPrompt(legacy)

		// 迁移后的数据应该失败验证（因为缺少author）
		err := migrated.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "author is required")
	})
}
