package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sandroJayas/storage-service/dto"
	"github.com/sandroJayas/storage-service/usecase"
	"github.com/sandroJayas/storage-service/utils"
	"go.uber.org/zap"
)

type ItemController struct {
	itemService *usecase.ItemService
	boxService  *usecase.BoxService
}

func NewItemController(itemService *usecase.ItemService, boxService *usecase.BoxService) *ItemController {
	return &ItemController{
		itemService: itemService,
		boxService:  boxService,
	}
}

// AddItem godoc
// @Summary Add an item to a sort-packed box
// @Description Only applicable for boxes packed by Sort staff.
// @Tags items
// @Accept json
// @Produce json
// @Param box_id path string true "Box ID"
// @Param body body dto.AddItemRequest true "Item data"
// @Success 201 {object} map[string]string "ID of the created item"
// @Failure 400 {object} map[string]string "Invalid box ID or request payload"
// @Failure 404 {object} map[string]string "Box not found or inaccessible"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /boxes/{id}/items [post]
func (ic *ItemController) AddItem(c *gin.Context) {
	boxID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("invalid box ID", zap.String("id", c.Param("box_id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}
	var req dto.AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("invalid add item input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	accountType := c.GetString("account_type")
	if err := assertBoxOwnership(ic.boxService, c, boxID, userID, &accountType); err != nil {
		return
	}
	itemID, err := ic.itemService.AddItem(c.Request.Context(), boxID, userID, req)
	if err != nil {
		utils.Logger.Error("add item failed", zap.String("box_id", boxID.String()), zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("item added", zap.String("box_id", boxID.String()), zap.String("item_id", itemID.String()), zap.String("user_id", userID.String()))
	c.JSON(http.StatusCreated, gin.H{"id": itemID})
}

// ListItems godoc
// @Summary List items in a sort-packed box
// @Description Returns all items for a given box ID. Only the box owner can access this.
// @Tags items
// @Produce json
// @Param box_id path string true "Box ID"
// @Success 200 {object} map[string][]dto.ItemDTO "List of items"
// @Failure 400 {object} map[string]string "Invalid box ID"
// @Failure 404 {object} map[string]string "Box not found or inaccessible"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /boxes/{id}/items [get]
func (ic *ItemController) ListItems(c *gin.Context) {
	boxID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("invalid box ID", zap.String("id", c.Param("box_id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	accountType := c.GetString("account_type")
	if err := assertBoxOwnership(ic.boxService, c, boxID, userID, &accountType); err != nil {
		return
	}
	items, err := ic.itemService.ListBoxItems(c.Request.Context(), boxID, userID)
	if err != nil {
		utils.Logger.Error("list items failed", zap.String("box_id", boxID.String()), zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("items listed", zap.String("box_id", boxID.String()), zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetItem godoc
// @Summary Get a single item by ID
// @Description Returns a single item's full details
// @Tags items
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} dto.ItemDTO "Item details"
// @Failure 400 {object} map[string]string "Invalid item ID"
// @Failure 404 {object} map[string]string "Item not found or inaccessible"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /items/{id} [get]
func (ic *ItemController) GetItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("invalid item ID", zap.String("item_id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	item, err := ic.itemService.GetItem(c.Request.Context(), itemID, userID)
	if err != nil {
		utils.Logger.Warn("get item failed", zap.String("item_id", itemID.String()), zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("item fetched", zap.String("item_id", itemID.String()), zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, item)
}

// UpdateItemByID godoc
// @Summary Update item by ID
// @Description Updates fields of an item (name, description, quantity, image_url)
// @Tags items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param body body dto.UpdateItemRequest true "Fields to update"
// @Success 200 {object} map[string]string "Item updated successfully"
// @Failure 400 {object} map[string]string "Invalid item ID or payload"
// @Failure 404 {object} map[string]string "Item not found or inaccessible"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /items/{id} [patch]
func (ic *ItemController) UpdateItemByID(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("Invalid item ID for update",
			zap.String("item_id", c.Param("id")),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := assertItemOwnership(ic.itemService, c, itemID, userID); err != nil {
		return
	}
	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("Invalid item update payload",
			zap.String("item_id", itemID.String()),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ic.itemService.UpdateItem(c.Request.Context(), itemID, userID, req); err != nil {
		utils.Logger.Error("item update failed",
			zap.String("item_id", itemID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("Item updated successfully",
		zap.String("item_id", itemID.String()),
		zap.String("user_id", userID.String()),
		zap.Any("fields", req),
	)
	c.JSON(http.StatusOK, gin.H{"message": "Item updated successfully"})
}

// DeleteItem godoc
// @Summary Delete an item
// @Description Deletes the item if it exists and belongs to the user
// @Tags items
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} map[string]string "Item deleted successfully"
// @Failure 400 {object} map[string]string "Invalid item ID"
// @Failure 404 {object} map[string]string "Item not found or inaccessible"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /items/{id} [delete]
func (ic *ItemController) DeleteItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("invalid item ID", zap.String("item_id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := assertItemOwnership(ic.itemService, c, itemID, userID); err != nil {
		return
	}
	if err := ic.itemService.DeleteItem(c.Request.Context(), itemID, userID); err != nil {
		utils.Logger.Error("item deletion failed", zap.String("item_id", itemID.String()), zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("item deleted", zap.String("item_id", itemID.String()), zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

func assertItemOwnership(service *usecase.ItemService, c *gin.Context, itemID, userID uuid.UUID) error {
	_, err := service.GetItem(c.Request.Context(), itemID, userID)
	if err != nil {
		utils.Logger.Warn("Unauthorized access or item not found",
			zap.String("item_id", itemID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found or not accessible"})
	}
	return err
}
