package backup

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"v/config"
	"v/logger"
	"v/model"
	"v/notification"
	"v/settings"
)

// Backup represents a system backup
type Backup struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// Manager represents a backup manager
type Manager struct {
	log        *logger.Logger
	settings   *settings.Manager
	notifier   *notification.Manager
	config     *config.Config
	backupPath string
	stopChan   chan struct{}
	db         model.DB
}

// New creates a new backup manager
func New(log *logger.Logger, settings *settings.Manager, notifier *notification.Manager, config *config.Config, db model.DB) *Manager {
	return &Manager{
		log:        log,
		settings:   settings,
		notifier:   notifier,
		config:     config,
		backupPath: filepath.Join("backups"),
		stopChan:   make(chan struct{}),
		db:         db,
	}
}

// Start starts the backup manager
func (m *Manager) Start() error {
	s := m.settings.Get()
	if !s.Backup.Enable {
		return nil
	}

	// Create backup directory
	if err := os.MkdirAll(m.backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Start backup routine
	go m.backupRoutine()

	m.log.Info("Backup manager started", logger.Fields{
		"backup_path": m.backupPath,
		"interval":    s.Backup.Interval,
		"retention":   s.Backup.Retention,
	})

	return nil
}

// Stop stops the backup manager
func (m *Manager) Stop() {
	close(m.stopChan)
}

// backupRoutine runs the backup routine
func (m *Manager) backupRoutine() {
	s := m.settings.Get()
	ticker := time.NewTicker(s.Backup.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			if err := m.CreateBackup(); err != nil {
				m.log.Error("Failed to create backup", logger.Fields{
					"error": err,
				})
			}
		}
	}
}

// CreateBackup creates a new backup
func (m *Manager) CreateBackup() error {
	backupID := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(m.backupPath, fmt.Sprintf("backup_%s.zip", backupID))

	// Create zip file
	file, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer file.Close()

	// Create zip writer
	writer := zip.NewWriter(file)
	defer writer.Close()

	// Backup database
	if err := m.backupDatabase(writer); err != nil {
		return fmt.Errorf("failed to backup database: %v", err)
	}

	// Backup certificates
	if err := m.backupCertificates(writer); err != nil {
		return fmt.Errorf("failed to backup certificates: %v", err)
	}

	// Backup configuration
	if err := m.backupConfiguration(writer); err != nil {
		return fmt.Errorf("failed to backup configuration: %v", err)
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get backup file info: %v", err)
	}

	// Create backup record
	backup := &model.Backup{
		Path:      backupFile,
		Size:      fileInfo.Size(),
		Status:    "completed",
		Timestamp: time.Now(),
	}

	// Save backup record
	if err := m.db.CreateBackup(backup); err != nil {
		return fmt.Errorf("failed to save backup record: %v", err)
	}

	// Clean up old backups
	if err := m.cleanupOldBackups(); err != nil {
		m.log.Error("Failed to cleanup old backups", logger.Fields{
			"error": err,
		})
	}

	// Send notification
	if err := m.notifier.SendBackupNotification(true, backupFile, fileInfo.Size()); err != nil {
		m.log.Error("Failed to send backup notification", logger.Fields{
			"error": err,
		})
	}

	m.log.Info("Backup created successfully", logger.Fields{
		"backup_id": backupID,
		"path":      backupFile,
		"size":      fileInfo.Size(),
	})

	return nil
}

// backupDatabase backs up the database
func (m *Manager) backupDatabase(writer *zip.Writer) error {
	// Get database configuration
	dbConfig := m.config.Database

	// Create database backup file
	dbFile, err := writer.Create("database.sql")
	if err != nil {
		return fmt.Errorf("failed to create database backup file: %v", err)
	}

	// Execute database dump
	cmd := fmt.Sprintf("pg_dump -h %s -p %d -U %s -d %s -Fp",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Name)

	if dbConfig.SSLMode != "" {
		cmd += fmt.Sprintf(" --sslmode=%s", dbConfig.SSLMode)
	}

	// Set PGPASSWORD environment variable
	os.Setenv("PGPASSWORD", dbConfig.Password)

	// Execute command
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("failed to execute database dump: %v", err)
	}

	// Write database dump to zip file
	if _, err := dbFile.Write(output); err != nil {
		return fmt.Errorf("failed to write database dump: %v", err)
	}

	return nil
}

// backupCertificates backs up SSL certificates
func (m *Manager) backupCertificates(writer *zip.Writer) error {
	certDir := m.config.SSL.CertDir
	if certDir == "" {
		return nil
	}

	// Walk through certificate directory
	return filepath.Walk(certDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Create file in zip
		relPath, err := filepath.Rel(certDir, path)
		if err != nil {
			return err
		}

		file, err := writer.Create(filepath.Join("certificates", relPath))
		if err != nil {
			return err
		}

		// Copy file content
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(file, src)
		return err
	})
}

// backupConfiguration backs up system configuration
func (m *Manager) backupConfiguration(writer *zip.Writer) error {
	// Create configuration file
	configFile, err := writer.Create("config.json")
	if err != nil {
		return fmt.Errorf("failed to create configuration backup file: %v", err)
	}

	// Get current configuration
	config := m.settings.Get()

	// Marshal configuration
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}

	// Write configuration to zip file
	if _, err := configFile.Write(data); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	return nil
}

