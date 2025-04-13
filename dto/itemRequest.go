package dto

type UpdateItemRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Quantity    *int    `json:"quantity,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
}
