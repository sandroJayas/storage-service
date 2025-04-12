package dto

type CreateBoxRequest struct {
	PackingMode string `json:"packing_mode" binding:"required,oneof=self sort"`
	ItemName    string `json:"item_name"` // Optional: only used in self mode
	ItemNote    string `json:"item_note"`
}
