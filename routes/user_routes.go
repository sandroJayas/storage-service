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

func RegisterStorageRoutes(r *gin.Engine, boxController *controllers.BoxController, db *gorm.DB) {

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

	routes := r.Group("/boxes")
	{
		routes.POST("", middleware.AuthMiddleware(), boxController.CreateBox)
		routes.GET("", middleware.AuthMiddleware(), boxController.ListUserBoxes)
		routes.GET(":id", middleware.AuthMiddleware(), boxController.GetBoxByID)
		routes.PATCH(":id/status", middleware.AuthMiddleware(), boxController.UpdateStatus)
		routes.DELETE(":id", middleware.AuthMiddleware(), boxController.DeleteBox)
		routes.PATCH(":id/items/:item_id", middleware.AuthMiddleware(), boxController.UpdateItem)
	}
}
