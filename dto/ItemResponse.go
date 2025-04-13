package dto

import "github.com/google/uuid"

type ItemDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	ImageURL    string    `json:"image_url"`
}
