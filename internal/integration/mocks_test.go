package integration

import (
	"context"
	"fmt"

	"github.com/google/go-github/v73/github"
)

// MockGistClient is a mock implementation of the gist client for integration tests
type MockGistClient struct {
	gists              map[string]*github.Gist
	userGists          []*github.Gist
	getGistError       error
	createGistError    error
	updateGistError    error
	nextGistID         int
	createGistCalls    int
	updateGistCalls    int
}

func (m *MockGistClient) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	if m.getGistError != nil {
		return nil, m.getGistError
	}
	gist, exists := m.gists[gistID]
	if !exists {
		return nil, fmt.Errorf("gist not found")
	}
	return gist, nil
}

func (m *MockGistClient) CreatePublicGist(ctx context.Context, gistName, description, content string) (string, string, error) {
	if m.createGistError != nil {
		return "", "", m.createGistError
	}
	m.createGistCalls++
	
	// Generate new gist ID
	m.nextGistID++
	gistID := fmt.Sprintf("public%d", m.nextGistID)
	gistURL := fmt.Sprintf("https://gist.github.com/testuser/%s", gistID)
	
	// Create the gist
	gist := &github.Gist{
		ID:          github.String(gistID),
		HTMLURL:     github.String(gistURL),
		Description: github.String(description),
		Public:      github.Bool(true),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(gistName): {
				Content: github.String(content),
			},
		},
	}
	
	m.gists[gistID] = gist
	m.userGists = append(m.userGists, gist)
	
	return gistID, gistURL, nil
}

func (m *MockGistClient) UpdateGist(ctx context.Context, gistID, gistName, description, content string) (string, error) {
	if m.updateGistError != nil {
		return "", m.updateGistError
	}
	m.updateGistCalls++
	
	gist, exists := m.gists[gistID]
	if !exists {
		return "", fmt.Errorf("gist not found")
	}
	
	// Update the gist
	gist.Description = github.String(description)
	gist.Files = map[github.GistFilename]github.GistFile{
		github.GistFilename(gistName): {
			Content: github.String(content),
		},
	}
	
	return *gist.HTMLURL, nil
}

func (m *MockGistClient) ListUserGists(ctx context.Context, username string) ([]*github.Gist, error) {
	return m.userGists, nil
}

// MockUI is a mock implementation of the UI interface for integration tests
type MockUI struct {
	confirmResponses []bool
	confirmIndex     int
	confirmCalls     []string
}

func (m *MockUI) Confirm(message string) (bool, error) {
	m.confirmCalls = append(m.confirmCalls, message)
	
	if m.confirmIndex < len(m.confirmResponses) {
		response := m.confirmResponses[m.confirmIndex]
		m.confirmIndex++
		return response, nil
	}
	
	return false, nil
}