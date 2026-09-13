package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"muma-ramen-backend/models"
)

type CategoryHandler struct {
	DB *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{DB: db}
}

// ListCategories returns all categories (public)
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	var categories []models.Category

	query := h.DB

	// Include menu items count or items
	if c.Query("include_items") == "true" {
		query = query.Preload("MenuItems")
	}

	if err := query.Order("sort_order, name").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// CreateCategory creates a new category (admin+)
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"category": cat})
}

// UpdateCategory updates a category (admin+)
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var cat models.Category
	if err := h.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&cat).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"category": cat})
}

// DeleteCategory deletes a category (admin+)
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	var cat models.Category
	if err := h.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Check if category has menu items
	var count int64
	h.DB.Model(&models.MenuItem{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category masih punya menu items. Hapus atau pindahkan dulu."})
		return
	}

	if err := h.DB.Delete(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// UploadImage handles image upload for a category
func (h *CategoryHandler) UploadImage(c *gin.Context) {
	id := c.Param("id")

	var cat models.Category
	if err := h.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}

	// Generate filename
	filename := "category_" + id + "_" + file.Filename
	filepath := "uploads/" + filename

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	imageURL := "/uploads/" + filename
	if err := h.DB.Model(&cat).Update("image_url", imageURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	cat.ImageURL = imageURL
	c.JSON(http.StatusOK, gin.H{"category": cat})
}
