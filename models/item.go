package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Item represents an item inside a box
type Item struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BoxID       uuid.UUID `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:text"`
	Quantity    int       `gorm:"default:1"`
	ImageURL    string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
