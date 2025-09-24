package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"rtims-backend/internal/models"
	"rtims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

func InitAuthHandlers(secret []byte) {
	jwtSecret = secret
}

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      models.RoleStaff,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// TODO: Save to database
	// This would be implemented when we create the database service

	// Generate tokens
	accessToken, refreshToken := generateTokens(user)

	response := models.AuthResponse{
		User:        user,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600, // 1 hour
	}

	c.JSON(http.StatusCreated, response)
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user from database
	// This would be implemented when we create the database service
	var user models.User
	var hashedPassword string

	// For now, create a mock user for demonstration
	if req.Email == "admin@example.com" && req.Password == "admin123" {
		user = models.User{
			ID:       uuid.New(),
			Name:     "Admin User",
			Email:    req.Email,
			Role:     models.RoleAdmin,
			IsActive: true,
		}
		hashedPassword = "$2a$10$mockhashedpassword"
	} else if req.Email == "staff@example.com" && req.Password == "staff123" {
		user = models.User{
			ID:       uuid.New(),
			Name:     "Staff User",
			Email:    req.Email,
			Role:     models.RoleStaff,
			IsActive: true,
		}
		hashedPassword = "$2a$10$mockhashedpassword"
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, refreshToken := generateTokens(user)

	response := models.AuthResponse{
		User:        user,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600, // 1 hour
	}

	// TODO: Save refresh token to Redis
	// This would be implemented when we create the Redis service

	c.JSON(http.StatusOK, response)
}

func RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Validate refresh token from Redis
	// This would be implemented when we create the Redis service

	// For now, create a mock validation
	if req.RefreshToken == "mock-refresh-token" {
		// Get user from token claims
		user := models.User{
			ID:       uuid.New(),
			Name:     "Mock User",
			Email:    "user@example.com",
			Role:     models.RoleStaff,
			IsActive: true,
		}

		accessToken, refreshToken := generateTokens(user)

		response := models.AuthResponse{
			User:        user,
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}

		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
}

func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Generate password reset token and send email
	// This would be implemented when we create the email service

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

func ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Validate reset token and update password
	// This would be implemented when we create the database service

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func GetProfile(c *gin.Context) {
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Get user profile from database
	// This would be implemented when we create the database service

	user := models.User{
		ID:       userID,
		Name:     "Mock User",
		Email:    "user@example.com",
		Role:     models.RoleStaff,
		IsActive: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, user)
}

func UpdateProfile(c *gin.Context) {
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Update user profile in database
	// This would be implemented when we create the database service

	user := models.User{
		ID:        userID,
		Name:      "Updated User",
		Email:     "user@example.com",
		Role:      models.RoleStaff,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, user)
}

func generateTokens(user models.User) (string, string) {
	// Generate access token (1 hour)
	accessClaims := models.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, _ := accessToken.SignedString(jwtSecret)

	// Generate refresh token (24 hours)
	refreshClaims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, _ := refreshToken.SignedString(jwtSecret)

	return accessTokenString, refreshTokenString
}