package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/interfaces"
	"github.com/grigri201/prompt-vault/internal/models"
)

// SyncManager defines the interface for synchronization operations
type SyncManager interface {
	interfaces.Manager

	// SynchronizeData performs bidirectional synchronization
	SynchronizeData(ctx context.Context) error

	// GetSyncStatus returns the current synchronization status
	GetSyncStatus() interfaces.SyncStatus
}

// Manager implements the SyncManager interface
type Manager struct {
	cache       interfaces.CacheManager
	authManager interfaces.AuthManager
	gistClient  *gist.Client
	initialized bool
	status      interfaces.SyncStatus
}

// NewManager creates a new sync manager
func NewManager(
	cache interfaces.CacheManager,
	authManager interfaces.AuthManager,
	gistClient *gist.Client,
) *Manager {
	return &Manager{
		cache:       cache,
		authManager: authManager,
		gistClient:  gistClient,
		status: interfaces.SyncStatus{
			Direction: interfaces.SyncDirectionNone,
		},
	}
}

// Initialize implements interfaces.Manager
func (m *Manager) Initialize(ctx context.Context) error {
	if m.initialized {
		return nil
	}

	// Ensure cache is initialized
	if err := m.cache.Initialize(ctx); err != nil {
		return errors.NewFileSystemError("Initialize", err)
	}

	m.initialized = true
	return nil
}

// Cleanup implements interfaces.Manager
func (m *Manager) Cleanup() error {
	m.initialized = false
	return nil
}

// IsInitialized implements interfaces.Manager
func (m *Manager) IsInitialized() bool {
	return m.initialized
}

// SynchronizeData performs bidirectional synchronization based on timestamps
func (m *Manager) SynchronizeData(ctx context.Context) error {
	if !m.initialized {
		return errors.NewValidationErrorMsg("SynchronizeData", "manager not initialized")
	}

	if m.gistClient == nil {
		return errors.NewAuthErrorMsg("SynchronizeData", "GitHub client not available")
	}

	// Get local index with timestamp
	localIndex, err := m.cache.GetIndex()
	if err != nil || localIndex == nil {
		// If no local index exists, create empty one
		localIndex = &models.Index{
			Entries:   []models.IndexEntry{},
			UpdatedAt: time.Time{}, // Zero time indicates never synced
		}
	}

	// Get the username from auth manager
	if m.authManager == nil {
		return errors.NewValidationErrorMsg("SynchronizeData", "auth manager not initialized")
	}

	username, err := m.authManager.GetUsername()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get username")
	}

	// Get remote index with timestamp
	remoteIndex, _, err := m.gistClient.GetIndexGist(ctx, username)
	if err != nil {
		// If no remote index exists, we'll sync local to remote
		if isIndexNotFoundError(err) {
			if len(localIndex.Entries) > 0 {
				return m.syncToRemote(ctx, localIndex)
			}
			// Both local and remote are empty - this is normal for new users
			// No need to sync anything, just return a helpful message
			return errors.NewValidationErrorMsg("SynchronizeData", "no prompts found. Use 'pv add <file>' or 'pv add <prompt-gist-url>' to get started")
		}
		return errors.WrapWithMessage(err, "failed to get remote index")
	}

	m.status.LocalTime = localIndex.UpdatedAt
	m.status.RemoteTime = remoteIndex.UpdatedAt

	// Compare timestamps and determine sync direction
	if remoteIndex.UpdatedAt.After(localIndex.UpdatedAt) {
		m.status.Direction = interfaces.SyncDirectionDownload
		m.status.NeedsSync = true
		return m.syncFromRemote(ctx, remoteIndex)
	} else if localIndex.UpdatedAt.After(remoteIndex.UpdatedAt) {
		m.status.Direction = interfaces.SyncDirectionUpload
		m.status.NeedsSync = true
		return m.syncToRemote(ctx, localIndex)
	}

	// Timestamps are equal, no sync needed
	m.status.Direction = interfaces.SyncDirectionNone
	m.status.NeedsSync = false
	return nil
}

// GetSyncStatus returns the current synchronization status
func (m *Manager) GetSyncStatus() interfaces.SyncStatus {
	return m.status
}

// syncFromRemote downloads data from remote to local
func (m *Manager) syncFromRemote(ctx context.Context, remoteIndex *models.Index) error {
	m.status.Progress.InProgress = true
	defer func() { m.status.Progress.InProgress = false }()

	localIndex, err := m.cache.GetIndex()
	if err != nil {
		localIndex = &models.Index{Entries: []models.IndexEntry{}}
	}

	// Calculate differences - what needs to be downloaded
	toDownload := m.calculateDifferences(localIndex, remoteIndex)
	m.status.Progress.Total = len(toDownload)
	m.status.Progress.Completed = 0

	// Create GistOperations for reliable fetching
	gistOps := gist.NewGistOperations(gist.GistOperationsConfig{
		Client:     m.gistClient,
		RetryCount: 3,
	})

	// Download changed prompts
	for _, gistID := range toDownload {
		prompt, err := gistOps.FetchPromptGist(ctx, gistID)
		if err != nil {
			// Skip deleted or inaccessible prompts
			if isGistNotFoundError(err) {
				continue
			}
			return errors.WrapWithMessage(err, "failed to fetch prompt")
		}

		if err := m.cache.SavePrompt(prompt); err != nil {
			return errors.WrapWithMessage(err, "failed to save prompt")
		}

		m.status.Progress.Completed++
	}

	// Update local index
	remoteIndex.UpdatedAt = time.Now()
	if err := m.cache.SaveIndex(remoteIndex); err != nil {
		return errors.WrapWithMessage(err, "failed to save index")
	}

	m.status.NeedsSync = false
	return nil
}

