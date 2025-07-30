package sync

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/grigri201/prompt-vault/internal/cache"
	"github.com/grigri201/prompt-vault/internal/config"
	"github.com/grigri201/prompt-vault/internal/errors"
	"github.com/grigri201/prompt-vault/internal/gist"
	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
)

// SyncResult contains the results of a sync operation
type SyncResult struct {
	PromptsUpdated int
	PromptsAdded   int
	PromptsRemoved int
	Duration       time.Duration
	TotalPrompts   int
}

// Service handles synchronization operations
type Service interface {
	// SyncWithTimeout performs sync with configurable timeout
	SyncWithTimeout(ctx context.Context, timeout time.Duration) (*SyncResult, error)
	
	// GetLastSyncTime returns when the last successful sync occurred
	GetLastSyncTime() (time.Time, error)
}

// serviceImpl is the default implementation
type serviceImpl struct {
	configManager *config.Manager
	cacheManager  *cache.Manager
	yamlParser    *parser.YAMLParser
}

// NewService creates a new sync service
func NewService(configManager *config.Manager, cacheManager *cache.Manager) Service {
	// Create YAML parser with lenient mode for sync
	yamlParser := parser.NewYAMLParser(parser.YAMLParserConfig{
		Strict: false, // Be lenient when syncing
	})
	
	return &serviceImpl{
		configManager: configManager,
		cacheManager:  cacheManager,
		yamlParser:    yamlParser,
	}
}

// SyncWithTimeout implements the Service interface
func (s *serviceImpl) SyncWithTimeout(ctx context.Context, timeout time.Duration) (*SyncResult, error) {
	startTime := time.Now()
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Get config
	cfg, err := s.configManager.GetConfig()
	if err != nil {
		return nil, errors.WrapWithMessage(err, "failed to get config")
	}
	
	if cfg.Token == "" {
		return nil, errors.NewAuthErrorMsg("sync", "not authenticated. Please run 'pv login' first")
	}
	
	// Create GitHub client
	client, err := gist.NewClient(cfg.Token)
	if err != nil {
		return nil, errors.WrapWithMessage(err, "failed to create GitHub client")
	}
	
	// Get all gists for the user
	gists, err := client.ListUserGists(ctx, cfg.Username)
	if err != nil {
		return nil, errors.WrapWithMessage(err, "failed to list gists")
	}
	
	result := &SyncResult{
		Duration: time.Since(startTime),
	}
	
	if len(gists) == 0 {
		// Create a new empty index
		index := &models.Index{
			Username:  cfg.Username,
			Entries:   []models.IndexEntry{},
			UpdatedAt: time.Now(),
		}
		if err := s.cacheManager.SaveIndex(index); err != nil {
			return nil, errors.WrapWithMessage(err, "failed to save empty index")
		}
		return result, nil
	}
	
	// Process gists and build index
	var entries []models.IndexEntry
	processedCount := 0
	
	for _, g := range gists {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, errors.NewNetworkErrorMsg("sync", "sync operation timed out")
		default:
		}
		
		if err := s.processGist(ctx, g, client, &entries, &processedCount); err != nil {
			// Log error but continue with other gists
			continue
		}
	}
	
	// Get existing index to compare
	existingIndex, err := s.cacheManager.GetIndex()
	var existingCount int
	if err == nil && existingIndex != nil {
		existingCount = len(existingIndex.Entries)
	}
	
	// Create new index
	index := &models.Index{
		Username:  cfg.Username,
		Entries:   entries,
		UpdatedAt: time.Now(),
	}
	
	// Save the index
	if err := s.cacheManager.SaveIndex(index); err != nil {
		return nil, errors.WrapWithMessage(err, "failed to save index")
	}
	
	// Calculate results
	result.TotalPrompts = len(entries)
	result.PromptsAdded = len(entries) - existingCount
	if result.PromptsAdded < 0 {
		result.PromptsRemoved = -result.PromptsAdded
		result.PromptsAdded = 0
	}
	result.Duration = time.Since(startTime)
	
	return result, nil
}

// processGist processes a single gist and adds it to entries if it's a valid prompt
func (s *serviceImpl) processGist(ctx context.Context, g *github.Gist, client *gist.Client, entries *[]models.IndexEntry, processedCount *int) error {
	// Skip if gist has no ID
	if g.ID == nil {
		return nil
	}
	
	// Skip if gist is public (we only sync private prompts)
	if g.Public != nil && *g.Public {
		return nil
	}
	
	// Skip index gists - check if any file name ends with -promptvault-index.json
	for filename := range g.Files {
		if strings.HasSuffix(string(filename), "-promptvault-index.json") {
			return nil
		}
	}
	
	// Get the full gist with file contents
	fullGist, err := client.GetGist(ctx, *g.ID)
	if err != nil {
		return err
	}
	
	// Look for a prompt file in the gist
	for _, file := range fullGist.Files {
		if file.Content == nil {
			continue
		}
		
		// Try to parse as a prompt
		prompt, err := s.yamlParser.ParsePromptFile(*file.Content)
		if err != nil {
			// Not a valid prompt, skip
			continue
		}
		
		// Set gist information
		prompt.GistID = *g.ID
		if g.HTMLURL != nil {
			prompt.GistURL = *g.HTMLURL
		}
		if g.UpdatedAt != nil {
			prompt.UpdatedAt = g.UpdatedAt.Time
		}
		
		// Cache the prompt
		if err := s.cacheManager.SavePrompt(prompt); err != nil {
			// Log but don't fail
			continue
		}
		
		// Add to index
		*entries = append(*entries, prompt.ToIndexEntry())
		*processedCount++
		
		// Only process first valid prompt file in the gist
		break
	}
	
	return nil
}

// GetLastSyncTime implements the Service interface
func (s *serviceImpl) GetLastSyncTime() (time.Time, error) {
	index, err := s.cacheManager.GetIndex()
	if err != nil {
		return time.Time{}, err
	}
	
	if index == nil {
		return time.Time{}, errors.NewFileSystemErrorMsg("GetLastSyncTime", "no index found")
	}
	
	return index.UpdatedAt, nil
}