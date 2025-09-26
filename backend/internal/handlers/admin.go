package handlers

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rtims-backend/internal/database"
	"rtims-backend/internal/models"
	"rtims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/crypto/bcrypt"
)

var now = time.Now()

type AdminHandler struct {
	userService     *database.UserService
	categoryService *database.CategoryService
	dashboardService *database.DashboardService
	settingsService *database.SettingsService
	auditService    *database.AuditService
	db              *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{
		userService:     database.NewUserService(db),
		categoryService: database.NewCategoryService(db),
		dashboardService: database.NewDashboardService(db),
		settingsService: database.NewSettingsService(db),
		auditService:    database.NewAuditService(db),
		db:              db,
	}
}

// Helper function to create audit log
func createAuditLog(c *gin.Context, tableName string, recordID uuid.UUID, action models.AuditAction, oldValues, newValues map[string]interface{}) {
	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		// For cases where user is not authenticated, use system user
		userID = uuid.Nil
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	auditLog := models.AuditLog{
		ID:         uuid.New(),
		TableName:  tableName,
		RecordID:   recordID,
		Action:     action,
		OldValues:  oldValues,
		NewValues:  newValues,
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	// Save to database using audit service
	// Note: This would be handled by the audit middleware in production
	// but keeping this helper for specific audit logging needs
	_ = auditLog
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.dashboardService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) GetDashboardAlerts(c *gin.Context) {
	alerts, err := h.dashboardService.GetAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard alerts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	role := c.Query("role")
	isActive := c.Query("is_active")

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filter := models.UserFilter{
		Page:     page,
		Limit:    limit,
		Search:   search,
		Role:     role,
		IsActive: isActive,
	}

	users, total, err := h.userService.GetUsers(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + limit - 1) / limit,
		},
	})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if req.Name == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, email, and password are required"})
		return
	}

	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user already exists
	existingUser, err := h.userService.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := &models.User{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      req.Role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = h.userService.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "users",
		RecordID:   user.ID,
		Action:     models.ActionCreate,
		OldValues:  nil,
		NewValues:  map[string]interface{}{"name": req.Name, "email": req.Email, "role": req.Role},
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

	c.JSON(http.StatusCreated, user)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
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

	// Get existing user from database
	oldUser, err := h.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// Update user in database
	err = h.userService.UpdateUser(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
		return
	}

	// Get updated user
	user, err := h.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "users",
		RecordID:   id,
		Action:     models.ActionUpdate,
		OldValues:  map[string]interface{}{"name": oldUser.Name, "email": oldUser.Email, "role": oldUser.Role, "is_active": oldUser.IsActive},
		NewValues:  map[string]interface{}{"name": user.Name, "email": user.Email, "role": user.Role, "is_active": user.IsActive},
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user data for audit log before deletion
	oldUser, err := h.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user from database
	err = h.userService.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "users",
		RecordID:   id,
		Action:     models.ActionDelete,
		OldValues:  map[string]interface{}{"name": oldUser.Name, "email": oldUser.Email, "role": oldUser.Role, "is_active": oldUser.IsActive},
		NewValues:  nil,
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *AdminHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category name is required"})
		return
	}

	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create category
	category := &models.Category{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	err = h.categoryService.CreateCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "categories",
		RecordID:   category.ID,
		Action:     models.ActionCreate,
		OldValues:  nil,
		NewValues:  map[string]interface{}{"name": req.Name, "description": req.Description},
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusCreated, category)
}

func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var req models.UpdateCategoryRequest
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

	// Get existing category from database
	oldCategory, err := h.categoryService.GetCategory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	// Update category in database
	err = h.categoryService.UpdateCategory(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category: " + err.Error()})
		return
	}

	// Get updated category
	category, err := h.categoryService.GetCategory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated category: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "categories",
		RecordID:   id,
		Action:     models.ActionUpdate,
		OldValues:  map[string]interface{}{"name": oldCategory.Name, "description": oldCategory.Description},
		NewValues:  map[string]interface{}{"name": category.Name, "description": category.Description},
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, category)
}

