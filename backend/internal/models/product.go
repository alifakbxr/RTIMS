package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID               uuid.UUID `json:"id" db:"id"`
	Name             string    `json:"name" db:"name" validate:"required,min=1,max=200"`
	SKU              string    `json:"sku" db:"sku" validate:"required,min=1,max=50"`
	Stock            int       `json:"stock" db:"stock" validate:"min=0"`
	Price            float64   `json:"price" db:"price" validate:"min=0"`
	Category         string    `json:"category" db:"category" validate:"required"`
	MinimumThreshold int       `json:"minimum_threshold" db:"minimum_threshold" validate:"min=0"`
	SupplierInfo     string    `json:"supplier_info" db:"supplier_info"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type CreateProductRequest struct {
	Name             string  `json:"name" validate:"required,min=1,max=200"`
	SKU              string  `json:"sku" validate:"required,min=1,max=50"`
	Stock            int     `json:"stock" validate:"min=0"`
	Price            float64 `json:"price" validate:"min=0"`
	Category         string  `json:"category" validate:"required"`
	MinimumThreshold int     `json:"minimum_threshold" validate:"min=0"`
	SupplierInfo     string  `json:"supplier_info"`
}

type UpdateProductRequest struct {
	Name             *string  `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	SKU              *string  `json:"sku,omitempty" validate:"omitempty,min=1,max=50"`
	Stock            *int     `json:"stock,omitempty" validate:"omitempty,min=0"`
	Price            *float64 `json:"price,omitempty" validate:"omitempty,min=0"`
	Category         *string  `json:"category,omitempty"`
	MinimumThreshold *int     `json:"minimum_threshold,omitempty" validate:"omitempty,min=0"`
	SupplierInfo     *string  `json:"supplier_info,omitempty"`
}

type ProductFilter struct {
	Search       string `form:"search"`
	Category     string `form:"category"`
	MinStock     *int   `form:"min_stock"`
	MaxStock     *int   `form:"max_stock"`
	MinPrice     *float64 `form:"min_price"`
	MaxPrice     *float64 `form:"max_price"`
	LowStockOnly bool   `form:"low_stock_only"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
	SortBy       string `form:"sort_by"`
	SortOrder    string `form:"sort_order"`
}