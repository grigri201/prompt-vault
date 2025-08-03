package infra

import (
	"github.com/grigri/pv/internal/model"
)

type Store interface {
	List() ([]model.Prompt, error)
}
