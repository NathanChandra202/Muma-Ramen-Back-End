package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"muma-ramen-backend/models"
)

type MenuHandler struct {
	DB *gorm.DB
}

func NewMenuHandler(db *gorm.DB) *MenuHandler {
	return &MenuHandler{DB: db}
}

// ListMenu returns all menu items (public)
func (h *MenuHandler) ListMenu(c *gin.Context) {
	var items []models.MenuItem

	query := h.DB.Preload("Category")

	// Filter by category
	if catID := c.Query("category_id"); catID != "" {
		query = query.Where("category_id = ?", catID)
	}

	// Filter available only
	if avail := c.Query("available"); avail == "true" {
		query = query.Where("is_available = ?", true)
	}

	// Search
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Order("category_id, name").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"menu": items})
}

// GetMenuItem returns a single menu item
func (h *MenuHandler) GetMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := h.DB.Preload("Category").First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"menu_item": item})
}

// CreateMenuItem creates a new menu item (admin+)
func (h *MenuHandler) CreateMenuItem(c *gin.Context) {
	var item models.MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify category exists
	var cat models.Category
	if err := h.DB.First(&cat, item.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found"})
		return
	}

	if err := h.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu item"})
		return
	}

	h.DB.Preload("Category").First(&item, item.ID)
	c.JSON(http.StatusCreated, gin.H{"menu_item": item})
}

// UpdateMenuItem updates a menu item (admin+)
func (h *MenuHandler) UpdateMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := h.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&item).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu item"})
		return
	}

	h.DB.Preload("Category").First(&item, id)
	c.JSON(http.StatusOK, gin.H{"menu_item": item})
}

// DeleteMenuItem deletes a menu item (admin+)
func (h *MenuHandler) DeleteMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := h.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	if err := h.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu item deleted"})
}

// UpdateStock updates stock for a menu item (kasir+)
func (h *MenuHandler) UpdateStock(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := h.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	var req struct {
		Stock int `json:"stock"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.Stock = req.Stock
	if req.Stock <= 0 {
		item.IsAvailable = false
	}
	h.DB.Save(&item)

	c.JSON(http.StatusOK, gin.H{"menu_item": item})
}

// ToggleAvailability toggles availability of a menu item (kasir+)
func (h *MenuHandler) ToggleAvailability(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := h.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	var req struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.IsAvailable = req.IsAvailable
	h.DB.Save(&item)

	c.JSON(http.StatusOK, gin.H{"menu_item": item})
}

// UploadImage handles image upload for menu items
func (h *MenuHandler) UploadImage(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := h.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image files required"})
		return
	}

	var urls []string
	if item.Images != "" {
		json.Unmarshal([]byte(item.Images), &urls)
	}
	
	for i, file := range files {
		filename := "menu_" + id + "_" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + strconv.Itoa(i) + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		urls = append(urls, "/uploads/"+filename)
	}

	if len(urls) > 0 {
		item.ImageURL = urls[0]
		bytes, _ := json.Marshal(urls)
		item.Images = string(bytes)
	}
	
	h.DB.Save(&item)

	c.JSON(http.StatusOK, gin.H{"menu_item": item})
}
