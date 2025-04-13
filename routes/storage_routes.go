package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sandroJayas/storage-service/controllers"
	"github.com/sandroJayas/storage-service/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"net/http"
)

func RegisterStorageRoutes(r *gin.Engine, boxController *controllers.BoxController, itemController *controllers.ItemController, db *gorm.DB) {

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		db, err := db.DB()
		if err != nil || db.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	boxes := r.Group("/boxes")
	boxes.Use(middleware.AuthMiddleware())
	boxes.Use(middleware.RateLimitMiddleware())
	{
		boxes.POST("", boxController.CreateBox)
		boxes.GET("", boxController.ListUserBoxes)
		boxes.GET(":id", boxController.GetBoxByID)
		boxes.PATCH(":id/status", boxController.UpdateStatus)
		boxes.DELETE(":id", boxController.DeleteBox)

		boxes.POST(":id/items", itemController.AddItem)
		boxes.GET(":id/items", itemController.ListItems)
	}

	items := r.Group("/items")
	items.Use(middleware.AuthMiddleware())
	boxes.Use(middleware.RateLimitMiddleware())
	{
		items.GET(":id", itemController.GetItem)
		items.PATCH(":id", itemController.UpdateItemByID)
		items.DELETE(":id", itemController.DeleteItem)
	}
}
