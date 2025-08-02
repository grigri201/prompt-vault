package infra

import (
	"github.com/grigri/pv/internal/model"
)

// GistInfo 包含 Gist 的基本信息
type GistInfo struct {
	ID          string
	URL         string
	IsPublic    bool
	HasAccess   bool
	Description string
	Owner       string
}

type Store interface {
	List() ([]model.Prompt, error)
	Add(model.Prompt) error
	Delete(keyword string) error
	Update(model.Prompt) error
	Get(keyword string) ([]model.Prompt, error)
	GetContent(gistID string) (string, error)
	GetGistInfo(gistURL string) (*GistInfo, error)
	FindExistingPromptByURL(gistURL string) (*model.Prompt, error)
}
