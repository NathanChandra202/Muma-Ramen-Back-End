package models

import (
	"time"

	"gorm.io/gorm"
)

// Role constants
const (
	RolePembeli    = "pembeli"
	RoleKasir      = "kasir"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "superadmin"
)

// Order status constants
const (
	OrderStatusPending   = "pending"
	OrderStatusPreparing = "preparing"
	OrderStatusReady     = "ready"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)

// Order type constants
const (
	OrderTypeDineIn   = "dine_in"
	OrderTypeTakeaway = "takeaway"
)

// User model
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Email        string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Role         string         `gorm:"size:20;not null;default:'pembeli'" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category model
type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	MenuItems   []MenuItem     `gorm:"foreignKey:CategoryID" json:"menu_items,omitempty"`
}

// MenuItem model
type MenuItem struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CategoryID  uint           `gorm:"not null" json:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Price       float64        `gorm:"not null" json:"price"`
	ImageURL    string         `gorm:"size:500" json:"image_url"`
	IsAvailable bool           `gorm:"default:true" json:"is_available"`
	Stock       int            `gorm:"default:100" json:"stock"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Order model
type Order struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	OrderNumber string         `gorm:"size:20;uniqueIndex;not null" json:"order_number"`
	OrderType   string         `gorm:"size:20;not null;default:'dine_in'" json:"order_type"`
	TableNumber string         `gorm:"size:10" json:"table_number"`
	Status      string         `gorm:"size:20;not null;default:'pending'" json:"status"`
	TotalAmount float64        `gorm:"not null;default:0" json:"total_amount"`
	Notes       string         `gorm:"size:500" json:"notes"`
	Items       []OrderItem    `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// OrderItem model
type OrderItem struct {
	ID         uint     `gorm:"primarykey" json:"id"`
	OrderID    uint     `gorm:"not null" json:"order_id"`
	MenuItemID uint     `gorm:"not null" json:"menu_item_id"`
	MenuItem   MenuItem `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
	Quantity   int      `gorm:"not null;default:1" json:"quantity"`
	Price      float64  `gorm:"not null" json:"price"`
	Subtotal   float64  `gorm:"not null" json:"subtotal"`
	Notes      string   `gorm:"size:255" json:"notes"`
}
