package infra

import (
	"pv/internal/model"
)

type Store interface {
	List() ([]model.Prompt, error)
}
