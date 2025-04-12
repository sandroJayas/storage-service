package models

import (
	"time"

	"github.com/google/uuid"
)

// StorageOrder represents an action to move or store a box
type StorageOrder struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID        uuid.UUID `gorm:"not null"`
	BoxID         uuid.UUID `gorm:"not null"`
	Type          string    `gorm:"type:varchar(20);not null"` // 'pickup', 'return', 'relocate'
	ScheduledDate time.Time `gorm:"not null"`
	Status        string    `gorm:"type:varchar(30);not null"` // 'requested', 'in_progress', 'completed', 'cancelled'
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
