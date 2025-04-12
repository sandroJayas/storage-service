package usecase

import (
	"context"
	"github.com/sandroJayas/storage-service/domain/repository"
	"github.com/sandroJayas/storage-service/models"

	"github.com/google/uuid"
	"github.com/sandroJayas/storage-service/dto"
)

type BoxService struct {
	repo repository.BoxRepository
}

func NewBoxService(repo repository.BoxRepository) *BoxService {
	return &BoxService{repo}
}

func (s *BoxService) CreateBox(ctx context.Context, userID uuid.UUID, req dto.CreateBoxRequest) (uuid.UUID, error) {
	box := &models.Box{
		ID:          uuid.New(),
		UserID:      userID,
		PackingMode: req.PackingMode,
		Status:      "pending_pickup",
	}

	if req.PackingMode == "self" {
		box.Items = []models.Item{
			{
				ID:          uuid.New(),
				Name:        req.ItemName,
				Description: req.ItemNote,
				Quantity:    1,
			},
		}
	}

	err := s.repo.Create(ctx, box)
	if err != nil {
		return uuid.Nil, err
	}
	return box.ID, nil
}

func (s *BoxService) ListUserBoxes(ctx context.Context, userID uuid.UUID) ([]dto.BoxResponse, error) {
	boxes, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.BoxResponse
	for _, box := range boxes {
		var items []dto.ItemDTO
		for _, item := range box.Items {
			items = append(items, dto.ItemDTO{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Quantity:    item.Quantity,
				ImageURL:    item.ImageURL,
			})
		}
		result = append(result, dto.BoxResponse{
			ID:          box.ID,
			PackingMode: box.PackingMode,
			Status:      box.Status,
			Items:       items,
		})
	}
	return result, nil
}

func (s *BoxService) GetBoxByID(ctx context.Context, boxID uuid.UUID) (*dto.BoxResponse, error) {
	box, err := s.repo.FindByID(ctx, boxID)
	if err != nil {
		return nil, err
	}

	var items []dto.ItemDTO
	for _, item := range box.Items {
		items = append(items, dto.ItemDTO{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Quantity:    item.Quantity,
			ImageURL:    item.ImageURL,
		})
	}

	resp := &dto.BoxResponse{
		ID:          box.ID,
		PackingMode: box.PackingMode,
		Status:      box.Status,
		Items:       items,
	}
	return resp, nil
}

func (s *BoxService) UpdateStatus(ctx context.Context, boxID uuid.UUID, newStatus string) error {
	return s.repo.UpdateStatus(ctx, boxID, newStatus)
}

func (s *BoxService) DeleteBox(ctx context.Context, boxID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, boxID)
}
