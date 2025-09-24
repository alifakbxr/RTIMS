package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		c.Next()
	}
}

func RateLimit() gin.HandlerFunc {
	// Simple in-memory rate limiting
	// In production, use Redis for distributed rate limiting
	limiter := make(map[string][]int64)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Check rate limit (100 requests per minute)
		now := int64(60) // 1 minute window
		limit := 100

		if requests, exists := limiter[clientIP]; exists {
			// Remove old requests outside the window
			var validRequests []int64
			for _, reqTime := range requests {
				if now-reqTime < 60 {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) >= limit {
				c.JSON(429, gin.H{"error": "Too many requests"})
				c.Abort()
				return
			}

			limiter[clientIP] = append(validRequests, now)
		} else {
			limiter[clientIP] = []int64{now}
		}

		c.Next()
	}
}