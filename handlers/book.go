package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yuriowindiatmoko-plai-robotic/bookstore-api/models"
	"gorm.io/gorm"
)

type BookHandler struct {
	db *gorm.DB
}

func NewBookHandler(db *gorm.DB) *BookHandler {
	return &BookHandler{db: db}
}

type createBookInput struct {
	Title       string  `json:"title" binding:"required"`
	Author      string  `json:"author" binding:"required"`
	ISBN        string  `json:"isbn" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock"`
	Description *string `json:"description"`
}

type updateBookInput struct {
	Title       *string  `json:"title"`
	Author      *string  `json:"author"`
	ISBN        *string  `json:"isbn"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	Description *string  `json:"description"`
}

func (h *BookHandler) GetBooks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	var total int64
	if err := h.db.Model(&models.Book{}).Count(&total).Error; err != nil {
		logDBError(c, "count books", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list books"})
		return
	}

	var books []models.Book
	if err := h.db.Offset((page - 1) * limit).Limit(limit).Find(&books).Error; err != nil {
		logDBError(c, "list books", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list books"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": books,
		"meta": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func (h *BookHandler) GetBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var book models.Book
	err := h.db.First(&book, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	if err != nil {
		logDBError(c, "get book", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": book})
}

func (h *BookHandler) CreateBook(c *gin.Context) {
	var in createBookInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	book := models.Book{
		Title:       in.Title,
		Author:      in.Author,
		ISBN:        in.ISBN,
		Price:       in.Price,
		Stock:       in.Stock,
		Description: in.Description,
	}

	if err := h.db.Create(&book).Error; err != nil {
		logDBError(c, "create book", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create book"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": book})
}

func (h *BookHandler) UpdateBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var book models.Book
	err := h.db.First(&book, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	if err != nil {
		logDBError(c, "get book for update", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update book"})
		return
	}

	var in updateBookInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	updates := map[string]any{}
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Author != nil {
		updates["author"] = *in.Author
	}
	if in.ISBN != nil {
		updates["isbn"] = *in.ISBN
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.Stock != nil {
		updates["stock"] = *in.Stock
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := h.db.Model(&book).Updates(updates).Error; err != nil {
		logDBError(c, "update book", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update book"})
		return
	}

	if err := h.db.First(&book, id).Error; err != nil {
		logDBError(c, "reload book after update", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": book})
}

func (h *BookHandler) DeleteBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var book models.Book
	err := h.db.First(&book, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	if err != nil {
		logDBError(c, "get book for delete", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete book"})
		return
	}

	if err := h.db.Delete(&book).Error; err != nil {
		logDBError(c, "delete book", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete book"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

func logDBError(c *gin.Context, action string, err error) {
	fmt.Fprintf(gin.DefaultErrorWriter, "[db] %s %s: %v\n", c.Request.Method, c.Request.URL.Path, err)
}
