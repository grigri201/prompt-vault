package infra

import "github.com/grigri/pv/internal/model"

type MemoryStore struct {
	prompts []model.Prompt
}

func (memoryStore *MemoryStore) List() ([]model.Prompt, error) {
	return memoryStore.prompts, nil
}

func NewMemoryStore() Store {
	return &MemoryStore{
		prompts: []model.Prompt{
			{ID: "1", Name: "Translate", Author: "GriGri", GistURL: "TestGistURL"},
		},
	}
}
