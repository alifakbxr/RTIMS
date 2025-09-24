package websocket

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"rtims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, implement proper origin checking
		return true
	},
}

func ServeWebSocket(hub *Hub, c *gin.Context, db *sql.DB, redisClient *redis.Client) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	// Get current user info
	userID, role, err := middleware.GetCurrentUser(c)
	if err != nil {
		log.Println("Failed to get user info:", err)
		conn.Close()
		return
	}

	client := &Client{
		ID:   userID.String(),
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  hub,
	}

	client.Hub.Register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()

	// Send initial data to the client
	go func() {
		// Send current stock levels
		sendStockUpdates(client, db)

		// Send notifications
		sendNotifications(client, db, userID)

		// Send system status
		sendSystemStatus(client, db)
	}()
}

func sendStockUpdates(client *Client, db *sql.DB) {
	// Query low stock products
	rows, err := db.Query(`
		SELECT id, name, sku, stock, minimum_threshold
		FROM products
		WHERE stock <= minimum_threshold AND minimum_threshold > 0
	`)
	if err != nil {
		log.Println("Failed to query low stock products:", err)
		return
	}
	defer rows.Close()

	var lowStockProducts []map[string]interface{}
	for rows.Next() {
		var id, name, sku string
		var stock, threshold int
		if err := rows.Scan(&id, &name, &sku, &stock, &threshold); err != nil {
			continue
		}

		lowStockProducts = append(lowStockProducts, map[string]interface{}{
			"id":                id,
			"name":              name,
			"sku":               sku,
			"stock":             stock,
			"minimum_threshold": threshold,
			"type":              "low_stock_alert",
		})
	}

	if len(lowStockProducts) > 0 {
		message := map[string]interface{}{
			"type":    "stock_update",
			"data":    lowStockProducts,
			"message": "Low stock alerts",
		}

		if jsonData, err := json.Marshal(message); err == nil {
			select {
			case client.Send <- jsonData:
			case <-time.After(time.Second):
			}
		}
	}
}

func sendNotifications(client *Client, db *sql.DB, userID uuid.UUID) {
	// Query unread notifications
	rows, err := db.Query(`
		SELECT id, message, type, created_at
		FROM notifications
		WHERE user_id = $1 AND is_read = false
		ORDER BY created_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		log.Println("Failed to query notifications:", err)
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	for rows.Next() {
		var id, message, notifType string
		var createdAt time.Time
		if err := rows.Scan(&id, &message, &notifType, &createdAt); err != nil {
			continue
		}

		notifications = append(notifications, map[string]interface{}{
			"id":         id,
			"message":    message,
			"type":       notifType,
			"created_at": createdAt,
		})
	}

	if len(notifications) > 0 {
		message := map[string]interface{}{
			"type": "notifications",
			"data": notifications,
		}

		if jsonData, err := json.Marshal(message); err == nil {
			select {
			case client.Send <- jsonData:
			case <-time.After(time.Second):
			}
		}
	}
}

func sendSystemStatus(client *Client, db *sql.DB) {
	// Get system statistics
	var totalProducts, lowStockCount, totalUsers int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	db.QueryRow("SELECT COUNT(*) FROM products WHERE stock <= minimum_threshold").Scan(&lowStockCount)
	db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&totalUsers)

	systemStatus := map[string]interface{}{
		"total_products":   totalProducts,
		"low_stock_count":  lowStockCount,
		"total_users":      totalUsers,
		"server_time":      time.Now(),
	}

	message := map[string]interface{}{
		"type": "system_status",
		"data": systemStatus,
	}

	if jsonData, err := json.Marshal(message); err == nil {
		select {
		case client.Send <- jsonData:
		case <-time.After(time.Second):
		}
	}
}

// BroadcastStockUpdate sends stock updates to all connected clients
func BroadcastStockUpdate(hub *Hub, productID uuid.UUID, newStock int) {
	message := map[string]interface{}{
		"type":      "stock_change",
		"product_id": productID,
		"new_stock": newStock,
		"timestamp": time.Now(),
	}

	if jsonData, err := json.Marshal(message); err == nil {
		select {
		case hub.Broadcast <- jsonData:
		default:
		}
	}
}

// BroadcastNotification sends notifications to specific users or all users
func BroadcastNotification(hub *Hub, userID uuid.UUID, message string, notifType string) {
	notification := map[string]interface{}{
		"type":      "notification",
		"user_id":   userID,
		"message":   message,
		"notif_type": notifType,
		"timestamp": time.Now(),
	}

	if jsonData, err := json.Marshal(notification); err == nil {
		select {
		case hub.Broadcast <- jsonData:
		default:
		}
	}
}