package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"muma-ramen-backend/models"
)

type DashboardHandler struct {
	DB *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{DB: db}
}

// GetStats returns dashboard statistics (admin+)
func (h *DashboardHandler) GetStats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	// Today's orders count
	var todayOrdersCount int64
	h.DB.Model(&models.Order{}).Where("DATE(created_at) = ?", today).Count(&todayOrdersCount)

	// Today's revenue
	var todayRevenue float64
	h.DB.Model(&models.Order{}).
		Where("DATE(created_at) = ? AND status != ?", today, models.OrderStatusCancelled).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&todayRevenue)

	// Active orders (pending + preparing + ready)
	var activeOrders int64
	h.DB.Model(&models.Order{}).
		Where("status IN ?", []string{models.OrderStatusPending, models.OrderStatusPreparing, models.OrderStatusReady}).
		Count(&activeOrders)

	// Completed orders today
	var completedToday int64
	h.DB.Model(&models.Order{}).
		Where("DATE(created_at) = ? AND status = ?", today, models.OrderStatusCompleted).
		Count(&completedToday)

	// Low stock items (stock < 10)
	var lowStockCount int64
	h.DB.Model(&models.MenuItem{}).Where("stock < 10 AND stock > 0").Count(&lowStockCount)

	// Out of stock items
	var outOfStockCount int64
	h.DB.Model(&models.MenuItem{}).Where("stock <= 0 OR is_available = ?", false).Count(&outOfStockCount)

	// Total menu items
	var totalMenuItems int64
	h.DB.Model(&models.MenuItem{}).Count(&totalMenuItems)

	// Total users by role
	type RoleCount struct {
		Role  string
		Count int64
	}
	var roleCounts []RoleCount
	h.DB.Model(&models.User{}).Select("role, count(*) as count").Group("role").Scan(&roleCounts)

	// Recent orders (last 10)
	var recentOrders []models.Order
	h.DB.Preload("Items.MenuItem").Preload("User").
		Order("created_at DESC").Limit(10).Find(&recentOrders)

	// Popular items (top 5 by order count)
	type PopularItem struct {
		MenuItemID uint    `json:"menu_item_id"`
		Name       string  `json:"name"`
		TotalQty   int     `json:"total_qty"`
		TotalSales float64 `json:"total_sales"`
	}
	var popularItems []PopularItem
	h.DB.Table("order_items").
		Select("order_items.menu_item_id, menu_items.name, SUM(order_items.quantity) as total_qty, SUM(order_items.subtotal) as total_sales").
		Joins("JOIN menu_items ON menu_items.id = order_items.menu_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.status != ? AND DATE(orders.created_at) = ?", models.OrderStatusCancelled, today).
		Group("order_items.menu_item_id").
		Order("total_qty DESC").
		Limit(5).
		Scan(&popularItems)

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"today_orders":    todayOrdersCount,
			"today_revenue":   todayRevenue,
			"active_orders":   activeOrders,
			"completed_today": completedToday,
			"low_stock":       lowStockCount,
			"out_of_stock":    outOfStockCount,
			"total_menu":      totalMenuItems,
			"users_by_role":   roleCounts,
			"recent_orders":   recentOrders,
			"popular_items":   popularItems,
		},
	})
}
