package logger

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"v/model"
)

// Manager 日志管理器
type Manager struct {
	log    *Logger
	db     model.DB
	stopCh chan struct{}
}

// NewManager 创建日志管理器
func NewManager(log *Logger, db model.DB) *Manager {
	return &Manager{
		log:    log,
		db:     db,
		stopCh: make(chan struct{}),
	}
}

// Start 启动日志管理器
func (m *Manager) Start() error {
	m.log.Info("Starting log manager", Fields{})
	go m.cleanupLoop()
	return nil
}

// Stop 停止日志管理器
func (m *Manager) Stop() error {
	m.log.Info("Stopping log manager", Fields{})
	close(m.stopCh)
	return nil
}

// cleanupLoop 清理循环
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// 删除 30 天前的日志
			if err := m.db.DeleteLogsBefore(time.Now().Add(-30 * 24 * time.Hour)); err != nil {
				m.log.Error("Failed to cleanup logs", Fields{
					"error": err.Error(),
				})
			}
		}
	}
}

// Log 记录日志
func (m *Manager) Log(level, module, message string, details interface{}, userID int64, username, ip, userAgent string) error {
	// 将 details 转换为 JSON 字符串
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %v", err)
	}

	log := &model.Log{
		Level:     level,
		Module:    module,
		Message:   message,
		Details:   string(detailsJSON),
		UserID:    userID,
		Username:  username,
		IP:        ip,
		UserAgent: userAgent,
	}

	if err := m.db.CreateLog(log); err != nil {
		return fmt.Errorf("failed to create log: %v", err)
	}

	return nil
}

// ListLogs 获取日志列表
func (m *Manager) ListLogs(query *model.LogQuery) ([]*model.Log, error) {
	return m.db.ListLogs(query)
}

// GetTotalLogs 获取日志总数
func (m *Manager) GetTotalLogs(query *model.LogQuery) (int64, error) {
	return m.db.GetTotalLogs(query)
}

// ExportLogs 导出日志
func (m *Manager) ExportLogs(query *model.LogQuery) (string, error) {
	// 获取日志列表
	logs, err := m.db.ListLogs(query)
	if err != nil {
		return "", fmt.Errorf("failed to list logs: %v", err)
	}

	// 创建导出目录
	exportDir := "exports"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %v", err)
	}

	// 创建导出文件
	filename := fmt.Sprintf("logs_%s.csv", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(exportDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create export file: %v", err)
	}
	defer file.Close()

	// 创建 CSV 写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"ID", "Level", "Module", "Message", "Details", "IP", "UserAgent", "UserID", "Username", "CreatedAt"}
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write headers: %v", err)
	}

	// 写入数据
	for _, log := range logs {
		record := []string{
			fmt.Sprintf("%d", log.ID),
			log.Level,
			log.Module,
			log.Message,
			log.Details,
			log.IP,
			log.UserAgent,
			fmt.Sprintf("%d", log.UserID),
			log.Username,
			log.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write record: %v", err)
		}
	}

	return filepath, nil
}

// DeleteLogs 删除日志
func (m *Manager) DeleteLogs(query *model.LogQuery) error {
	logs, err := m.db.ListLogs(query)
	if err != nil {
		return fmt.Errorf("failed to list logs: %v", err)
	}

	for _, log := range logs {
		if err := m.db.DeleteLog(log.ID); err != nil {
			return fmt.Errorf("failed to delete log: %v", err)
		}
	}

	return nil
}
