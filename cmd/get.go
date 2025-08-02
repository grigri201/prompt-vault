package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/clipboard"
	"github.com/grigri/pv/internal/model"
	"github.com/grigri/pv/internal/service"
	"github.com/grigri/pv/internal/tui"
	"github.com/grigri/pv/internal/variable"
)

type GetCmd = *cobra.Command

type get struct {
	promptService  service.PromptService
	clipboardUtil  clipboard.Util
	variableParser variable.Parser
	tuiInterface   tui.TUIInterface
}

func (g *get) execute(cmd *cobra.Command, args []string) {
	// Validate arguments - we accept 0-1 arguments
	if len(args) > 1 {
		fmt.Println("❌ Error: Too many arguments provided")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  pv get                       # 交互式选择获取")
		fmt.Println("  pv get <keyword>             # 根据关键字筛选获取")
		fmt.Println("  pv get <gist-url>            # 直接获取指定URL的提示")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  pv get                       # 显示所有提示供选择")
		fmt.Println("  pv get golang                # 筛选包含 'golang' 的提示")
		fmt.Println("  pv get https://gist.github.com/user/abc123")
		return
	}

	// Route to appropriate mode based on arguments
	switch len(args) {
	case 0:
		// 交互式模式 - 显示所有提示供用户选择
		g.handleInteractiveMode()
	case 1:
		arg := args[0]
		// Check if it looks like a URL first
		if g.looksLikeURL(arg) {
			// It's a URL-like string, check if it's a valid Gist URL
			if g.isGistURL(arg) {
				// Valid Gist URL - direct get mode
				g.handleDirectMode(arg)
			} else {
				// Invalid Gist URL - show error
				g.handleInvalidURL(arg)
			}
		} else {
			// Not URL-like - keyword filtering mode
			g.handleFilterMode(arg)
		}
	}
}

// isGistURL validates if the provided string is a valid GitHub Gist URL
func (g *get) isGistURL(input string) bool {
	// First check if it looks like a URL at all
	if !g.looksLikeURL(input) {
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
func (g *get) looksLikeURL(input string) bool {
	if len(input) < 7 {
		return false
	}
	return (len(input) >= 8 && input[:8] == "https://") || 
		   (len(input) >= 7 && input[:7] == "http://")
}

// handleInteractiveMode handles the interactive get mode (no arguments)
func (g *get) handleInteractiveMode() {
	fmt.Println("🔄 Interactive mode - loading all prompts...")
	
	// Step 1: Call promptService.ListPrompts() to get all prompts
	prompts, err := g.promptService.ListPrompts()
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
	
	// Step 3: Get user selected prompt from TUI
	selectedPrompt, err := g.tuiInterface.ShowPromptList(prompts)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 获取操作已取消")
			return
		}
		// Handle other TUI errors
		fmt.Printf("❌ Error displaying prompt list: %v\n", err)
		return
	}
	
	fmt.Printf("📝 Selected prompt: %s (by %s)\n", selectedPrompt.Name, selectedPrompt.Author)
	fmt.Println()
	
	// Step 4: Process the selected prompt (get content, handle variables, copy to clipboard)
	g.processSelectedPrompt(selectedPrompt)
}

