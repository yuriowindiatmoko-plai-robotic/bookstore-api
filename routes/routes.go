package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuriowindiatmoko-plai-robotic/bookstore-api/handlers"
	"github.com/yuriowindiatmoko-plai-robotic/bookstore-api/middleware"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	router.GET("/health", healthHandler)

	authH := handlers.NewAuthHandler(db)
	bookH := handlers.NewBookHandler(db)

	v1 := router.Group("/api/v1")
	v1.GET("/health", healthHandler)

	v1.POST("/register", authH.Register)
	v1.POST("/login", authH.Login)
	v1.GET("/books", bookH.GetBooks)
	v1.GET("/books/:id", bookH.GetBook)

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/books", bookH.CreateBook)
	protected.PUT("/books/:id", bookH.UpdateBook)
	protected.DELETE("/books/:id", bookH.DeleteBook)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}})
}
