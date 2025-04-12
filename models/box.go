package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Box represents a stored container belonging to a user
type Box struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID      uuid.UUID  `gorm:"not null"`
	PackingMode string     `gorm:"type:varchar(20);not null"` // 'self' or 'sort'
	Status      string     `gorm:"type:varchar(30);not null"` // 'pending_pickup', 'stored', 'in_transit', etc.
	LocationID  *uuid.UUID `gorm:"type:uuid"`
	Location    *StorageLocation
	Items       []Item `gorm:"foreignKey:BoxID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
