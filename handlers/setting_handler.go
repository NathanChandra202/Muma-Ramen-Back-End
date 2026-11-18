package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

// hexColorRegex validates #RRGGBB format
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// isValidHexColor returns true if s is a valid #RRGGBB hex color
func isValidHexColor(s string) bool {
	return hexColorRegex.MatchString(s)
}

// GetSettings returns all settings as a key-value map (public, no auth required)
func (h *SettingHandler) GetSettings(c *gin.Context) {
	var settings []models.Setting
	if err := h.DB.Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	// Default values — includes new branding fields
	result := map[string]string{
		"store_name":      "Muma Cibubur",
		"store_hours":     "Buka • 10:00 - 22:00",
		"tax_rate":        "10",
		"store_logo":      "",
		"primary_color":   "",
		"secondary_color": "",
	}

	for _, s := range settings {
		result[s.Key] = s.Value
	}

	c.JSON(http.StatusOK, result)
}

// UpdateSettings updates settings — only accessible by superadmin.
// Validates hex color format for primary_color and secondary_color.
func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate hex color fields
	colorFields := []string{"primary_color", "secondary_color"}
	for _, field := range colorFields {
		if val, ok := req[field]; ok && val != "" {
			if !isValidHexColor(val) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Format %s tidak valid, harus #RRGGBB (contoh: #F59E0B)", field),
				})
				return
			}
		}
	}

	for k, v := range req {
		setting := models.Setting{Key: k, Value: v}
		if err := h.DB.Save(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting: " + k})
			return
		}
	}

	// Return updated settings
	var allSettings []models.Setting
	h.DB.Find(&allSettings)
	result := map[string]string{
		"store_name":      "Muma Cibubur",
		"store_hours":     "Buka • 10:00 - 22:00",
		"tax_rate":        "10",
		"store_logo":      "",
		"primary_color":   "",
		"secondary_color": "",
	}
	for _, s := range allSettings {
		result[s.Key] = s.Value
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Settings updated successfully",
		"settings": result,
	})
}

// UploadLogo handles store_logo upload.
// Supports both multipart file upload AND plain URL string input.
//
// Multipart: POST with field "file" (image file)
// URL input:  POST with field "url" (string)
func (h *SettingHandler) UploadLogo(c *gin.Context) {
	// Check if a URL string was provided directly
	if logoURL := c.PostForm("url"); logoURL != "" {
		setting := models.Setting{Key: "store_logo", Value: logoURL}
		if err := h.DB.Save(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save logo URL"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":    "Logo URL updated successfully",
			"store_logo": logoURL,
		})
		return
	}

	// Otherwise handle file upload
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provide either 'file' (image) or 'url' (string)"})
		return
	}
	defer file.Close()

	// Validate extension
	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".svg": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung. Gunakan jpg, png, svg, atau webp"})
		return
	}

	// Ensure upload directory exists
	uploadDir := "uploads/settings"
	os.MkdirAll(uploadDir, 0755)

	// Unique filename
	filename := fmt.Sprintf("store_logo_%d%s", time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write image file"})
		return
	}

	logoURL := "/api/uploads/settings/" + filename
	setting := models.Setting{Key: "store_logo", Value: logoURL}
	if err := h.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save logo in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Logo uploaded successfully",
		"store_logo": logoURL,
	})
}

// UploadSettingImage uploads an image for a generic setting key (kept for backward compatibility)
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

	uploadDir := "uploads/settings"
	os.MkdirAll(uploadDir, 0755)

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s_%d%s", key, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

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
