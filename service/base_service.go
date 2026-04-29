package service

import (
	"github.com/dimas292/go-pkg/model"
	"github.com/dimas292/go-pkg/repository"
)

// Service defines the generic business-logic contract.
type Service[T any, PT model.ModelPtr[T]] interface {
	Create(entity PT) error
	FindByID(id string) (PT, error)
	FindAll(page, perPage int) ([]T, int64, error)
	Update(entity PT) error
	Delete(id string) error
}

// BaseService is a generic service that delegates to BaseRepository.
// Embed this in concrete services and override methods to add business logic.
type BaseService[T any, PT model.ModelPtr[T]] struct {
	Repo *repository.BaseRepository[T, PT]
}

// NewBaseService creates a new BaseService for the given model type.
func NewBaseService[T any, PT model.ModelPtr[T]](repo *repository.BaseRepository[T, PT]) *BaseService[T, PT] {
	return &BaseService[T, PT]{Repo: repo}
}

// Create delegates to the repository.
func (s *BaseService[T, PT]) Create(entity PT) error {
	return s.Repo.Create(entity)
}

// FindByID delegates to the repository.
func (s *BaseService[T, PT]) FindByID(id string) (PT, error) {
	return s.Repo.FindByID(id)
}

// FindAll delegates to the repository.
func (s *BaseService[T, PT]) FindAll(page, perPage int) ([]T, int64, error) {
	return s.Repo.FindAll(page, perPage)
}

// Update delegates to the repository.
func (s *BaseService[T, PT]) Update(entity PT) error {
	return s.Repo.Update(entity)
}

// Delete delegates to the repository.
func (s *BaseService[T, PT]) Delete(id string) error {
	return s.Repo.Delete(id)
}