// handleFilterMode handles the keyword filtering get mode
func (g *get) handleFilterMode(keyword string) {
	fmt.Printf("🔄 Filter mode - searching for prompts matching '%s'...\n", keyword)
	
	// Step 1: Call promptService.FilterPrompts(keyword) to filter prompts
	filteredPrompts, err := g.promptService.FilterPrompts(keyword)
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
		fmt.Println("  pv get golang       # Search for prompts containing 'golang'")
		fmt.Println("  pv get john         # Search for prompts by author 'john'")
		fmt.Println("  pv get review       # Search for prompts about 'review'")
		return
	}
	
	// Step 3: Display filtering results statistics and information
	fmt.Printf("🎯 Found %d prompt(s) matching '%s':\n", len(filteredPrompts), keyword)
	fmt.Println()
	
	// Show a preview of the filtered results
	fmt.Println("Matching prompts:")
	for i, prompt := range filteredPrompts {
		fmt.Printf("  %d. %s (by %s)\n", i+1, prompt.Name, prompt.Author)
	}
	fmt.Println()
	fmt.Printf("✨ Keyword '%s' will be highlighted in the selection interface.\n", keyword)
	fmt.Println()
	
	// Step 4: Get user selected prompt from TUI
	selectedPrompt, err := g.tuiInterface.ShowPromptList(filteredPrompts)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 获取操作已取消")
			return
		}
		// Handle other TUI errors
		fmt.Printf("❌ Error displaying filtered prompt list: %v\n", err)
		return
	}
	
	fmt.Printf("📝 Selected prompt: %s (by %s)\n", selectedPrompt.Name, selectedPrompt.Author)
	fmt.Printf("    Matches keyword: '%s'\n", keyword)
	fmt.Println()
	
	// Step 5: Process the selected prompt
	g.processSelectedPrompt(selectedPrompt)
}

// handleDirectMode handles the direct URL get mode
func (g *get) handleDirectMode(gistURL string) {
	fmt.Printf("🔄 Direct mode - processing URL: %s\n", gistURL)
	
	// Step 1: Validate and parse the GitHub Gist URL
	if !g.isGistURL(gistURL) {
		fmt.Println("❌ Error: Invalid GitHub Gist URL format")
		fmt.Println()
		fmt.Println("Valid URL formats:")
		fmt.Println("  https://gist.github.com/username/gist-id")
		fmt.Println("  https://gist.github.com/gist-id")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  pv get https://gist.github.com/user/1234567890abcdef")
		return
	}
	
	// Step 2: Get the prompt directly by URL
	prompt, err := g.promptService.GetPromptByURL(gistURL)
	if err != nil {
		fmt.Printf("❌ Error getting prompt: %v\n", err)
		fmt.Println()
		fmt.Println("This could be due to:")
		fmt.Println("  • The Gist URL is not in your Prompt Vault")
		fmt.Println("  • The URL may be incorrect or the Gist may have been deleted")
		fmt.Println("  • You may not have access to this Gist")
		fmt.Println("  • Network connectivity issues")
		fmt.Println()
		fmt.Println("To see all prompts in your vault, use:")
		fmt.Println("  pv list")
		return
	}
	
	fmt.Printf("✅ Found prompt: %s (by %s)\n", prompt.Name, prompt.Author)
	fmt.Printf("   Gist URL: %s\n", prompt.GistURL)
	fmt.Println()
	
	// Step 3: Process the prompt directly
	g.processSelectedPrompt(*prompt)
}

// handleInvalidURL handles invalid URL input and shows helpful error messages
func (g *get) handleInvalidURL(invalidURL string) {
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
		fmt.Printf("  pv get \"%s\"\n", invalidURL)
		fmt.Println()
		fmt.Println("For GitHub Gist URLs, use the format:")
		fmt.Println("  https://gist.github.com/username/gist-id")
	}
}

