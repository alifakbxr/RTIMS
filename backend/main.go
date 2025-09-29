package main

import (
	"log"
	"net/http"

	"rtims-backend/config"
	"rtims-backend/internal/database"
	"rtims-backend/internal/handlers"
	"rtims-backend/internal/middleware"
	"rtims-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize JWT secret with logging
		log.Printf("Initializing JWT secret...")
		if cfg.JWTSecret == "" {
			log.Fatal("JWT_SECRET is not set in environment variables")
		}
		if len(cfg.JWTSecret) < 32 {
			log.Printf("Warning: JWT_SECRET is shorter than recommended (32 characters). Current length: %d", len(cfg.JWTSecret))
		}
		middleware.InitJWTSecret(cfg)
		log.Printf("JWT secret initialized successfully (length: %d characters)", len(cfg.JWTSecret))

	// Database and Redis are already initialized above

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize database with enhanced validation
		log.Println("Initializing database connection...")
		db := database.InitDB(cfg.DatabaseURL)
		defer db.Close()

		// Validate database connection and required tables
		if err := database.ValidateDatabaseConnection(db); err != nil {
			log.Fatal("Database validation failed:", err)
		}
		log.Println("Database connection validated successfully")

		// Initialize Redis client with enhanced validation
		log.Println("Initializing Redis connection...")
		redisClient := database.InitRedis(cfg.RedisURL)
		defer redisClient.Close()

		// Validate Redis connection
		if err := database.ValidateRedisConnection(redisClient); err != nil {
			log.Fatal("Redis validation failed:", err)
		}
		log.Println("Redis connection validated successfully")

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
	r.Use(middleware.RateLimit())

	// Initialize audit middleware with database
	auditMiddleware := middleware.NewAuditMiddleware(db)

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Initialize auth handlers
		handlers.InitAuthHandlers([]byte(cfg.JWTSecret), db, redisClient)

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
			protected.Use(auditMiddleware.AuditLog())
			{
				// Test endpoint for JWT middleware verification
				protected.GET("/test-auth", func(c *gin.Context) {
					userID, exists := c.Get("user_id")
					if !exists {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
						return
					}

					email, exists := c.Get("email")
					if !exists {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found in context"})
						return
					}

					role, exists := c.Get("role")
					if !exists {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
						return
					}

					c.JSON(http.StatusOK, gin.H{
						"message": "JWT authentication successful",
						"user_id": userID,
						"email":   email,
						"role":    role,
						"path":    c.Request.URL.Path,
						"method":  c.Request.Method,
					})
				})

				// User routes
				protected.GET("/profile", handlers.GetProfile)
				protected.PUT("/profile", handlers.UpdateProfile)

			// Initialize product handler
			productHandler := handlers.NewProductHandler(db, redisClient, wsHub)

			// Initialize notification handler
			notificationHandler := handlers.NewNotificationHandler(db, wsHub)

			// Initialize admin handler
			adminHandler := handlers.NewAdminHandler(db)

			// Dashboard routes
			protected.GET("/dashboard/stats", adminHandler.GetDashboardStats)
			protected.GET("/dashboard/alerts", adminHandler.GetDashboardAlerts)

			// Product routes
			products := protected.Group("/products")
			{
				products.GET("/", productHandler.GetProducts)
				products.GET("/:id", productHandler.GetProduct)
				products.POST("/", productHandler.CreateProduct)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.DELETE("/:id", productHandler.DeleteProduct)
				products.POST("/:id/stock", productHandler.UpdateStock)
			}

			// Stock movement routes
			movements := protected.Group("/stock-movements")
			{
				movements.GET("/", productHandler.GetStockMovements)
				movements.GET("/:id", productHandler.GetStockMovement)
			}

			// Category routes
			categories := protected.Group("/categories")
			{
				categories.GET("/", adminHandler.GetCategories)
				categories.POST("/", adminHandler.CreateCategory)
				categories.PUT("/:id", adminHandler.UpdateCategory)
				categories.DELETE("/:id", adminHandler.DeleteCategory)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
				// User management
				admin.GET("/users", adminHandler.GetUsers)
				admin.POST("/users", adminHandler.CreateUser)
				admin.PUT("/users/:id", adminHandler.UpdateUser)
				admin.DELETE("/users/:id", adminHandler.DeleteUser)

				// Category management
				admin.GET("/categories", adminHandler.GetCategories)
				admin.POST("/categories", adminHandler.CreateCategory)
				admin.PUT("/categories/:id", adminHandler.UpdateCategory)
				admin.DELETE("/categories/:id", adminHandler.DeleteCategory)

				// Reports
				admin.GET("/reports/stats", adminHandler.GetReportStats)
				admin.GET("/reports/types", adminHandler.GetReportTypes)
				admin.GET("/reports/recent", adminHandler.GetRecentReports)
				admin.GET("/reports/inventory", adminHandler.GenerateReport)
				admin.GET("/reports/movements", adminHandler.GenerateReport)
				admin.GET("/reports/users", adminHandler.GenerateReport)
				admin.GET("/reports/financial", adminHandler.GenerateReport)
				admin.GET("/reports/:type", adminHandler.GenerateReport)

				// System settings
				admin.GET("/settings", adminHandler.GetSettings)
				admin.PUT("/settings", adminHandler.UpdateSettings)
				admin.GET("/settings/status", adminHandler.GetSystemStatus)
				admin.POST("/settings/backup", adminHandler.TriggerBackup)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("/", notificationHandler.GetNotifications)
				notifications.PUT("/:id/read", notificationHandler.MarkNotificationRead)
				notifications.POST("/", notificationHandler.CreateNotification)
			}

			// Audit log routes
			auditLogs := protected.Group("/audit-logs")
			{
				auditLogs.GET("/", notificationHandler.GetAuditLogs)
				auditLogs.GET("/:id", notificationHandler.GetAuditLog)
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