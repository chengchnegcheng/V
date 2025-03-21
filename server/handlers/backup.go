package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"

	"v/backup"
	"v/config"
	"v/logger"
	"v/model"
	"v/notification"
	"v/settings"

	"github.com/gin-gonic/gin"
)

var backupMgr *backup.Manager

// InitBackupHandlers 初始化备份处理器
func InitBackupHandlers(log *logger.Logger, settingsMgr *settings.Manager, notifyMgr *notification.Manager, cfg *config.Config, db model.DB) {
	backupMgr = backup.New(log, settingsMgr, notifyMgr, cfg, db)
}

// HandleCreateBackup 处理创建备份的请求
func HandleCreateBackup(c *gin.Context) {
	// 创建备份
	backup, err := backupMgr.CreateBackup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create backup",
		})
		return
	}

	c.JSON(http.StatusCreated, backup)
}

// HandleListBackups 处理获取备份列表的请求
func HandleListBackups(c *gin.Context) {
	// 获取备份列表
	backups, err := backupMgr.ListBackups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list backups",
		})
		return
	}

	c.JSON(http.StatusOK, backups)
}

// HandleRestoreBackup 处理恢复备份的请求
func HandleRestoreBackup(c *gin.Context) {
	// 获取备份ID
	backupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid backup ID",
		})
		return
	}

	// 恢复备份
	if err := backupMgr.RestoreBackup(backupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restore backup",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Backup restored successfully",
	})
}

// HandleDeleteBackup 处理删除备份的请求
func HandleDeleteBackup(c *gin.Context) {
	// 获取备份ID
	backupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid backup ID",
		})
		return
	}

	// 删除备份
	if err := backupMgr.DeleteBackup(backupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete backup",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Backup deleted successfully",
	})
}

// HandleDownloadBackup 处理下载备份的请求
func HandleDownloadBackup(c *gin.Context) {
	// 获取备份ID
	backupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid backup ID",
		})
		return
	}

	// 获取备份信息
	backup, err := backupMgr.GetBackup(backupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Backup not found",
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(backup.Path))

	// 下载备份文件
	if err := backupMgr.DownloadBackup(backupID, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to download backup",
		})
		return
	}
}
