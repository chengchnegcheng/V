package auth

import (
	"net/http"
	"v/model"

	"github.com/gin-gonic/gin"
)

// Handler handles authentication-related HTTP requests
type Handler struct {
	authService *Service
}

// NewHandler creates a new authentication handler
func NewHandler(authService *Service) *Handler {
	return &Handler{authService: authService}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User  *model.User `json:"user"`
	Token string      `json:"token"`
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		case ErrAccountLocked:
			c.JSON(http.StatusForbidden, gin.H{"error": "account is locked"})
		case ErrAccountExpired:
			c.JSON(http.StatusForbidden, gin.H{"error": "account has expired"})
		case ErrAccountDisabled:
			c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		User:  user,
		Token: token,
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		if err := h.authService.Logout(token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ChangePassword(user.(*model.User).ID, req.OldPassword, req.NewPassword); err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// ResetPassword handles password reset
func (h *Handler) ResetPassword(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.authService.ResetPassword(user.(*model.User).ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}
