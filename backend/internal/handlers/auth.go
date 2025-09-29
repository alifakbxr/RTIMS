package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"rtims-backend/internal/database"
	"rtims-backend/internal/models"
	"rtims-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte
var userService *database.UserService
var auditService *database.AuditService
var redisClient *redis.Client
var emailService *EmailService
var ctx = context.Background()

// Simple email service for sending password reset emails
type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (es *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	// TODO: Implement real email service integration
	// This should integrate with SMTP, SendGrid, AWS SES, or similar service
	// For now, return an error to indicate this needs to be implemented
	return fmt.Errorf("email service not implemented - please configure SMTP or email service provider")
}

func InitAuthHandlers(secret []byte, db *sql.DB, redis *redis.Client) {
	jwtSecret = secret
	userService = database.NewUserService(db)
	auditService = database.NewAuditService(db)
	redisClient = redis
	emailService = NewEmailService()
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

	// Save to database
	err = userService.CreateUser(&user)
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
		NewValues:  map[string]interface{}{"name": req.Name, "email": req.Email, "role": user.Role},
		ChangedBy:  user.ID, // User created themselves
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = auditService.CreateAuditLog(auditLog)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to create audit log: %v", err)
	}

	// Generate tokens
	accessToken, _, err := generateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

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

  // Get user from database
  user, err := userService.GetUserByEmail(req.Email)
  if err != nil {
  	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
  	return
  }

  // Verify password against hashed password in database
  err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
  if err != nil {
  	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
  	return
  }

  // Check if user is active
  if !user.IsActive {
  	c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
  	return
  }

  // Generate tokens
  accessToken, refreshTokenString, err := generateTokens(*user)
  if err != nil {
  	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
  	return
  }

  response := models.AuthResponse{
  	User:        *user,
  	AccessToken: accessToken,
  	TokenType:   "Bearer",
  	ExpiresIn:   3600, // 1 hour
  }

  // Save refresh token to Redis (24 hours expiry)
  refreshTokenKey := "refresh_token:" + refreshTokenString
  err = redisClient.Set(ctx, refreshTokenKey, user.ID.String(), 24*time.Hour).Err()
  if err != nil {
  	log.Printf("Failed to save refresh token to Redis: %v", err)
  }

  c.JSON(http.StatusOK, response)
}

func RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token from Redis
		tokenKey := "refresh_token:" + req.RefreshToken
		userIDStr, err := redisClient.Get(ctx, tokenKey).Result()
	if err != nil || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Parse user ID from Redis
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Get user from database
	user, err := userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Generate new access token
	accessToken, _, err := generateTokens(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	response := models.AuthResponse{
		User:        *user,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	c.JSON(http.StatusOK, response)
}

func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate reset token
	resetToken := uuid.New().String()

	// Store token in Redis with 1 hour expiry
	resetTokenKey := "password_reset:" + resetToken
	err := redisClient.Set(ctx, resetTokenKey, req.Email, time.Hour).Err()
	if err != nil {
		log.Printf("Failed to store password reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password reset request"})
		return
	}

	// Send password reset email using the email service
	err = emailService.SendPasswordResetEmail(req.Email, resetToken)
	if err != nil {
		log.Printf("Failed to send password reset email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send password reset email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent successfully"})
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

	// Validate password strength
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		return
	}

	// Validate reset token from Redis
	resetTokenKey := "password_reset:" + req.Token
	email, err := redisClient.Get(ctx, resetTokenKey).Result()
	if err != nil || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Get user by email
	user, err := userService.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update user password in database
	updates := map[string]interface{}{
		"password": string(hashedPassword),
	}
	err = userService.UpdateUser(user.ID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password: " + err.Error()})
		return
	}

	// Delete used reset token
	redisClient.Del(ctx, resetTokenKey)

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TableName:  "users",
		RecordID:   user.ID,
		Action:     models.ActionUpdate,
		OldValues:  map[string]interface{}{"password": "[REDACTED]"},
		NewValues:  map[string]interface{}{"password": "[REDACTED]"},
		ChangedBy:  user.ID,
		ChangedAt:  time.Now(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	err = auditService.CreateAuditLog(auditLog)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func GetProfile(c *gin.Context) {
	userID, _, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user profile from database
	user, err := userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
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

	// Get old user data for audit log
	oldUser, err := userService.GetUser(userID)
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

	// Update user profile in database
	err = userService.UpdateUser(userID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile: " + err.Error()})
		return
	}

	// Get updated user
	user, err := userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user: " + err.Error()})
		return
	}

	// Create audit log
	createAuditLog(c, "users", userID, models.ActionUpdate,
		map[string]interface{}{
			"name":     oldUser.Name,
			"email":    oldUser.Email,
			"role":     oldUser.Role,
			"is_active": oldUser.IsActive,
		},
		map[string]interface{}{
			"name":     user.Name,
			"email":    user.Email,
			"role":     user.Role,
			"is_active": user.IsActive,
		})

	c.JSON(http.StatusOK, user)
}

func generateTokens(user models.User) (string, string, error) {
 	// Generate access token (1 hour)
 	accessClaims := models.Claims{
 		UserID: user.ID,
 		Email:  user.Email,
 		Role:   user.Role,
 		RegisteredClaims: jwt.RegisteredClaims{
 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
 			IssuedAt:  jwt.NewNumericDate(time.Now()),
 			Subject:   user.ID.String(),
 		},
 	}

 	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
 	accessTokenString, err := accessToken.SignedString(jwtSecret)
 	if err != nil {
 		return "", "", fmt.Errorf("failed to generate access token: %w", err)
 	}

 	// Generate refresh token (24 hours) - using different secret for security
 	refreshClaims := jwt.RegisteredClaims{
 		Subject:   user.ID.String(),
 		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
 		IssuedAt:  jwt.NewNumericDate(time.Now()),
 		ID:       uuid.New().String(), // Unique token ID
 	}

 	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
 	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
 	if err != nil {
 		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
 	}

 	return accessTokenString, refreshTokenString, nil
 }