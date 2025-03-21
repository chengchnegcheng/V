package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"v/config"
	"v/logger"
	"v/model"
	"v/notification"
	"v/settings"
)

// Manager 备份管理器
type Manager struct {
	log         *logger.Logger
	db          model.DB
	backupDir   string
	settingsMgr *settings.Manager
	notifyMgr   *notification.Manager
	config      *config.Config
}

// New 创建备份管理器
func New(log *logger.Logger, settingsMgr *settings.Manager, notifyMgr *notification.Manager, cfg *config.Config, db model.DB) *Manager {
	return &Manager{
		log:         log,
		db:          db,
		backupDir:   cfg.BackupDir,
		settingsMgr: settingsMgr,
		notifyMgr:   notifyMgr,
		config:      cfg,
	}
}

// CreateBackup 创建备份
func (m *Manager) CreateBackup() (*model.Backup, error) {
	// 创建备份目录
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %v", err)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.json", timestamp)
	filepath := filepath.Join(m.backupDir, filename)

	// 创建备份文件
	file, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %v", err)
	}
	defer file.Close()

	// 获取所有需要备份的数据
	backupData := struct {
		Users        []*model.User          `json:"users"`
		Protocols    []*model.Protocol      `json:"protocols"`
		Certificates []*model.Certificate   `json:"certificates"`
		Settings     map[string]interface{} `json:"settings"`
	}{}

	// 获取用户数据
	users, err := m.db.ListUsers(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	backupData.Users = users

	// 获取协议数据
	protocols, err := m.db.ListProtocols(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get protocols: %v", err)
	}
	backupData.Protocols = protocols

	// 获取证书数据
	certificates, err := m.db.ListCertificates()
	if err != nil {
		return nil, fmt.Errorf("failed to get certificates: %v", err)
	}
	backupData.Certificates = certificates

	// 获取系统设置
	settings, err := m.settingsMgr.GetAllSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %v", err)
	}
	backupData.Settings = settings

	// 将数据写入备份文件
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backupData); err != nil {
		return nil, fmt.Errorf("failed to write backup data: %v", err)
	}

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// 创建备份记录
	backup := &model.Backup{
		Path:      filepath,
		Size:      fileInfo.Size(),
		Status:    "completed",
		Timestamp: time.Now(),
	}

	// 记录备份信息
	m.log.Info("Backup created successfully", logger.Fields{
		"path":      backup.Path,
		"size":      backup.Size,
		"timestamp": backup.Timestamp,
	})

	// 发送通知
	m.notifyMgr.SendNotification("backup_created", map[string]interface{}{
		"path":      backup.Path,
		"size":      backup.Size,
		"timestamp": backup.Timestamp,
	})

	return backup, nil
}

// GetBackup 获取备份信息
func (m *Manager) GetBackup(backupID int64) (*model.Backup, error) {
	// 获取备份列表
	backups, err := m.ListBackups()
	if err != nil {
		return nil, err
	}

	// 查找指定ID的备份
	for _, backup := range backups {
		if backup.ID == backupID {
			return backup, nil
		}
	}

	return nil, fmt.Errorf("backup not found: %d", backupID)
}

// RestoreBackup 恢复备份
func (m *Manager) RestoreBackup(backupID int64) error {
	// 获取备份信息
	backup, err := m.GetBackup(backupID)
	if err != nil {
		return err
	}

	// 检查备份文件是否存在
	if _, err := os.Stat(backup.Path); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backup.Path)
	}

	// 打开备份文件
	file, err := os.Open(backup.Path)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer file.Close()

	// 读取备份数据
	var backupData struct {
		Users        []*model.User          `json:"users"`
		Protocols    []*model.Protocol      `json:"protocols"`
		Certificates []*model.Certificate   `json:"certificates"`
		Settings     map[string]interface{} `json:"settings"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&backupData); err != nil {
		return fmt.Errorf("failed to decode backup data: %v", err)
	}

	// 开始事务
	if err := m.db.Begin(); err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// 恢复用户数据
	for _, user := range backupData.Users {
		if err := m.db.CreateUser(user); err != nil {
			m.db.Rollback()
			return fmt.Errorf("failed to restore user: %v", err)
		}
	}

	// 恢复协议数据
	for _, protocol := range backupData.Protocols {
		if err := m.db.CreateProtocol(protocol); err != nil {
			m.db.Rollback()
			return fmt.Errorf("failed to restore protocol: %v", err)
		}
	}

	// 恢复证书数据
	for _, cert := range backupData.Certificates {
		if err := m.db.CreateCertificate(cert); err != nil {
			m.db.Rollback()
			return fmt.Errorf("failed to restore certificate: %v", err)
		}
	}

	// 恢复系统设置
	if err := m.settingsMgr.UpdateSettings(backupData.Settings); err != nil {
		m.db.Rollback()
		return fmt.Errorf("failed to restore settings: %v", err)
	}

	// 提交事务
	if err := m.db.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// 记录恢复信息
	m.log.Info("Backup restored successfully", logger.Fields{
		"backup_id": backupID,
		"timestamp": time.Now(),
	})

	// 发送通知
	m.notifyMgr.SendNotification("backup_restored", map[string]interface{}{
		"backup_id": backupID,
		"timestamp": time.Now(),
	})

	return nil
}

// ListBackups 获取备份列表
func (m *Manager) ListBackups() ([]*model.Backup, error) {
	// 获取备份目录下的所有文件
	files, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %v", err)
	}

	var backups []*model.Backup
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filepath := filepath.Join(m.backupDir, file.Name())
		fileInfo, err := file.Info()
		if err != nil {
			continue
		}

		backup := &model.Backup{
			Path:      filepath,
			Size:      fileInfo.Size(),
			Status:    "completed",
			Timestamp: fileInfo.ModTime(),
		}
		backups = append(backups, backup)
	}

	return backups, nil
}

// DeleteBackup 删除备份
func (m *Manager) DeleteBackup(backupID int64) error {
	// 获取备份信息
	backup, err := m.GetBackup(backupID)
	if err != nil {
		return err
	}

	// 删除备份文件
	if err := os.Remove(backup.Path); err != nil {
		return fmt.Errorf("failed to delete backup file: %v", err)
	}

	// 记录删除信息
	m.log.Info("Backup deleted successfully", logger.Fields{
		"backup_id": backupID,
		"path":      backup.Path,
	})

	// 发送通知
	m.notifyMgr.SendNotification("backup_deleted", map[string]interface{}{
		"backup_id": backupID,
		"path":      backup.Path,
	})

	return nil
}

// DownloadBackup 下载备份文件
func (m *Manager) DownloadBackup(backupID int64, writer io.Writer) error {
	// 获取备份信息
	backup, err := m.GetBackup(backupID)
	if err != nil {
		return err
	}

	// 打开备份文件
	file, err := os.Open(backup.Path)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer file.Close()

	// 复制文件内容到写入器
	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("failed to copy backup file: %v", err)
	}

	return nil
}
