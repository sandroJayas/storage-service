package usecase

import (
	"context"
	"errors"
	"github.com/sandroJayas/storage-service/domain/repository"
	"github.com/sandroJayas/storage-service/models"

	"github.com/google/uuid"
	"github.com/sandroJayas/storage-service/dto"
)

type ItemService struct {
	itemRepo repository.ItemRepository
	boxRepo  repository.BoxRepository
}

func NewItemService(itemRepo repository.ItemRepository, boxRepo repository.BoxRepository) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
		boxRepo:  boxRepo,
	}
}

func (s *ItemService) AddItem(ctx context.Context, boxID, userID uuid.UUID, req dto.AddItemRequest) (uuid.UUID, error) {
	box, err := s.boxRepo.FindByID(ctx, boxID, userID)
	if err != nil {
		return uuid.Nil, err
	}
	if box.PackingMode != "sort" {
		return uuid.Nil, errors.New("can only add items to sort-packed boxes")
	}
	item := &models.Item{
		BoxID:       box.ID,
		Name:        req.Name,
		Description: req.Description,
		Quantity:    req.Quantity,
		ImageURL:    req.ImageURL,
	}

	err = s.itemRepo.Create(ctx, item)
	if err != nil {
		return uuid.Nil, err
	}

	return item.ID, nil
}

func (s *ItemService) ListBoxItems(ctx context.Context, boxID, userID uuid.UUID) ([]dto.ItemDTO, error) {
	box, err := s.boxRepo.FindByID(ctx, boxID, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.itemRepo.ListByBoxID(ctx, box.ID)
	if err != nil {
		return nil, err
	}

	var result []dto.ItemDTO
	for _, item := range items {
		result = append(result, dto.ItemDTO{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Quantity:    item.Quantity,
			ImageURL:    item.ImageURL,
		})
	}

	return result, nil
}

func (s *ItemService) GetItem(ctx context.Context, itemID, userID uuid.UUID) (dto.ItemDTO, error) {
	item, err := s.itemRepo.GetByID(ctx, itemID, userID)
	if err != nil {
		return dto.ItemDTO{}, err
	}

	return dto.ItemDTO{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Quantity:    item.Quantity,
		ImageURL:    item.ImageURL,
	}, nil
}

func (s *ItemService) UpdateItem(ctx context.Context, itemID, userID uuid.UUID, req dto.UpdateItemRequest) error {
	item, err := s.itemRepo.GetByID(ctx, itemID, userID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if req.ImageURL != nil {
		item.ImageURL = *req.ImageURL
	}

	return s.itemRepo.Update(ctx, item)
}

func (s *ItemService) DeleteItem(ctx context.Context, itemID, userID uuid.UUID) error {
	return s.itemRepo.Delete(ctx, itemID)
}
