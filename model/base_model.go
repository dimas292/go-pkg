package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model is the generic constraint that all domain models must satisfy.
// Any struct embedding BaseModel automatically satisfies this interface.
type Model interface {
	GetID() string
	SetID(id string)
}

// ModelPtr is a constraint ensuring T is a pointer to a type embedding BaseModel.
// Usage: BaseRepository[T, PT ModelPtr[T]] ensures *YourStruct satisfies the interface.
type ModelPtr[T any] interface {
	Model
	*T
}

// BaseModel provides common fields for all GORM models.
// Embed this in your domain structs to get UUID ID, timestamps, and soft-delete.
type BaseModel struct {
	ID        string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate is a GORM hook that auto-generates a UUID before inserting a new record.
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// GetID returns the model's primary key.
func (b *BaseModel) GetID() string {
	return b.ID
}

// SetID sets the model's primary key.
func (b *BaseModel) SetID(id string) {
	b.ID = id
}