// processSelectedPrompt handles the core processing workflow
func (g *get) processSelectedPrompt(prompt model.Prompt) {
	fmt.Printf("🔄 Processing prompt: %s...\n", prompt.Name)
	
	// Step 1: Get prompt content
	content, err := g.promptService.GetPromptContent(&prompt)
	if err != nil {
		fmt.Printf("❌ Failed to get prompt content: %v\n", err)
		fmt.Println()
		fmt.Println("This could be due to:")
		fmt.Println("  • Network connectivity issues")
		fmt.Println("  • GitHub authentication problems")
		fmt.Println("  • The prompt may have been deleted")
		fmt.Println()
		fmt.Println("Please check your connection and authentication, then try again.")
		return
	}
	
	// Step 2: Check for variables and handle them
	if !g.variableParser.HasVariables(content) {
		// No variables, directly copy to clipboard
		fmt.Println("📄 Prompt has no variables, copying directly to clipboard...")
		g.copyToClipboard(content, prompt.Name)
		return
	}
	
	// Step 3: Extract variables and show form
	fmt.Println("🔧 Prompt contains variables, collecting user input...")
	variables := g.variableParser.ExtractVariables(content)
	
	fmt.Printf("📋 Found %d variable(s): %v\n", len(variables), variables)
	fmt.Println()
	
	values, err := g.tuiInterface.ShowVariableForm(variables)
	if err != nil {
		// Handle user cancellation gracefully
		if err.Error() == tui.ErrMsgUserCancelled {
			fmt.Println("🚫 获取操作已取消")
			return
		}
		// Handle other form errors
		fmt.Printf("❌ Error collecting variable values: %v\n", err)
		return
	}
	
	// Step 4: Replace variables with user values
	fmt.Println("🔄 Replacing variables with your values...")
	finalContent := g.variableParser.ReplaceVariables(content, values)
	
	// Display what was replaced
	fmt.Println("✅ Variable replacement completed:")
	for variable, value := range values {
		fmt.Printf("  • {%s} → %s\n", variable, value)
	}
	fmt.Println()
	
	// Step 5: Copy to clipboard
	g.copyToClipboard(finalContent, prompt.Name)
}

// copyToClipboard handles clipboard operations with appropriate feedback
func (g *get) copyToClipboard(content, promptName string) {
	// Check if clipboard is available
	if !g.clipboardUtil.IsAvailable() {
		fmt.Println("⚠️  Clipboard is not available on this system")
		fmt.Println()
		fmt.Println("Prompt content:")
		fmt.Println("================")
		fmt.Println(content)
		fmt.Println("================")
		fmt.Println()
		fmt.Printf("Please manually copy the above content for prompt: %s\n", promptName)
		return
	}
	
	// Copy to clipboard
	err := g.clipboardUtil.Copy(content)
	if err != nil {
		fmt.Printf("❌ Failed to copy to clipboard: %v\n", err)
		fmt.Println()
		fmt.Println("Prompt content:")
		fmt.Println("================")
		fmt.Println(content)
		fmt.Println("================")
		fmt.Println()
		fmt.Printf("Please manually copy the above content for prompt: %s\n", promptName)
		return
	}
	
	// Success message
	fmt.Printf("✅ Successfully copied prompt '%s' to clipboard!\n", promptName)
	fmt.Printf("   Content length: %d characters\n", len(content))
	fmt.Println()
	fmt.Println("💡 The prompt is now ready to paste into your target application.")
}

// NewGetCommand creates a new get command with proper Cobra configuration
func NewGetCommand(
	promptService service.PromptService,
	clipboardUtil clipboard.Util,
	variableParser variable.Parser,
	tuiInterface tui.TUIInterface,
) GetCmd {
	g := &get{
		promptService:  promptService,
		clipboardUtil:  clipboardUtil,
		variableParser: variableParser,
		tuiInterface:   tuiInterface,
	}

	return &cobra.Command{
		Use:   "get [keyword|gist-url]",
		Short: "获取存储的提示并复制到剪贴板",
		Long: `从 Prompt Vault 中获取提示，支持变量替换，并将最终结果复制到剪贴板。

支持三种获取模式：

1. 交互式获取 (无参数):
   显示所有提示的列表，允许你通过数字选择要获取的提示。

2. 关键字筛选获取:
   根据关键字筛选提示（匹配名称、作者或描述），然后选择要获取的提示。

3. 直接URL获取:
   直接获取指定 GitHub Gist URL 对应的提示。

如果提示包含 {variable} 占位符，系统会显示表单供您填写变量值，
然后将变量替换为您的输入内容后复制到剪贴板。`,
		Example: `  # 交互式获取 - 显示所有提示供选择
  pv get

  # 关键字筛选获取 - 筛选包含 'golang' 的提示
  pv get golang

  # 直接获取指定URL的提示
  pv get https://gist.github.com/user/abc123

  # 处理包含变量的提示
  # 系统会自动检测 {name} 和 {role} 变量并要求您填写`,
		Args: cobra.MaximumNArgs(1), // 0-1 arguments allowed
		Run:  g.execute,
	}
}
