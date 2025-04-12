package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sandroJayas/storage-service/dto"
	"github.com/sandroJayas/storage-service/usecase"
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	boxID, err := bc.service.CreateBox(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}

	box, err := bc.service.GetBoxByID(c.Request.Context(), boxID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := bc.service.UpdateStatus(c.Request.Context(), boxID, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid box ID"})
		return
	}

	if err := bc.service.DeleteBox(c.Request.Context(), boxID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Box deleted"})
}
