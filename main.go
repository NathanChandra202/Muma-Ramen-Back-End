package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"muma-ramen-backend/config"
	"muma-ramen-backend/handlers"
	"muma-ramen-backend/middleware"
	"muma-ramen-backend/models"
	"muma-ramen-backend/seed"
)

func main() {
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.MenuItem{},
		&models.Order{},
		&models.OrderItem{},
	)

	// Create uploads directory
	os.MkdirAll("uploads", 0755)

	// Seed database
	seed.SeedDatabase(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	menuHandler := handlers.NewMenuHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	orderHandler := handlers.NewOrderHandler(db)
	userHandler := handlers.NewUserHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	// Setup router
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true }, // Allow all origins for local dev
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files
	r.Static("/uploads", "./uploads")
	r.Static("/images", "./images")

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.Me)
		}

		// Menu routes (public read, auth for write)
		menu := api.Group("/menu")
		{
			menu.GET("", menuHandler.ListMenu)
			menu.GET("/:id", menuHandler.GetMenuItem)

			// Protected menu routes
			menuAuth := menu.Group("")
			menuAuth.Use(middleware.AuthMiddleware(cfg))
			{
				menuAuth.POST("", middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin), menuHandler.CreateMenuItem)
				menuAuth.PUT("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin), menuHandler.UpdateMenuItem)
				menuAuth.DELETE("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin), menuHandler.DeleteMenuItem)
				menuAuth.PUT("/:id/stock", middleware.RequireRole(models.RoleKasir, models.RoleAdmin, models.RoleSuperAdmin), menuHandler.UpdateStock)
				menuAuth.PUT("/:id/availability", middleware.RequireRole(models.RoleKasir, models.RoleAdmin, models.RoleSuperAdmin), menuHandler.ToggleAvailability)
				menuAuth.POST("/:id/image", middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin), menuHandler.UploadImage)
			}
		}

		// Category routes
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.ListCategories)

			catAuth := categories.Group("")
			catAuth.Use(middleware.AuthMiddleware(cfg))
			{
				catAuth.POST("", middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin), categoryHandler.CreateCategory)
				catAuth.PUT("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin), categoryHandler.UpdateCategory)
				catAuth.DELETE("/:id", middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin), categoryHandler.DeleteCategory)
			}
		}

		// Order routes
		orders := api.Group("/orders")
		{
			// Optional Auth for creating and viewing a specific order (guest checkout)
			ordersGuest := orders.Group("")
			ordersGuest.Use(middleware.OptionalAuthMiddleware(cfg))
			{
				ordersGuest.POST("", orderHandler.CreateOrder)
				ordersGuest.GET("/:id", orderHandler.GetOrder)
			}

			// Require Auth for management
			ordersAuth := orders.Group("")
			ordersAuth.Use(middleware.AuthMiddleware(cfg))
			{
				ordersAuth.GET("", orderHandler.ListOrders)
				ordersAuth.PUT("/:id/status", middleware.RequireRole(models.RoleKasir, models.RoleAdmin, models.RoleSuperAdmin), orderHandler.UpdateOrderStatus)
				ordersAuth.DELETE("/:id", orderHandler.CancelOrder)
			}
		}

		// User management routes (admin+)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg))
		users.Use(middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin))
		{
			users.GET("", userHandler.ListUsers)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole(models.RoleSuperAdmin), userHandler.DeleteUser)
		}

		// Dashboard routes (admin+)
		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware(cfg))
		dashboard.Use(middleware.RequireRole(models.RoleAdmin, models.RoleSuperAdmin))
		{
			dashboard.GET("/stats", dashboardHandler.GetStats)
		}
	}

	log.Printf("🍜 Muma Ramen Backend running on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
