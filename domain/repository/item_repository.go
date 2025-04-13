package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sandroJayas/storage-service/models"
)

type ItemRepository interface {
	Create(ctx context.Context, item *models.Item) error
	ListByBoxID(ctx context.Context, boxID uuid.UUID) ([]models.Item, error)
	GetByID(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) (*models.Item, error)
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, itemID uuid.UUID) error
}