func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get category data for audit log before deletion
	oldCategory, err := h.categoryService.GetCategory(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Check if category has products
	var productCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM products WHERE category = $1", oldCategory.Name).Scan(&productCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check category usage: " + err.Error()})
		return
	}

	if productCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category with existing products"})
		return
	}

	// Delete category from database
	err = h.categoryService.DeleteCategory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "categories",
		RecordID:   id,
		Action:     models.ActionDelete,
		OldValues:  map[string]interface{}{"name": oldCategory.Name, "description": oldCategory.Description},
		NewValues:  nil,
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

func (h *AdminHandler) GenerateInventoryReport(c *gin.Context) {
	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	category := c.Query("category")
	format := c.DefaultQuery("format", "json") // json, csv, pdf

	// Build query based on filters
	query := `
		SELECT p.id, p.name, p.sku, p.stock, p.price, p.category, p.minimum_threshold,
		       c.name as category_name, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category = c.name
	`
	args := []interface{}{}
	argCount := 0

	conditions := []string{}

	if startDate != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.created_at >= $%d", argCount))
		args = append(args, startDate)
	}

	if endDate != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.created_at <= $%d", argCount))
		args = append(args, endDate)
	}

	if category != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("p.category = $%d", argCount))
		args = append(args, category)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY p.name"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate inventory report: " + err.Error()})
		return
	}
	defer rows.Close()

	var products []gin.H
	var totalValue float64
	var lowStockCount int

	for rows.Next() {
		var id, name, sku, categoryName string
		var stock int
		var price float64
		var minimumThreshold int
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &name, &sku, &stock, &price, &categoryName, &minimumThreshold, &categoryName, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		product := gin.H{
			"id":                id,
			"name":              name,
			"sku":               sku,
			"stock":             stock,
			"price":             price,
			"category":          categoryName,
			"minimum_threshold": minimumThreshold,
			"created_at":        createdAt,
			"updated_at":        updatedAt,
		}
		products = append(products, product)
		totalValue += price * float64(stock)

		if stock <= minimumThreshold {
			lowStockCount++
		}
	}

	report := gin.H{
		"report_type":    "inventory",
		"generated_at":   time.Now(),
		"date_range": gin.H{
			"start": startDate,
			"end":   endDate,
		},
		"filters": gin.H{
			"category": category,
		},
		"summary": gin.H{
			"total_products":   len(products),
			"total_value":      totalValue,
			"low_stock_items":  lowStockCount,
			"average_price":    totalValue / float64(len(products)),
		},
		"data": products,
	}

	if format == "json" {
		c.JSON(http.StatusOK, report)
	} else if format == "csv" {
		// Generate CSV export
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=inventory_report_%s.csv", time.Now().Format("2006-01-02_15-04-05")))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write CSV header
		writer.Write([]string{"ID", "Name", "SKU", "Stock", "Price", "Category", "Minimum Threshold", "Created At", "Updated At"})

		// Write product data
		for _, product := range products {
			writer.Write([]string{
				fmt.Sprintf("%v", product["id"]),
				fmt.Sprintf("%v", product["name"]),
				fmt.Sprintf("%v", product["sku"]),
				fmt.Sprintf("%v", product["stock"]),
				fmt.Sprintf("%.2f", product["price"]),
				fmt.Sprintf("%v", product["category"]),
				fmt.Sprintf("%v", product["minimum_threshold"]),
				fmt.Sprintf("%v", product["created_at"]),
				fmt.Sprintf("%v", product["updated_at"]),
			})
		}
	} else if format == "pdf" {
		// Generate PDF export
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)

		// Title
		pdf.Cell(40, 10, "Inventory Report")
		pdf.Ln(12)

		pdf.SetFont("Arial", "", 10)

		// Report metadata
		pdf.Cell(40, 6, fmt.Sprintf("Generated At: %s", time.Now().Format("2006-01-02 15:04:05")))
		pdf.Ln(6)
		pdf.Cell(40, 6, fmt.Sprintf("Total Products: %d", len(products)))
		pdf.Ln(6)
		pdf.Cell(40, 6, fmt.Sprintf("Total Value: %.2f", totalValue))
		pdf.Ln(6)
		pdf.Cell(40, 6, fmt.Sprintf("Low Stock Items: %d", lowStockCount))
		pdf.Ln(10)

		// Table header
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(20, 8, "ID", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 8, "Name", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 8, "SKU", "1", 0, "C", true, 0, "")
		pdf.CellFormat(15, 8, "Stock", "1", 0, "C", true, 0, "")
		pdf.CellFormat(20, 8, "Price", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 8, "Category", "1", 0, "C", true, 0, "")
		pdf.CellFormat(20, 8, "Min Threshold", "1", 0, "C", true, 0, "")
		pdf.Ln(8)

		// Table data
		pdf.SetFont("Arial", "", 7)
		pdf.SetFillColor(255, 255, 255)

		for _, product := range products {
			pdf.CellFormat(20, 6, fmt.Sprintf("%v", product["id"]), "1", 0, "L", false, 0, "")
			pdf.CellFormat(40, 6, fmt.Sprintf("%v", product["name"]), "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 6, fmt.Sprintf("%v", product["sku"]), "1", 0, "L", false, 0, "")
			pdf.CellFormat(15, 6, fmt.Sprintf("%v", product["stock"]), "1", 0, "C", false, 0, "")
			pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", product["price"]), "1", 0, "R", false, 0, "")
			pdf.CellFormat(30, 6, fmt.Sprintf("%v", product["category"]), "1", 0, "L", false, 0, "")
			pdf.CellFormat(20, 6, fmt.Sprintf("%v", product["minimum_threshold"]), "1", 0, "C", false, 0, "")
			pdf.Ln(6)
		}

		// Set headers for PDF download
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=inventory_report_%s.pdf", time.Now().Format("2006-01-02_15-04-05")))

		// Output PDF to response writer
		err := pdf.Output(c.Writer)
		if err != nil {
			log.Printf("Failed to generate PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF report"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Supported formats: json, csv, pdf"})
	}
}

func (h *AdminHandler) GenerateMovementReport(c *gin.Context) {
	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	productID := c.Query("product_id")
	reason := c.Query("reason")
	format := c.DefaultQuery("format", "json")
	reportType := "movements" // Define reportType for this function

	// Build query based on filters
	query := `
		SELECT sm.id, sm.product_id, sm.change, sm.reason, sm.created_at, sm.notes,
		       p.name as product_name, u.name as user_name
		FROM stock_movements sm
		LEFT JOIN products p ON sm.product_id = p.id
		LEFT JOIN users u ON sm.created_by = u.id
	`
	args := []interface{}{}
	argCount := 0

	conditions := []string{}

	if startDate != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("sm.created_at >= $%d", argCount))
		args = append(args, startDate)
	}

	if endDate != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("sm.created_at <= $%d", argCount))
		args = append(args, endDate)
	}

	if productID != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("sm.product_id = $%d", argCount))
		args = append(args, productID)
	}

	if reason != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("sm.reason = $%d", argCount))
		args = append(args, reason)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY sm.created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate movement report: " + err.Error()})
		return
	}
	defer rows.Close()

	var movements []gin.H
	var totalIn, totalOut int

	for rows.Next() {
		var id, productID, reason, productName, userName, notes string
		var change int
		var createdAt time.Time

		err := rows.Scan(&id, &productID, &change, &reason, &createdAt, &notes, &productName, &userName)
		if err != nil {
			continue
		}

		movement := gin.H{
			"id":           id,
			"product_id":   productID,
			"product_name": productName,
			"change":       change,
			"reason":       reason,
			"user_name":    userName,
			"created_at":   createdAt,
			"notes":        notes,
		}
		movements = append(movements, movement)

		if change > 0 {
			totalIn += change
		} else {
			totalOut += int(-change)
		}
	}

	report := gin.H{
		"report_type":    "stock_movements",
		"generated_at":   time.Now(),
		"date_range": gin.H{
			"start": startDate,
			"end":   endDate,
		},
		"filters": gin.H{
			"product_id": productID,
			"reason":     reason,
		},
		"summary": gin.H{
			"total_movements": len(movements),
			"total_in":        totalIn,
			"total_out":       totalOut,
			"net_change":      totalIn - totalOut,
		},
		"data": movements,
	}

	if format == "json" {
		c.JSON(http.StatusOK, report)
	} else if format == "csv" {
		// Generate CSV export
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_report_%s.csv", reportType, time.Now().Format("2006-01-02_15-04-05")))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write CSV header based on report type
		switch reportType {
		case "inventory":
			writer.Write([]string{"ID", "Name", "SKU", "Stock", "Price", "Category", "Minimum Threshold"})
			for _, item := range report["data"].([]gin.H) {
				writer.Write([]string{
					fmt.Sprintf("%v", item["id"]),
					fmt.Sprintf("%v", item["name"]),
					fmt.Sprintf("%v", item["sku"]),
					fmt.Sprintf("%v", item["stock"]),
					fmt.Sprintf("%.2f", item["price"]),
					fmt.Sprintf("%v", item["category"]),
					fmt.Sprintf("%v", item["minimum_threshold"]),
				})
			}
		case "movements":
			writer.Write([]string{"ID", "Product ID", "Product Name", "Change", "Reason", "Created At"})
			for _, item := range report["data"].([]gin.H) {
				writer.Write([]string{
					fmt.Sprintf("%v", item["id"]),
					fmt.Sprintf("%v", item["product_id"]),
					fmt.Sprintf("%v", item["product_name"]),
					fmt.Sprintf("%v", item["change"]),
					fmt.Sprintf("%v", item["reason"]),
					fmt.Sprintf("%v", item["created_at"]),
				})
			}
		case "users":
			writer.Write([]string{"User ID", "Actions", "Last Action"})
			for _, item := range report["data"].([]gin.H) {
				writer.Write([]string{
					fmt.Sprintf("%v", item["user_id"]),
					fmt.Sprintf("%v", item["actions"]),
					fmt.Sprintf("%v", item["last_action"]),
				})
			}
		}
	} else if format == "pdf" {
		// Generate PDF export
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)

		// Title
		pdf.Cell(40, 10, fmt.Sprintf("%s Report", strings.Title(reportType)))
		pdf.Ln(12)

		pdf.SetFont("Arial", "", 10)

		// Report metadata
		pdf.Cell(40, 6, fmt.Sprintf("Generated At: %s", time.Now().Format("2006-01-02 15:04:05")))
		pdf.Ln(6)
		pdf.Cell(40, 6, fmt.Sprintf("Report Type: %s", reportType))
		pdf.Ln(6)
		pdf.Cell(40, 6, fmt.Sprintf("Format: %s", format))
		pdf.Ln(10)

		// Table header and data based on report type
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(240, 240, 240)

		switch reportType {
		case "inventory":
			pdf.CellFormat(20, 8, "ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(40, 8, "Name", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "SKU", "1", 0, "C", true, 0, "")
			pdf.CellFormat(15, 8, "Stock", "1", 0, "C", true, 0, "")
			pdf.CellFormat(20, 8, "Price", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Category", "1", 0, "C", true, 0, "")
			pdf.CellFormat(20, 8, "Min Threshold", "1", 0, "C", true, 0, "")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 7)
			pdf.SetFillColor(255, 255, 255)
			for _, item := range report["data"].([]gin.H) {
				pdf.CellFormat(20, 6, fmt.Sprintf("%v", item["id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(40, 6, fmt.Sprintf("%v", item["name"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(25, 6, fmt.Sprintf("%v", item["sku"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(15, 6, fmt.Sprintf("%v", item["stock"]), "1", 0, "C", false, 0, "")
				pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", item["price"]), "1", 0, "R", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["category"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(20, 6, fmt.Sprintf("%v", item["minimum_threshold"]), "1", 0, "C", false, 0, "")
				pdf.Ln(6)
			}
		case "movements":
			pdf.CellFormat(25, 8, "ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Product ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(40, 8, "Product Name", "1", 0, "C", true, 0, "")
			pdf.CellFormat(15, 8, "Change", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Reason", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Created At", "1", 0, "C", true, 0, "")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 7)
			pdf.SetFillColor(255, 255, 255)
			for _, item := range report["data"].([]gin.H) {
				pdf.CellFormat(25, 6, fmt.Sprintf("%v", item["id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["product_id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(40, 6, fmt.Sprintf("%v", item["product_name"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(15, 6, fmt.Sprintf("%v", item["change"]), "1", 0, "C", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["reason"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["created_at"]), "1", 0, "L", false, 0, "")
				pdf.Ln(6)
			}
		case "users":
			pdf.CellFormat(50, 8, "User ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "Actions", "1", 0, "C", true, 0, "")
			pdf.CellFormat(50, 8, "Last Action", "1", 0, "C", true, 0, "")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 7)
			pdf.SetFillColor(255, 255, 255)
			for _, item := range report["data"].([]gin.H) {
				pdf.CellFormat(50, 6, fmt.Sprintf("%v", item["user_id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(25, 6, fmt.Sprintf("%v", item["actions"]), "1", 0, "C", false, 0, "")
				pdf.CellFormat(50, 6, fmt.Sprintf("%v", item["last_action"]), "1", 0, "L", false, 0, "")
				pdf.Ln(6)
			}
		}

		// Set headers for PDF download
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_report_%s.pdf", reportType, time.Now().Format("2006-01-02_15-04-05")))

		// Output PDF to response writer
		err := pdf.Output(c.Writer)
		if err != nil {
			log.Printf("Failed to generate PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF report"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Supported formats: json, csv, pdf"})
	}
}

func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsService.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var req map[string]interface{}
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

	// Get old settings for audit log
	oldSettings, err := h.settingsService.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current settings: " + err.Error()})
		return
	}

	// Update settings in database
	err = h.settingsService.UpdateSettings(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings: " + err.Error()})
		return
	}

	// Get updated settings
	newSettings, err := h.settingsService.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated settings: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "system_settings",
		RecordID:   uuid.New(), // Using new UUID since settings don't have a specific ID
		Action:     models.ActionUpdate,
		OldValues:  oldSettings,
		NewValues:  newSettings,
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, newSettings)
}

func (h *AdminHandler) GetReportStats(c *gin.Context) {
	// Get report statistics from audit logs
	var totalReports int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs
		WHERE table_name = 'reports' OR action = 'report_generated'
	`).Scan(&totalReports)
	if err != nil {
		totalReports = 0
	}

	// Get this month's reports
	var thisMonth int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM audit_logs
		WHERE (table_name = 'reports' OR action = 'report_generated')
		AND changed_at >= date_trunc('month', CURRENT_DATE)
	`).Scan(&thisMonth)
	if err != nil {
		thisMonth = 0
	}

	// Get total data points (approximate from products and movements)
	var dataPoints int
	err = h.db.QueryRow("SELECT (SELECT COUNT(*) FROM products) + (SELECT COUNT(*) FROM stock_movements)").Scan(&dataPoints)
	if err != nil {
		dataPoints = 0
	}

	// Get most popular report type from actual data
	var mostPopularType string
	err = h.db.QueryRow(`
		SELECT table_name, COUNT(*) as count
		FROM audit_logs
		WHERE table_name IN ('reports', 'products', 'stock_movements', 'users')
		GROUP BY table_name
		ORDER BY count DESC
		LIMIT 1
	`).Scan(&mostPopularType)
	if err != nil {
		mostPopularType = "inventory" // fallback
	}

	// Calculate average report size from actual data
	var avgSize float64
	err = h.db.QueryRow(`
		SELECT AVG(LENGTH(COALESCE(old_values::text, '')) + LENGTH(COALESCE(new_values::text, '')))
		FROM audit_logs
		WHERE table_name IN ('reports', 'products', 'stock_movements', 'users')
	`).Scan(&avgSize)
	if err != nil {
		avgSize = 0
	}

	// Format average size
	var avgSizeStr string
	if avgSize >= 1024*1024 {
		avgSizeStr = fmt.Sprintf("%.1fMB", avgSize/(1024*1024))
	} else if avgSize >= 1024 {
		avgSizeStr = fmt.Sprintf("%.1fKB", avgSize/1024)
	} else {
		avgSizeStr = fmt.Sprintf("%.0fB", avgSize)
	}

	stats := gin.H{
		"total_reports":     totalReports,
		"this_month":        thisMonth,
		"data_points":       dataPoints,
		"last_generated":    time.Now(),
		"most_popular_type": mostPopularType,
		"average_size":      avgSizeStr,
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) GetReportTypes(c *gin.Context) {
	// Get available report types from database tables and system capabilities
	reportTypes := []gin.H{
		{
			"id":          "inventory",
			"name":        "Inventory Report",
			"description": "Complete overview of all products and stock levels",
			"available":   true,
			"formats":     []string{"json", "csv", "pdf"},
			"frequency":   "daily",
		},
		{
			"id":          "movements",
			"name":        "Stock Movements",
			"description": "Track all inventory changes and transactions",
			"available":   true,
			"formats":     []string{"json", "csv", "pdf"},
			"frequency":   "daily",
		},
		{
			"id":          "users",
			"name":        "User Activity",
			"description": "User actions and system usage statistics",
			"available":   true,
			"formats":     []string{"json", "csv"},
			"frequency":   "weekly",
		},
	}

	// Check if financial data is available
	var productCount int
	err := h.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&productCount)
	if err == nil && productCount > 0 {
		// Add financial report if we have products
		financialReport := gin.H{
			"id":          "financial",
			"name":        "Financial Summary",
			"description": "Revenue, costs, and profit analysis",
			"available":   true,
			"formats":     []string{"json", "pdf"},
			"frequency":   "monthly",
		}
		reportTypes = append(reportTypes, financialReport)
	}

	c.JSON(http.StatusOK, reportTypes)
}

func (h *AdminHandler) GetRecentReports(c *gin.Context) {
	// Get recent reports from audit logs
	query := `
		SELECT id, table_name, action, changed_at, changed_by
		FROM audit_logs
		WHERE table_name IN ('reports', 'products', 'stock_movements')
		ORDER BY changed_at DESC
		LIMIT 10
	`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent reports: " + err.Error()})
		return
	}
	defer rows.Close()

	var reports []gin.H
	for rows.Next() {
		var id, tableName, action string
		var changedAt time.Time
		var changedBy uuid.UUID

		err := rows.Scan(&id, &tableName, &action, &changedAt, &changedBy)
		if err != nil {
			continue
		}

		// Calculate approximate size based on table name and record count
		var estimatedSize int
		switch tableName {
		case "products":
			estimatedSize = 1024 // ~1KB per product record
		case "stock_movements":
			estimatedSize = 512 // ~512B per movement record
		case "users":
			estimatedSize = 256 // ~256B per user record
		default:
			estimatedSize = 1024 // default estimate
		}

		report := gin.H{
			"id":           id,
			"name":         fmt.Sprintf("%s Report", strings.Title(tableName)),
			"type":         tableName,
			"format":       "json",
			"generated_at": changedAt,
			"size":         estimatedSize,
			"status":       "completed",
			"download_url": fmt.Sprintf("/api/admin/reports/%s/download/%s", tableName, id),
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, reports)
}

func (h *AdminHandler) GenerateReport(c *gin.Context) {
	reportType := c.Param("type")
	format := c.DefaultQuery("format", "json")

	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	report := gin.H{
		"report_type":  reportType,
		"format":       format,
		"generated_at": time.Now(),
		"data":         []gin.H{},
	}

	switch reportType {
	case "inventory":
		// Get all products
		query := "SELECT id, name, sku, stock, price, category, minimum_threshold FROM products ORDER BY name"
		rows, err := h.db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate inventory report: " + err.Error()})
			return
		}
		defer rows.Close()

		var products []gin.H
		for rows.Next() {
			var id, name, sku, category string
			var stock int
			var price float64
			var minimumThreshold int

			err := rows.Scan(&id, &name, &sku, &stock, &price, &category, &minimumThreshold)
			if err != nil {
				continue
			}

			product := gin.H{
				"id":                id,
				"name":              name,
				"sku":               sku,
				"stock":             stock,
				"price":             price,
				"category":          category,
				"minimum_threshold": minimumThreshold,
			}
			products = append(products, product)
		}
		report["data"] = products

	case "movements":
		// Get recent stock movements
		query := `
			SELECT sm.id, sm.product_id, sm.change, sm.reason, sm.created_at, p.name
			FROM stock_movements sm
			JOIN products p ON sm.product_id = p.id
			ORDER BY sm.created_at DESC
			LIMIT 100
		`
		rows, err := h.db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate movements report: " + err.Error()})
			return
		}
		defer rows.Close()

		var movements []gin.H
		for rows.Next() {
			var id, productID, reason, productName string
			var change int
			var createdAt time.Time

			err := rows.Scan(&id, &productID, &change, &reason, &createdAt, &productName)
			if err != nil {
				continue
			}

			movement := gin.H{
				"id":         id,
				"product_id": productID,
				"product_name": productName,
				"change":     change,
				"reason":     reason,
				"created_at": createdAt,
			}
			movements = append(movements, movement)
		}
		report["data"] = movements

	case "users":
		// Get user activity from audit logs
		query := `
			SELECT changed_by, COUNT(*) as actions, MAX(changed_at) as last_action
			FROM audit_logs
			WHERE changed_by IS NOT NULL
			GROUP BY changed_by
			ORDER BY actions DESC
		`
		rows, err := h.db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate users report: " + err.Error()})
			return
		}
		defer rows.Close()

		var userActivities []gin.H
		for rows.Next() {
			var userID uuid.UUID
			var actions int
			var lastAction time.Time

			err := rows.Scan(&userID, &actions, &lastAction)
			if err != nil {
				continue
			}

			activity := gin.H{
				"user_id":     userID,
				"actions":     actions,
				"last_action": lastAction,
			}
			userActivities = append(userActivities, activity)
		}
		report["data"] = userActivities

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report type"})
		return
	}

	// Create audit log for report generation
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "reports",
		RecordID:   uuid.New(),
		Action:     models.ActionCreate,
		OldValues:  nil,
		NewValues:  map[string]interface{}{"report_type": reportType, "format": format, "data_count": len(report["data"].([]gin.H))},
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	if format == "json" {
		c.JSON(http.StatusOK, report)
	} else if format == "csv" {
		// Generate CSV export
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_report_%s.csv", reportType, time.Now().Format("2006-01-02_15-04-05")))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write CSV header based on report type
		switch reportType {
		case "inventory":
			writer.Write([]string{"ID", "Name", "SKU", "Stock", "Price", "Category", "Minimum Threshold"})
			for _, item := range report["data"].([]gin.H) {
				writer.Write([]string{
					fmt.Sprintf("%v", item["id"]),
					fmt.Sprintf("%v", item["name"]),
					fmt.Sprintf("%v", item["sku"]),
					fmt.Sprintf("%v", item["stock"]),
					fmt.Sprintf("%.2f", item["price"]),
					fmt.Sprintf("%v", item["category"]),
					fmt.Sprintf("%v", item["minimum_threshold"]),
				})
			}
		case "movements":
			writer.Write([]string{"ID", "Product ID", "Product Name", "Change", "Reason", "Created At"})
			for _, item := range report["data"].([]gin.H) {
				writer.Write([]string{
					fmt.Sprintf("%v", item["id"]),
					fmt.Sprintf("%v", item["product_id"]),
					fmt.Sprintf("%v", item["product_name"]),
					fmt.Sprintf("%v", item["change"]),
					fmt.Sprintf("%v", item["reason"]),
					fmt.Sprintf("%v", item["created_at"]),
				})
			}
		case "users":
			writer.Write([]string{"User ID", "Actions", "Last Action"})
			for _, item := range report["data"].([]gin.H) {
				writer.Write([]string{
					fmt.Sprintf("%v", item["user_id"]),
					fmt.Sprintf("%v", item["actions"]),
					fmt.Sprintf("%v", item["last_action"]),
				})
			}
		}
	} else if format == "pdf" {
		// Generate PDF export
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)

		// Title
		pdf.Cell(40, 10, fmt.Sprintf("%s Report", strings.Title(reportType)))
		pdf.Ln(12)

		pdf.SetFont("Arial", "", 10)

		// Report metadata
		pdf.Cell(40, 6, fmt.Sprintf("Generated At: %s", time.Now().Format("2006-01-02 15:04:05")))
		pdf.Ln(6)
		pdf.Cell(40, 6, fmt.Sprintf("Report Type: %s", reportType))
		pdf.Ln(6)
		pdf.Cell(40, 6, fmt.Sprintf("Format: %s", format))
		pdf.Ln(10)

		// Table header and data based on report type
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(240, 240, 240)

		switch reportType {
		case "inventory":
			pdf.CellFormat(20, 8, "ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(40, 8, "Name", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "SKU", "1", 0, "C", true, 0, "")
			pdf.CellFormat(15, 8, "Stock", "1", 0, "C", true, 0, "")
			pdf.CellFormat(20, 8, "Price", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Category", "1", 0, "C", true, 0, "")
			pdf.CellFormat(20, 8, "Min Threshold", "1", 0, "C", true, 0, "")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 7)
			pdf.SetFillColor(255, 255, 255)
			for _, item := range report["data"].([]gin.H) {
				pdf.CellFormat(20, 6, fmt.Sprintf("%v", item["id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(40, 6, fmt.Sprintf("%v", item["name"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(25, 6, fmt.Sprintf("%v", item["sku"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(15, 6, fmt.Sprintf("%v", item["stock"]), "1", 0, "C", false, 0, "")
				pdf.CellFormat(20, 6, fmt.Sprintf("%.2f", item["price"]), "1", 0, "R", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["category"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(20, 6, fmt.Sprintf("%v", item["minimum_threshold"]), "1", 0, "C", false, 0, "")
				pdf.Ln(6)
			}
		case "movements":
			pdf.CellFormat(25, 8, "ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Product ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(40, 8, "Product Name", "1", 0, "C", true, 0, "")
			pdf.CellFormat(15, 8, "Change", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Reason", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Created At", "1", 0, "C", true, 0, "")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 7)
			pdf.SetFillColor(255, 255, 255)
			for _, item := range report["data"].([]gin.H) {
				pdf.CellFormat(25, 6, fmt.Sprintf("%v", item["id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["product_id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(40, 6, fmt.Sprintf("%v", item["product_name"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(15, 6, fmt.Sprintf("%v", item["change"]), "1", 0, "C", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["reason"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(30, 6, fmt.Sprintf("%v", item["created_at"]), "1", 0, "L", false, 0, "")
				pdf.Ln(6)
			}
		case "users":
			pdf.CellFormat(50, 8, "User ID", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "Actions", "1", 0, "C", true, 0, "")
			pdf.CellFormat(50, 8, "Last Action", "1", 0, "C", true, 0, "")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 7)
			pdf.SetFillColor(255, 255, 255)
			for _, item := range report["data"].([]gin.H) {
				pdf.CellFormat(50, 6, fmt.Sprintf("%v", item["user_id"]), "1", 0, "L", false, 0, "")
				pdf.CellFormat(25, 6, fmt.Sprintf("%v", item["actions"]), "1", 0, "C", false, 0, "")
				pdf.CellFormat(50, 6, fmt.Sprintf("%v", item["last_action"]), "1", 0, "L", false, 0, "")
				pdf.Ln(6)
			}
		}

		// Set headers for PDF download
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_report_%s.pdf", reportType, time.Now().Format("2006-01-02_15-04-05")))

		// Output PDF to response writer
		err := pdf.Output(c.Writer)
		if err != nil {
			log.Printf("Failed to generate PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF report"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Supported formats: json, csv, pdf"})
	}
}

func (h *AdminHandler) GetSystemStatus(c *gin.Context) {
	status, err := h.settingsService.GetSystemStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *AdminHandler) TriggerBackup(c *gin.Context) {
	// Get current user for audit logging
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	backup, err := h.settingsService.TriggerBackup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger backup: " + err.Error()})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "system",
		RecordID:   uuid.New(),
		Action:     models.ActionCreate,
		OldValues:  nil,
		NewValues:  map[string]interface{}{"backup_id": backup["backup_id"], "action": "backup_triggered"},
		ChangedBy:  userID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = h.auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, backup)
}