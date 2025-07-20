package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
)

// newLoginCmd creates the login command
func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with GitHub using a Personal Access Token",
		Long: `Authenticate with GitHub using a Personal Access Token (PAT).
The token will be stored securely in your configuration file.

To create a Personal Access Token:
1. Go to https://github.com/settings/tokens
2. Click "Generate new token" (classic)
3. Give it a descriptive name
4. Select the "gist" scope
5. Click "Generate token"
6. Copy the token and use it with this command`,
		RunE: runLogin,
	}
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Show instructions
	fmt.Fprintln(cmd.OutOrStdout(), "GitHub Personal Access Token Setup")
	fmt.Fprintln(cmd.OutOrStdout(), "==================================")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "To create a Personal Access Token:")
	fmt.Fprintln(cmd.OutOrStdout(), "1. Go to https://github.com/settings/tokens")
	fmt.Fprintln(cmd.OutOrStdout(), "2. Click 'Generate new token' (classic)")
	fmt.Fprintln(cmd.OutOrStdout(), "3. Give it a descriptive name")
	fmt.Fprintln(cmd.OutOrStdout(), "4. Select the 'gist' scope")
	fmt.Fprintln(cmd.OutOrStdout(), "5. Click 'Generate token'")
	fmt.Fprintln(cmd.OutOrStdout(), "6. Copy the token and paste it below")
	fmt.Fprintln(cmd.OutOrStdout())

	// Prompt for token
	fmt.Fprint(cmd.OutOrStdout(), "Enter your GitHub Personal Access Token: ")

	// Read token securely
	token, err := readToken(cmd)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to read token")
	}

	// Validate token is not empty
	token = strings.TrimSpace(token)
	if token == "" {
		fmt.Fprintln(cmd.OutOrStderr(), "\nError: Token cannot be empty")
		return errors.NewValidationErrorMsg("login", "token cannot be empty")
	}

	// Create auth manager
	authManager := auth.NewManager()

	// Validate token with GitHub API
	fmt.Fprintln(cmd.OutOrStdout(), "\nValidating token...")

	ctx := context.Background()
	username, err := validateAndSaveToken(ctx, authManager, token)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to validate token")
	}

	// Success message
	fmt.Fprintf(cmd.OutOrStdout(), "\nSuccessfully authenticated as %s!\n", username)
	fmt.Fprintln(cmd.OutOrStdout(), "Your token has been saved securely.")

	return nil
}

// readToken reads the token from input using visible input
func readToken(cmd *cobra.Command) (string, error) {
	// Always use visible input for better compatibility
	reader := bufio.NewReader(cmd.InOrStdin())
	token, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(token), nil
}

// validateAndSaveToken validates the token and saves it if valid
func validateAndSaveToken(ctx context.Context, authManager *auth.Manager, token string) (string, error) {
	// Create a temporary gist client to validate the token
	client, err := gist.NewClient(token)
	if err != nil {
		return "", errors.WrapWithMessage(err, "failed to create GitHub client")
	}

	// Validate token and get username
	username, err := client.ValidateToken(ctx)
	if err != nil {
		return "", err
	}

	// Save the validated token with username
	if err := authManager.SaveToken(token, username); err != nil {
		return "", errors.WrapWithMessage(err, "failed to save token")
	}

	return username, nil
}
