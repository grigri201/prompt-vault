package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/infra"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/tui"
)

type DeleteCmd = *cobra.Command

type delete struct {
	store         infra.Store
	promptService service.PromptService
}

func (dc *delete) execute(cmd *cobra.Command, args []string) {
	// Validate arguments - we accept 0-1 arguments
	if len(args) > 1 {
		fmt.Println("❌ Error: Too many arguments provided")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  pv delete                    # 交互式选择删除")
		fmt.Println("  pv delete <keyword>          # 根据关键字筛选删除")
		fmt.Println("  pv delete <gist-url>         # 直接删除指定URL的提示")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  pv delete                    # 显示所有提示供选择")
		fmt.Println("  pv delete golang             # 筛选包含 'golang' 的提示")
		fmt.Println("  pv delete https://gist.github.com/user/abc123")
		return
	}

	// Route to appropriate mode based on arguments
	switch len(args) {
	case 0:
		// 交互式模式 - 显示所有提示供用户选择
		dc.handleInteractiveMode()
	case 1:
		arg := args[0]
		// Check if it looks like a URL first
		if dc.looksLikeURL(arg) {
			// It's a URL-like string, check if it's a valid Gist URL
			if dc.isGistURL(arg) {
				// Valid Gist URL - direct deletion mode
				dc.handleDirectMode(arg)
			} else {
				// Invalid Gist URL - show error
				dc.handleInvalidURL(arg)
			}
		} else {
			// Not URL-like - keyword filtering mode
			dc.handleFilterMode(arg)
		}
	}
}

// isGistURL validates if the provided string is a valid GitHub Gist URL
func (dc *delete) isGistURL(input string) bool {
	// First check if it looks like a URL at all
	if !dc.looksLikeURL(input) {
		return false
	}
	
	// Check if it contains gist.github.com
	if !contains(input, "gist.github.com") {
		return false
	}
	
	// Additional validation: check if it has a gist ID pattern
	// GitHub gist IDs are typically 20 or 32 character hex strings
	return containsGistID(input)
}

// looksLikeURL checks if the input string looks like a URL
func (dc *delete) looksLikeURL(input string) bool {
	if len(input) < 7 {
		return false
	}
	return (len(input) >= 8 && input[:8] == "https://") || 
		   (len(input) >= 7 && input[:7] == "http://")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr)))
}

// containsSubstring is a simple substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// containsGistID checks if the URL contains a valid gist ID pattern
func containsGistID(url string) bool {
	// Look for 20 or 32 character hex strings in the URL
	parts := splitString(url, "/")
	for _, part := range parts {
		if len(part) == 20 || len(part) == 32 {
			if isHexString(part) {
				return true
			}
		}
	}
	return false
}

// splitString splits a string by delimiter
func splitString(s, delimiter string) []string {
	if len(s) == 0 {
		return []string{}
	}
	
	var result []string
	start := 0
	
	for i := 0; i <= len(s)-len(delimiter); i++ {
		if s[i:i+len(delimiter)] == delimiter {
			result = append(result, s[start:i])
			start = i + len(delimiter)
		}
	}
	result = append(result, s[start:])
	return result
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// extractGistID extracts the Gist ID from a GitHub Gist URL
func (dc *delete) extractGistID(url string) string {
	// Split the URL by "/" to find the Gist ID
	parts := splitString(url, "/")
	
	// Look for 20 or 32 character hex strings in the URL parts
	for _, part := range parts {
		if (len(part) == 20 || len(part) == 32) && isHexString(part) {
			return part
		}
	}
	
	return ""
}

// handleInteractiveMode handles the interactive deletion mode (no arguments)
func (dc *delete) handleInteractiveMode() {
	fmt.Println("🔄 Interactive mode - loading all prompts...")
	
	// Step 1: Call promptService.ListForDeletion() to get all prompts
	prompts, err := dc.promptService.ListForDeletion()
	if err != nil {
		fmt.Printf("❌ Error loading prompts: %v\n", err)
		return
	}
	
	// Step 2: Handle empty list situation with friendly message
	if len(prompts) == 0 {
		fmt.Println("📭 No prompts found in your vault.")
		fmt.Println()
		fmt.Println("To add prompts to your vault, use:")
		fmt.Println("  pv add <path-to-yaml-file>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  pv add my-prompt.yaml")
		return
	}
	
	fmt.Printf("📋 Found %d prompt(s) in your vault.\n", len(prompts))
	fmt.Println()
	
	// Step 3: Create TUI factory and launch prompt selection interface
	tuiFactory := tui.NewBubbleTeaTUI()
	
	// Step 4: Get user selected prompt from TUI
	selectedPrompt, err := tuiFactory.ShowPromptList(prompts)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 删除操作已取消")
			return
		}
		// Handle other TUI errors
		fmt.Printf("❌ Error displaying prompt list: %v\n", err)
		return
	}
	
	// Step 5: Integrate confirmation interface showing prompt details
	fmt.Printf("📝 Selected prompt: %s (by %s)\n", selectedPrompt.Name, selectedPrompt.Author)
	fmt.Println()
	
	confirmed, err := tuiFactory.ShowConfirm(selectedPrompt)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 删除操作已取消")
			return
		}
		// Handle other confirmation errors
		fmt.Printf("❌ Error displaying confirmation dialog: %v\n", err)
		return
	}
	
	// Check if user cancelled the confirmation
	if !confirmed {
		fmt.Println("🚫 删除操作已取消")
		return
	}
	
	// Step 6: Execute deletion operation and display results
	fmt.Printf("🗑️  Deleting prompt '%s'...\n", selectedPrompt.Name)
	
	// Use the Gist URL to identify the prompt for deletion
	err = dc.promptService.DeleteByURL(selectedPrompt.GistURL)
	if err != nil {
		fmt.Printf("❌ Failed to delete prompt: %v\n", err)
		fmt.Println()
		fmt.Println("This could be due to:")
		fmt.Println("  • Network connectivity issues")
		fmt.Println("  • GitHub authentication problems")
		fmt.Println("  • The prompt may have been already deleted")
		fmt.Println()
		fmt.Println("Please check your connection and authentication, then try again.")
		return
	}
	
	// Step 7: Display success message
	fmt.Printf("✅ Successfully deleted prompt '%s'\n", selectedPrompt.Name)
	fmt.Printf("   Author: %s\n", selectedPrompt.Author)
	fmt.Printf("   Gist URL: %s\n", selectedPrompt.GistURL)
	fmt.Println()
	fmt.Println("The prompt has been removed from both GitHub Gist and your local index.")
}

