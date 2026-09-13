package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"muma-ramen-backend/models"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// ListUsers returns all users (admin+)
func (h *UserHandler) ListUsers(c *gin.Context) {
	var users []models.User
	query := h.DB.Order("created_at DESC")

	// Filter by role
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Map to safe response (exclude password hash)
	var result []gin.H
	for _, u := range users {
		result = append(result, gin.H{
			"id":         u.ID,
			"name":       u.Name,
			"email":      u.Email,
			"role":       u.Role,
			"created_at": u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": result})
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

// CreateUser creates a new user with a specific role (admin+)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	validRoles := map[string]bool{
		models.RolePembeli: true,
		models.RoleKasir:   true,
		models.RoleAdmin:   true,
	}

	currentRole, _ := c.Get("userRole")

	// Only superadmin can create admin users
	if req.Role == models.RoleAdmin && currentRole.(string) != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya super admin yang bisa membuat akun admin"})
		return
	}

	// Cannot create superadmin
	if req.Role == models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tidak bisa membuat akun super admin"})
		return
	}

	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Check if email already exists
	var existing models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Email, // using email as default or need to add Phone to CreateUserRequest
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// UpdateUser updates a user (admin+)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cannot modify superadmin
	if user.Role == models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify super admin"})
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Role != "" && req.Role != models.RoleSuperAdmin {
		user.Role = req.Role
	}

	h.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		},
	})
}

// UpdateProfile updates the currently authenticated user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		},
	})
}

// DeleteUser deletes a user (superadmin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Role == models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete super admin"})
		return
	}

	if err := h.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
