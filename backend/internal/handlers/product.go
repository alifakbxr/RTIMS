package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"rtims-backend/internal/database"
	"rtims-backend/internal/models"
	"rtims-backend/internal/middleware"
	"rtims-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-redis/redis/v8"
)

type ProductHandler struct {
	productService      *database.ProductService
	auditService        *database.AuditService
	notificationService *database.NotificationService
	db                  *sql.DB
	redisClient         *redis.Client
	hub                 *websocket.Hub
}

func NewProductHandler(db *sql.DB, redisClient *redis.Client, hub *websocket.Hub) *ProductHandler {
	return &ProductHandler{
		productService:      database.NewProductService(db),
		auditService:        database.NewAuditService(db),
		notificationService: database.NewNotificationService(db),
		db:                  db,
		redisClient:         redisClient,
		hub:                 hub,
	}
}

// Helper function to create audit log
func (h *ProductHandler) createAuditLog(c *gin.Context, recordID uuid.UUID, action models.AuditAction, oldValues, newValues map[string]interface{}) {
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		log.Printf("Failed to get user for audit log: %v", err)
		return
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "products",
		RecordID:   recordID,
		Action:     action,
		OldValues:  oldValues,
		NewValues:  newValues,
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
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

	// Get products from database
	products, total, err := h.productService.GetProducts(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products: " + err.Error()})
		return
	}

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

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.productService.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
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

	product := &models.Product{
		ID:               uuid.New(),
		Name:             req.Name,
		SKU:              req.SKU,
		Stock:            req.Stock,
		Price:            req.Price,
		Category:         req.Category,
		MinimumThreshold: req.MinimumThreshold,
		SupplierInfo:     req.SupplierInfo,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save product to database
	err = h.productService.CreateProduct(product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
		return
	}

	// Create audit log
	h.createAuditLog(c, product.ID, models.ActionCreate, nil, map[string]interface{}{
		"name":              req.Name,
		"sku":               req.SKU,
		"stock":             req.Stock,
		"price":             req.Price,
		"category":          req.Category,
		"minimum_threshold": req.MinimumThreshold,
		"supplier_info":     req.SupplierInfo,
	})

	// Create stock movement if initial stock is provided
	if req.Stock > 0 {
		err = h.productService.UpdateProductStock(product.ID, req.Stock, models.ReasonPurchase, userID, "Initial stock")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create initial stock movement: " + err.Error()})
			return
		}

		// Send WebSocket notification if stock is low
		if req.Stock <= req.MinimumThreshold {
			websocket.BroadcastStockUpdate(h.hub, product.ID, req.Stock)
		}
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
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

	_, _, err = middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.SKU != nil {
		updates["sku"] = *req.SKU
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.MinimumThreshold != nil {
		updates["minimum_threshold"] = *req.MinimumThreshold
	}
	if req.SupplierInfo != nil {
		updates["supplier_info"] = *req.SupplierInfo
	}

	// Get old product for audit logging
	oldProduct, err := h.productService.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current product: " + err.Error()})
		return
	}

	// Update product in database
	err = h.productService.UpdateProduct(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product: " + err.Error()})
		return
	}

	// Get updated product
	product, err := h.productService.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated product: " + err.Error()})
		return
	}

	// Create audit log
	h.createAuditLog(c, id, models.ActionUpdate, map[string]interface{}{
		"name":              oldProduct.Name,
		"sku":               oldProduct.SKU,
		"stock":             oldProduct.Stock,
		"price":             oldProduct.Price,
		"category":          oldProduct.Category,
		"minimum_threshold": oldProduct.MinimumThreshold,
		"supplier_info":     oldProduct.SupplierInfo,
	}, map[string]interface{}{
		"name":              product.Name,
		"sku":               product.SKU,
		"stock":             product.Stock,
		"price":             product.Price,
		"category":          product.Category,
		"minimum_threshold": product.MinimumThreshold,
		"supplier_info":     product.SupplierInfo,
	})

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	_, _, err = middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get product for audit logging before deletion
	product, err := h.productService.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product: " + err.Error()})
		return
	}

	// Delete product from database
	err = h.productService.DeleteProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product: " + err.Error()})
		return
	}

	// Create audit log
	h.createAuditLog(c, id, models.ActionDelete, map[string]interface{}{
		"name":              product.Name,
		"sku":               product.SKU,
		"stock":             product.Stock,
		"price":             product.Price,
		"category":          product.Category,
		"minimum_threshold": product.MinimumThreshold,
		"supplier_info":     product.SupplierInfo,
	}, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *ProductHandler) UpdateStock(c *gin.Context) {
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

	// Get current product for audit logging
	product, err := h.productService.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product: " + err.Error()})
		return
	}

	oldStock := product.Stock

	// Update product stock in database
	err = h.productService.UpdateProductStock(id, req.Change, req.Reason, userID, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock: " + err.Error()})
		return
	}

	// Get updated product
	updatedProduct, err := h.productService.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated product: " + err.Error()})
		return
	}

	// Create audit log
	h.createAuditLog(c, id, models.ActionUpdate, map[string]interface{}{
		"stock": oldStock,
	}, map[string]interface{}{
		"stock": updatedProduct.Stock,
	})

	// Send WebSocket notification
	websocket.BroadcastStockUpdate(h.hub, id, updatedProduct.Stock)

	// Create notification if stock is low
	if updatedProduct.Stock <= updatedProduct.MinimumThreshold && updatedProduct.MinimumThreshold > 0 {
		notification := &models.Notification{
			ID:        uuid.New(),
			UserID:    userID,
			Message:   fmt.Sprintf("Product '%s' stock is low (%d remaining)", updatedProduct.Name, updatedProduct.Stock),
			Type:      models.NotificationLowStock,
			IsRead:    false,
			CreatedAt: time.Now(),
		}

		// Save notification to database
		err = h.notificationService.CreateNotification(notification)
		if err != nil {
			log.Printf("Failed to create low stock notification: %v", err)
		} else {
			// Send WebSocket notification for low stock
			websocket.BroadcastNotification(h.hub, userID, notification.Message, string(notification.Type))
		}
	}

	stockMovement := models.StockMovement{
		ID:        uuid.New(),
		ProductID: id,
		Change:    req.Change,
		Reason:    req.Reason,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		Notes:     req.Notes,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Stock updated successfully",
		"stock_movement": stockMovement,
	})
}

func (h *ProductHandler) GetStockMovements(c *gin.Context) {
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

	// Get stock movements from database
	movements, total, err := h.productService.GetStockMovements(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stock movements: " + err.Error()})
		return
	}

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

func (h *ProductHandler) GetStockMovement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movement ID"})
		return
	}

	movement, err := h.productService.GetStockMovement(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock movement not found"})
		return
	}

	c.JSON(http.StatusOK, movement)
}