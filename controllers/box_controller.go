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

type BoxController struct {
	service *usecase.BoxService
}

func NewBoxController(service *usecase.BoxService) *BoxController {
	return &BoxController{service: service}
}

// CreateBox godoc
// @Summary Create a new box
// @Description Create a new box (self or sort packing)
// @Tags boxes
// @Accept json
// @Produce json
// @Param body body dto.CreateBoxRequest true "Box data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /boxes [post]
func (bc *BoxController) CreateBox(c *gin.Context) {
	var req dto.CreateBoxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("Invalid box creation input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	boxID, err := bc.service.CreateBox(c.Request.Context(), userID, req)
	if err != nil {
		utils.Logger.Error("Box creation failed",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	utils.Logger.Info("Box successfully created",
		zap.String("box_id", boxID.String()),
		zap.String("user_id", userID.String()))
	c.JSON(http.StatusCreated, gin.H{"id": boxID})
}

// ListUserBoxes godoc
// @Summary List user boxes
// @Description Get all boxes owned by the user
// @Tags boxes
// @Produce json
// @Success 200 {object} map[string][]dto.BoxResponse
// @Failure 500 {object} map[string]string
// @Router /boxes [get]
func (bc *BoxController) ListUserBoxes(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	boxes, err := bc.service.ListUserBoxes(c.Request.Context(), userID)
	if err != nil {
		utils.Logger.Error("Failed to list user boxes",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("User boxes retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(boxes)))
	c.JSON(http.StatusOK, gin.H{"boxes": boxes})
}

// GetBoxByID godoc
// @Summary Get a box by ID
// @Description Get box and items by ID
// @Tags boxes
// @Produce json
// @Param id path string true "Box ID"
// @Success 200 {object} map[string]dto.BoxResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /boxes/{id} [get]
func (bc *BoxController) GetBoxByID(c *gin.Context) {
	boxID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("Invalid box ID format",
			zap.String("id", c.Param("id")),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	box, err := bc.service.GetBoxByID(c.Request.Context(), boxID, userID)
	if err != nil {
		utils.Logger.Warn("Box not found or not accessible",
			zap.String("id", boxID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("Box retrieved successfully",
		zap.String("id", boxID.String()),
		zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, gin.H{"box": box})
}

// UpdateStatus godoc
// @Summary Update box status
// @Description Update status of a box (admin action)
// @Tags boxes
// @Accept json
// @Produce json
// @Param id path string true "Box ID"
// @Param body body map[string]string true "New status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /boxes/{id}/status [patch]
func (bc *BoxController) UpdateStatus(c *gin.Context) {
	boxID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("Invalid box ID for status update",
			zap.String("id", c.Param("id")),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required,oneof=pending_pickup stored in_transit returned disposed"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Logger.Warn("Invalid status update payload",
			zap.String("id", boxID.String()),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := bc.assertBoxOwnership(c, boxID, userID); err != nil {
		return
	}

	if err := bc.service.UpdateStatus(c.Request.Context(), boxID, body.Status); err != nil {
		utils.Logger.Error("Failed to update box status", zap.String("id", boxID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("Box status updated successfully",
		zap.String("id", boxID.String()),
		zap.String("new_status", body.Status))
	c.JSON(http.StatusOK, gin.H{"message": "Box status updated"})
}

// DeleteBox godoc
// @Summary Delete a box (soft delete)
// @Description Soft delete a box before pickup
// @Tags boxes
// @Produce json
// @Param id path string true "Box ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /boxes/{id} [delete]
func (bc *BoxController) DeleteBox(c *gin.Context) {
	boxID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("Invalid box ID for deletion",
			zap.String("id", c.Param("id")),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := bc.assertBoxOwnership(c, boxID, userID); err != nil {
		return
	}

	if err := bc.service.DeleteBox(c.Request.Context(), boxID); err != nil {
		utils.Logger.Error("Failed to delete box",
			zap.String("id", boxID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("Box soft deleted successfully", zap.String("id", boxID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Box deleted"})
}

// UpdateItem godoc
// @Summary Update an item inside a box
// @Description Update fields like name, description, quantity, image_url
// @Tags boxes
// @Accept json
// @Produce json
// @Param id path string true "Box ID"
// @Param item_id path string true "Item ID"
// @Param body body dto.UpdateItemRequest true "Fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /boxes/{id}/items/{item_id} [patch]
func (bc *BoxController) UpdateItem(c *gin.Context) {
	boxID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Logger.Warn("Invalid box ID for item update",
			zap.String("id", c.Param("id")),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		utils.Logger.Warn("Invalid item ID for update",
			zap.String("item_id", c.Param("item_id")),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
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

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := bc.assertBoxOwnership(c, boxID, userID); err != nil {
		return
	}

	if err := bc.service.UpdateItem(c.Request.Context(), boxID, itemID, req); err != nil {
		utils.Logger.Error("Failed to update item",
			zap.String("id", boxID.String()),
			zap.String("item_id", itemID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Logger.Info("Item updated successfully",
		zap.String("id", boxID.String()),
		zap.String("item_id", itemID.String()),
		zap.String("user_id", userID.String()),
		zap.Any("fields", req),
	)
	c.JSON(http.StatusOK, gin.H{"message": "Item updated successfully"})
}

func (bc *BoxController) assertBoxOwnership(c *gin.Context, boxID uuid.UUID, userID uuid.UUID) error {
	_, err := bc.service.GetBoxByID(c.Request.Context(), boxID, userID)
	if err != nil {
		utils.Logger.Warn("Unauthorized access or box not found",
			zap.String("id", boxID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "box not found or not accessible"})
	}
	return err
}
