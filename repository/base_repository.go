package repository

import (
	"errors"

	"github.com/dimas292/go-pkg/apperror"
	"github.com/dimas292/go-pkg/model"
	"gorm.io/gorm"
)

// Repository defines the generic CRUD contract.
type Repository[T any, PT model.ModelPtr[T]] interface {
	Create(entity PT) error
	FindByID(id string) (PT, error)
	FindAll(page, perPage int) ([]T, int64, error)
	Update(entity PT) error
	Delete(id string) error
}

// BaseRepository is a generic GORM-backed repository implementing CRUD operations.
// T is the value type (e.g. URL), PT is *T satisfying Model.
type BaseRepository[T any, PT model.ModelPtr[T]] struct {
	DB *gorm.DB
}

// NewBaseRepository creates a new BaseRepository for the given model type.
func NewBaseRepository[T any, PT model.ModelPtr[T]](db *gorm.DB) *BaseRepository[T, PT] {
	return &BaseRepository[T, PT]{DB: db}
}

// Create inserts a new record.
func (r *BaseRepository[T, PT]) Create(entity PT) error {
	if err := r.DB.Create(entity).Error; err != nil {
		return apperror.Internal("failed to create record", err)
	}
	return nil
}

// FindByID retrieves a single record by primary key.
func (r *BaseRepository[T, PT]) FindByID(id string) (PT, error) {
	entity := PT(new(T))
	if err := r.DB.First(entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("record not found")
		}
		return nil, apperror.Internal("failed to find record", err)
	}
	return entity, nil
}

// FindAll retrieves paginated records and total count.
func (r *BaseRepository[T, PT]) FindAll(page, perPage int) ([]T, int64, error) {
	var entities []T
	var total int64

	if err := r.DB.Model(PT(new(T))).Count(&total).Error; err != nil {
		return nil, 0, apperror.Internal("failed to count records", err)
	}

	offset := (page - 1) * perPage
	if err := r.DB.Offset(offset).Limit(perPage).Find(&entities).Error; err != nil {
		return nil, 0, apperror.Internal("failed to retrieve records", err)
	}

	return entities, total, nil
}

// Update saves changes to an existing record.
func (r *BaseRepository[T, PT]) Update(entity PT) error {
	if err := r.DB.Save(entity).Error; err != nil {
		return apperror.Internal("failed to update record", err)
	}
	return nil
}

// Delete soft-deletes a record by primary key.
func (r *BaseRepository[T, PT]) Delete(id string) error {
	entity := PT(new(T))
	if err := r.DB.Where("id = ?", id).Delete(entity).Error; err != nil {
		return apperror.Internal("failed to delete record", err)
	}
	return nil
}
