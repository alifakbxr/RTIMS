package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationLowStock NotificationType = "low_stock"
	NotificationSystem   NotificationType = "system"
	NotificationUser     NotificationType = "user"
)

type Notification struct {
	ID        uuid.UUID         `json:"id" db:"id"`
	UserID    uuid.UUID         `json:"user_id" db:"user_id"`
	Message   string            `json:"message" db:"message" validate:"required"`
	Type      NotificationType  `json:"type" db:"type" validate:"required"`
	IsRead    bool              `json:"is_read" db:"is_read"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
}

type CreateNotificationRequest struct {
	UserID  uuid.UUID        `json:"user_id" validate:"required"`
	Message string           `json:"message" validate:"required"`
	Type    NotificationType `json:"type" validate:"required"`
}

type NotificationFilter struct {
	UserID    *uuid.UUID        `form:"user_id"`
	Type      *NotificationType `form:"type"`
	IsRead    *bool             `form:"is_read"`
	Page      int               `form:"page"`
	Limit     int               `form:"limit"`
	SortBy    string            `form:"sort_by"`
	SortOrder string            `form:"sort_order"`
}