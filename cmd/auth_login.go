package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/service"
)

// AuthLoginCmd is the auth login command type
type AuthLoginCmd = *cobra.Command

// authLogin holds the dependencies for the login command
type authLogin struct {
	service service.AuthService
}

// NewAuthLoginCommand creates the auth login command
func NewAuthLoginCommand(authService service.AuthService) AuthLoginCmd {
	al := &authLogin{service: authService}

	return &cobra.Command{
		Use:   "login",
		Short: "Login with GitHub Personal Access Token",
		Long: `Login with GitHub Personal Access Token to access GitHub Gist API.

You'll need to create a Personal Access Token with 'gist' scope:
1. Go to https://github.com/settings/tokens
2. Click "Generate new token"
3. Select 'gist' scope
4. Generate and copy the token`,
		RunE: al.execute,
	}
}

// execute runs the login command
func (al *authLogin) execute(cmd *cobra.Command, args []string) error {
	// Prompt for token
	fmt.Print("Enter your GitHub Personal Access Token: ")

	token, err := readToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	// Validate and save token
	fmt.Println("\nValidating token...")
	if err := al.service.Login(token); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}

	// Get user info to display success message
	status, err := al.service.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Login succeeded but failed to get user info: %v\n", err)
		fmt.Println("✓ Token validated successfully")
		return nil
	}

	fmt.Println("✓ Token validated successfully")
	fmt.Println(service.GetLoginMessage(status.Username))

	return nil
}

// readToken reads the token from stdin with visible input
func readToken() (string, error) {
	// 使用标准输入，显示输入内容
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read token input: %w", err)
	}

	// Trim any whitespace
	token = strings.TrimSpace(token)

	// Validate token format and length
	if err := validateToken(token); err != nil {
		return "", err
	}

	return token, nil
}

// validateToken validates the token format and provides helpful error messages
func validateToken(token string) error {
	// Check if token is empty
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Check token length (GitHub PATs are typically 40+ characters)
	if len(token) < 40 {
		return fmt.Errorf("token appears too short (%d characters). GitHub Personal Access Tokens are typically 40+ characters. Please ensure you copied the entire token", len(token))
	}

	// Check for valid GitHub token format (alphanumeric, underscore, hyphen)
	validTokenRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validTokenRegex.MatchString(token) {
		// Check if it might be truncated or contain invalid characters
		if strings.Contains(token, " ") {
			return fmt.Errorf("token contains spaces. Please ensure you copied only the token without any surrounding text")
		}
		return fmt.Errorf("token contains invalid characters. GitHub tokens should only contain letters, numbers, underscores, and hyphens")
	}

	// Warning for very long tokens that might indicate paste issues
	if len(token) > 300 {
		fmt.Fprintf(os.Stderr, "Warning: Token is unusually long (%d characters). Please verify it was pasted correctly.\n", len(token))
	}

	return nil
}
