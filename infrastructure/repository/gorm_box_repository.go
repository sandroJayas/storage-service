package repository

import (
	"context"
	"github.com/sandroJayas/storage-service/dto"
	"github.com/sandroJayas/storage-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormBoxRepository struct {
	db *gorm.DB
}

func NewGormBoxRepository(db *gorm.DB) *GormBoxRepository {
	return &GormBoxRepository{db}
}

func (r *GormBoxRepository) Create(ctx context.Context, box *models.Box) error {
	return r.db.WithContext(ctx).Create(box).Error
}

func (r *GormBoxRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Box, error) {
	var box models.Box
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ? AND user_id = ?", id, userID).
		First(&box).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

func (r *GormBoxRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Box, error) {
	var boxes []models.Box
	err := r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).Find(&boxes).Error
	return boxes, err
}

func (r *GormBoxRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&models.Box{}).Where("id = ?", id).Update("status", status).Error
}

func (r *GormBoxRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Box{}).Error
}

func (r *GormBoxRepository) UpdateItem(ctx context.Context, boxID, itemID uuid.UUID, req dto.UpdateItemRequest) error {
	var updates = map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Quantity != nil {
		updates["quantity"] = *req.Quantity
	}
	if req.ImageURL != nil {
		updates["image_url"] = *req.ImageURL
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	return r.db.WithContext(ctx).
		Model(&models.Item{}).
		Where("id = ? AND box_id = ?", itemID, boxID).
		Updates(updates).Error
}