// handleFilterMode handles the keyword filtering deletion mode
func (dc *delete) handleFilterMode(keyword string) {
	fmt.Printf("🔄 Filter mode - searching for prompts matching '%s'...\n", keyword)
	
	// Step 1: Call promptService.FilterForDeletion(keyword) to filter prompts
	filteredPrompts, err := dc.promptService.FilterForDeletion(keyword)
	if err != nil {
		fmt.Printf("❌ Error filtering prompts: %v\n", err)
		fmt.Println()
		fmt.Println("This could be due to:")
		fmt.Println("  • Network connectivity issues")
		fmt.Println("  • Data access problems")
		fmt.Println("  • Storage service unavailable")
		fmt.Println()
		fmt.Println("Please check your connection and try again.")
		return
	}
	
	// Step 2: Handle no matching results situation with appropriate messages
	if len(filteredPrompts) == 0 {
		fmt.Printf("📭 No prompts found matching '%s'\n", keyword)
		fmt.Println()
		fmt.Println("Tips for better search results:")
		fmt.Println("  • Try a shorter or more general keyword")
		fmt.Println("  • Check your spelling")
		fmt.Println("  • Try searching by author name")
		fmt.Println("  • Use 'pv list' to see all available prompts")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  pv delete golang       # Search for prompts containing 'golang'")
		fmt.Println("  pv delete john          # Search for prompts by author 'john'")
		fmt.Println("  pv delete review        # Search for prompts about 'review'")
		return
	}
	
	// Step 3: Display filtering results statistics and information
	fmt.Printf("🎯 Found %d prompt(s) matching '%s':\n", len(filteredPrompts), keyword)
	fmt.Println()
	
	// Show a preview of the filtered results with keyword highlighting info
	fmt.Println("Matching prompts:")
	for i, prompt := range filteredPrompts {
		// Add basic highlighting indication (we'll display actual highlighting in TUI)
		fmt.Printf("  %d. %s (by %s)\n", i+1, prompt.Name, prompt.Author)
	}
	fmt.Println()
	fmt.Printf("✨ Keyword '%s' will be highlighted in the selection interface.\n", keyword)
	fmt.Println()
	
	// Step 4: Launch TUI selection interface for filtered results
	tuiFactory := tui.NewBubbleTeaTUI()
	
	// Use the specialized filtered prompt list method
	selectedPrompt, err := tuiFactory.ShowPromptListFiltered(filteredPrompts, keyword)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 删除操作已取消")
			return
		}
		// Handle other TUI errors
		fmt.Printf("❌ Error displaying filtered prompt list: %v\n", err)
		return
	}
	
	// Step 5: Integrate subsequent confirmation and deletion workflow
	fmt.Printf("📝 Selected prompt: %s (by %s)\n", selectedPrompt.Name, selectedPrompt.Author)
	fmt.Printf("    Matches keyword: '%s'\n", keyword)
	fmt.Println()
	
	// Show confirmation interface with detailed information
	confirmed, err := tuiFactory.ShowConfirm(selectedPrompt)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 删除操作已取消")
			return
		}
		// Handle other confirmation errors
		fmt.Printf("❌ Error displaying confirmation dialog: %v\n", err)
		return
	}
	
	// Check if user cancelled the confirmation
	if !confirmed {
		fmt.Println("🚫 删除操作已取消")
		return
	}
	
	// Step 6: Execute deletion operation and display comprehensive results
	fmt.Printf("🗑️  Deleting prompt '%s' (matched by keyword '%s')...\n", selectedPrompt.Name, keyword)
	
	// Use the Gist URL to identify the prompt for deletion
	err = dc.promptService.DeleteByURL(selectedPrompt.GistURL)
	if err != nil {
		fmt.Printf("❌ Failed to delete prompt: %v\n", err)
		fmt.Println()
		fmt.Println("This could be due to:")
		fmt.Println("  • Network connectivity issues")
		fmt.Println("  • GitHub authentication problems")
		fmt.Println("  • The prompt may have been already deleted")
		fmt.Println("  • Insufficient permissions to delete the GitHub Gist")
		fmt.Println()
		fmt.Println("Please check your connection and authentication, then try again.")
		return
	}
	
	// Step 7: Display success message with filtering context
	fmt.Printf("✅ Successfully deleted prompt '%s'\n", selectedPrompt.Name)
	fmt.Printf("   Author: %s\n", selectedPrompt.Author)
	fmt.Printf("   Matched keyword: '%s'\n", keyword)
	fmt.Printf("   Gist URL: %s\n", selectedPrompt.GistURL)
	fmt.Println()
	fmt.Println("The prompt has been removed from both GitHub Gist and your local index.")
	fmt.Printf("You can search for other prompts using: pv delete \"%s\"\n", keyword)
}

