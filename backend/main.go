package main

import (
	"log"
	"os"

	"rtims-backend/config"
	"rtims-backend/internal/database"
	"rtims-backend/internal/handlers"
	"rtims-backend/internal/middleware"
	"rtims-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db := database.InitDB(cfg.DatabaseURL)
	defer db.Close()

	// Initialize Redis client
	redisClient := database.InitRedis(cfg.RedisURL)
	defer redisClient.Close()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/refresh", handlers.RefreshToken)
			auth.POST("/forgot-password", handlers.ForgotPassword)
			auth.POST("/reset-password", handlers.ResetPassword)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth())
		{
			// User routes
			protected.GET("/profile", handlers.GetProfile)
			protected.PUT("/profile", handlers.UpdateProfile)

			// Product routes
			products := protected.Group("/products")
			{
				products.GET("/", handlers.GetProducts)
				products.GET("/:id", handlers.GetProduct)
				products.POST("/", handlers.CreateProduct)
				products.PUT("/:id", handlers.UpdateProduct)
				products.DELETE("/:id", handlers.DeleteProduct)
				products.POST("/:id/stock", handlers.UpdateStock)
			}

			// Stock movement routes
			movements := protected.Group("/stock-movements")
			{
				movements.GET("/", handlers.GetStockMovements)
				movements.GET("/:id", handlers.GetStockMovement)
			}

			// Category routes
			categories := protected.Group("/categories")
			{
				categories.GET("/", handlers.GetCategories)
				categories.POST("/", handlers.CreateCategory)
				categories.PUT("/:id", handlers.UpdateCategory)
				categories.DELETE("/:id", handlers.DeleteCategory)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
				// User management
				admin.GET("/users", handlers.GetUsers)
				admin.POST("/users", handlers.CreateUser)
				admin.PUT("/users/:id", handlers.UpdateUser)
				admin.DELETE("/users/:id", handlers.DeleteUser)

				// Reports
				admin.GET("/reports/inventory", handlers.GenerateInventoryReport)
				admin.GET("/reports/movements", handlers.GenerateMovementReport)

				// System settings
				admin.GET("/settings", handlers.GetSettings)
				admin.PUT("/settings", handlers.UpdateSettings)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("/", handlers.GetNotifications)
				notifications.PUT("/:id/read", handlers.MarkNotificationRead)
			}
		}

		// WebSocket endpoint
		r.GET("/ws", func(c *gin.Context) {
			websocket.ServeWebSocket(wsHub, c, db, redisClient)
		})
	}

	// Swagger documentation
	if cfg.Environment != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Start server
	log.Printf("Server starting on port %s in %s mode", cfg.Port, cfg.Environment)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}