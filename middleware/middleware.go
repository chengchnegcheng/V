package middleware

import (
	"net/http"
	"strings"
	"time"

	"v/errors"
	"v/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// AuthMiddleware handles authentication
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Check token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		// Validate token
		token := parts[1]
		claims, err := validateToken(token, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			return
		}
		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting
type RateLimitMiddleware struct {
	limiter *rate.Limiter
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(r rate.Limit, b int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: rate.NewLimiter(r, b),
	}
}

// RateLimit handles rate limiting
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// LogMiddleware logs request details
func LogMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(start)
		log.Info("Request completed", logger.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})
	}
}

// ErrorMiddleware handles errors
func ErrorMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			log.Error("Request error", logger.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"error":  err.Error(),
			})

			// Handle custom errors
			if customErr, ok := err.(*errors.Error); ok {
				c.JSON(customErr.Code, gin.H{
					"error": customErr.Message,
				})
				return
			}

			// Handle other errors
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered", logger.Fields{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"error":  err,
				})

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()

		c.Next()
	}
}
