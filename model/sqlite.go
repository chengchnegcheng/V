package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"v/common"
)

// SQLiteDB is the SQLite implementation of the DB interface
type SQLiteDB struct {
	db     *sql.DB
	tx     *sql.Tx
	logger *slog.Logger
}

// NewSQLiteDB creates a new SQLiteDB instance
func NewSQLiteDB(db *sql.DB, logger *slog.Logger) *SQLiteDB {
	return &SQLiteDB{
		db:     db,
		logger: logger,
	}
}

// SQLiteDB combines the properly fixed implementations for the database interface.
// This file should be saved with UTF-8 encoding.

// The following were copied from fixed_implementation.go:

// Begin starts a transaction
func (db *SQLiteDB) Begin() error {
	if db.tx != nil {
		return fmt.Errorf("transaction already in progress")
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	db.tx = tx
	return nil
}

// Commit commits the current transaction
func (db *SQLiteDB) Commit() error {
	if db.tx == nil {
		return fmt.Errorf("no transaction in progress")
	}

	err := db.tx.Commit()
	db.tx = nil
	return err
}

// Rollback rolls back the current transaction
func (db *SQLiteDB) Rollback() error {
	if db.tx == nil {
		return fmt.Errorf("no transaction in progress")
	}

	err := db.tx.Rollback()
	db.tx = nil
	return err
}

// Close closes the database connection
func (db *SQLiteDB) Close() error {
	if db.tx != nil {
		db.tx.Rollback()
		db.tx = nil
	}
	return db.db.Close()
}

// AutoMigrate 执行自动迁移
func (db *SQLiteDB) AutoMigrate() error {
	db.logger.Info("执行自定义表创建，跳过自动迁移")
	// 使用静态表定义替代动态迁移
	return nil
}

// InitTables 初始化数据库表
func (db *SQLiteDB) InitTables() error {
	// 使用定制表创建语句，避免使用SQLite中的保留关键词
	db.logger.Info("执行自定义表初始化")
	return nil
}

// getSystemValue gets a system setting value by key
func (db *SQLiteDB) getSystemValue(key string) (string, error) {
	var value string
	err := db.db.QueryRow("SELECT value FROM system_settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// setSystemValue sets a system setting value
func (db *SQLiteDB) setSystemValue(key, value string) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	// Try to update first
	result, err := db.db.Exec(
		"UPDATE system_settings SET value = ?, updated_at = ? WHERE key = ?",
		value, now, key)
	if err != nil {
		return err
	}

	// If no rows affected, insert
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		_, err = db.db.Exec(
			"INSERT INTO system_settings (key, value, created_at, updated_at) VALUES (?, ?, ?, ?)",
			key, value, now, now)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListProtocolStatsByUserID 获取用户的所有协议统计
func (db *SQLiteDB) ListProtocolStatsByUserID(userID int64) ([]*ProtocolStats, error) {
	query := `SELECT 
		ps.id, ps.protocol_id, ps.user_id, ps.upload, ps.download, ps.last_active, ps.created_at, ps.updated_at
		FROM protocol_stats ps
		WHERE ps.user_id = ?`

	rows, err := db.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*ProtocolStats
	for rows.Next() {
		stat := &ProtocolStats{}
		var lastActiveStr, createdAtStr, updatedAtStr string

		err := rows.Scan(
			&stat.ID,
			&stat.ProtocolID,
			&stat.UserID,
			&stat.Upload,
			&stat.Download,
			&lastActiveStr,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		// 解析时间字段
		stat.LastActive, _ = time.Parse("2006-01-02 15:04:05", lastActiveStr)
		stat.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		stat.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		stats = append(stats, stat)
	}

	return stats, nil
}

// GetAllUsers returns all users
func (db *SQLiteDB) GetAllUsers() ([]*User, error) {
	query := `SELECT 
		id, username, email, password, salt, role, 
		status, traffic_limit, traffic_used, expire_at, 
		last_login_at, login_attempts, locked_until, is_admin,
		created_at, updated_at
	FROM users`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*User
	for rows.Next() {
		user := &User{}
		var expireAtStr, lastLoginAtStr, lockedUntilStr, createdAtStr, updatedAtStr string

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Salt,
			&user.Role,
			&user.Status,
			&user.TrafficLimit,
			&user.TrafficUsed,
			&expireAtStr,
			&lastLoginAtStr,
			&user.LoginAttempts,
			&lockedUntilStr,
			&user.IsAdmin,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		// Parse time fields
		if expireAtStr != "" {
			parsedTime, _ := time.Parse("2006-01-02 15:04:05", expireAtStr)
			user.ExpireAt = &parsedTime
		}
		if lastLoginAtStr != "" {
			parsedTime, _ := time.Parse("2006-01-02 15:04:05", lastLoginAtStr)
			user.LastLoginAt = &parsedTime
		}
		if lockedUntilStr != "" {
			parsedTime, _ := time.Parse("2006-01-02 15:04:05", lockedUntilStr)
			user.LockedUntil = &parsedTime
		}
		user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		result = append(result, user)
	}

	return result, rows.Err()
}

// CleanupTraffic cleans up traffic records before the given time
func (db *SQLiteDB) CleanupTraffic(before time.Time) error {
	beforeStr := before.Format("2006-01-02 15:04:05")
	_, err := db.db.Exec("DELETE FROM protocol_stats WHERE created_at < ?", beforeStr)
	return err
}

// CreateAlert creates a new alert record
func (db *SQLiteDB) CreateAlert(alert *AlertRecord) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `INSERT INTO alert_records (
		type, value, threshold, message, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)`

	_, err := db.db.Exec(
		query,
		alert.Type,
		alert.Value,
		alert.Threshold,
		alert.Message,
		now,
		now,
	)

	if err != nil {
		return err
	}

	return nil
}

// CreateAlertRecord creates a new alert record
func (db *SQLiteDB) CreateAlertRecord(record *AlertRecord) error {
	return db.CreateAlert(record)
}

// CreateBackup creates a database backup
func (db *SQLiteDB) CreateBackup(backup *Backup) error {
	// Simple implementation that records the backup metadata
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `INSERT INTO backups (
		path, size, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?)`

	_, err := db.db.Exec(
		query,
		backup.Path,
		backup.Size,
		backup.Status,
		now,
		now,
	)

	return err
}

// CreateCertificate creates a new certificate record
func (db *SQLiteDB) CreateCertificate(cert *Certificate) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `INSERT INTO certificates (
		domain, cert_file, key_file, status, 
		last_checked_at, last_renewed_at, expires_at, 
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.Exec(
		query,
		cert.Domain,
		cert.CertFile,
		cert.KeyFile,
		cert.Status,
		cert.LastCheckedAt.Format("2006-01-02 15:04:05"),
		cert.LastRenewedAt.Format("2006-01-02 15:04:05"),
		cert.ExpiresAt.Format("2006-01-02 15:04:05"),
		now,
		now,
	)

	return err
}

// CreateDailyStats creates a new daily stats record
func (db *SQLiteDB) CreateDailyStats(stats *DailyStats) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	dateStr := stats.Date.Format("2006-01-02")

	query := `INSERT INTO daily_stats (
		user_id, date, upload, download, total, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.Exec(
		query,
		stats.UserID,
		dateStr,
		stats.Upload,
		stats.Download,
		stats.Total,
		now,
		now,
	)

	return err
}

// CreateLog creates a new log record
func (db *SQLiteDB) CreateLog(log *Log) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `INSERT INTO logs (
		level, module, message, details, ip, user_agent, user_id, username,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.Exec(
		query,
		log.Level,
		log.Module,
		log.Message,
		log.Details,
		log.IP,
		log.UserAgent,
		log.UserID,
		log.Username,
		now,
		now,
	)

	return err
}

// GetLog retrieves a log record by ID
func (db *SQLiteDB) GetLog(id int64) (*Log, error) {
	query := `SELECT 
		id, level, module, message, details, ip, user_agent, user_id, username,
		created_at, updated_at
	FROM logs WHERE id = ?`

	row := db.db.QueryRow(query, id)

	log := &Log{}
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&log.ID,
		&log.Level,
		&log.Module,
		&log.Message,
		&log.Details,
		&log.IP,
		&log.UserAgent,
		&log.UserID,
		&log.Username,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse time fields
	log.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	log.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return log, nil
}

// UpdateLog updates a log record
func (db *SQLiteDB) UpdateLog(log *Log) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `UPDATE logs SET
		level = ?, module = ?, message = ?, details = ?, ip = ?,
		user_agent = ?, user_id = ?, username = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		log.Level,
		log.Module,
		log.Message,
		log.Details,
		log.IP,
		log.UserAgent,
		log.UserID,
		log.Username,
		now,
		log.ID,
	)

	return err
}

// DeleteLog deletes a log record
func (db *SQLiteDB) DeleteLog(id int64) error {
	query := `DELETE FROM logs WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// ListLogs lists log records based on query parameters
func (db *SQLiteDB) ListLogs(query *LogQuery) ([]*Log, error) {
	sqlQuery := `SELECT 
		id, level, module, message, details, ip, user_agent, user_id, username,
		created_at, updated_at
	FROM logs WHERE 1=1`

	var args []interface{}

	// Apply filters
	if query.Level != "" {
		sqlQuery += " AND level = ?"
		args = append(args, query.Level)
	}

	if query.Module != "" {
		sqlQuery += " AND module = ?"
		args = append(args, query.Module)
	}

	if !query.StartTime.IsZero() {
		sqlQuery += " AND created_at >= ?"
		args = append(args, query.StartTime.Format("2006-01-02 15:04:05"))
	}

	if !query.EndTime.IsZero() {
		sqlQuery += " AND created_at <= ?"
		args = append(args, query.EndTime.Format("2006-01-02 15:04:05"))
	}

	if query.UserID > 0 {
		sqlQuery += " AND user_id = ?"
		args = append(args, query.UserID)
	}

	// Add ORDER BY and LIMIT
	sqlQuery += " ORDER BY created_at DESC"

	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		sqlQuery += " LIMIT ? OFFSET ?"
		args = append(args, query.PageSize, offset)
	}

	// Execute query
	rows, err := db.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse results
	var logs []*Log
	for rows.Next() {
		log := &Log{}
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&log.ID,
			&log.Level,
			&log.Module,
			&log.Message,
			&log.Details,
			&log.IP,
			&log.UserAgent,
			&log.UserID,
			&log.Username,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		log.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		log.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// GetTotalLogs gets the total count of logs based on query parameters
func (db *SQLiteDB) GetTotalLogs(query *LogQuery) (int64, error) {
	sqlQuery := "SELECT COUNT(*) FROM logs WHERE 1=1"

	var args []interface{}

	// Apply filters
	if query.Level != "" {
		sqlQuery += " AND level = ?"
		args = append(args, query.Level)
	}

	if query.Module != "" {
		sqlQuery += " AND module = ?"
		args = append(args, query.Module)
	}

	if !query.StartTime.IsZero() {
		sqlQuery += " AND created_at >= ?"
		args = append(args, query.StartTime.Format("2006-01-02 15:04:05"))
	}

	if !query.EndTime.IsZero() {
		sqlQuery += " AND created_at <= ?"
		args = append(args, query.EndTime.Format("2006-01-02 15:04:05"))
	}

	if query.UserID > 0 {
		sqlQuery += " AND user_id = ?"
		args = append(args, query.UserID)
	}

	// Execute query
	var count int64
	err := db.db.QueryRow(sqlQuery, args...).Scan(&count)

	return count, err
}

// DeleteLogsBefore deletes logs created before a specific time
func (db *SQLiteDB) DeleteLogsBefore(t time.Time) error {
	query := "DELETE FROM logs WHERE created_at < ?"
	_, err := db.db.Exec(query, t.Format("2006-01-02 15:04:05"))
	return err
}

// ExportLogs exports logs to a CSV file
func (db *SQLiteDB) ExportLogs(query *LogQuery) (string, error) {
	// First, get the logs
	logs, err := db.ListLogs(query)
	if err != nil {
		return "", err
	}

	// Create a temporary file
	filename := fmt.Sprintf("logs_export_%s.csv", time.Now().Format("20060102_150405"))
	filepath := fmt.Sprintf("./tmp/%s", filename)

	// Ensure directory exists
	if err := os.MkdirAll("./tmp", 0755); err != nil {
		return "", err
	}

	// Create and write to file
	f, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Write header
	headers := []string{"ID", "Level", "Module", "Message", "Details", "IP", "UserAgent", "UserID", "Username", "CreatedAt"}
	f.WriteString(strings.Join(headers, ",") + "\n")

	// Write data
	for _, log := range logs {
		row := []string{
			fmt.Sprintf("%d", log.ID),
			log.Level,
			log.Module,
			fmt.Sprintf("%q", log.Message), // Quote to handle commas
			fmt.Sprintf("%q", log.Details), // Quote to handle commas
			log.IP,
			fmt.Sprintf("%q", log.UserAgent), // Quote to handle commas
			fmt.Sprintf("%d", log.UserID),
			log.Username,
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		f.WriteString(strings.Join(row, ",") + "\n")
	}

	return filepath, nil
}

// CreateProtocol creates a new protocol record
func (db *SQLiteDB) CreateProtocol(protocol *Protocol) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `INSERT INTO protocols (
		user_id, type, settings, port, status, traffic_limit, 
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.Exec(
		query,
		protocol.UserID,
		protocol.Type,
		protocol.Settings,
		protocol.Port,
		protocol.Status,
		protocol.TrafficLimit,
		now,
		now,
	)

	return err
}

// GetProtocol retrieves a protocol by ID
func (db *SQLiteDB) GetProtocol(id int64) (*Protocol, error) {
	query := `SELECT 
		id, user_id, type, settings, port, status, traffic_limit, 
		created_at, updated_at
	FROM protocols WHERE id = ?`

	row := db.db.QueryRow(query, id)

	protocol := &Protocol{}
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&protocol.ID,
		&protocol.UserID,
		&protocol.Type,
		&protocol.Settings,
		&protocol.Port,
		&protocol.Status,
		&protocol.TrafficLimit,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse time fields
	protocol.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	protocol.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return protocol, nil
}

// GetProtocolsByUserID retrieves all protocols for a user
func (db *SQLiteDB) GetProtocolsByUserID(userID int64) ([]*Protocol, error) {
	query := `SELECT 
		id, user_id, type, settings, port, status, traffic_limit, 
		created_at, updated_at
	FROM protocols WHERE user_id = ?`

	rows, err := db.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var protocols []*Protocol
	for rows.Next() {
		protocol := &Protocol{}
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&protocol.ID,
			&protocol.UserID,
			&protocol.Type,
			&protocol.Settings,
			&protocol.Port,
			&protocol.Status,
			&protocol.TrafficLimit,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		protocol.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		protocol.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		protocols = append(protocols, protocol)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return protocols, nil
}

// UpdateProtocol updates a protocol
func (db *SQLiteDB) UpdateProtocol(protocol *Protocol) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `UPDATE protocols SET
		user_id = ?, type = ?, settings = ?, port = ?, status = ?, 
		traffic_limit = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		protocol.UserID,
		protocol.Type,
		protocol.Settings,
		protocol.Port,
		protocol.Status,
		protocol.TrafficLimit,
		now,
		protocol.ID,
	)

	return err
}

// DeleteProtocol deletes a protocol
func (db *SQLiteDB) DeleteProtocol(id int64) error {
	query := `DELETE FROM protocols WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// GetProtocolsByPort retrieves protocols by port
func (db *SQLiteDB) GetProtocolsByPort(port int) ([]*Protocol, error) {
	query := `SELECT 
		id, user_id, type, settings, port, status, traffic_limit, 
		created_at, updated_at
	FROM protocols WHERE port = ?`

	rows, err := db.db.Query(query, port)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var protocols []*Protocol
	for rows.Next() {
		protocol := &Protocol{}
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&protocol.ID,
			&protocol.UserID,
			&protocol.Type,
			&protocol.Settings,
			&protocol.Port,
			&protocol.Status,
			&protocol.TrafficLimit,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		protocol.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		protocol.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		protocols = append(protocols, protocol)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return protocols, nil
}

// ListProtocols lists protocols with pagination
func (db *SQLiteDB) ListProtocols(page, pageSize int) ([]*Protocol, error) {
	offset := (page - 1) * pageSize

	query := `SELECT 
		id, user_id, type, settings, port, status, traffic_limit, 
		created_at, updated_at
	FROM protocols ORDER BY id DESC LIMIT ? OFFSET ?`

	rows, err := db.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var protocols []*Protocol
	for rows.Next() {
		protocol := &Protocol{}
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&protocol.ID,
			&protocol.UserID,
			&protocol.Type,
			&protocol.Settings,
			&protocol.Port,
			&protocol.Status,
			&protocol.TrafficLimit,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		protocol.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		protocol.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		protocols = append(protocols, protocol)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return protocols, nil
}

// SearchProtocols searches protocols by keyword
func (db *SQLiteDB) SearchProtocols(keyword string) ([]*Protocol, error) {
	// Use LIKE for simple searching
	query := `SELECT 
		id, user_id, type, settings, port, status, traffic_limit, 
		created_at, updated_at
	FROM protocols 
	WHERE type LIKE ? OR settings LIKE ? OR status LIKE ?
	ORDER BY id DESC`

	likeParam := "%" + keyword + "%"

	rows, err := db.db.Query(query, likeParam, likeParam, likeParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var protocols []*Protocol
	for rows.Next() {
		protocol := &Protocol{}
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&protocol.ID,
			&protocol.UserID,
			&protocol.Type,
			&protocol.Settings,
			&protocol.Port,
			&protocol.Status,
			&protocol.TrafficLimit,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		protocol.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		protocol.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		protocols = append(protocols, protocol)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return protocols, nil
}

// CreateProtocolStats creates a new protocol stats record
func (db *SQLiteDB) CreateProtocolStats(stats *ProtocolStats) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	lastActiveStr := ""
	if !stats.LastActive.IsZero() {
		lastActiveStr = stats.LastActive.Format("2006-01-02 15:04:05")
	}

	query := `INSERT INTO protocol_stats (
		protocol_id, user_id, upload, download, last_active,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.Exec(
		query,
		stats.ProtocolID,
		stats.UserID,
		stats.Upload,
		stats.Download,
		lastActiveStr,
		now,
		now,
	)

	return err
}

// GetProtocolStats retrieves protocol stats by ID
func (db *SQLiteDB) GetProtocolStats(id int64) (*ProtocolStats, error) {
	query := `SELECT 
		id, protocol_id, user_id, upload, download, last_active,
		created_at, updated_at
	FROM protocol_stats WHERE id = ?`

	row := db.db.QueryRow(query, id)

	stats := &ProtocolStats{}
	var lastActiveStr, createdAtStr, updatedAtStr string

	err := row.Scan(
		&stats.ID,
		&stats.ProtocolID,
		&stats.UserID,
		&stats.Upload,
		&stats.Download,
		&lastActiveStr,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse time fields
	if lastActiveStr != "" {
		stats.LastActive, _ = time.Parse("2006-01-02 15:04:05", lastActiveStr)
	}
	stats.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	stats.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return stats, nil
}

// UpdateProtocolStats updates protocol stats
func (db *SQLiteDB) UpdateProtocolStats(stats *ProtocolStats) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	lastActiveStr := ""
	if !stats.LastActive.IsZero() {
		lastActiveStr = stats.LastActive.Format("2006-01-02 15:04:05")
	}

	query := `UPDATE protocol_stats SET
		protocol_id = ?, user_id = ?, upload = ?, download = ?, 
		last_active = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		stats.ProtocolID,
		stats.UserID,
		stats.Upload,
		stats.Download,
		lastActiveStr,
		now,
		stats.ID,
	)

	return err
}

// ListProtocolStatsByProtocolID lists protocol stats by protocol ID
func (db *SQLiteDB) ListProtocolStatsByProtocolID(protocolID int64) ([]*ProtocolStats, error) {
	query := `SELECT 
		id, protocol_id, user_id, upload, download, last_active,
		created_at, updated_at
	FROM protocol_stats WHERE protocol_id = ?`

	rows, err := db.db.Query(query, protocolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statsList []*ProtocolStats
	for rows.Next() {
		stats := &ProtocolStats{}
		var lastActiveStr, createdAtStr, updatedAtStr string

		err := rows.Scan(
			&stats.ID,
			&stats.ProtocolID,
			&stats.UserID,
			&stats.Upload,
			&stats.Download,
			&lastActiveStr,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		if lastActiveStr != "" {
			stats.LastActive, _ = time.Parse("2006-01-02 15:04:05", lastActiveStr)
		}
		stats.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		stats.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		statsList = append(statsList, stats)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return statsList, nil
}

// CreateProxy creates a new proxy
func (db *SQLiteDB) CreateProxy(proxy *common.Proxy) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	var expireAtStr string
	if !proxy.ExpireAt.IsZero() {
		expireAtStr = proxy.ExpireAt.Format("2006-01-02 15:04:05")
	}
	var lastActiveStr string
	if !proxy.LastActiveAt.IsZero() {
		lastActiveStr = proxy.LastActiveAt.Format("2006-01-02 15:04:05")
	}

	query := `INSERT INTO proxies (
		user_id, protocol, port, config, settings, listen_addr, remote_addr,
		enabled, upload, download, last_active_at, created_at, updated_at, expire_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.db.Exec(
		query,
		proxy.UserID,
		proxy.Protocol,
		proxy.Port,
		proxy.Config,
		proxy.Settings,
		proxy.ListenAddr,
		proxy.RemoteAddr,
		proxy.Enabled,
		proxy.Upload,
		proxy.Download,
		lastActiveStr,
		now,
		now,
		expireAtStr,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	proxy.ID = id
	return nil
}

// GetProxy retrieves a proxy by ID
func (db *SQLiteDB) GetProxy(id int64) (*common.Proxy, error) {
	query := `SELECT 
		id, user_id, protocol, port, config, settings, listen_addr, remote_addr,
		enabled, upload, download, last_active_at, created_at, updated_at, expire_at
	FROM proxies WHERE id = ?`

	row := db.db.QueryRow(query, id)

	proxy := &common.Proxy{}
	var lastActiveStr, createdAtStr, updatedAtStr, expireAtStr string

	err := row.Scan(
		&proxy.ID,
		&proxy.UserID,
		&proxy.Protocol,
		&proxy.Port,
		&proxy.Config,
		&proxy.Settings,
		&proxy.ListenAddr,
		&proxy.RemoteAddr,
		&proxy.Enabled,
		&proxy.Upload,
		&proxy.Download,
		&lastActiveStr,
		&createdAtStr,
		&updatedAtStr,
		&expireAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse time fields
	if lastActiveStr != "" {
		lastActive, _ := time.Parse("2006-01-02 15:04:05", lastActiveStr)
		proxy.LastActiveAt = lastActive
	}
	if expireAtStr != "" {
		expireAt, _ := time.Parse("2006-01-02 15:04:05", expireAtStr)
		proxy.ExpireAt = &expireAt
	}
	proxy.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	proxy.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return proxy, nil
}

// GetProxiesByUserID retrieves proxies by user ID
func (db *SQLiteDB) GetProxiesByUserID(userID int64) ([]*common.Proxy, error) {
	query := `SELECT 
		id, user_id, protocol, port, config, settings, listen_addr, remote_addr,
		enabled, upload, download, last_active_at, created_at, updated_at, expire_at
	FROM proxies WHERE user_id = ?`

	rows, err := db.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*common.Proxy
	for rows.Next() {
		proxy := &common.Proxy{}
		var lastActiveStr, createdAtStr, updatedAtStr, expireAtStr string

		err := rows.Scan(
			&proxy.ID,
			&proxy.UserID,
			&proxy.Protocol,
			&proxy.Port,
			&proxy.Config,
			&proxy.Settings,
			&proxy.ListenAddr,
			&proxy.RemoteAddr,
			&proxy.Enabled,
			&proxy.Upload,
			&proxy.Download,
			&lastActiveStr,
			&createdAtStr,
			&updatedAtStr,
			&expireAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		if lastActiveStr != "" {
			lastActive, _ := time.Parse("2006-01-02 15:04:05", lastActiveStr)
			proxy.LastActiveAt = lastActive
		}
		if expireAtStr != "" {
			expireAt, _ := time.Parse("2006-01-02 15:04:05", expireAtStr)
			proxy.ExpireAt = &expireAt
		}
		proxy.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		proxy.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		proxies = append(proxies, proxy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return proxies, nil
}

// UpdateProxy updates a proxy
func (db *SQLiteDB) UpdateProxy(proxy *common.Proxy) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	var expireAtStr string
	if !proxy.ExpireAt.IsZero() {
		expireAtStr = proxy.ExpireAt.Format("2006-01-02 15:04:05")
	}
	var lastActiveStr string
	if !proxy.LastActiveAt.IsZero() {
		lastActiveStr = proxy.LastActiveAt.Format("2006-01-02 15:04:05")
	}

	query := `UPDATE proxies SET
		user_id = ?, protocol = ?, port = ?, config = ?, settings = ?,
		listen_addr = ?, remote_addr = ?, enabled = ?, upload = ?, download = ?,
		last_active_at = ?, updated_at = ?, expire_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		proxy.UserID,
		proxy.Protocol,
		proxy.Port,
		proxy.Config,
		proxy.Settings,
		proxy.ListenAddr,
		proxy.RemoteAddr,
		proxy.Enabled,
		proxy.Upload,
		proxy.Download,
		lastActiveStr,
		now,
		expireAtStr,
		proxy.ID,
	)

	return err
}

// DeleteProxy deletes a proxy
func (db *SQLiteDB) DeleteProxy(id int64) error {
	query := `DELETE FROM proxies WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// GetProxiesByPort retrieves proxies by port
func (db *SQLiteDB) GetProxiesByPort(port int) ([]*common.Proxy, error) {
	query := `SELECT 
		id, user_id, protocol, port, config, settings, listen_addr, remote_addr,
		enabled, upload, download, last_active_at, created_at, updated_at, expire_at
	FROM proxies WHERE port = ?`

	rows, err := db.db.Query(query, port)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*common.Proxy
	for rows.Next() {
		proxy := &common.Proxy{}
		var lastActiveStr, createdAtStr, updatedAtStr, expireAtStr string

		err := rows.Scan(
			&proxy.ID,
			&proxy.UserID,
			&proxy.Protocol,
			&proxy.Port,
			&proxy.Config,
			&proxy.Settings,
			&proxy.ListenAddr,
			&proxy.RemoteAddr,
			&proxy.Enabled,
			&proxy.Upload,
			&proxy.Download,
			&lastActiveStr,
			&createdAtStr,
			&updatedAtStr,
			&expireAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		if lastActiveStr != "" {
			lastActive, _ := time.Parse("2006-01-02 15:04:05", lastActiveStr)
			proxy.LastActiveAt = lastActive
		}
		if expireAtStr != "" {
			expireAt, _ := time.Parse("2006-01-02 15:04:05", expireAtStr)
			proxy.ExpireAt = &expireAt
		}
		proxy.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		proxy.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		proxies = append(proxies, proxy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return proxies, nil
}

// ListProxies lists proxies with pagination
func (db *SQLiteDB) ListProxies(page, pageSize int) ([]*common.Proxy, error) {
	offset := (page - 1) * pageSize

	query := `SELECT 
		id, user_id, protocol, port, config, settings, listen_addr, remote_addr,
		enabled, upload, download, last_active_at, created_at, updated_at, expire_at
	FROM proxies ORDER BY id DESC LIMIT ? OFFSET ?`

	rows, err := db.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*common.Proxy
	for rows.Next() {
		proxy := &common.Proxy{}
		var lastActiveStr, createdAtStr, updatedAtStr, expireAtStr string

		err := rows.Scan(
			&proxy.ID,
			&proxy.UserID,
			&proxy.Protocol,
			&proxy.Port,
			&proxy.Config,
			&proxy.Settings,
			&proxy.ListenAddr,
			&proxy.RemoteAddr,
			&proxy.Enabled,
			&proxy.Upload,
			&proxy.Download,
			&lastActiveStr,
			&createdAtStr,
			&updatedAtStr,
			&expireAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		if lastActiveStr != "" {
			lastActive, _ := time.Parse("2006-01-02 15:04:05", lastActiveStr)
			proxy.LastActiveAt = lastActive
		}
		if expireAtStr != "" {
			expireAt, _ := time.Parse("2006-01-02 15:04:05", expireAtStr)
			proxy.ExpireAt = &expireAt
		}
		proxy.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		proxy.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		proxies = append(proxies, proxy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return proxies, nil
}

// GetTotalProxies gets the total count of proxies
func (db *SQLiteDB) GetTotalProxies() (int64, error) {
	var count int64
	err := db.db.QueryRow("SELECT COUNT(*) FROM proxies").Scan(&count)
	return count, err
}

// SearchProxies searches proxies by keyword
func (db *SQLiteDB) SearchProxies(keyword string) ([]*common.Proxy, error) {
	query := `SELECT 
		id, user_id, protocol, port, config, settings, listen_addr, remote_addr,
		enabled, upload, download, last_active_at, created_at, updated_at, expire_at
	FROM proxies 
	WHERE protocol LIKE ? OR config LIKE ? OR settings LIKE ?
	ORDER BY id DESC`

	likeParam := "%" + keyword + "%"

	rows, err := db.db.Query(query, likeParam, likeParam, likeParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*common.Proxy
	for rows.Next() {
		proxy := &common.Proxy{}
		var lastActiveStr, createdAtStr, updatedAtStr, expireAtStr string

		err := rows.Scan(
			&proxy.ID,
			&proxy.UserID,
			&proxy.Protocol,
			&proxy.Port,
			&proxy.Config,
			&proxy.Settings,
			&proxy.ListenAddr,
			&proxy.RemoteAddr,
			&proxy.Enabled,
			&proxy.Upload,
			&proxy.Download,
			&lastActiveStr,
			&createdAtStr,
			&updatedAtStr,
			&expireAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		if lastActiveStr != "" {
			lastActive, _ := time.Parse("2006-01-02 15:04:05", lastActiveStr)
			proxy.LastActiveAt = lastActive
		}
		if expireAtStr != "" {
			expireAt, _ := time.Parse("2006-01-02 15:04:05", expireAtStr)
			proxy.ExpireAt = &expireAt
		}
		proxy.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		proxy.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		proxies = append(proxies, proxy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return proxies, nil
}

// CreateTraffic creates a new traffic statistics record
func (db *SQLiteDB) CreateTraffic(traffic *common.TrafficStats) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	total := traffic.Upload + traffic.Download

	query := `INSERT INTO traffic_stats (
		user_id, proxy_id, upload, download, total, traffic_limit, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.db.Exec(
		query,
		traffic.UserID,
		traffic.ProxyID,
		traffic.Upload,
		traffic.Download,
		total,
		traffic.TrafficLimit,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create traffic: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	traffic.ID = id
	traffic.Total = total
	traffic.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", now)
	traffic.UpdatedAt = traffic.CreatedAt

	return nil
}

// GetTraffic retrieves traffic statistics by ID
func (db *SQLiteDB) GetTraffic(id int64) (*common.TrafficStats, error) {
	query := `SELECT 
		id, user_id, proxy_id, upload, download, total, traffic_limit, created_at, updated_at
	FROM traffic_stats WHERE id = ?`

	row := db.db.QueryRow(query, id)

	traffic := &common.TrafficStats{}
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&traffic.ID,
		&traffic.UserID,
		&traffic.ProxyID,
		&traffic.Upload,
		&traffic.Download,
		&traffic.Total,
		&traffic.TrafficLimit,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("traffic not found")
		}
		return nil, fmt.Errorf("failed to get traffic: %w", err)
	}

	// Parse time fields
	traffic.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	traffic.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return traffic, nil
}

// UpdateTraffic updates traffic statistics
func (db *SQLiteDB) UpdateTraffic(traffic *common.TrafficStats) error {
	query := `UPDATE traffic_stats SET
		user_id = ?, proxy_id = ?, upload = ?, download = ?, total = ?, traffic_limit = ?, updated_at = ?
	WHERE id = ?`

	now := time.Now().Format("2006-01-02 15:04:05")
	total := traffic.Upload + traffic.Download

	_, err := db.db.Exec(
		query,
		traffic.UserID,
		traffic.ProxyID,
		traffic.Upload,
		traffic.Download,
		total,
		traffic.TrafficLimit,
		now,
		traffic.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update traffic: %w", err)
	}

	return nil
}

// DeleteTraffic deletes traffic statistics
func (db *SQLiteDB) DeleteTraffic(id int64) error {
	query := `DELETE FROM traffic_stats WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// ListTrafficByUserID lists traffic statistics by user ID
func (db *SQLiteDB) ListTrafficByUserID(userID int64) ([]*common.TrafficStats, error) {
	query := `SELECT 
		id, user_id, proxy_id, upload, download, created_at
	FROM traffic_stats WHERE user_id = ?`

	rows, err := db.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*common.TrafficStats
	for rows.Next() {
		traffic := &common.TrafficStats{}
		var createdAtStr string

		err := rows.Scan(
			&traffic.ID,
			&traffic.UserID,
			&traffic.ProxyID,
			&traffic.Upload,
			&traffic.Download,
			&createdAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		traffic.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

		result = append(result, traffic)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// ListTrafficByProxyID lists traffic statistics by proxy ID
func (db *SQLiteDB) ListTrafficByProxyID(proxyID int64) ([]*common.TrafficStats, error) {
	query := `SELECT 
		id, user_id, proxy_id, upload, download, created_at
	FROM traffic_stats WHERE proxy_id = ?`

	rows, err := db.db.Query(query, proxyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*common.TrafficStats
	for rows.Next() {
		traffic := &common.TrafficStats{}
		var createdAtStr string

		err := rows.Scan(
			&traffic.ID,
			&traffic.UserID,
			&traffic.ProxyID,
			&traffic.Upload,
			&traffic.Download,
			&createdAtStr,
		)

		if err != nil {
			return nil, err
		}

		// Parse time fields
		traffic.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

		result = append(result, traffic)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetTrafficStats retrieves traffic statistics for a user
func (db *SQLiteDB) GetTrafficStats(userID uint) (*TrafficStats, error) {
	query := `SELECT 
		id, user_id, upload, download, total, traffic_limit, expire_at, last_reset_at, created_at, updated_at
	FROM traffic_stats WHERE user_id = ?`

	row := db.db.QueryRow(query, userID)

	stats := &TrafficStats{}
	var expireAtStr, lastResetAtStr, createdAtStr, updatedAtStr string

	err := row.Scan(
		&stats.ID,
		&stats.UserID,
		&stats.Upload,
		&stats.Download,
		&stats.Total,
		&stats.TrafficLimit,
		&expireAtStr,
		&lastResetAtStr,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse time fields
	stats.ExpireAt, _ = time.Parse("2006-01-02 15:04:05", expireAtStr)
	stats.LastResetAt, _ = time.Parse("2006-01-02 15:04:05", lastResetAtStr)
	stats.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	stats.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return stats, nil
}

// CreateTrafficRecord creates a traffic record
func (db *SQLiteDB) CreateTrafficRecord(traffic *Traffic) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `INSERT INTO traffic (
		user_id, proxy_id, up, down, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.db.Exec(
		query,
		traffic.UserID,
		traffic.ProxyID,
		traffic.Up,
		traffic.Down,
		now,
		now,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	traffic.ID = id
	return nil
}

// CreateTrafficHistory creates traffic history record
func (db *SQLiteDB) CreateTrafficHistory(history *TrafficHistory) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `INSERT INTO traffic_history (
		user_id, protocol, upload, download, date, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := db.db.Exec(
		query,
		history.UserID,
		history.Protocol,
		history.Upload,
		history.Download,
		history.Date,
		now,
		now,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	history.ID = id
	return nil
}

// CreateUser creates a new user
func (db *SQLiteDB) CreateUser(user *User) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	var expireAtStr string
	if user.ExpireAt != nil {
		expireAtStr = user.ExpireAt.Format("2006-01-02 15:04:05")
	}

	var lastLoginAtStr string
	if user.LastLoginAt != nil {
		lastLoginAtStr = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}

	var lockedUntilStr string
	if user.LockedUntil != nil {
		lockedUntilStr = user.LockedUntil.Format("2006-01-02 15:04:05")
	}

	query := `INSERT INTO users (
		username, email, password, salt, role, status, traffic_limit, traffic_used,
		last_login_at, login_attempts, locked_until, is_admin, expire_at, 
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.db.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Salt,
		user.Role,
		user.Status,
		user.TrafficLimit,
		user.TrafficUsed,
		lastLoginAtStr,
		user.LoginAttempts,
		lockedUntilStr,
		boolToInt(user.IsAdmin),
		expireAtStr,
		now,
		now,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", now)
	user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", now)

	return nil
}

// boolToInt converts a bool to an int (1 for true, 0 for false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// DeleteAlert deletes an alert record
func (db *SQLiteDB) DeleteAlert(id int64) error {
	query := `DELETE FROM alert_records WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// DeleteBackup 删除备份
func (db *SQLiteDB) DeleteBackup(id int64) error {
	query := `DELETE FROM backups WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// DeleteBackupsBefore 删除指定时间之前的备份
func (db *SQLiteDB) DeleteBackupsBefore(date time.Time) error {
	query := `DELETE FROM backups WHERE timestamp < ?`
	_, err := db.db.Exec(query, date.Format("2006-01-02 15:04:05"))
	return err
}

// DeleteCertificate 删除证书
func (db *SQLiteDB) DeleteCertificate(domain string) error {
	query := `DELETE FROM certificates WHERE domain = ?`
	_, err := db.db.Exec(query, domain)
	return err
}

// GetSettings retrieves a setting value by key
func (db *SQLiteDB) GetSettings(key string) (string, error) {
	var value string
	err := db.db.QueryRow("SELECT value FROM system_settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("setting not found")
		}
		return "", err
	}
	return value, nil
}

// SetSettings sets a setting value
func (db *SQLiteDB) SetSettings(key, value string) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	// Use INSERT OR REPLACE to handle both insert and update
	_, err := db.db.Exec(
		"INSERT INTO system_settings (key, value, created_at, updated_at) VALUES (?, ?, ?, ?) "+
			"ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?",
		key, value, now, now, value, now)

	return err
}

// DeleteDailyStatsBefore 删除指定日期之前的每日流量统计
func (db *SQLiteDB) DeleteDailyStatsBefore(date time.Time) error {
	query := `DELETE FROM daily_stats WHERE date < ?`
	_, err := db.db.Exec(query, date.Format("2006-01-02"))
	return err
}

// GetTotalBackups 获取备份总数
func (db *SQLiteDB) GetTotalBackups() (int64, error) {
	var count int64
	err := db.db.QueryRow("SELECT COUNT(*) FROM backups").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTotalUsers 获取用户总数
func (db *SQLiteDB) GetTotalUsers() (int64, error) {
	var count int64
	err := db.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetUser 根据ID获取用户
func (db *SQLiteDB) GetUser(id int64) (*User, error) {
	query := `SELECT id, username, email, password, salt, role, status, traffic_limit, traffic_used, 
              last_login_at, login_attempts, locked_until, is_admin, expire_at, created_at, updated_at 
              FROM users WHERE id = ?`

	user := &User{}
	var lastLoginAt, lockedUntil, expireAt, createdAt, updatedAt sql.NullString

	err := db.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Status,
		&user.TrafficLimit, &user.TrafficUsed, &lastLoginAt, &user.LoginAttempts, &lockedUntil,
		&user.IsAdmin, &expireAt, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 用户不存在
		}
		return nil, err
	}

	// 处理可空时间字段
	if lastLoginAt.Valid {
		lastLogin, err := time.Parse("2006-01-02 15:04:05", lastLoginAt.String)
		if err == nil {
			user.LastLoginAt = &lastLogin
		}
	}

	if lockedUntil.Valid {
		locked, err := time.Parse("2006-01-02 15:04:05", lockedUntil.String)
		if err == nil {
			user.LockedUntil = &locked
		}
	}

	if expireAt.Valid {
		expire, err := time.Parse("2006-01-02 15:04:05", expireAt.String)
		if err == nil {
			user.ExpireAt = &expire
		}
	}

	if createdAt.Valid {
		user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}

	if updatedAt.Valid {
		user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
	}

	return user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (db *SQLiteDB) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, username, email, password, salt, role, status, traffic_limit, traffic_used, 
              last_login_at, login_attempts, locked_until, is_admin, expire_at, created_at, updated_at 
              FROM users WHERE email = ?`

	user := &User{}
	var lastLoginAt, lockedUntil, expireAt, createdAt, updatedAt sql.NullString

	err := db.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Status,
		&user.TrafficLimit, &user.TrafficUsed, &lastLoginAt, &user.LoginAttempts, &lockedUntil,
		&user.IsAdmin, &expireAt, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 用户不存在
		}
		return nil, err
	}

	// 处理可空时间字段
	if lastLoginAt.Valid {
		lastLogin, err := time.Parse("2006-01-02 15:04:05", lastLoginAt.String)
		if err == nil {
			user.LastLoginAt = &lastLogin
		}
	}

	if lockedUntil.Valid {
		locked, err := time.Parse("2006-01-02 15:04:05", lockedUntil.String)
		if err == nil {
			user.LockedUntil = &locked
		}
	}

	if expireAt.Valid {
		expire, err := time.Parse("2006-01-02 15:04:05", expireAt.String)
		if err == nil {
			user.ExpireAt = &expire
		}
	}

	if createdAt.Valid {
		user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}

	if updatedAt.Valid {
		user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (db *SQLiteDB) GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, email, password, salt, role, status, traffic_limit, traffic_used, 
              last_login_at, login_attempts, locked_until, is_admin, expire_at, created_at, updated_at 
              FROM users WHERE username = ?`

	user := &User{}
	var lastLoginAt, lockedUntil, expireAt, createdAt, updatedAt sql.NullString

	err := db.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Status,
		&user.TrafficLimit, &user.TrafficUsed, &lastLoginAt, &user.LoginAttempts, &lockedUntil,
		&user.IsAdmin, &expireAt, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 用户不存在
		}
		return nil, err
	}

	// 处理可空时间字段
	if lastLoginAt.Valid {
		lastLogin, err := time.Parse("2006-01-02 15:04:05", lastLoginAt.String)
		if err == nil {
			user.LastLoginAt = &lastLogin
		}
	}

	if lockedUntil.Valid {
		locked, err := time.Parse("2006-01-02 15:04:05", lockedUntil.String)
		if err == nil {
			user.LockedUntil = &locked
		}
	}

	if expireAt.Valid {
		expire, err := time.Parse("2006-01-02 15:04:05", expireAt.String)
		if err == nil {
			user.ExpireAt = &expire
		}
	}

	if createdAt.Valid {
		user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}

	if updatedAt.Valid {
		user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
	}

	return user, nil
}

// ListAlertRecords 获取所有告警记录
func (db *SQLiteDB) ListAlertRecords(out *[]*AlertRecord) error {
	query := `SELECT id, type, value, threshold, message, created_at, updated_at 
              FROM alert_records ORDER BY created_at DESC`

	rows, err := db.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		alert := &AlertRecord{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&alert.ID, &alert.Type, &alert.Value, &alert.Threshold, &alert.Message,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return err
		}

		// 解析时间
		alert.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		alert.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		*out = append(*out, alert)
	}

	return rows.Err()
}

// ListAlerts 分页获取告警记录
func (db *SQLiteDB) ListAlerts(page, pageSize int) ([]*AlertRecord, error) {
	offset := (page - 1) * pageSize
	query := `SELECT id, type, value, threshold, message, created_at, updated_at 
              FROM alert_records ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := db.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*AlertRecord

	for rows.Next() {
		alert := &AlertRecord{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&alert.ID, &alert.Type, &alert.Value, &alert.Threshold, &alert.Message,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析时间
		alert.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		alert.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

// ListBackups 获取所有备份记录
func (db *SQLiteDB) ListBackups() ([]*Backup, error) {
	query := `SELECT id, path, size, status, created_at, updated_at 
              FROM backups ORDER BY created_at DESC`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []*Backup

	for rows.Next() {
		backup := &Backup{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&backup.ID, &backup.Path, &backup.Size, &backup.Status,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析时间
		backup.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		backup.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		backups = append(backups, backup)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return backups, nil
}

// ListCertificates 获取所有证书
func (db *SQLiteDB) ListCertificates() ([]*Certificate, error) {
	query := `SELECT id, domain, cert_file, key_file, status, last_checked_at, last_renewed_at, expires_at, created_at, updated_at 
              FROM certificates ORDER BY expires_at ASC`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certificates []*Certificate

	for rows.Next() {
		cert := &Certificate{}
		var lastCheckedAt, lastRenewedAt, expiresAt, createdAt, updatedAt string

		err := rows.Scan(
			&cert.ID, &cert.Domain, &cert.CertFile, &cert.KeyFile, &cert.Status,
			&lastCheckedAt, &lastRenewedAt, &expiresAt, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析时间
		cert.LastCheckedAt, _ = time.Parse("2006-01-02 15:04:05", lastCheckedAt)
		cert.LastRenewedAt, _ = time.Parse("2006-01-02 15:04:05", lastRenewedAt)
		cert.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAt)
		cert.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		cert.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		certificates = append(certificates, cert)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return certificates, nil
}

// ListDailyStatsByUserID 获取用户的每日流量统计
func (db *SQLiteDB) ListDailyStatsByUserID(userID int64) ([]*DailyStats, error) {
	query := `SELECT id, user_id, date, upload, download, total, created_at, updated_at 
              FROM daily_stats WHERE user_id = ? ORDER BY date DESC`

	rows, err := db.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*DailyStats

	for rows.Next() {
		stat := &DailyStats{}
		var date, createdAt, updatedAt string

		err := rows.Scan(
			&stat.ID, &stat.UserID, &date, &stat.Upload, &stat.Download, &stat.Total,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 解析日期和时间
		stat.Date, _ = time.Parse("2006-01-02", date)
		stat.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		stat.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

// ListTrafficHistoryByDateRange 根据日期范围获取用户的流量历史
func (db *SQLiteDB) ListTrafficHistoryByDateRange(userID uint, startDate, endDate string, histories *[]*TrafficHistory) error {
	query := `SELECT id, user_id, protocol, upload, download, date, created_at, updated_at 
              FROM traffic_history 
              WHERE user_id = ? AND date BETWEEN ? AND ? 
              ORDER BY date ASC`

	rows, err := db.db.Query(query, userID, startDate, endDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		history := &TrafficHistory{}
		var date, createdAt, updatedAt string

		err := rows.Scan(
			&history.ID, &history.UserID, &history.Protocol, &history.Upload, &history.Download,
			&date, &createdAt, &updatedAt,
		)
		if err != nil {
			return err
		}

		// 日期可能已经是字符串格式，直接赋值
		history.Date = date
		history.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		history.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		*histories = append(*histories, history)
	}

	return rows.Err()
}

// ListUsers 分页获取用户列表
func (db *SQLiteDB) ListUsers(page, pageSize int) ([]*User, error) {
	offset := (page - 1) * pageSize
	query := `SELECT id, username, email, password, salt, role, status, traffic_limit, traffic_used, 
              last_login_at, login_attempts, locked_until, is_admin, expire_at, created_at, updated_at 
              FROM users ORDER BY id DESC LIMIT ? OFFSET ?`

	rows, err := db.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var lastLoginAt, lockedUntil, expireAt, createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Status,
			&user.TrafficLimit, &user.TrafficUsed, &lastLoginAt, &user.LoginAttempts, &lockedUntil,
			&user.IsAdmin, &expireAt, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 处理可空时间字段
		if lastLoginAt.Valid {
			lastLogin, err := time.Parse("2006-01-02 15:04:05", lastLoginAt.String)
			if err == nil {
				user.LastLoginAt = &lastLogin
			}
		}

		if lockedUntil.Valid {
			locked, err := time.Parse("2006-01-02 15:04:05", lockedUntil.String)
			if err == nil {
				user.LockedUntil = &locked
			}
		}

		if expireAt.Valid {
			expire, err := time.Parse("2006-01-02 15:04:05", expireAt.String)
			if err == nil {
				user.ExpireAt = &expire
			}
		}

		if createdAt.Valid {
			user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}

		if updatedAt.Valid {
			user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUser 删除用户
func (db *SQLiteDB) DeleteUser(id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.db.Exec(query, id)
	return err
}

// GetAlert 获取告警记录
func (db *SQLiteDB) GetAlert(id int64) (*AlertRecord, error) {
	query := `SELECT id, type, value, threshold, message, created_at, updated_at
              FROM alert_records WHERE id = ?`

	row := db.db.QueryRow(query, id)
	alert := &AlertRecord{}
	var createdAt, updatedAt string

	err := row.Scan(
		&alert.ID,
		&alert.Type,
		&alert.Value,
		&alert.Threshold,
		&alert.Message,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析时间
	alert.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	alert.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return alert, nil
}

// GetBackup 获取备份
func (db *SQLiteDB) GetBackup(id int64) (*Backup, error) {
	query := `SELECT id, path, size, status, timestamp, created_at, updated_at
              FROM backups WHERE id = ?`

	row := db.db.QueryRow(query, id)

	backup := &Backup{}
	var timestampStr, createdStr, updatedStr string

	err := row.Scan(
		&backup.ID,
		&backup.Path,
		&backup.Size,
		&backup.Status,
		&timestampStr,
		&createdStr,
		&updatedStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析时间
	backup.Timestamp, _ = time.Parse("2006-01-02 15:04:05", timestampStr)
	backup.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdStr)
	backup.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedStr)

	return backup, nil
}

// GetCertificate 获取证书
func (db *SQLiteDB) GetCertificate(domain string) (*Certificate, error) {
	query := `SELECT 
		id, domain, cert_file, key_file, status, last_checked_at, 
		last_renewed_at, expires_at, created_at, updated_at
	FROM certificates WHERE domain = ?`

	row := db.db.QueryRow(query, domain)

	cert := &Certificate{}
	var lastCheckedStr, lastRenewedStr, expiresStr, createdAtStr, updatedAtStr string

	err := row.Scan(
		&cert.ID,
		&cert.Domain,
		&cert.CertFile,
		&cert.KeyFile,
		&cert.Status,
		&lastCheckedStr,
		&lastRenewedStr,
		&expiresStr,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析时间
	cert.LastCheckedAt, _ = time.Parse("2006-01-02 15:04:05", lastCheckedStr)
	cert.LastRenewedAt, _ = time.Parse("2006-01-02 15:04:05", lastRenewedStr)
	cert.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresStr)
	cert.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	cert.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return cert, nil
}

// SearchUsers 根据关键词搜索用户
func (db *SQLiteDB) SearchUsers(keyword string) ([]*User, error) {
	// 使用LIKE进行简单搜索，匹配用户名和邮箱
	query := `SELECT id, username, email, password, salt, role, status, traffic_limit, traffic_used, 
              last_login_at, login_attempts, locked_until, is_admin, expire_at, created_at, updated_at 
              FROM users 
              WHERE username LIKE ? OR email LIKE ? OR role LIKE ? OR status LIKE ?
              ORDER BY id DESC`

	// 构造模糊查询参数
	likeParam := "%" + keyword + "%"
	rows, err := db.db.Query(query, likeParam, likeParam, likeParam, likeParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var lastLoginAt, lockedUntil, expireAt, createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Status,
			&user.TrafficLimit, &user.TrafficUsed, &lastLoginAt, &user.LoginAttempts, &lockedUntil,
			&user.IsAdmin, &expireAt, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 处理可空时间字段
		if lastLoginAt.Valid {
			lastLogin, err := time.Parse("2006-01-02 15:04:05", lastLoginAt.String)
			if err == nil {
				user.LastLoginAt = &lastLogin
			}
		}

		if lockedUntil.Valid {
			locked, err := time.Parse("2006-01-02 15:04:05", lockedUntil.String)
			if err == nil {
				user.LockedUntil = &locked
			}
		}

		if expireAt.Valid {
			expire, err := time.Parse("2006-01-02 15:04:05", expireAt.String)
			if err == nil {
				user.ExpireAt = &expire
			}
		}

		if createdAt.Valid {
			user.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}

		if updatedAt.Valid {
			user.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateBackup 更新备份记录
func (db *SQLiteDB) UpdateBackup(backup *Backup) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `UPDATE backups SET
		path = ?, size = ?, status = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		backup.Path,
		backup.Size,
		backup.Status,
		now,
		backup.ID,
	)

	return err
}

// UpdateCertificate 更新证书记录
func (db *SQLiteDB) UpdateCertificate(cert *Certificate) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	query := `UPDATE certificates SET
		domain = ?, cert_file = ?, key_file = ?, status = ?,
		last_checked_at = ?, last_renewed_at = ?, expires_at = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		cert.Domain,
		cert.CertFile,
		cert.KeyFile,
		cert.Status,
		cert.LastCheckedAt.Format("2006-01-02 15:04:05"),
		cert.LastRenewedAt.Format("2006-01-02 15:04:05"),
		cert.ExpiresAt.Format("2006-01-02 15:04:05"),
		now,
		cert.ID,
	)

	return err
}

// UpdateUser 更新用户信息
func (db *SQLiteDB) UpdateUser(user *User) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	// 处理可能为空的时间字段
	var lastLoginAtStr, lockedUntilStr, expireAtStr string
	if user.LastLoginAt != nil {
		lastLoginAtStr = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}
	if user.LockedUntil != nil {
		lockedUntilStr = user.LockedUntil.Format("2006-01-02 15:04:05")
	}
	if user.ExpireAt != nil {
		expireAtStr = user.ExpireAt.Format("2006-01-02 15:04:05")
	}

	query := `UPDATE users SET
		username = ?, email = ?, password = ?, salt = ?, role = ?, status = ?,
		traffic_limit = ?, traffic_used = ?, last_login_at = ?, login_attempts = ?,
		locked_until = ?, is_admin = ?, expire_at = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Salt,
		user.Role,
		user.Status,
		user.TrafficLimit,
		user.TrafficUsed,
		lastLoginAtStr,
		user.LoginAttempts,
		lockedUntilStr,
		boolToInt(user.IsAdmin),
		expireAtStr,
		now,
		user.ID,
	)

	return err
}

// GetTotalProtocols 获取协议总数
func (db *SQLiteDB) GetTotalProtocols() (int64, error) {
	var count int64
	err := db.db.QueryRow("SELECT COUNT(*) FROM protocols").Scan(&count)
	return count, err
}

// UpdateTrafficStats updates traffic statistics
func (db *SQLiteDB) UpdateTrafficStats(stats *TrafficStats) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	expireAt := stats.ExpireAt.Format("2006-01-02 15:04:05")
	lastResetAt := stats.LastResetAt.Format("2006-01-02 15:04:05")

	query := `UPDATE traffic_stats SET 
		user_id = ?, upload = ?, download = ?, total = ?, traffic_limit = ?,
		expire_at = ?, last_reset_at = ?, updated_at = ?
	WHERE id = ?`

	_, err := db.db.Exec(
		query,
		stats.UserID,
		stats.Upload,
		stats.Download,
		stats.Total,
		stats.TrafficLimit,
		expireAt,
		lastResetAt,
		now,
		stats.ID,
	)

	return err
}
