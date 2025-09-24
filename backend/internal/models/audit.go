package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	ActionCreate  AuditAction = "create"
	ActionUpdate  AuditAction = "update"
	ActionDelete  AuditAction = "delete"
	ActionLogin   AuditAction = "login"
	ActionLogout  AuditAction = "logout"
	ActionView    AuditAction = "view"
)

type AuditLog struct {
	ID         uuid.UUID            `json:"id" db:"id"`
	TableName  string               `json:"table_name" db:"table_name" validate:"required"`
	RecordID   uuid.UUID            `json:"record_id" db:"record_id"`
	Action     AuditAction          `json:"action" db:"action" validate:"required"`
	OldValues  map[string]interface{} `json:"old_values" db:"old_values"`
	NewValues  map[string]interface{} `json:"new_values" db:"new_values"`
	ChangedBy  uuid.UUID            `json:"changed_by" db:"changed_by"`
	ChangedAt  time.Time            `json:"changed_at" db:"changed_at"`
	IPAddress  string               `json:"ip_address" db:"ip_address"`
	UserAgent  string               `json:"user_agent" db:"user_agent"`
}

type CreateAuditLogRequest struct {
	TableName string                 `json:"table_name" validate:"required"`
	RecordID  uuid.UUID              `json:"record_id"`
	Action    AuditAction            `json:"action" validate:"required"`
	OldValues map[string]interface{} `json:"old_values,omitempty"`
	NewValues map[string]interface{} `json:"new_values,omitempty"`
	ChangedBy uuid.UUID              `json:"changed_by" validate:"required"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
}

type AuditLogFilter struct {
	TableName *string      `form:"table_name"`
	Action    *AuditAction `form:"action"`
	ChangedBy *uuid.UUID   `form:"changed_by"`
	StartDate *time.Time   `form:"start_date"`
	EndDate   *time.Time   `form:"end_date"`
	Page      int          `form:"page"`
	Limit     int          `form:"limit"`
	SortBy    string       `form:"sort_by"`
	SortOrder string       `form:"sort_order"`
}