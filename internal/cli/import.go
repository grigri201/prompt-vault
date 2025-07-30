package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/auth"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/imports"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/ui"
	"github.com/spf13/cobra"
)

type importOptions struct {
	gistURL string
}

func newImportCommand() *cobra.Command {
	opts := &importOptions{}

	cmd := &cobra.Command{
		Use:   "import <gist-url>",
		Short: "Import a public prompt gist into your collection",
		Long: `Import a public prompt gist into your collection. This adds an existing prompt gist
to your index, allowing you to use it with 'pv get'.

If the prompt is already imported, you'll be prompted to update it if the version has changed.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.gistURL = args[0]
			return runImport(cmd, opts)
		},
	}

	// Apply auto-sync middleware
	WrapWithAutoSync(cmd)

	return cmd
}

// newImportCmd creates an import command with injected dependencies (for testing)
func newImportCmd(manager importManager, gistClient gistClientForImport) *cobra.Command {
	opts := &importOptions{}

	cmd := &cobra.Command{
		Use:   "import <gist-url>",
		Short: "Import a public prompt gist into your collection",
		Long: `Import a public prompt gist into your collection. This adds an existing prompt gist
to your index, allowing you to use it with 'pv get'.

If the prompt is already imported, you'll be prompted to update it if the version has changed.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.gistURL = args[0]
			return runImportWithDeps(cmd, opts, manager, gistClient)
		},
	}

	return cmd
}

// importManager interface for dependency injection
type importManager interface {
	ImportPrompt(ctx context.Context, gistURL string, index *models.Index) (*imports.ImportResult, error)
}

// gistClientForImport interface for gist operations needed by import command
type gistClientForImport interface {
	GetIndex(ctx context.Context) (*models.Index, error)
	UpdateIndex(ctx context.Context, index *models.Index) error
}

func runImport(cmd *cobra.Command, opts *importOptions) error {
	// Perform pre-sync to ensure we have latest index and avoid conflicts
	if err := performPreSync(cmd, false); err != nil {
		return err
	}

	// Initialize dependencies
	authManager := auth.NewManager()
	token, err := authManager.GetToken()
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgAuthRequired)
	}
	username, err := authManager.GetUsername()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get username")
	}

	gistClient, err := gist.NewClient(token)
	if err != nil {
		return errors.WrapWithMessage(err, errors.ErrMsgNetworkTimeout)
	}

	// Create UI adapter
	uiAdapter := &importUIAdapter{cmd: cmd}

	// Create import manager
	manager := imports.NewManager(gistClient, uiAdapter)

	// Create gist client adapter
	gistAdapter := &gistClientAdapter{
		client:   gistClient,
		username: username,
	}

	return runImportWithDeps(cmd, opts, manager, gistAdapter)
}

// gistClientAdapter adapts gist.Client to gistClientForImport interface
type gistClientAdapter struct {
	client   *gist.Client
	username string
}

func (a *gistClientAdapter) GetIndex(ctx context.Context) (*models.Index, error) {
	// List all gists to find the index
	gists, err := a.client.ListUserGists(ctx, a.username)
	if err != nil {
		return nil, err
	}

	// Find index gist
	var indexGist *github.Gist
	for _, g := range gists {
		if g.Description != nil && *g.Description == "Prompt Vault Index" {
			for filename := range g.Files {
				if filename == "index.json" {
					indexGist = g
					break
				}
			}
		}
	}

	if indexGist == nil {
		// Return empty index if not found
		return &models.Index{
			Username:        a.username,
			Entries:         []models.IndexEntry{},
			ImportedEntries: []models.IndexEntry{},
			UpdatedAt:       time.Now(),
		}, nil
	}

	// Get the full gist content
	fullGist, err := a.client.GetGist(ctx, *indexGist.ID)
	if err != nil {
		return nil, err
	}

	// Parse index from gist
	indexFile, ok := fullGist.Files["index.json"]
	if !ok {
		return nil, errors.NewValidationErrorMsg("parsePromptVaultIndex", "index.json not found in gist")
	}

	content := *indexFile.Content
	var index models.Index
	if err := json.Unmarshal([]byte(content), &index); err != nil {
		return nil, errors.WrapError("parsePromptVaultIndex", err)
	}

	return &index, nil
}

func (a *gistClientAdapter) UpdateIndex(ctx context.Context, index *models.Index) error {
	_, err := a.client.UpdateIndexGist(ctx, a.username, index)
	return err
}

func runImportWithDeps(cmd *cobra.Command, opts *importOptions, manager importManager, gistClient gistClientForImport) error {
	ctx := context.Background()

	// Create cache manager to work with local index (like upload does)
	_, cacheManager := createManagers()

	// Get current local index after preSync
	index, err := cacheManager.GetIndex()
	if err != nil {
		// If index doesn't exist, create a new one (like upload does)
		index = &models.Index{
			Username:        "", // Will be set during sync
			Entries:         []models.IndexEntry{},
			ImportedEntries: []models.IndexEntry{},
			UpdatedAt:       time.Now(),
		}
	}

	// Import the prompt (this will modify the index's ImportedEntries)
	result, err := manager.ImportPrompt(ctx, opts.gistURL, index)
	if err != nil {
		errMsg := errors.GetImportErrorMessage(err)
		cmd.PrintErrf("%s\n", errMsg)
		return err
	}

	// Update the local index timestamp
	index.UpdatedAt = time.Now()

	// Save the updated local index (like upload does)
	if err := cacheManager.SaveIndex(index); err != nil {
		// Don't fail the import if index saving fails (like upload does)
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update local index: %v\n", err)
	}

	// GitHub index update will be handled by auto-sync middleware (like upload does)

	// Display result
	if result.IsUpdate {
		cmd.Printf("Successfully updated prompt (gist: %s) from version %s to %s\n",
			result.GistID, result.OldVersion, result.NewVersion)
	} else {
		cmd.Printf("Successfully imported prompt (gist: %s)\n", result.GistID)
	}

	return nil
}

// importUIAdapter implements the UI interface for imports.Manager
type importUIAdapter struct {
	cmd *cobra.Command
}

func (u *importUIAdapter) Confirm(message string) (bool, error) {
	// Print the message
	u.cmd.Println(message)

	// Create a selector model for confirmation
	choices := []string{"Yes", "No"}
	model := ui.NewSelector(choices)

	// Run the bubble tea program
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, errors.WrapError("showConfirmation", err)
	}

	// Check if user selected "Yes"
	selectorModel := finalModel.(*ui.SelectorModel)
	return selectorModel.IsConfirmed() && selectorModel.GetSelection() == "Yes", nil
}