// handleDirectMode handles the direct URL deletion mode
func (dc *delete) handleDirectMode(gistURL string) {
	fmt.Printf("🔄 Direct mode - processing URL: %s\n", gistURL)
	
	// Step 1: Validate and parse the GitHub Gist URL
	if !dc.isGistURL(gistURL) {
		fmt.Println("❌ Error: Invalid GitHub Gist URL format")
		fmt.Println()
		fmt.Println("Valid URL formats:")
		fmt.Println("  https://gist.github.com/username/gist-id")
		fmt.Println("  https://gist.github.com/gist-id")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  pv delete https://gist.github.com/user/1234567890abcdef1234567890abcdef")
		return
	}
	
	// Extract Gist ID from URL for validation
	gistID := dc.extractGistID(gistURL)
	if gistID == "" {
		fmt.Println("❌ Error: Unable to extract Gist ID from URL")
		fmt.Println()
		fmt.Println("Please ensure the URL contains a valid Gist ID (20 or 32 character hex string)")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  https://gist.github.com/user/1234567890abcdef1234567890abcdef")
		return
	}
	
	fmt.Printf("🔍 Searching for prompt with Gist ID: %s\n", gistID)
	fmt.Println()
	
	// Step 2: Call promptService.DeleteByURL(gistURL) to find the prompt
	// First, we need to find the prompt to show details before deletion
	// We'll use the store to find the prompt by Gist ID
	prompts, err := dc.store.Get(gistID)
	if err != nil {
		fmt.Printf("❌ Error searching for prompt: %v\n", err)
		fmt.Println()
		fmt.Println("This could be due to:")
		fmt.Println("  • Network connectivity issues")
		fmt.Println("  • Data access problems")
		fmt.Println("  • Storage service unavailable")
		fmt.Println()
		fmt.Println("Please check your connection and try again.")
		return
	}
	
	// Step 3: Handle prompt not found scenarios
	if len(prompts) == 0 {
		fmt.Printf("❌ Prompt not found for URL: %s\n", gistURL)
		fmt.Println()
		fmt.Println("Possible reasons:")
		fmt.Println("  • The Gist URL is not in your Prompt Vault")
		fmt.Println("  • The URL may be incorrect or the Gist may have been deleted")
		fmt.Println("  • You may not have access to this Gist")
		fmt.Println()
		fmt.Println("To see all prompts in your vault, use:")
		fmt.Println("  pv list")
		fmt.Println()
		fmt.Println("To add a new prompt, use:")
		fmt.Println("  pv add <path-to-yaml-file>")
		return
	}
	
	// Use the first matching prompt (should only be one due to unique URLs)
	targetPrompt := prompts[0]
	fmt.Printf("✅ Found prompt: %s (by %s)\n", targetPrompt.Name, targetPrompt.Author)
	fmt.Printf("   Gist URL: %s\n", targetPrompt.GistURL)
	fmt.Println()
	
	// Step 4: Display confirmation interface directly (skip TUI list selection)
	fmt.Println("🎯 Direct URL deletion mode - proceeding to confirmation...")
	fmt.Println()
	
	// Create TUI factory for confirmation dialog
	tuiFactory := tui.NewBubbleTeaTUI()
	
	// Show confirmation interface with prompt details
	confirmed, err := tuiFactory.ShowConfirm(targetPrompt)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 删除操作已取消")
			return
		}
		// Handle other confirmation errors
		fmt.Printf("❌ Error displaying confirmation dialog: %v\n", err)
		return
	}
	
	// Check if user cancelled the confirmation
	if !confirmed {
		fmt.Println("🚫 删除操作已取消")
		return
	}
	
	// Step 5: Execute deletion operation if confirmed
	fmt.Printf("🗑️  Deleting prompt '%s' from URL: %s...\n", targetPrompt.Name, gistURL)
	
	// Use the PromptService DeleteByURL method for deletion
	err = dc.promptService.DeleteByURL(gistURL)
	if err != nil {
		fmt.Printf("❌ Failed to delete prompt: %v\n", err)
		fmt.Println()
		fmt.Println("This could be due to:")
		fmt.Println("  • Network connectivity issues")
		fmt.Println("  • GitHub authentication problems")
		fmt.Println("  • Insufficient permissions to delete the GitHub Gist")
		fmt.Println("  • The Gist may have been already deleted externally")
		fmt.Println()
		fmt.Println("Please check your connection and authentication, then try again.")
		return
	}
	
	// Step 6: Display success message
	fmt.Printf("✅ Successfully deleted prompt '%s'\n", targetPrompt.Name)
	fmt.Printf("   Author: %s\n", targetPrompt.Author)
	fmt.Printf("   Gist URL: %s\n", targetPrompt.GistURL)
	fmt.Printf("   Gist ID: %s\n", gistID)
	fmt.Println()
	fmt.Println("The prompt has been removed from both GitHub Gist and your local index.")
	fmt.Println()
	fmt.Println("💡 Tip: Use 'pv list' to see your remaining prompts.")
}

