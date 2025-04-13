package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sandroJayas/storage-service/dto"
	"github.com/sandroJayas/storage-service/models"
)

type BoxRepository interface {
	Create(ctx context.Context, box *models.Box) error
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Box, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Box, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	UpdateItem(ctx context.Context, boxID, itemID uuid.UUID, req dto.UpdateItemRequest) error
}