// syncToRemote uploads data from local to remote
func (m *Manager) syncToRemote(ctx context.Context, localIndex *models.Index) error {
	m.status.Progress.InProgress = true
	defer func() { m.status.Progress.InProgress = false }()

	// Get the username from auth manager
	if m.authManager == nil {
		return errors.NewValidationErrorMsg("SynchronizeData", "auth manager not initialized")
	}

	username, err := m.authManager.GetUsername()
	if err != nil {
		return errors.WrapWithMessage(err, "failed to get username")
	}

	remoteIndex, _, err := m.gistClient.GetIndexGist(ctx, username)
	if err != nil && !isIndexNotFoundError(err) {
		return errors.WrapWithMessage(err, "failed to get remote index")
	}
	if remoteIndex == nil {
		remoteIndex = &models.Index{Entries: []models.IndexEntry{}}
	}

	// Calculate differences - what needs to be uploaded
	toUpload := m.calculateDifferences(remoteIndex, localIndex)
	m.status.Progress.Total = len(toUpload)
	m.status.Progress.Completed = 0

	// Upload changed prompts
	for _, gistID := range toUpload {
		prompt, err := m.cache.GetPrompt(gistID)
		if err != nil {
			continue // Skip if prompt not found locally
		}

		// Update or create gist for this prompt
		if err := m.updateOrCreatePromptGist(ctx, prompt); err != nil {
			return errors.WrapWithMessage(err, "failed to update prompt")
		}

		m.status.Progress.Completed++
	}

	// Update remote index
	localIndex.UpdatedAt = time.Now()
	_, err = m.gistClient.UpdateIndexGist(ctx, "", localIndex)
	if err != nil {
		return errors.WrapWithMessage(err, "failed to update remote index")
	}

	m.status.NeedsSync = false
	return nil
}

// calculateDifferences returns gist IDs that exist in newIndex but not in oldIndex
// or have different timestamps
func (m *Manager) calculateDifferences(oldIndex, newIndex *models.Index) []string {
	oldMap := make(map[string]models.IndexEntry)
	for _, entry := range oldIndex.Entries {
		oldMap[entry.GistID] = entry
	}

	var differences []string
	for _, newEntry := range newIndex.Entries {
		oldEntry, exists := oldMap[newEntry.GistID]
		if !exists || newEntry.UpdatedAt.After(oldEntry.UpdatedAt) {
			differences = append(differences, newEntry.GistID)
		}
	}

	return differences
}

// updateOrCreatePromptGist updates an existing gist or creates a new one for the prompt
func (m *Manager) updateOrCreatePromptGist(ctx context.Context, prompt *models.Prompt) error {
	// Convert prompt to gist format - recreate the YAML frontmatter format
	content, err := m.promptToGistContent(prompt)
	if err != nil {
		return errors.WrapError("Manager.updateOrCreatePromptGist", err)
	}

	// Create filename based on prompt name
	filename := fmt.Sprintf("%s.yaml", prompt.Name)
	description := fmt.Sprintf("Prompt: %s", prompt.Name)
	if prompt.Description != "" {
		description = prompt.Description
	}

	if prompt.GistID != "" {
		// Update existing gist
		_, err = m.gistClient.UpdateGist(ctx, prompt.GistID, filename, description, content)
		return err
	}

	// Create new gist
	gistID, gistURL, err := m.gistClient.CreateGist(ctx, filename, description, content)
	if err != nil {
		return err
	}

	// Update prompt with new gist information
	prompt.GistID = gistID
	prompt.GistURL = gistURL
	return m.cache.SavePrompt(prompt)
}

// promptToGistContent converts a prompt to the YAML frontmatter format used in gists
func (m *Manager) promptToGistContent(prompt *models.Prompt) (string, error) {
	// Create YAML metadata
	metaYAML := fmt.Sprintf(`name: "%s"
author: "%s"
tags: [%s]`, prompt.Name, prompt.Author, formatTags(prompt.Tags))

	if prompt.Version != "" {
		metaYAML += fmt.Sprintf(`
version: "%s"`, prompt.Version)
	}

	if prompt.Description != "" {
		metaYAML += fmt.Sprintf(`
description: "%s"`, prompt.Description)
	}

	if prompt.Parent != "" {
		metaYAML += fmt.Sprintf(`
parent: "%s"`, prompt.Parent)
	}

	if prompt.ID != "" {
		metaYAML += fmt.Sprintf(`
id: "%s"`, prompt.ID)
	}

	// Combine YAML frontmatter with content
	return fmt.Sprintf("---\n%s\n---\n%s", metaYAML, prompt.Content), nil
}

// formatTags formats tags for YAML array
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	quotedTags := make([]string, len(tags))
	for i, tag := range tags {
		quotedTags[i] = fmt.Sprintf(`"%s"`, tag)
	}

	// Join with comma and space
	result := ""
	for i, tag := range quotedTags {
		if i > 0 {
			result += ", "
		}
		result += tag
	}

	return result
}

// Helper functions to check error types
func isIndexNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Use errors.As to properly unwrap and check the error
	var appErr *errors.AppError
	if errors.As(err, &appErr) {
		// Check if it's a FileSystemError with "index gist not found" message
		if appErr.Type == errors.ErrTypeFileSystem &&
			(contains(appErr.Message, "index gist not found") || contains(appErr.Error(), "index gist not found")) {
			return true
		}
	}

	// Fall back to string matching for any other error types
	errStr := err.Error()
	return contains(errStr, "index gist not found") ||
		(contains(errStr, "FileSystemError") && contains(errStr, "index gist not found"))
}

func isGistNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "404") || contains(errStr, "not found")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
