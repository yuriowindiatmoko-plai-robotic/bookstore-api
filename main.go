package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/bookstore-api/config"
	"github.com/yourusername/bookstore-api/models"
)

func main() {
	db, err := config.Connect()
	if err != nil {
		fmt.Fprintf(gin.DefaultErrorWriter, "database connection failed: %v\n", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&models.Book{}, &models.User{}); err != nil {
		fmt.Fprintf(gin.DefaultErrorWriter, "auto-migrate failed: %v\n", err)
		os.Exit(1)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Fprintf(gin.DefaultWriter, "Server running on :%s\n", port)

	if err := r.Run(":" + port); err != nil {
		fmt.Fprintf(gin.DefaultErrorWriter, "server error: %v\n", err)
		os.Exit(1)
	}
}
