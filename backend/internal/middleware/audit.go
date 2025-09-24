package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"rtims-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit logging for health checks and swagger docs
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/swagger" {
			c.Next()
			return
		}

		// Get current user info
		userID, role, err := GetCurrentUser(c)
		if err != nil {
			// For unauthenticated requests, use system user
			userID = uuid.Nil
			role = models.RoleStaff
		}

		// Capture request body for create/update operations
		var requestBody map[string]interface{}
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				json.Unmarshal(bodyBytes, &requestBody)
				// Restore the request body
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Process the request
		c.Next()

		// Log the action
		go func() {
			auditLog := models.CreateAuditLogRequest{
				TableName: extractTableName(c.Request.URL.Path),
				RecordID:  extractRecordID(c.Request.URL.Path),
				Action:    mapMethodToAction(c.Request.Method),
				NewValues: requestBody,
				ChangedBy: userID,
				IPAddress: c.ClientIP(),
				UserAgent: c.GetHeader("User-Agent"),
			}

			// TODO: Save to database
			// This would be implemented when we create the audit service
			_ = auditLog
		}()
	}
}

func extractTableName(path string) string {
	// Extract table name from URL path
	// e.g., /api/v1/products -> products
	// e.g., /api/v1/users/123 -> users
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "v1" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "unknown"
}

func extractRecordID(path string) uuid.UUID {
	// Extract UUID from URL path
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if id, err := uuid.Parse(part); err == nil {
			return id
		}
	}
	return uuid.Nil
}

func mapMethodToAction(method string) models.AuditAction {
	switch method {
	case "GET":
		return models.ActionView
	case "POST":
		return models.ActionCreate
	case "PUT":
		return models.ActionUpdate
	case "DELETE":
		return models.ActionDelete
	default:
		return models.ActionView
	}
}