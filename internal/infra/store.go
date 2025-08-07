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
	// 现有方法
	List() ([]model.Prompt, error)
	Add(model.Prompt) error
	Delete(keyword string) error
	Update(model.Prompt) error
	Get(keyword string) ([]model.Prompt, error)
	GetContent(gistID string) (string, error)
	
	// 新增 gist 管理方法
	CreatePublicGist(prompt model.Prompt) (string, error)
	UpdateGist(gistURL string, prompt model.Prompt) error
	GetGistInfo(gistURL string) (*GistInfo, error)
	
	// 新增 export 管理方法
	AddExport(prompt model.IndexedPrompt) error
	UpdateExport(prompt model.IndexedPrompt) error
	GetExports() ([]model.IndexedPrompt, error)
	
	// 新增 URL 重复检查方法
	FindExistingPromptByURL(gistURL string) (*model.Prompt, error)
}