// cleanupOldBackups removes old backups
func (m *Manager) cleanupOldBackups() error {
	s := m.settings.Get()
	recordFile := filepath.Join(m.backupPath, "backups.json")

	// Read backup records
	data, err := os.ReadFile(recordFile)
	if err != nil {
		return fmt.Errorf("failed to read backup records: %v", err)
	}

	var records []*model.Backup
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to unmarshal backup records: %v", err)
	}

	// Sort records by timestamp (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	// Remove old backups
	for i := s.Backup.Retention; i < len(records); i++ {
		backup := records[i]
		if err := os.Remove(backup.Path); err != nil {
			m.log.Error("Failed to remove old backup", logger.Fields{
				"backup_id": backup.ID,
				"path":      backup.Path,
				"error":     err,
			})
		}
		records = append(records[:i], records[i+1:]...)
		i--
	}

	// Save updated records
	data, err = json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup records: %v", err)
	}

	if err := os.WriteFile(recordFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup records: %v", err)
	}

	return nil
}

// ListBackups returns a list of backups
func (m *Manager) ListBackups() ([]*model.Backup, error) {
	return m.db.ListBackups(0, 10)
}

// RestoreBackup restores a backup
func (m *Manager) RestoreBackup(backupID int64) error {
	// Get backup record
	backup, err := m.db.GetBackup(backupID)
	if err != nil {
		return err
	}

	if backup == nil {
		return fmt.Errorf("backup not found: %d", backupID)
	}

	// Open backup file
	reader, err := zip.OpenReader(backup.Path)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer reader.Close()

	// Restore database
	if err := m.restoreDatabase(reader); err != nil {
		return fmt.Errorf("failed to restore database: %v", err)
	}

	// Restore certificates
	if err := m.restoreCertificates(reader); err != nil {
		return fmt.Errorf("failed to restore certificates: %v", err)
	}

	// Restore configuration
	if err := m.restoreConfiguration(reader); err != nil {
		return fmt.Errorf("failed to restore configuration: %v", err)
	}

	m.log.Info("Backup restored successfully", logger.Fields{
		"backup_id": backupID,
	})

	return nil
}

// restoreDatabase restores the database from backup
func (m *Manager) restoreDatabase(reader *zip.ReadCloser) error {
	// Find database backup file
	var dbFile *zip.File
	for _, f := range reader.File {
		if f.Name == "database.sql" {
			dbFile = f
			break
		}
	}

	if dbFile == nil {
		return fmt.Errorf("database backup file not found")
	}

	// Open database backup file
	rc, err := dbFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open database backup file: %v", err)
	}
	defer rc.Close()

	// Read database dump
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read database backup: %v", err)
	}

	// Get database configuration
	dbConfig := m.config.Database

	// Set PGPASSWORD environment variable
	os.Setenv("PGPASSWORD", dbConfig.Password)

	// Execute database restore
	cmd := fmt.Sprintf("psql -h %s -p %d -U %s -d %s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Name)

	if dbConfig.SSLMode != "" {
		cmd += fmt.Sprintf(" --sslmode=%s", dbConfig.SSLMode)
	}

	// Execute command
	process := exec.Command("sh", "-c", cmd)
	process.Stdin = bytes.NewReader(data)
	if output, err := process.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to execute database restore: %v\n%s", err, output)
	}

	return nil
}

// restoreCertificates restores SSL certificates from backup
func (m *Manager) restoreCertificates(reader *zip.ReadCloser) error {
	certDir := m.config.SSL.CertDir
	if certDir == "" {
		return nil
	}

	// Remove existing certificates
	if err := os.RemoveAll(certDir); err != nil {
		return fmt.Errorf("failed to remove existing certificates: %v", err)
	}

	// Create certificate directory
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %v", err)
	}

	// Extract certificate files
	for _, f := range reader.File {
		if !strings.HasPrefix(f.Name, "certificates/") {
			continue
		}

		// Create file
		path := filepath.Join(certDir, strings.TrimPrefix(f.Name, "certificates/"))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create certificate directory: %v", err)
		}

		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create certificate file: %v", err)
		}

		// Copy file content
		rc, err := f.Open()
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to open certificate file: %v", err)
		}

		_, err = io.Copy(file, rc)
		rc.Close()
		file.Close()

		if err != nil {
			return fmt.Errorf("failed to copy certificate file: %v", err)
		}
	}

	return nil
}

// restoreConfiguration restores system configuration from backup
func (m *Manager) restoreConfiguration(reader *zip.ReadCloser) error {
	// Find configuration file
	var configFile *zip.File
	for _, f := range reader.File {
		if f.Name == "config.json" {
			configFile = f
			break
		}
	}

	if configFile == nil {
		return fmt.Errorf("configuration file not found")
	}

	// Open configuration file
	rc, err := configFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open configuration file: %v", err)
	}
	defer rc.Close()

	// Read configuration
	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %v", err)
	}

	// Unmarshal configuration
	var config settings.Settings
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %v", err)
	}

	// Update configuration
	if err := m.settings.Update(&config); err != nil {
		return fmt.Errorf("failed to update configuration: %v", err)
	}

	return nil
}

// DeleteBackup deletes a backup
func (m *Manager) DeleteBackup(backupID int64) error {
	// Get backup record
	backup, err := m.db.GetBackup(backupID)
	if err != nil {
		return err
	}

	// Delete backup file
	if err := os.Remove(backup.Path); err != nil {
		return fmt.Errorf("failed to delete backup file: %v", err)
	}

	// Delete backup info from database
	return m.db.DeleteBackup(backupID)
}
