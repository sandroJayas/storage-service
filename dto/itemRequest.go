package dto

type UpdateItemRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Quantity    *int    `json:"quantity,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
}

type AddItemRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
	ImageURL    string `json:"image_url"`
}
