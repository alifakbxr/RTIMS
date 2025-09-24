package handlers

import (
	"net/http"
	"time"

	"rtims-backend/internal/models"
	"rtims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var now = time.Now()

func GetNotifications(c *gin.Context) {
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters
	var filter models.NotificationFilter
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

	// TODO: Get notifications from database with filters
	// This would be implemented when we create the database service

	// Mock data for demonstration
	notifications := []models.Notification{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Message:   "Product 'Sample Product' stock is low (5 remaining)",
			Type:      models.NotificationLowStock,
			IsRead:    false,
			CreatedAt: now,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Message:   "System maintenance scheduled for tonight",
			Type:      models.NotificationSystem,
			IsRead:    true,
			CreatedAt: now.Add(-time.Hour),
		},
	}

	total := len(notifications)

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"pagination": gin.H{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
			"pages": (total + filter.Limit - 1) / filter.Limit,
		},
	})
}

func MarkNotificationRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Mark notification as read in database
	// This would be implemented when we create the database service

	// TODO: Create audit log
	// This would be implemented when we create the audit service

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
		"id":      id,
		"user_id": userID,
	})
}

func CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Save notification to database
	// This would be implemented when we create the database service

	notification := models.Notification{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Message:   req.Message,
		Type:      req.Type,
		IsRead:    false,
		CreatedAt: now,
	}

	// TODO: Create audit log
	// This would be implemented when we create the audit service

	// TODO: Send WebSocket notification
	// This would be implemented when we integrate WebSocket

	c.JSON(http.StatusCreated, notification)
}

func GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	var filter models.AuditLogFilter
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

	// TODO: Get audit logs from database with filters
	// This would be implemented when we create the database service

	// Mock data for demonstration
	auditLogs := []models.AuditLog{
		{
			ID:         uuid.New(),
			TableName:  "products",
			RecordID:   uuid.New(),
			Action:     models.ActionCreate,
			OldValues:  nil,
			NewValues:  gin.H{"name": "Sample Product", "sku": "SP001"},
			ChangedBy:  uuid.New(),
			ChangedAt:  now,
			IPAddress:  "192.168.1.1",
			UserAgent:  "Mozilla/5.0",
		},
		{
			ID:         uuid.New(),
			TableName:  "users",
			RecordID:   uuid.New(),
			Action:     models.ActionLogin,
			OldValues:  nil,
			NewValues:  nil,
			ChangedBy:  uuid.New(),
			ChangedAt:  now.Add(-time.Hour),
			IPAddress:  "192.168.1.1",
			UserAgent:  "Mozilla/5.0",
		},
	}

	total := len(auditLogs)

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": auditLogs,
		"pagination": gin.H{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
			"pages": (total + filter.Limit - 1) / filter.Limit,
		},
	})
}

func GetAuditLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit log ID"})
		return
	}

	// TODO: Get audit log from database
	// This would be implemented when we create the database service

	auditLog := models.AuditLog{
		ID:         id,
		TableName:  "products",
		RecordID:   uuid.New(),
		Action:     models.ActionCreate,
		OldValues:  nil,
		NewValues:  gin.H{"name": "Sample Product", "sku": "SP001"},
		ChangedBy:  uuid.New(),
		ChangedAt:  now,
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
	}

	c.JSON(http.StatusOK, auditLog)
}