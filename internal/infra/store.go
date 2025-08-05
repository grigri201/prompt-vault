package infra

import (
	"github.com/grigri/pv/internal/model"
)

type Store interface {
	List() ([]model.Prompt, error)
	Add(model.Prompt) error
	Delete(keyword string) error
	Update(model.Prompt) error
	Get(keyword string) ([]model.Prompt, error)
	GetContent(gistID string) (string, error)
}
