package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/yuriowindiatmoko-plai-robotic/bookstore-api/config"
	"github.com/yuriowindiatmoko-plai-robotic/bookstore-api/models"
	"github.com/yuriowindiatmoko-plai-robotic/bookstore-api/routes"
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

	routes.SetupRoutes(r, db)

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
