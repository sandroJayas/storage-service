package repository

import (
	"context"
	"errors"
	"github.com/sandroJayas/storage-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormItemRepository struct {
	db *gorm.DB
}

func NewGormItemRepository(db *gorm.DB) *GormItemRepository {
	return &GormItemRepository{db}
}

func (r *GormItemRepository) Create(ctx context.Context, item *models.Item) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *GormItemRepository) ListByBoxID(ctx context.Context, boxID uuid.UUID) ([]models.Item, error) {
	var items []models.Item
	err := r.db.WithContext(ctx).Where("box_id = ?", boxID).Find(&items).Error
	return items, err
}

func (r *GormItemRepository) GetByID(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) (*models.Item, error) {
	var item models.Item
	err := r.db.WithContext(ctx).
		Joins("JOIN boxes ON boxes.id = items.box_id").
		Where("items.id = ? AND boxes.user_id = ?", itemID, userID).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &item, err
}

func (r *GormItemRepository) Update(ctx context.Context, item *models.Item) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *GormItemRepository) Delete(ctx context.Context, itemID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Item{}, itemID).Error
}
