package middleware

import (
	"time"

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
 	// Simple in-memory rate limiting with cleanup
 	// In production, use Redis for distributed rate limiting
 	limiter := make(map[string][]int64)
 	lastCleanup := time.Now()

 	return func(c *gin.Context) {
 		// Get client IP
 		clientIP := c.ClientIP()

 		// Cleanup old entries every 5 minutes to prevent memory leaks
 		now := time.Now().Unix()
 		if now-lastCleanup > 300 {
 			for ip, requests := range limiter {
 				var validRequests []int64
 				window := int64(60) // 1 minute window
 				for _, reqTime := range requests {
 					if now-reqTime < window {
 						validRequests = append(validRequests, reqTime)
 					}
 				}
 				if len(validRequests) == 0 {
 					delete(limiter, ip)
 				} else {
 					limiter[ip] = validRequests
 				}
 			}
 			lastCleanup = now
 		}

 		// Check rate limit (100 requests per minute)
 		limit := 100
 		window := int64(60) // 1 minute window

 		if requests, exists := limiter[clientIP]; exists {
 			// Remove old requests outside the window
 			var validRequests []int64
 			for _, reqTime := range requests {
 				if now-reqTime < window {
 					validRequests = append(validRequests, reqTime)
 				}
 			}

 			if len(validRequests) >= limit {
 				c.JSON(429, gin.H{"error": "Too many requests", "retry_after": 60})
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