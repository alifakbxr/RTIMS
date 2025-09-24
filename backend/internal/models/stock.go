package models

import (
	"time"

	"github.com/google/uuid"
)

type MovementReason string

const (
	ReasonPurchase   MovementReason = "purchase"
	ReasonSale       MovementReason = "sale"
	ReasonAdjustment MovementReason = "adjustment"
	ReasonReturn     MovementReason = "return"
	ReasonDamage     MovementReason = "damage"
	ReasonTransfer   MovementReason = "transfer"
)

type StockMovement struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	ProductID uuid.UUID      `json:"product_id" db:"product_id"`
	Change    int            `json:"change" db:"change"` // positive for in, negative for out
	Reason    MovementReason `json:"reason" db:"reason" validate:"required"`
	CreatedBy uuid.UUID      `json:"created_by" db:"created_by"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	Notes     string         `json:"notes" db:"notes"`
}

type CreateStockMovementRequest struct {
	ProductID uuid.UUID      `json:"product_id" validate:"required"`
	Change    int            `json:"change" validate:"required"` // positive for in, negative for out
	Reason    MovementReason `json:"reason" validate:"required,oneof=purchase sale adjustment return damage transfer"`
	Notes     string         `json:"notes"`
}

type StockMovementFilter struct {
	ProductID *uuid.UUID      `form:"product_id"`
	Reason    *MovementReason `form:"reason"`
	StartDate *time.Time      `form:"start_date"`
	EndDate   *time.Time      `form:"end_date"`
	Page      int             `form:"page"`
	Limit     int             `form:"limit"`
	SortBy    string          `form:"sort_by"`
	SortOrder string          `form:"sort_order"`
}