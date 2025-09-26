package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"rtims-backend/internal/database"
	"rtims-backend/internal/models"
	"rtims-backend/internal/middleware"
	"rtims-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationService *database.NotificationService
	auditService        *database.AuditService
	db                  *sql.DB
	hub                 *websocket.Hub
}

func NewNotificationHandler(db *sql.DB, hub *websocket.Hub) *NotificationHandler {
	return &NotificationHandler{
		notificationService: database.NewNotificationService(db),
		auditService:        database.NewAuditService(db),
		db:                  db,
		hub:                 hub,
	}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
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

	filter.UserID = &userID

	// Get notifications from database
	notifications, total, err := h.notificationService.GetNotifications(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications: " + err.Error()})
		return
	}

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

func (h *NotificationHandler) MarkNotificationRead(c *gin.Context) {
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

	// Mark notification as read in database
	err = h.notificationService.MarkAsRead(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "notifications",
		RecordID:   id,
		Action:     models.ActionUpdate,
		OldValues:  gin.H{"is_read": false},
		NewValues:  gin.H{"is_read": true},
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
		"id":      id,
		"user_id": userID,
	})
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create notification object
	notification := &models.Notification{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Message:   req.Message,
		Type:      req.Type,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// Save notification to database
	err = h.notificationService.CreateNotification(notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "notifications",
		RecordID:   notification.ID,
		Action:     models.ActionCreate,
		OldValues:  nil,
		NewValues:  gin.H{"user_id": req.UserID, "message": req.Message, "type": req.Type},
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create audit log: %v", err)
	}

	// Send WebSocket notification
	websocket.BroadcastNotification(h.hub, req.UserID, req.Message, string(req.Type))

	c.JSON(http.StatusCreated, notification)
}

func (h *NotificationHandler) GetAuditLogs(c *gin.Context) {
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

	// Get audit logs from database
	auditLogs, total, err := h.auditService.GetAuditLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit logs: " + err.Error()})
		return
	}

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

func (h *NotificationHandler) GetAuditLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit log ID"})
		return
	}

	auditLog, err := h.auditService.GetAuditLog(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit log not found"})
		return
	}

	c.JSON(http.StatusOK, auditLog)
}