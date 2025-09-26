package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rtims-backend/internal/models"

	"github.com/google/uuid"
)

type ProductService struct {
	db *sql.DB
}

func NewProductService(db *sql.DB) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) GetProducts(filter models.ProductFilter) ([]models.Product, int, error) {
	// Build query
	query := `SELECT id, name, sku, stock, price, category, minimum_threshold, supplier_info, created_at, updated_at FROM products`
	countQuery := `SELECT COUNT(*) FROM products`
	var args []interface{}
	var conditions []string

	// Add filters
	if filter.Search != "" {
		conditions = append(conditions, "(name ILIKE $%d OR sku ILIKE $%d OR category ILIKE $%d)")
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.Category != "" {
		conditions = append(conditions, "category = $%d")
		args = append(args, filter.Category)
	}

	if filter.MinStock != nil {
		conditions = append(conditions, "stock >= $%d")
		args = append(args, *filter.MinStock)
	}

	if filter.MaxStock != nil {
		conditions = append(conditions, "stock <= $%d")
		args = append(args, *filter.MaxStock)
	}

	if filter.MinPrice != nil {
		conditions = append(conditions, "price >= $%d")
		args = append(args, *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		conditions = append(conditions, "price <= $%d")
		args = append(args, *filter.MaxPrice)
	}

	if filter.LowStockOnly {
		conditions = append(conditions, "stock <= minimum_threshold")
	}

	// Add WHERE clause if conditions exist
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// Add sorting
	sortBy := "created_at"
	sortOrder := "DESC"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "name", "sku", "stock", "price", "category", "created_at", "updated_at":
			sortBy = filter.SortBy
		}
	}
	if filter.SortOrder != "" && (filter.SortOrder == "ASC" || filter.SortOrder == "DESC") {
		sortOrder = filter.SortOrder
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	offset := (filter.Page - 1) * filter.Limit
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, offset)

	// Get total count
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get product count: %w", err)
	}

	// Get products
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.SKU,
			&product.Stock,
			&product.Price,
			&product.Category,
			&product.MinimumThreshold,
			&product.SupplierInfo,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	return products, total, nil
}

func (s *ProductService) GetProduct(id uuid.UUID) (*models.Product, error) {
	query := `SELECT id, name, sku, stock, price, category, minimum_threshold, supplier_info, created_at, updated_at
			  FROM products WHERE id = $1`

	var product models.Product
	err := s.db.QueryRow(query, id).Scan(
		&product.ID,
		&product.Name,
		&product.SKU,
		&product.Stock,
		&product.Price,
		&product.Category,
		&product.MinimumThreshold,
		&product.SupplierInfo,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (s *ProductService) CreateProduct(product *models.Product) error {
	query := `INSERT INTO products (id, name, sku, stock, price, category, minimum_threshold, supplier_info, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := s.db.Exec(query,
		product.ID,
		product.Name,
		product.SKU,
		product.Stock,
		product.Price,
		product.Category,
		product.MinimumThreshold,
		product.SupplierInfo,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (s *ProductService) UpdateProduct(id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	var setParts []string
	var args []interface{}
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "name", "sku", "category", "supplier_info":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		case "stock", "minimum_threshold":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		case "price":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no valid updates provided")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (s *ProductService) DeleteProduct(id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (s *ProductService) UpdateProductStock(productID uuid.UUID, change int, reason models.MovementReason, createdBy uuid.UUID, notes string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update product stock
	query := `UPDATE products SET stock = stock + $1, updated_at = $2 WHERE id = $3`
	_, err = tx.Exec(query, change, time.Now(), productID)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	// Create stock movement record
	movementQuery := `INSERT INTO stock_movements (id, product_id, change, reason, created_by, created_at, notes)
					  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	movementID := uuid.New()
	_, err = tx.Exec(movementQuery, movementID, productID, change, reason, createdBy, time.Now(), notes)
	if err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}

	return tx.Commit()
}

func (s *ProductService) GetStockMovements(filter models.StockMovementFilter) ([]models.StockMovement, int, error) {
	// Build query
	query := `SELECT id, product_id, change, reason, created_by, created_at, notes FROM stock_movements`
	countQuery := `SELECT COUNT(*) FROM stock_movements`
	var args []interface{}
	var conditions []string

	// Add filters
	if filter.ProductID != nil {
		conditions = append(conditions, "product_id = $1")
		args = append(args, *filter.ProductID)
	}

	if filter.Reason != nil {
		conditions = append(conditions, "reason = $%d")
		args = append(args, *filter.Reason)
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "created_at >= $%d")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "created_at <= $%d")
		args = append(args, *filter.EndDate)
	}

	// Add WHERE clause if conditions exist
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// Add sorting
	sortBy := "created_at"
	sortOrder := "DESC"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "created_at", "change", "reason":
			sortBy = filter.SortBy
		}
	}
	if filter.SortOrder != "" && (filter.SortOrder == "ASC" || filter.SortOrder == "DESC") {
		sortOrder = filter.SortOrder
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	offset := (filter.Page - 1) * filter.Limit
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, offset)

	// Get total count
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get stock movements count: %w", err)
	}

	// Get stock movements
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get stock movements: %w", err)
	}
	defer rows.Close()

	var movements []models.StockMovement
	for rows.Next() {
		var movement models.StockMovement
		err := rows.Scan(
			&movement.ID,
			&movement.ProductID,
			&movement.Change,
			&movement.Reason,
			&movement.CreatedBy,
			&movement.CreatedAt,
			&movement.Notes,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan stock movement: %w", err)
		}
		movements = append(movements, movement)
	}

	return movements, total, nil
}

func (s *ProductService) GetStockMovement(id uuid.UUID) (*models.StockMovement, error) {
	query := `SELECT id, product_id, change, reason, created_by, created_at, notes
			  FROM stock_movements WHERE id = $1`

	var movement models.StockMovement
	err := s.db.QueryRow(query, id).Scan(
		&movement.ID,
		&movement.ProductID,
		&movement.Change,
		&movement.Reason,
		&movement.CreatedBy,
		&movement.CreatedAt,
		&movement.Notes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("stock movement not found")
		}
		return nil, fmt.Errorf("failed to get stock movement: %w", err)
	}

	return &movement, nil
}