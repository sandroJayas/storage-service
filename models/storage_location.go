package models

import (
	"time"

	"github.com/google/uuid"
)

// StorageLocation represents a physical location where boxes are stored
type StorageLocation struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Address     string    `gorm:"type:text;not null"`
	Capacity    int       `gorm:"not null"`
	CurrentLoad int       `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