// handleInvalidURL handles invalid URL input and shows helpful error messages
func (dc *delete) handleInvalidURL(invalidURL string) {
	fmt.Printf("❌ Error: Invalid GitHub Gist URL format: %s\n", invalidURL)
	fmt.Println()
	
	// Provide specific guidance based on the URL pattern
	if contains(invalidURL, "gist.github.com") {
		fmt.Println("The URL contains 'gist.github.com' but doesn't match the expected format.")
		fmt.Println()
		fmt.Println("Valid GitHub Gist URL formats:")
		fmt.Println("  https://gist.github.com/username/gist-id")
		fmt.Println("  https://gist.github.com/gist-id")
		fmt.Println()
		fmt.Println("Where gist-id is a 20 or 32 character hexadecimal string.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  https://gist.github.com/user/1234567890abcdef1234567890abcdef")
		fmt.Println("  https://gist.github.com/abcdef1234567890abcd")
	} else {
		fmt.Println("This appears to be a URL but not a GitHub Gist URL.")
		fmt.Println()
		fmt.Println("If you meant to search for prompts containing this text, try:")
		fmt.Printf("  pv delete \"%s\"\n", invalidURL)
		fmt.Println()
		fmt.Println("For GitHub Gist URLs, use the format:")
		fmt.Println("  https://gist.github.com/username/gist-id")
	}
}

// NewDeleteCommand creates a new delete command with proper Cobra configuration
func NewDeleteCommand(store infra.Store, promptService service.PromptService) DeleteCmd {
	dc := &delete{
		store:         store,
		promptService: promptService,
	}

	return &cobra.Command{
		Use:   "delete [keyword|gist-url]",
		Short: "删除存储的提示",
		Long: `删除存储在 Prompt Vault 中的提示。

支持三种删除模式：

1. 交互式删除 (无参数):
   显示所有提示的列表，允许你通过数字选择要删除的提示。

2. 关键字筛选删除:
   根据关键字筛选提示（匹配名称、作者或描述），然后选择要删除的提示。

3. 直接URL删除:
   直接删除指定 GitHub Gist URL 对应的提示。

所有删除操作都需要确认，防止误删重要数据。删除操作会同时从 GitHub Gist 和本地索引中移除提示。`,
		Example: `  # 交互式删除 - 显示所有提示供选择
  pv delete

  # 关键字筛选删除 - 筛选包含 'golang' 的提示
  pv delete golang

  # 直接删除指定URL的提示
  pv delete https://gist.github.com/user/abc123

  # 使用别名
  pv del golang`,
		Args: cobra.MaximumNArgs(1), // 0-1 arguments allowed
		Run:  dc.execute,
	}
}