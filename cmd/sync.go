package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	apperrors "github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/service"
)

type SyncCmd *cobra.Command

type sync struct {
	promptService service.PromptService
}

// SyncStats holds sync operation statistics
type SyncStats struct {
	Total     int
	Success   int
	Failed    int
	Skipped   int
	Errors    []string
}

func (sc *sync) execute(cmd *cobra.Command, args []string) {
	// Initialize sync statistics
	stats := &SyncStats{
		Errors: make([]string, 0),
	}

	fmt.Println("🔄 开始同步提示词...")
	fmt.Println()

	// Get verbose flag for detailed output
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Step 1: Sync raw index.json first
	fmt.Println("📥 正在同步索引文件...")
	
	if err := sc.promptService.Sync(); err != nil {
		// Handle different types of errors with user-friendly messages
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			switch appErr.Type {
			case apperrors.ErrAuth:
				fmt.Printf("❌ 认证错误: %s\n", appErr.Message)
				fmt.Println()
				fmt.Println("请运行 'pv auth login' 重新登录 GitHub。")
				return
			case apperrors.ErrNetwork:
				fmt.Printf("❌ 网络错误: %s\n", appErr.Message)
				fmt.Println()
				fmt.Println("请检查网络连接后重试。")
				return
			default:
				fmt.Printf("❌ 索引同步失败: %s\n", appErr.Message)
				fmt.Println()
				fmt.Println("请检查网络连接和 GitHub 认证状态。")
				return
			}
		}

		// For other unexpected errors, show the original error and exit
		fmt.Printf("❌ 索引同步失败: %v\n", err)
		fmt.Println()
		fmt.Println("请检查网络连接和 GitHub 认证状态。")
		return
	}
	
	fmt.Println("  ✅ 索引同步成功")
	fmt.Println()

	// Step 2: Get list of all prompts to sync
	prompts, err := sc.promptService.ListPrompts()
	if err != nil {
		// Handle different types of errors with user-friendly messages
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			switch appErr.Type {
			case apperrors.ErrAuth:
				fmt.Printf("❌ 认证错误: %s\n", appErr.Message)
				fmt.Println()
				fmt.Println("请运行 'pv auth login' 重新登录 GitHub。")
				return
			case apperrors.ErrNetwork:
				fmt.Printf("❌ 网络错误: %s\n", appErr.Message)
				fmt.Println()
				fmt.Println("请检查网络连接后重试。")
				return
			default:
				fmt.Printf("❌ 错误: %s\n", appErr.Message)
				return
			}
		}

		// For other unexpected errors, show the original error and exit
		log.Fatalf("获取提示词列表时发生意外错误: %v", err)
	}

	stats.Total = len(prompts)
	
	if stats.Total == 0 {
		fmt.Println("📝 没有找到需要同步的提示词。")
		fmt.Println()
		fmt.Println("要添加提示词:")
		fmt.Println("  • 在 GitHub Gists 中直接创建提示词")
		fmt.Println("  • 使用 'pv add <file>' 添加新提示词")
		return
	}

	fmt.Printf("📋 发现 %d 个提示词需要同步\n", stats.Total)
	fmt.Println()

	// Step 3: Serial download of prompt content with progress display
	for i, prompt := range prompts {
		// Display progress in the exact format specified: "正在下载 X/Y"
		fmt.Printf("⬇️  正在下载 %d/%d: %s", i+1, stats.Total, prompt.Name)
		if verbose {
			fmt.Printf(" (%s)", prompt.ID)
		}
		fmt.Println()

		// Try to get content for this prompt
		// This call will go through the CachedStore which will:
		// 1. Try to fetch from remote (GitHub Gist)
		// 2. Cache the content locally on success
		// 3. Return the content
		_, err := sc.promptService.GetPromptContent(&prompt)
		if err != nil {
			stats.Failed++
			errorMsg := fmt.Sprintf("下载 '%s' 失败: %v", prompt.Name, err)
			stats.Errors = append(stats.Errors, errorMsg)
			
			if verbose {
				fmt.Printf("  ❌ %s\n", errorMsg)
			} else {
				fmt.Printf("  ❌ 失败\n")
			}
			// Continue processing other prompts even if this one fails
			continue
		}
		
		// Success: content has been downloaded and cached
		stats.Success++
		if verbose {
			fmt.Printf("  ✅ 成功缓存\n")
		}
	}

	fmt.Println()
	sc.displaySyncResults(stats, verbose)
}

func (sc *sync) displaySyncResults(stats *SyncStats, verbose bool) {
	fmt.Println("📊 同步完成统计:")
	fmt.Println()
	fmt.Printf("  总计: %d\n", stats.Total)
	fmt.Printf("  成功: %d\n", stats.Success)
	fmt.Printf("  失败: %d\n", stats.Failed)
	fmt.Printf("  跳过: %d\n", stats.Skipped)
	fmt.Println()

	if stats.Success == stats.Total {
		fmt.Println("🎉 所有提示词同步并缓存成功!")
	} else if stats.Success > 0 {
		fmt.Printf("✅ %d 个提示词同步并缓存成功", stats.Success)
		if stats.Failed > 0 {
			fmt.Printf("，%d 个失败", stats.Failed)
		}
		fmt.Println()
	} else {
		fmt.Println("❌ 同步失败，没有提示词被成功下载和缓存")
	}

	// Show errors if verbose mode is enabled or if there are critical failures
	if verbose && len(stats.Errors) > 0 {
		fmt.Println()
		fmt.Println("详细错误信息:")
		for _, errMsg := range stats.Errors {
			fmt.Printf("  • %s\n", errMsg)
		}
	} else if len(stats.Errors) > 0 {
		fmt.Println()
		fmt.Println("使用 --verbose 查看详细错误信息")
	}

	if stats.Success > 0 {
		fmt.Println()
		fmt.Println("提示词已缓存到本地，现在可以离线使用。")
		fmt.Println("运行 'pv list' 查看同步的提示词。")
	}
}

func NewSyncCommand(promptService service.PromptService) SyncCmd {
	sc := &sync{promptService: promptService}
	
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "将远程 GitHub Gist 数据同步到本地缓存",
		Long: `同步命令实现完整的缓存同步流程，从 GitHub Gist 下载最新的提示词数据到本地缓存。

完整同步流程：
  • 获取远程提示词索引列表
  • 串行下载所有提示词内容到本地缓存
  • 显示 "正在下载 X/Y" 进度信息
  • 单个提示词失败时继续处理其他提示词
  • 显示最终同步统计信息（成功/失败数量）

适用场景：
  • 首次使用时建立本地缓存
  • 网络恢复后更新本地数据
  • 手动刷新缓存确保数据最新
  • 为离线使用准备数据

注意：该命令需要有效的 GitHub 认证。如果认证失败，请先运行 'pv auth login'。`,
		Example: `  # 同步所有提示词到本地缓存
  pv sync

  # 同步并显示详细信息
  pv sync --verbose`,
		Run: sc.execute,
	}

	// Add verbose flag for detailed output
	cmd.Flags().BoolP("verbose", "v", false, "显示详细的同步过程和错误信息")

	return cmd
}