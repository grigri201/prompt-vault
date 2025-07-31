package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grigri201/prompt-vault/internal/models"
)

func TestCompleteUserWorkflow(t *testing.T) {
	// 创建隔离的测试环境
	testDir := t.TempDir()
	setupTestEnvironment(t, testDir)

	t.Run("Complete Workflow", func(t *testing.T) {
		// 1. 验证基本结构创建
		cacheDir := filepath.Join(testDir, ".cache", "prompt-vault")
		configDir := filepath.Join(testDir, ".config", "prompt-vault")

		err := os.MkdirAll(cacheDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// 2. 测试索引创建和管理
		index := &models.Index{
			Username:        "testuser",
			ImportedEntries: []models.IndexEntry{},
		}

		// 3. 测试添加提示条目
		entry := models.IndexEntry{
			GistID:  "test123",
			GistURL: "https://gist.github.com/user/test123",
			Name:    "Test Entry",
			Author:  "testuser",
			Tags:    []string{"test"},
		}

		index.AddImportedEntry(entry)
		assert.Len(t, index.ImportedEntries, 1)

		// 4. 验证搜索功能（基础逻辑）
		found, exists := index.FindImportedEntry("test123")
		assert.True(t, exists)
		assert.Equal(t, "Test Entry", found.Name)

		// 5. 测试更新功能
		entry.Name = "Updated Entry"
		updated := index.UpdateImportedEntry(entry)
		assert.True(t, updated)

		// 6. 验证更新后的状态
		found, exists = index.FindImportedEntry("test123")
		assert.True(t, exists)
		assert.Equal(t, "Updated Entry", found.Name)

		// 7. 测试多条目管理
		entry2 := models.IndexEntry{
			GistID: "test456",
			Name:   "Second Entry",
			Author: "testuser",
			Tags:   []string{"test", "second"},
		}
		index.AddImportedEntry(entry2)
		assert.Len(t, index.ImportedEntries, 2)

		// 8. 验证独立条目查找
		found2, exists2 := index.FindImportedEntry("test456")
		assert.True(t, exists2)
		assert.Equal(t, "Second Entry", found2.Name)
	})
}

func TestWorkflowErrorHandling(t *testing.T) {
	testDir := t.TempDir()
	setupTestEnvironment(t, testDir)

	t.Run("Invalid Data Handling", func(t *testing.T) {
		// 测试无效数据的处理
		meta := models.PromptMeta{
			Name: "", // 空名称应该失败
		}

		err := meta.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("Missing Fields", func(t *testing.T) {
		// 测试缺失字段
		meta := models.PromptMeta{
			Name:   "Test",
			Author: "testuser",
			// 缺少 Tags
		}

		err := meta.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one tag is required")
	})
}

func TestDataConsistency(t *testing.T) {
	t.Run("Prompt to IndexEntry Conversion", func(t *testing.T) {
		prompt := models.Prompt{
			PromptMeta: models.PromptMeta{
				Name:        "Test Prompt",
				Author:      "testuser",
				Tags:        []string{"test"},
				Version:     "1.0",
				Description: "Test description",
			},
			GistID:  "test123",
			GistURL: "https://gist.github.com/user/test123",
			Content: "Test content",
		}

		entry := prompt.ToIndexEntry()
		assert.Equal(t, prompt.Name, entry.Name)
		assert.Equal(t, prompt.Author, entry.Author)
		assert.Equal(t, prompt.GistID, entry.GistID)
		assert.Equal(t, prompt.Tags, entry.Tags)
		// Content should not be in IndexEntry
		assert.Empty(t, entry.ID) // ID should be empty in this case
	})
}

func setupTestEnvironment(t *testing.T, testDir string) {
	os.Setenv("HOME", testDir)
	os.Setenv("PV_CACHE_DIR", filepath.Join(testDir, ".cache", "prompt-vault"))
	os.Setenv("PV_CONFIG_DIR", filepath.Join(testDir, ".config", "prompt-vault"))
}
