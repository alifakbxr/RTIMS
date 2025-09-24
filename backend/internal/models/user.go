package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleStaff UserRole = "staff"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required,min=2,max=100"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Password  string    `json:"-" db:"password" validate:"required,min=8"`
	Role      UserRole  `json:"role" db:"role" validate:"required,oneof=staff admin"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

type CreateUserRequest struct {
	Name     string   `json:"name" validate:"required,min=2,max=100"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Role     UserRole `json:"role" validate:"required,oneof=staff admin"`
}

type UpdateUserRequest struct {
	Name     *string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email    *string   `json:"email,omitempty" validate:"omitempty,email"`
	Role     *UserRole `json:"role,omitempty" validate:"omitempty,oneof=staff admin"`
	IsActive *bool     `json:"is_active,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}