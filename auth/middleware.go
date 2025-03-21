package auth

import (
	"net/http"
	"strings"
	"v/model"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	authService *Service
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *Service) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth ensures that the request is authenticated
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.getTokenFromHeader(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// TODO: Validate token and get user
		// user, err := m.authService.ValidateToken(token)
		// if err != nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		// 	c.Abort()
		// 	return
		// }

		// c.Set("user", user)
		c.Next()
	}
}

// RequireAdmin ensures that the request is from an admin user
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !user.(*model.User).IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission ensures that the user has the required permission
func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// TODO: Check user permissions
		if !m.hasPermission(user.(*model.User), permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getTokenFromHeader extracts the token from the Authorization header
func (m *AuthMiddleware) getTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// hasPermission checks if a user has a specific permission
func (m *AuthMiddleware) hasPermission(user *model.User, permission string) bool {
	// TODO: Implement permission checking logic
	return user.IsAdmin
}
