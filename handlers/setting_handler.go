package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"muma-ramen-backend/models"
)

type SettingHandler struct {
	DB *gorm.DB
}

func NewSettingHandler(db *gorm.DB) *SettingHandler {
	return &SettingHandler{DB: db}
}

// GetSettings returns all settings as a key-value map
func (h *SettingHandler) GetSettings(c *gin.Context) {
	var settings []models.Setting
	if err := h.DB.Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	result := make(map[string]string)
	
	// Default values
	result["store_name"] = "Muma Cibubur"
	result["store_hours"] = "Buka • 10:00 - 22:00"

	for _, s := range settings {
		result[s.Key] = s.Value
	}

	c.JSON(http.StatusOK, result)
}

// UpdateSettings updates multiple settings from a JSON object
func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	for k, v := range req {
		setting := models.Setting{Key: k, Value: v}
		if err := h.DB.Save(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting: " + k})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// UploadSettingImage uploads an image for a specific setting key (e.g. promo_image)
func (h *SettingHandler) UploadSettingImage(c *gin.Context) {
	key := c.PostForm("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}
	defer file.Close()

	// Ensure directory exists
	uploadDir := "uploads/settings"
	os.MkdirAll(uploadDir, 0755)

	// Create unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s_%d%s", key, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write image"})
		return
	}

	// Update setting
	imageURL := "/" + filepath.ToSlash(filePath)
	setting := models.Setting{Key: key, Value: imageURL}
	if err := h.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image uploaded successfully",
		"url":     imageURL,
	})
}
