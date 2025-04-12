package dto

import "github.com/google/uuid"

// BoxResponse defines the structure returned when viewing a box
type BoxResponse struct {
	ID          uuid.UUID `json:"id"`
	PackingMode string    `json:"packing_mode"`
	Status      string    `json:"status"`
	Items       []ItemDTO `json:"items"`
}
