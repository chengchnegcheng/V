package handlers

import (
	"net/http"
	"strconv"
	"v/auth"
	"v/database"
	"v/logger"
	"v/model"

	"github.com/gin-gonic/gin"
)

var (
	userLogger *logger.Logger
	userMgr    model.DB
	authMgr    *auth.Manager
)

func init() {
	// Initialize database manager
	userMgr = database.GetWrappedDB()

	// Initialize logger
	userLogger = logger.NewLogger()

	// Initialize auth manager
	authMgr = auth.New(userLogger, userMgr)
}

// HandleGetCurrentUser returns the current user's information
func HandleGetCurrentUser(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := userMgr.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// HandleUpdateCurrentUser updates the current user's information
func HandleUpdateCurrentUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	userID := c.GetInt64("user_id")
	user, err := userMgr.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Update user fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Password != nil {
		// Hash the new password
		hashedPassword, err := auth.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}

	if err := userMgr.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user information"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// HandleUpdatePassword updates the current user's password
func HandleUpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	userID := c.GetInt64("user_id")
	user, err := userMgr.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Verify old password
	if !auth.CheckPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// Hash the new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = hashedPassword
	if err := userMgr.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// HandleGetTraffic returns the user's traffic statistics
func HandleGetTraffic(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Get traffic summary
	var summary struct {
		TotalUpload   int64 `json:"total_upload"`
		TotalDownload int64 `json:"total_download"`
		TrafficLimit  int64 `json:"traffic_limit"`
		UsedTraffic   int64 `json:"used_traffic"`
	}

	user, err := userMgr.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	summary.TrafficLimit = user.TrafficLimit
	summary.UsedTraffic = user.TrafficUsed

	// Get recent traffic logs
	stats, err := userMgr.ListProtocolStatsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get traffic logs"})
		return
	}

	// Calculate total upload and download
	for _, stat := range stats {
		summary.TotalUpload += stat.Upload
		summary.TotalDownload += stat.Download
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
		"logs":    stats,
	})
}

// HandleListUsers returns a list of all users
func HandleListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, err := userMgr.ListUsers((page-1)*pageSize, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// HandleGetUser returns a specific user's information
func HandleGetUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := userMgr.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// HandleUpdateUser updates a specific user's information
func HandleUpdateUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := userMgr.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Password != nil {
		// Hash the new password
		hashedPassword, err := auth.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}
	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}
	if req.ExpireAt != nil {
		user.ExpireAt = req.ExpireAt
	}
	if req.TrafficLimit != nil {
		user.TrafficLimit = *req.TrafficLimit
	}

	if err := userMgr.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user information"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// HandleDeleteUser deletes a specific user
func HandleDeleteUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := userMgr.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
