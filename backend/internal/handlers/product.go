package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"rtims-backend/internal/models"
	"rtims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var now = time.Now()

func GetProducts(c *gin.Context) {
	// Parse query parameters
	var filter models.ProductFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	offset := (filter.Page - 1) * filter.Limit

	// TODO: Get products from database with filters
	// This would be implemented when we create the database service

	// Mock data for demonstration
	products := []models.Product{
		{
			ID:               uuid.New(),
			Name:             "Sample Product 1",
			SKU:              "SP001",
			Stock:            50,
			Price:            29.99,
			Category:         "Electronics",
			MinimumThreshold: 10,
			SupplierInfo:     "Supplier A",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               uuid.New(),
			Name:             "Sample Product 2",
			SKU:              "SP002",
			Stock:            5,
			Price:            19.99,
			Category:         "Clothing",
			MinimumThreshold: 15,
			SupplierInfo:     "Supplier B",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	total := len(products)

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"pagination": gin.H{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
			"pages": (total + filter.Limit - 1) / filter.Limit,
		},
	})
}

func GetProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// TODO: Get product from database
	// This would be implemented when we create the database service

	// Mock data for demonstration
	product := models.Product{
		ID:               id,
		Name:             "Sample Product",
		SKU:              "SP001",
		Stock:            50,
		Price:            29.99,
		Category:         "Electronics",
		MinimumThreshold: 10,
		SupplierInfo:     "Supplier A",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	c.JSON(http.StatusOK, product)
}

func CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Save product to database
	// This would be implemented when we create the database service

	product := models.Product{
		ID:               uuid.New(),
		Name:             req.Name,
		SKU:              req.SKU,
		Stock:            req.Stock,
		Price:            req.Price,
		Category:         req.Category,
		MinimumThreshold: req.MinimumThreshold,
		SupplierInfo:     req.SupplierInfo,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// TODO: Create audit log
	// This would be implemented when we create the audit service

	// TODO: Send WebSocket notification if stock is low
	// This would be implemented when we integrate WebSocket

	c.JSON(http.StatusCreated, product)
}

func UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Get existing product from database
	// This would be implemented when we create the database service

	// TODO: Update product in database
	// This would be implemented when we create the database service

	product := models.Product{
		ID:               id,
		Name:             "Updated Product",
		SKU:              "SP001",
		Stock:            50,
		Price:            29.99,
		Category:         "Electronics",
		MinimumThreshold: 10,
		SupplierInfo:     "Supplier A",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// TODO: Create audit log
	// This would be implemented when we create the audit service

	c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Delete product from database
	// This would be implemented when we create the database service

	// TODO: Create audit log
	// This would be implemented when we create the audit service

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func UpdateStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req models.CreateStockMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Update product stock in database
	// This would be implemented when we create the database service

	// TODO: Create stock movement record
	// This would be implemented when we create the database service

	// TODO: Create audit log
	// This would be implemented when we create the audit service

	// TODO: Send WebSocket notification
	// This would be implemented when we integrate WebSocket

	// TODO: Create notification if stock is low
	// This would be implemented when we create the notification service

	stockMovement := models.StockMovement{
		ID:        uuid.New(),
		ProductID: id,
		Change:    req.Change,
		Reason:    req.Reason,
		CreatedBy: userID,
		CreatedAt: now,
		Notes:     req.Notes,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Stock updated successfully",
		"stock_movement": stockMovement,
	})
}

func GetStockMovements(c *gin.Context) {
	// Parse query parameters
	var filter models.StockMovementFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	offset := (filter.Page - 1) * filter.Limit

	// TODO: Get stock movements from database with filters
	// This would be implemented when we create the database service

	// Mock data for demonstration
	movements := []models.StockMovement{
		{
			ID:        uuid.New(),
			ProductID: uuid.New(),
			Change:    10,
			Reason:    models.ReasonPurchase,
			CreatedBy: uuid.New(),
			CreatedAt: now,
			Notes:     "Initial stock",
		},
	}

	total := len(movements)

	c.JSON(http.StatusOK, gin.H{
		"movements": movements,
		"pagination": gin.H{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
			"pages": (total + filter.Limit - 1) / filter.Limit,
		},
	})
}

func GetStockMovement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movement ID"})
		return
	}

	// TODO: Get stock movement from database
	// This would be implemented when we create the database service

	movement := models.StockMovement{
		ID:        id,
		ProductID: uuid.New(),
		Change:    10,
		Reason:    models.ReasonPurchase,
		CreatedBy: uuid.New(),
		CreatedAt: now,
		Notes:     "Sample movement",
	}

	c.JSON(http.StatusOK, movement)
}