package cmd

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/config"
	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/model"
)

type ListCmd *cobra.Command

type list struct {
	store       infra.Store
	configStore config.Store
	remote      bool
}

// ExportStatus 保存 prompt 的导出信息
type ExportStatus struct {
	IsExported bool
	ExportURL  string
	ExportedBy string
}

// buildExportMap 创建从 GistURL 到 ExportStatus 的查找映射
func buildExportMap(exports []model.IndexedPrompt) map[string]ExportStatus {
	exportMap := make(map[string]ExportStatus)
	for _, export := range exports {
		if export.Parent != nil && *export.Parent != "" {
			exportMap[*export.Parent] = ExportStatus{
				IsExported: true,
				ExportURL:  export.GistURL,
				ExportedBy: export.Author,
			}
		}
	}
	return exportMap
}

// formatExportInfo 格式化导出状态显示
func formatExportInfo(status ExportStatus) string {
	if !status.IsExported {
		return "[not exported]"
	}
	if status.ExportURL == "" {
		return "[export status unknown]"
	}
	return fmt.Sprintf("[✓ exported: %s]", status.ExportURL)
}

// formatPromptWithExport 创建包含导出信息的增强显示字符串
func formatPromptWithExport(prompt model.Prompt, exportMap map[string]ExportStatus) string {
	status := ExportStatus{IsExported: false}
	if s, exists := exportMap[prompt.GistURL]; exists {
		status = s
	}
	
	exportInfo := formatExportInfo(status)
	return fmt.Sprintf("  %s - author: %s : %s %s", 
		prompt.Name, prompt.Author, prompt.GistURL, exportInfo)
}

func (lc *list) execute(cmd *cobra.Command, args []string) {
	// Create appropriate store based on --remote flag
	var store infra.Store
	var cacheManager *infra.CacheManager

	if lc.remote {
		// Use remote store directly when --remote flag is set
		store = lc.store
	} else {
		// Create cached store for default behavior
		var err error
		cacheManager, err = infra.NewCacheManager()
		if err != nil {
			// If cache manager creation fails, fallback to remote store
			store = lc.store
		} else {
			// Create cached store with forceRemote=false for cache-first behavior
			store = infra.NewCachedStore(lc.store, cacheManager, lc.configStore, false)
		}
	}

	var prompts, err = store.List()
	if err != nil {
		// Handle friendly error messages for empty/missing index
		if errors.Is(err, infra.ErrNoIndex) {
			fmt.Println("📝 Welcome to Prompt Vault!")
			fmt.Println()
			fmt.Println("It looks like this is your first time using pv. Your prompt collection is empty.")
			fmt.Println()
			fmt.Println("To get started:")
			fmt.Println("  • Create prompts directly in GitHub Gists")
			fmt.Println("  • Use 'pv add <name>' to create a new prompt")
			fmt.Println("  • Run 'pv list' again to see your prompts")
			return
		}

		if errors.Is(err, infra.ErrEmptyIndex) {
			fmt.Println("📝 Your prompt collection is currently empty.")
			fmt.Println()
			fmt.Println("To add prompts:")
			fmt.Println("  • Create prompts directly in GitHub Gists")
			fmt.Println("  • Use 'pv add <name>' to create a new prompt")
			fmt.Println("  • Run 'pv list' again to see your prompts")
			return
		}

		// For other errors, show the original error
		log.Fatalf("error in get prompts: %v", err)
	}

	// Display prompts if we have any
	if len(prompts) == 0 {
		fmt.Println("📝 No prompts found in your collection.")
		return
	}

	// 获取导出数据并构建查找映射
	var exports []model.IndexedPrompt
	var exportMap map[string]ExportStatus

	if exports, err = store.GetExports(); err == nil {
		exportMap = buildExportMap(exports)
	} else {
		// 优雅降级：继续显示但不包含导出信息
		exportMap = make(map[string]ExportStatus)
	}

	fmt.Printf("📝 Found %d prompt(s):\n\n", len(prompts))
	for i := range prompts {
		var prompt = prompts[i]
		fmt.Printf("%s\n", formatPromptWithExport(prompt, exportMap))
	}

	// Display cache information when using cached data (requirement 5.5)
	if !lc.remote && cacheManager != nil {
		if cacheInfo, err := cacheManager.GetCacheInfo(); err == nil && !cacheInfo.LastUpdated.IsZero() {
			fmt.Printf("\n📋 Cache last updated: %s\n", cacheInfo.LastUpdated.Format(time.RFC3339))
		}
	}
}

func NewListCommand(store infra.Store, configStore config.Store) ListCmd {
	lc := &list{store: store, configStore: configStore}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all prompts in your collection",
		Long: `List all prompts in your collection.

By default, this command uses local cache for better performance.
Use --remote to fetch the latest data directly from GitHub Gist.`,
		Run: lc.execute,
	}

	// Add --remote flag (requirement 5.2)
	listCmd.Flags().BoolVarP(&lc.remote, "remote", "r", false, "Force fetch from remote GitHub Gist instead of using cache")

	return listCmd
}
