package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"rtims-backend/config"
	"rtims-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtSecret []byte

func InitJWTSecret(cfg *config.Config) {
 	log.Printf("Setting JWT secret from config (length: %d)", len(cfg.JWTSecret))
 	jwtSecret = []byte(cfg.JWTSecret)
 	log.Println("JWT secret initialized successfully")
 }

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth() gin.HandlerFunc {
 	return func(c *gin.Context) {
 		authHeader := c.GetHeader("Authorization")
 		if authHeader == "" {
 			log.Printf("JWT Auth: Missing Authorization header for request to %s", c.Request.URL.Path)
 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
 			c.Abort()
 			return
 		}

 		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
 		if tokenString == authHeader {
 			log.Printf("JWT Auth: Bearer token missing for request to %s", c.Request.URL.Path)
 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
 			c.Abort()
 			return
 		}

 		log.Printf("JWT Auth: Validating token for request to %s", c.Request.URL.Path)
 	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
 			return jwtSecret, nil
 		})

 		if err != nil {
 			log.Printf("JWT Auth: Token parsing failed for request to %s: %v", c.Request.URL.Path, err)

 			// Provide specific error messages based on the type of error
 			var errorMessage string
 			if strings.Contains(err.Error(), "expired") {
 				errorMessage = "Token has expired"
 			} else if strings.Contains(err.Error(), "malformed") {
 				errorMessage = "Token is malformed"
 			} else if strings.Contains(err.Error(), "signature") {
 				errorMessage = "Invalid token signature"
 			} else {
 				errorMessage = "Invalid token"
 			}

 			c.JSON(http.StatusUnauthorized, gin.H{
 				"error":   errorMessage,
 				"details": err.Error(),
 			})
 			c.Abort()
 			return
 		}

 		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
 			log.Printf("JWT Auth: Token validated successfully for user %s (role: %s) accessing %s", claims.Email, claims.Role, c.Request.URL.Path)
 			c.Set("user_id", claims.UserID)
 			c.Set("email", claims.Email)
 			c.Set("role", claims.Role)
 			c.Next()
 		} else {
 			log.Printf("JWT Auth: Invalid token claims for request to %s", c.Request.URL.Path)
 			c.JSON(http.StatusUnauthorized, gin.H{
 				"error":   "Invalid token claims",
 				"details": "Token claims could not be validated",
 			})
 			c.Abort()
 			return
 		}
 	}
 }

func AdminOnly() gin.HandlerFunc {
 	return func(c *gin.Context) {
 	role, exists := c.Get("role")
 	if !exists {
 		log.Printf("AdminOnly: User role not found for request to %s", c.Request.URL.Path)
 		c.JSON(http.StatusUnauthorized, gin.H{
 			"error":   "User role not found",
 			"details": "Authentication required before accessing admin resources",
 		})
 		c.Abort()
 		return
 	}

 	userRole, ok := role.(models.UserRole)
 	if !ok {
 		log.Printf("AdminOnly: Invalid role type for request to %s", c.Request.URL.Path)
 		c.JSON(http.StatusUnauthorized, gin.H{
 			"error":   "Invalid user role",
 			"details": "User role could not be determined",
 		})
 		c.Abort()
 		return
 	}

 	if userRole != models.RoleAdmin {
 		log.Printf("AdminOnly: Access denied for user with role %v (not admin) accessing %s", userRole, c.Request.URL.Path)
 		c.JSON(http.StatusForbidden, gin.H{
 			"error":   "Admin access required",
 			"details": fmt.Sprintf("User role '%s' does not have admin privileges", userRole),
 			"role":    userRole,
 		})
 		c.Abort()
 		return
 	}

 		log.Printf("AdminOnly: Admin access granted for user with role %v accessing %s", userRole, c.Request.URL.Path)
 		c.Next()
 	}
 }

func GetCurrentUser(c *gin.Context) (uuid.UUID, models.UserRole, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, "", fmt.Errorf("user ID not found in context")
	}

	role, exists := c.Get("role")
	if !exists {
		return uuid.Nil, "", fmt.Errorf("user role not found in context")
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, "", fmt.Errorf("invalid user ID type")
	}

	userRole, ok := role.(models.UserRole)
	if !ok {
		return uuid.Nil, "", fmt.Errorf("invalid user role type")
	}

	return userUUID, userRole, nil
}