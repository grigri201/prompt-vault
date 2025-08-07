package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/grigri/pv/internal/errors"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/tui"
	"github.com/spf13/cobra"
)

// ShareCmd represents the share command
type ShareCmd struct {
	promptService service.PromptService
	tui           tui.TUIInterface
}

// NewShareCommand creates a new share command
func NewShareCommand(promptService service.PromptService, tui tui.TUIInterface) *cobra.Command {
	shareCmd := &ShareCmd{
		promptService: promptService,
		tui:           tui,
	}

	cmd := &cobra.Command{
		Use:   "share [keyword|gist_url]",
		Short: "分享私有提示词为公开 Gist",
		Long: `分享命令允许将私有 GitHub Gists 转换为公开 Gists，让提示词可以与他人分享。

支持三种使用模式:

1. 交互式模式（无参数）:
   pv share
   显示所有私有提示词列表，可通过 TUI 界面选择要分享的提示词

2. 关键字筛选模式:
   pv share "keyword"
   根据关键字筛选私有提示词并显示选择界面

3. 直接 URL 模式:
   pv share https://gist.github.com/user/gist_id
   直接分享指定 URL 的私有提示词`,
		RunE: shareCmd.run,
	}

	return cmd
}

// run 执行 share 命令的主要逻辑
func (s *ShareCmd) run(cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 0:
		return s.handleInteractiveMode()
	case 1:
		arg := strings.TrimSpace(args[0])
		if s.isGistURL(arg) {
			return s.handleDirectMode(arg)
		}
		return s.handleFilterMode(arg)
	default:
		return fmt.Errorf("share 命令只接受 0 或 1 个参数")
	}
}

// handleInteractiveMode 处理交互式模式
func (s *ShareCmd) handleInteractiveMode() error {
	// 获取所有私有提示词
	privatePrompts, err := s.promptService.ListPrivatePrompts()
	if err != nil {
		return fmt.Errorf("获取私有提示词列表失败: %w", err)
	}

	if len(privatePrompts) == 0 {
		fmt.Println("没有找到私有提示词")
		return nil
	}

	// 显示 TUI 选择界面
	selectedPrompt, err := s.tui.ShowPromptList(privatePrompts)
	if err != nil {
		if err.Error() == "用户取消操作" {
			fmt.Println("取消分享操作")
			return nil
		}
		return fmt.Errorf("显示选择界面失败: %w", err)
	}

	return s.executeShare(&selectedPrompt)
}

// handleFilterMode 处理关键字筛选模式
func (s *ShareCmd) handleFilterMode(keyword string) error {
	// 根据关键字筛选私有提示词
	filteredPrompts, err := s.promptService.FilterPrivatePrompts(keyword)
	if err != nil {
		return fmt.Errorf("筛选私有提示词失败: %w", err)
	}

	if len(filteredPrompts) == 0 {
		fmt.Printf("没有找到匹配关键字 '%s' 的私有提示词\n", keyword)
		return nil
	}

	// 显示筛选结果的选择界面
	selectedPrompt, err := s.tui.ShowPromptList(filteredPrompts)
	if err != nil {
		if err.Error() == "用户取消操作" {
			fmt.Println("取消分享操作")
			return nil
		}
		return fmt.Errorf("显示选择界面失败: %w", err)
	}

	return s.executeShare(&selectedPrompt)
}

// handleDirectMode 处理直接 URL 模式
func (s *ShareCmd) handleDirectMode(gistURL string) error {
	// 验证 URL 格式
	if !s.isValidGistURL(gistURL) {
		return errors.ErrInvalidGistURL
	}

	// 获取指定 URL 的提示词
	prompt, err := s.promptService.GetPromptByURL(gistURL)
	if err != nil {
		return fmt.Errorf("获取提示词失败: %w", err)
	}

	// 验证是否为私有 gist
	gistInfo, err := s.promptService.ValidateGistAccess(gistURL)
	if err != nil {
		return fmt.Errorf("验证 gist 访问权限失败: %w", err)
	}

	if gistInfo.IsPublic {
		fmt.Println("该 Gist 已经是公开的，不需要分享")
		return nil
	}

	if !gistInfo.HasAccess {
		return errors.ErrGistAccessDenied
	}

	// 显示确认界面
	confirmed, err := s.tui.ShowConfirm(*prompt)
	if err != nil {
		return fmt.Errorf("显示确认界面失败: %w", err)
	}

	if !confirmed {
		fmt.Println("取消分享操作")
		return nil
	}

	return s.executeShare(prompt)
}

// executeShare 执行实际的分享操作
func (s *ShareCmd) executeShare(prompt *model.Prompt) error {
	fmt.Printf("正在分享提示词 '%s'...\n", prompt.Name)
	fmt.Println(prompt)
	sharedPrompt, err := s.promptService.SharePrompt(prompt)
	if err != nil {
		return fmt.Errorf("分享提示词失败: %w", err)
	}

	fmt.Printf("✅ 分享成功!\n")
	fmt.Printf("公开 Gist URL: %s\n", sharedPrompt.GistURL)

	if sharedPrompt.Parent != nil {
		fmt.Printf("原始私有 Gist: %s\n", *sharedPrompt.Parent)
	}

	return nil
}

// isGistURL 判断字符串是否为 Gist URL
func (s *ShareCmd) isGistURL(str string) bool {
	return s.looksLikeURL(str) && s.isValidGistURL(str)
}

// looksLikeURL 判断字符串是否看起来像 URL
func (s *ShareCmd) looksLikeURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// isValidGistURL 验证是否为有效的 GitHub Gist URL
func (s *ShareCmd) isValidGistURL(gistURL string) bool {
	parsedURL, err := url.Parse(gistURL)
	if err != nil {
		return false
	}

	// 检查是否为 GitHub Gist URL
	if parsedURL.Host != "gist.github.com" {
		return false
	}

	// 检查路径格式: /username/gist_id
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return false
	}

	// 验证 gist ID 格式（32位十六进制字符）
	gistID := pathParts[len(pathParts)-1]
	if len(gistID) != 32 {
		return false
	}

	for _, char := range gistID {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}
