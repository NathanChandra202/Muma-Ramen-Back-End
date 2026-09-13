package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"muma-ramen-backend/models"
)

type OrderHandler struct {
	DB *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{DB: db}
}

func (h *OrderHandler) generateOrderNumber() string {
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD

	var count int64
	// Count orders created today
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	h.DB.Model(&models.Order{}).Where("created_at >= ?", todayStart).Count(&count)

	return fmt.Sprintf("MRB%s%05d", dateStr, count+1)
}

type CreateOrderRequest struct {
	OrderType     string `json:"order_type" binding:"required"`
	TableNumber   string `json:"table_number"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	PaymentMethod string `json:"payment_method"`
	Notes         string `json:"notes"`
	Items         []struct {
		MenuItemID uint   `json:"menu_item_id" binding:"required"`
		Quantity   int    `json:"quantity" binding:"required,min=1"`
		Notes      string `json:"notes"`
	} `json:"items" binding:"required,min=1"`
}

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	// Validate order type
	if req.OrderType != models.OrderTypeDineIn && req.OrderType != models.OrderTypeTakeaway {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order type"})
		return
	}

	// Build order items and calculate total
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		var menuItem models.MenuItem
		if err := h.DB.First(&menuItem, item.MenuItemID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Menu item %d not found", item.MenuItemID)})
			return
		}

		if !menuItem.IsAvailable {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s sedang tidak tersedia", menuItem.Name)})
			return
		}

		if menuItem.Stock < item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Stok %s tidak cukup (sisa: %d)", menuItem.Name, menuItem.Stock)})
			return
		}

		subtotal := menuItem.Price * float64(item.Quantity)
		totalAmount += subtotal

		orderItems = append(orderItems, models.OrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      menuItem.Price,
			Subtotal:   subtotal,
			Notes:      item.Notes,
		})
	}

	var parsedUserID *uint
	if userID != nil {
		id := userID.(uint)
		parsedUserID = &id
	}

	userRole, _ := c.Get("userRole")
	isPembeli := userRole == nil || userRole.(string) == models.RolePembeli

	if isPembeli {
		if req.CustomerPhone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor telepon wajib diisi"})
			return
		}
		if req.PaymentMethod == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Metode pembayaran wajib dipilih"})
			return
		}
	}

	order := models.Order{
		UserID:        parsedUserID,
		OrderNumber:   h.generateOrderNumber(),
		OrderType:     req.OrderType,
		TableNumber:   req.TableNumber,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		PaymentMethod: req.PaymentMethod,
		Status:        models.OrderStatusPending,
		TotalAmount:   totalAmount,
		Notes:         req.Notes,
		Items:         orderItems,
	}

	// Use transaction to create order and update stock
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Decrease stock
		for _, item := range req.Items {
			if err := tx.Model(&models.MenuItem{}).Where("id = ?", item.MenuItemID).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Reload with associations
	h.DB.Preload("Items.MenuItem").Preload("User").First(&order, order.ID)

	c.JSON(http.StatusCreated, gin.H{"order": order})
}

// ListOrders returns orders based on role
func (h *OrderHandler) ListOrders(c *gin.Context) {
	var orders []models.Order
	query := h.DB.Preload("Items.MenuItem").Preload("User").Order("created_at DESC")

	userRole, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	// Pembeli can only see their own orders
	if userRole.(string) == models.RolePembeli {
		query = query.Where("user_id = ?", userID)
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date
	if date := c.Query("date"); date != "" {
		query = query.Where("DATE(created_at) = ?", date)
	}

	// Filter by order type
	if orderType := c.Query("order_type"); orderType != "" {
		query = query.Where("order_type = ?", orderType)
	}

	// Limit
	limit := 50
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	query = query.Limit(limit)

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetOrder returns a single order
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	var order models.Order
	if err := h.DB.Preload("Items.MenuItem").Preload("User").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Pembeli can only see their own orders
	userRole, roleExists := c.Get("userRole")
	userID, userExists := c.Get("userID")

	// If the order belongs to a user, but the requester is a guest, forbid
	if order.UserID != nil && (!userExists || (roleExists && userRole.(string) == models.RolePembeli && *order.UserID != userID.(uint))) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Note: If order.UserID is nil (guest order), anyone with the ID can view it.
	// This is acceptable since IDs can be considered temporary or we could switch to OrderNumber.

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// UpdateOrderStatus updates the status of an order (kasir+)
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")

	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status transition
	validTransitions := map[string][]string{
		models.OrderStatusPending:   {models.OrderStatusPreparing, models.OrderStatusCancelled},
		models.OrderStatusPreparing: {models.OrderStatusReady, models.OrderStatusCancelled},
		models.OrderStatusReady:     {models.OrderStatusCompleted},
	}

	allowed, exists := validTransitions[order.Status]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order sudah selesai atau dibatalkan"})
		return
	}

	valid := false
	for _, s := range allowed {
		if s == req.Status {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot change from %s to %s", order.Status, req.Status)})
		return
	}

	// If cancelling, restore stock
	if req.Status == models.OrderStatusCancelled {
		var items []models.OrderItem
		h.DB.Where("order_id = ?", order.ID).Find(&items)
		for _, item := range items {
			h.DB.Model(&models.MenuItem{}).Where("id = ?", item.MenuItemID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity))
		}
	}

	order.Status = req.Status
	h.DB.Save(&order)

	h.DB.Preload("Items.MenuItem").Preload("User").First(&order, order.ID)
	c.JSON(http.StatusOK, gin.H{"order": order})
}

// CancelOrder cancels an order
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	userRole, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	// Pembeli can only cancel their own pending orders
	if userRole.(string) == models.RolePembeli {
		if order.UserID == nil || *order.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}
		if order.Status != models.OrderStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya pesanan pending yang bisa dibatalkan"})
			return
		}
	}

	if order.Status == models.OrderStatusCompleted || order.Status == models.OrderStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order sudah selesai atau dibatalkan"})
		return
	}

	// Restore stock
	var items []models.OrderItem
	h.DB.Where("order_id = ?", order.ID).Find(&items)
	for _, item := range items {
		h.DB.Model(&models.MenuItem{}).Where("id = ?", item.MenuItemID).
			UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity))
	}

	order.Status = models.OrderStatusCancelled
	h.DB.Save(&order)

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled", "order": order})
}
