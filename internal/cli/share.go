package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/share"
	"github.com/grigri201/prompt-vault/internal/ui"
	"github.com/spf13/cobra"
)

type shareOptions struct {
	gistID string
}

func newShareCommand() *cobra.Command {
	opts := &shareOptions{}

	cmd := &cobra.Command{
		Use:   "share <gist-id>",
		Short: "Share a private prompt as a public gist",
		Long: `Share a private prompt as a public gist. This creates a public copy of your private prompt
that can be shared with others. The original private gist remains unchanged.

If a public version already exists, you'll be prompted to update it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.gistID = args[0]
			return runShare(cmd, opts)
		},
	}

	return cmd
}

// newShareCmd creates a share command with injected dependencies (for testing)
func newShareCmd(manager shareManager) *cobra.Command {
	opts := &shareOptions{}

	cmd := &cobra.Command{
		Use:   "share <gist-id>",
		Short: "Share a private prompt as a public gist",
		Long: `Share a private prompt as a public gist. This creates a public copy of your private prompt
that can be shared with others. The original private gist remains unchanged.

If a public version already exists, you'll be prompted to update it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.gistID = args[0]
			return runShareWithManager(cmd, opts, manager)
		},
	}

	return cmd
}

// shareManager interface for dependency injection
type shareManager interface {
	SharePrompt(ctx context.Context, privateGistID string) (*share.ShareResult, error)
}

func runShare(cmd *cobra.Command, opts *shareOptions) error {
	// Initialize dependencies
	authManager := auth.NewManager()
	token, username, err := authManager.GetToken()
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgAuthRequired)
	}

	gistClient, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgShareNetworkError)
	}

	// Create UI adapter
	uiAdapter := &cliUI{cmd: cmd}

	// Create share manager
	manager := share.NewManager(gistClient, uiAdapter, username)

	return runShareWithManager(cmd, opts, manager)
}

// cliUI implements the UI interface for share.Manager
type cliUI struct {
	cmd *cobra.Command
}

func (u *cliUI) Confirm(message string) (bool, error) {
	// Print the message
	u.cmd.Println(message)
	
	// Create a selector model for confirmation
	choices := []string{"Yes", "No"}
	model := ui.NewSelector(choices)

	// Run the bubble tea program
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("failed to run confirmation prompt: %w", err)
	}

	// Check if user selected "Yes"
	selectorModel := finalModel.(*ui.SelectorModel)
	return selectorModel.IsConfirmed() && selectorModel.GetSelection() == "Yes", nil
}

func runShareWithManager(cmd *cobra.Command, opts *shareOptions, manager shareManager) error {
	ctx := context.Background()

	// Share the prompt
	result, err := manager.SharePrompt(ctx, opts.gistID)
	if err != nil {
		errMsg := errors.GetShareErrorMessage(err)
		cmd.PrintErrf("%s\n", errMsg)
		return err
	}

	// Display result
	if result.IsUpdate {
		cmd.Printf("Successfully updated public gist: %s\n", result.PublicGistURL)
	} else {
		cmd.Printf("Successfully created public gist: %s\n", result.PublicGistURL)
	}

	return nil
}