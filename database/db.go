package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"v/common"
	"v/model"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed schema.sql
var schemaFS embed.FS

// DBInstance is the global database connection
var DBInstance *Database

// DB 是全局数据库连接对象，用于兼容性目的
var DB *sql.DB

// Database represents a database implementation
type Database struct {
	*gorm.DB
}

// NewDatabase creates a new database instance
func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return &Database{db}, nil
}

// Begin starts a new transaction
func (db *Database) Begin() (*gorm.DB, error) {
	tx := db.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

// Commit commits a transaction
func (db *Database) Commit() error {
	return db.DB.Commit().Error
}

// Rollback rolls back a transaction
func (db *Database) Rollback() error {
	return db.DB.Rollback().Error
}

// Close closes the database connection
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateBackup creates a backup of the database
func (db *Database) CreateBackup(backup *model.Backup) error {
	return db.BackupDB(backup.Path)
}

// RestoreBackup restores the database from a backup
func (db *Database) RestoreBackup(backup *model.Backup) error {
	return db.RestoreDB(backup.Path)
}

// CreateCertificate creates a new SSL certificate
func (db *Database) CreateCertificate(cert *model.Certificate) error {
	return db.Create(cert).Error
}

// GetCertificate gets a certificate by ID
func (db *Database) GetCertificate(id int64) (*model.Certificate, error) {
	var cert model.Certificate
	if err := db.First(&cert, id).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// UpdateCertificate updates a certificate
func (db *Database) UpdateCertificate(cert *model.Certificate) error {
	return db.Save(cert).Error
}

// DeleteCertificate deletes a certificate
func (db *Database) DeleteCertificate(id int64) error {
	return db.Delete(&model.Certificate{}, id).Error
}

// ListCertificates lists all certificates
func (db *Database) ListCertificates() ([]*model.Certificate, error) {
	var certs []*model.Certificate
	if err := db.Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// DBConfig represents database configuration
type DBConfig struct {
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
}

// DefaultDBConfig returns default database configuration
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		MaxOpenConns:    25,              // Maximum number of open connections
		MaxIdleConns:    10,              // Maximum number of idle connections
		ConnMaxLifetime: 5 * time.Minute, // Maximum lifetime of a connection
		ConnMaxIdleTime: 5 * time.Minute, // Maximum idle time of a connection
	}
}

// InitDBWithConfig initializes the database connection with custom configuration
func InitDBWithConfig(dbPath string, config *DBConfig) error {
	if config == nil {
		config = DefaultDBConfig()
	}

	// Create database directory if not exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// Initialize database instance
	db, err := NewDatabase(dbPath)
	if err != nil {
		return err
	}
	DBInstance = db

	// Configure connection pool
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Initialize schema
	if err := initSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Run migrations
	if err := DBInstance.AutoMigrate(&common.Proxy{}); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

// initSchema initializes the database schema
func initSchema() error {
	// Read schema file
	schemaBytes, err := fs.ReadFile(schemaFS, "schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	// Execute schema
	err = DBInstance.DB.Exec(string(schemaBytes)).Error
	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DBInstance != nil {
		return DBInstance.Close()
	}
	return nil
}

// GetDB returns the global database instance
func GetDB() *Database {
	return DBInstance
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := DBInstance.DB.Raw(`
		SELECT id, username, password, email, is_admin, "traffic_limit", traffic_used,
			expire_at, last_login_at, login_attempts, locked_until
		FROM users
		WHERE username = ?
	`, username).Scan(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func CreateUser(user *model.User) error {
	return DBInstance.DB.Create(user).Error
}

// GetUser retrieves a user by ID
func GetUser(id int64) (*model.User, error) {
	var user model.User
	err := DBInstance.DB.Raw(`
		SELECT id, username, email, password, salt, is_admin, "traffic_limit", traffic_used,
			expire_at, last_login_at, login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user
func UpdateUser(user *model.User) error {
	return DBInstance.DB.Save(user).Error
}

// DeleteUser deletes a user by ID
func DeleteUser(id int64) error {
	return DBInstance.DB.Delete(&model.User{}, id).Error
}

// ListUsers retrieves all users
func ListUsers() ([]model.User, error) {
	var users []model.User
	err := DBInstance.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// CreateUser creates a new user
func (db *Database) CreateUser(user *model.User) error {
	query := `
		INSERT INTO users (
			username, email, password, salt, is_admin, "traffic_limit", traffic_used,
			expire_at, last_login_at, login_attempts, locked_until, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query,
		user.Username,
		user.Email,
		user.Password,
		user.Salt,
		user.IsAdmin,
		user.TrafficLimit,
		user.TrafficUsed,
		user.ExpireAt,
		user.LastLoginAt,
		user.LoginAttempts,
		user.LockedUntil,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

// GetUser returns a user by ID
func (db *Database) GetUser(id int64) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, username, email, password, salt, is_admin, "traffic_limit", traffic_used,
			expire_at, last_login_at, login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Salt,
		&user.IsAdmin,
		&user.TrafficLimit,
		&user.TrafficUsed,
		&user.ExpireAt,
		&user.LastLoginAt,
		&user.LoginAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername returns a user by username
func (db *Database) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, username, email, password, salt, is_admin, "traffic_limit", traffic_used,
			expire_at, last_login_at, login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Salt,
		&user.IsAdmin,
		&user.TrafficLimit,
		&user.TrafficUsed,
		&user.ExpireAt,
		&user.LastLoginAt,
		&user.LoginAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail returns a user by email
func (db *Database) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, username, email, password, salt, is_admin, "traffic_limit", traffic_used,
			expire_at, last_login_at, login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE email = ?
	`
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Salt,
		&user.IsAdmin,
		&user.TrafficLimit,
		&user.TrafficUsed,
		&user.ExpireAt,
		&user.LastLoginAt,
		&user.LoginAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (db *Database) UpdateUser(user *model.User) error {
	query := `
		UPDATE users SET
			username = ?, email = ?, password = ?, salt = ?, is_admin = ?,
			"traffic_limit" = ?, traffic_used = ?, expire_at = ?, last_login_at = ?,
			login_attempts = ?, locked_until = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := db.Exec(query,
		user.Username,
		user.Email,
		user.Password,
		user.Salt,
		user.IsAdmin,
		user.TrafficLimit,
		user.TrafficUsed,
		user.ExpireAt,
		user.LastLoginAt,
		user.LoginAttempts,
		user.LockedUntil,
		time.Now(),
		user.ID,
	)
	return err
}

// DeleteUser deletes a user
func (db *Database) DeleteUser(id int64) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// ListUsers returns a list of users
func (db *Database) ListUsers(offset, limit int) ([]*model.User, error) {
	query := `
		SELECT id, username, email, password, salt, is_admin, "traffic_limit", traffic_used,
			expire_at, last_login_at, login_attempts, locked_until, created_at, updated_at
		FROM users
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Salt,
			&user.IsAdmin,
			&user.TrafficLimit,
			&user.TrafficUsed,
			&user.ExpireAt,
			&user.LastLoginAt,
			&user.LoginAttempts,
			&user.LockedUntil,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

// CreateProxy creates a new proxy
func (db *Database) CreateProxy(proxy *common.Proxy) error {
	return db.Create(proxy).Error
}

// GetProxy returns a proxy by ID
func (db *Database) GetProxy(id uint) (*common.Proxy, error) {
	var proxy common.Proxy
	if err := db.First(&proxy, id).Error; err != nil {
		return nil, err
	}
	return &proxy, nil
}

// GetProxiesByUser returns proxies by user ID
func (db *Database) GetProxiesByUser(userID uint) ([]*common.Proxy, error) {
	var proxies []*common.Proxy
	if err := db.Where("user_id = ?", userID).Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// UpdateProxy updates a proxy
func (db *Database) UpdateProxy(proxy *common.Proxy) error {
	return db.Save(proxy).Error
}

// DeleteProxy deletes a proxy
func (db *Database) DeleteProxy(id uint) error {
	return db.Delete(&common.Proxy{}, id).Error
}

// ListProxies returns a list of proxies
func (db *Database) ListProxies(offset, limit int) ([]*common.Proxy, error) {
	var proxies []*common.Proxy
	if err := db.Offset(offset).Limit(limit).Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// UpdateTraffic updates traffic statistics for a proxy
func (db *Database) UpdateTraffic(id uint, upload, download int64) error {
	return db.Model(&common.Proxy{}).Where("id = ?", id).Updates(map[string]interface{}{
		"upload":   upload,
		"download": download,
	}).Error
}

// Enable enables a proxy
func (db *Database) Enable(id uint) error {
	return db.Model(&common.Proxy{}).Where("id = ?", id).Update("enabled", true).Error
}

// Disable disables a proxy
func (db *Database) Disable(id uint) error {
	return db.Model(&common.Proxy{}).Where("id = ?", id).Update("enabled", false).Error
}

// UpdateLastActive updates the last active time for a proxy
func (db *Database) UpdateLastActive(id uint) error {
	return db.Model(&common.Proxy{}).Where("id = ?", id).Update("last_active_at", time.Now()).Error
}

// BackupDB creates a backup of the database
func (db *Database) BackupDB(backupPath string) error {
	// Create backup directory if not exists
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Create backup file
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer backupFile.Close()

	// Get SQL DB
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Begin transaction
	sqlTx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	// Dump database to backup file
	rows, err := sqlTx.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return fmt.Errorf("failed to get tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %v", err)
		}

		// Skip sqlite_sequence table
		if tableName == "sqlite_sequence" {
			continue
		}

		// Get table schema
		var schema string
		err = sqlTx.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&schema)
		if err != nil {
			return fmt.Errorf("failed to get table schema: %v", err)
		}

		// Write schema to backup file
		if _, err := backupFile.WriteString(schema + ";\n\n"); err != nil {
			return fmt.Errorf("failed to write schema: %v", err)
		}

		// Get table data
		tableRows, err := sqlTx.Query("SELECT * FROM " + tableName)
		if err != nil {
			return fmt.Errorf("failed to get table data: %v", err)
		}
		defer tableRows.Close()

		columns, err := tableRows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %v", err)
		}

		for tableRows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := tableRows.Scan(valuePtrs...); err != nil {
				return fmt.Errorf("failed to scan row: %v", err)
			}

			// Build INSERT statement
			insertStmt := fmt.Sprintf("INSERT INTO %s VALUES (", tableName)
			for i, v := range values {
				if i > 0 {
					insertStmt += ","
				}
				switch val := v.(type) {
				case nil:
					insertStmt += "NULL"
				case []byte:
					insertStmt += fmt.Sprintf("'%s'", string(val))
				case string:
					insertStmt += fmt.Sprintf("'%s'", val)
				default:
					insertStmt += fmt.Sprintf("%v", val)
				}
			}
			insertStmt += ");\n"

			if _, err := backupFile.WriteString(insertStmt); err != nil {
				return fmt.Errorf("failed to write insert statement: %v", err)
			}
		}
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// RestoreDB restores the database from a backup file
func (db *Database) RestoreDB(backupPath string) error {
	// Read backup file
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %v", err)
	}

	// Get SQL DB
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Begin transaction
	sqlTx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	// Execute backup SQL statements
	statements := strings.Split(string(backupData), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := sqlTx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %v", err)
		}
	}

	// Commit transaction
	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// Migration represents a database migration
type Migration struct {
	Version int
	Up      string
	Down    string
}

// Migrations contains all database migrations
var Migrations = []Migration{
	{
		Version: 1,
		Up: `
			CREATE TABLE IF NOT EXISTS migrations (
				version INTEGER PRIMARY KEY,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`,
		Down: `
			DROP TABLE IF EXISTS migrations;
		`,
	},
	{
		Version: 2,
		Up: `
			ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
			ALTER TABLE users ADD COLUMN login_attempts INTEGER DEFAULT 0;
			ALTER TABLE users ADD COLUMN locked_until TIMESTAMP;
		`,
		Down: `
			ALTER TABLE users DROP COLUMN last_login;
			ALTER TABLE users DROP COLUMN login_attempts;
			ALTER TABLE users DROP COLUMN locked_until;
		`,
	},
	{
		Version: 3,
		Up: `
			CREATE TABLE IF NOT EXISTS user_sessions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				token VARCHAR(255) NOT NULL UNIQUE,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP NOT NULL,
				ip_address VARCHAR(45),
				user_agent TEXT,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);
			CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
			CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
		`,
		Down: `
			DROP TABLE IF EXISTS user_sessions;
		`,
	},
}

// GetCurrentVersion returns the current database version
func (db *Database) GetCurrentVersion() (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM migrations").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current version: %v", err)
	}
	return version, nil
}

// MigrateUp migrates the database up
func (db *Database) MigrateUp() error {
	currentVersion, err := db.GetCurrentVersion()
	if err != nil {
		return err
	}

	// Get SQL DB
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Begin transaction
	sqlTx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	for _, migration := range Migrations {
		if migration.Version > currentVersion {
			// Apply migration
			_, err := sqlTx.Exec(migration.Up)
			if err != nil {
				return fmt.Errorf("failed to apply migration %d: %v", migration.Version, err)
			}

			// Record migration
			_, err = sqlTx.Exec("INSERT INTO migrations (version) VALUES (?)", migration.Version)
			if err != nil {
				return fmt.Errorf("failed to record migration %d: %v", migration.Version, err)
			}
		}
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %v", err)
	}

	return nil
}

// MigrateDown rolls back the last migration
func (db *Database) MigrateDown() error {
	currentVersion, err := db.GetCurrentVersion()
	if err != nil {
		return err
	}

	if currentVersion == 0 {
		return nil
	}

	// Get SQL DB
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Begin transaction
	sqlTx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	// Find the last migration
	var lastMigration Migration
	for _, m := range Migrations {
		if m.Version == currentVersion {
			lastMigration = m
			break
		}
	}

	// Roll back migration
	_, err = sqlTx.Exec(lastMigration.Down)
	if err != nil {
		return fmt.Errorf("failed to roll back migration %d: %v", currentVersion, err)
	}

	// Remove migration record
	_, err = sqlTx.Exec("DELETE FROM migrations WHERE version = ?", currentVersion)
	if err != nil {
		return fmt.Errorf("failed to remove migration record %d: %v", currentVersion, err)
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %v", err)
	}

	return nil
}

// PrepareStatements prepares commonly used SQL statements
func (db *Database) PrepareStatements() error {
	// Prepare user-related statements
	userStmt, err := db.Prepare(`
		SELECT id, username, email, password, is_admin, enabled, created_at, expire_at, 
		       "traffic_limit", used_traffic, last_login, login_attempts, locked_until
		FROM users WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare user statement: %v", err)
	}
	defer userStmt.Close()

	// Prepare proxy-related statements
	proxyStmt, err := db.Prepare(`
		SELECT id, user_id, protocol, settings, enabled, upload, download, created_at
		FROM proxy_configs WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare proxy statement: %v", err)
	}
	defer proxyStmt.Close()

	// Prepare traffic-related statements
	trafficStmt, err := db.Prepare(`
		SELECT id, user_id, proxy_id, upload, download, timestamp
		FROM traffic_logs WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare traffic statement: %v", err)
	}
	defer trafficStmt.Close()

	return nil
}

// OptimizeDB performs database optimization
func (db *Database) OptimizeDB() error {
	// Get SQL DB
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Begin transaction
	sqlTx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	// Analyze tables
	_, err = sqlTx.Exec("ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to analyze tables: %v", err)
	}

	// Vacuum database
	_, err = sqlTx.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %v", err)
	}

	// Rebuild indexes
	_, err = sqlTx.Exec("REINDEX")
	if err != nil {
		return fmt.Errorf("failed to rebuild indexes: %v", err)
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit optimization: %v", err)
	}

	return nil
}

// Manager represents the database manager
type Manager struct {
	db *gorm.DB
	tx *gorm.DB
}

var defaultManager *Manager

// InitDB initializes the database manager
func InitDB(db *gorm.DB) {
	defaultManager = &Manager{db: db}
}

// CreateUser creates a new user
func (m *Manager) CreateUser(user *model.User) error {
	return m.db.Create(user).Error
}

// GetUser gets a user by ID
func (m *Manager) GetUser(id int64) (*model.User, error) {
	var user model.User
	if err := m.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername gets a user by username
func (m *Manager) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := m.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail gets a user by email
func (m *Manager) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := m.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (m *Manager) UpdateUser(user *model.User) error {
	return m.db.Save(user).Error
}

// DeleteUser deletes a user
func (m *Manager) DeleteUser(id int64) error {
	return m.db.Delete(&model.User{}, id).Error
}

// ListUsers lists users with pagination
func (m *Manager) ListUsers(offset, limit int) ([]*model.User, error) {
	var users []*model.User
	if err := m.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// CreateProxy creates a new proxy
func (m *Manager) CreateProxy(proxy *common.Proxy) error {
	return m.db.Create(proxy).Error
}

// GetProxy gets a proxy by ID
func (m *Manager) GetProxy(id int64) (*common.Proxy, error) {
	var proxy common.Proxy
	if err := m.db.First(&proxy, id).Error; err != nil {
		return nil, err
	}
	return &proxy, nil
}

// GetProxiesByUser gets proxies by user ID
func (m *Manager) GetProxiesByUser(userID int64) ([]*common.Proxy, error) {
	var proxies []*common.Proxy
	if err := m.db.Where("user_id = ?", userID).Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// GetProxiesByUserID gets proxies by user ID (alias for GetProxiesByUser)
func (m *Manager) GetProxiesByUserID(userID int64) ([]*common.Proxy, error) {
	return m.GetProxiesByUser(userID)
}

// UpdateProxy updates a proxy
func (m *Manager) UpdateProxy(proxy *common.Proxy) error {
	return m.db.Save(proxy).Error
}

// DeleteProxy deletes a proxy
func (m *Manager) DeleteProxy(id int64) error {
	return m.db.Delete(&common.Proxy{}, id).Error
}

// ListProxies lists proxies with pagination
func (m *Manager) ListProxies(offset, limit int) ([]*common.Proxy, error) {
	var proxies []*common.Proxy
	if err := m.db.Offset(offset).Limit(limit).Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// Begin starts a transaction
func (m *Manager) Begin() error {
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	m.tx = tx
	return nil
}

// Commit commits a transaction
func (m *Manager) Commit() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction in progress")
	}
	err := m.tx.Commit().Error
	m.tx = nil
	return err
}

// Rollback rolls back a transaction
func (m *Manager) Rollback() error {
	if m.tx == nil {
		return fmt.Errorf("no transaction in progress")
	}
	err := m.tx.Rollback().Error
	m.tx = nil
	return err
}

// BackupDB backs up the database
func (m *Manager) BackupDB(backupPath string) error {
	// TODO: Implement database backup
	return nil
}

// RestoreDB restores the database
func (m *Manager) RestoreDB(backupPath string) error {
	// TODO: Implement database restore
	return nil
}

// MigrateUp migrates the database up
func (m *Manager) MigrateUp() error {
	// TODO: Implement database migration
	return nil
}

// MigrateDown migrates the database down
func (m *Manager) MigrateDown() error {
	// TODO: Implement database migration
	return nil
}

// OptimizeDB optimizes the database
func (m *Manager) OptimizeDB() error {
	// TODO: Implement database optimization
	return nil
}

// QueryRow executes a query that is expected to return at most one row
func (db *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil
	}
	return sqlDB.QueryRow(query, args...)
}

// GetAllProxies 获取所有代理配置
func (db *Database) GetAllProxies() ([]*common.Proxy, error) {
	var proxies []*common.Proxy
	if err := db.Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// UpdateProxyStats 更新代理统计信息
func (db *Database) UpdateProxyStats(proxy *common.Proxy) error {
	return db.Save(proxy).Error
}

// GetProxyByID 获取指定ID的代理
func (db *Database) GetProxyByID(id int64) (*common.Proxy, error) {
	var proxy common.Proxy
	if err := db.First(&proxy, id).Error; err != nil {
		return nil, err
	}
	return &proxy, nil
}

// GetUserProxies 获取用户的所有代理
func (db *Database) GetUserProxies(userID int64) ([]*common.Proxy, error) {
	var proxies []*common.Proxy
	if err := db.Where("user_id = ?", userID).Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// CreateDailyStats 创建日流量统计
func (db *Database) CreateDailyStats(stats *model.DailyStats) error {
	return db.Create(stats).Error
}

// DeleteDailyStatsBefore 删除指定日期之前的日流量统计
func (db *Database) DeleteDailyStatsBefore(date time.Time) error {
	return db.Where("date < ?", date).Delete(&model.DailyStats{}).Error
}

// ListDailyStatsByUserID 获取用户的日流量统计
func (db *Database) ListDailyStatsByUserID(userID int64, startDate, endDate time.Time) ([]*model.DailyStats, error) {
	var stats []*model.DailyStats
	err := db.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).Find(&stats).Error
	return stats, err
}

// GetDBInstance returns the global database instance
func GetDBInstance() *Database {
	return DBInstance
}

// Query executes a query that returns rows
func (db *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Query(query, args...)
}

// Exec executes a query without returning any rows
func (db *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Exec(query, args...)
}

// Prepare creates a prepared statement for later queries or executions
func (db *Database) Prepare(query string) (*sql.Stmt, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Prepare(query)
}

// AutoMigrate automatically migrates the database schema
func (m *Manager) AutoMigrate() error {
	// Implement database migration for common.Proxy
	return m.db.AutoMigrate(&common.Proxy{})
}

// CleanupTraffic cleans up traffic records before a specified time
func (m *Manager) CleanupTraffic(before time.Time) error {
	// TODO: Implement traffic cleanup
	return nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// MockDB 是一个完整的model.DB实现，用于在不初始化真实数据库的情况下运行程序
type MockDB struct{}

// CreateUser 创建用户
func (m *MockDB) CreateUser(user *model.User) error {
	return nil
}

// GetUser 获取用户
func (m *MockDB) GetUser(id int64) (*model.User, error) {
	return &model.User{}, nil
}

// GetUserByUsername 通过用户名获取用户
func (m *MockDB) GetUserByUsername(username string) (*model.User, error) {
	return &model.User{}, nil
}

// GetUserByEmail 通过邮箱获取用户
func (m *MockDB) GetUserByEmail(email string) (*model.User, error) {
	return &model.User{}, nil
}

// UpdateUser 更新用户
func (m *MockDB) UpdateUser(user *model.User) error {
	return nil
}

// DeleteUser 删除用户
func (m *MockDB) DeleteUser(id int64) error {
	return nil
}

// ListUsers 列出用户
func (m *MockDB) ListUsers(page, pageSize int) ([]*model.User, error) {
	return nil, nil
}

// GetTotalUsers 获取用户总数
func (m *MockDB) GetTotalUsers() (int64, error) {
	return 0, nil
}

// SearchUsers 搜索用户
func (m *MockDB) SearchUsers(keyword string) ([]*model.User, error) {
	return nil, nil
}

// CreateProxy 创建代理
func (m *MockDB) CreateProxy(proxy *common.Proxy) error {
	return nil
}

// GetProxy 获取代理
func (m *MockDB) GetProxy(id int64) (*common.Proxy, error) {
	return &common.Proxy{}, nil
}

// GetProxiesByUserID 获取用户代理
func (m *MockDB) GetProxiesByUserID(userID int64) ([]*common.Proxy, error) {
	return nil, nil
}

// UpdateProxy 更新代理
func (m *MockDB) UpdateProxy(proxy *common.Proxy) error {
	return nil
}

// DeleteProxy 删除代理
func (m *MockDB) DeleteProxy(id int64) error {
	return nil
}

// GetProxiesByPort 获取指定端口的代理
func (m *MockDB) GetProxiesByPort(port int) ([]*common.Proxy, error) {
	return nil, nil
}

// ListProxies 列出代理
func (m *MockDB) ListProxies(page, pageSize int) ([]*common.Proxy, error) {
	return nil, nil
}

// GetTotalProxies 获取代理总数
func (m *MockDB) GetTotalProxies() (int64, error) {
	return 0, nil
}

// SearchProxies 搜索代理
func (m *MockDB) SearchProxies(keyword string) ([]*common.Proxy, error) {
	return nil, nil
}

// CreateTraffic 创建流量统计
func (m *MockDB) CreateTraffic(traffic *common.TrafficStats) error {
	return nil
}

// GetTraffic 获取流量统计
func (m *MockDB) GetTraffic(id int64) (*common.TrafficStats, error) {
	return &common.TrafficStats{}, nil
}

// UpdateTraffic 更新流量统计
func (m *MockDB) UpdateTraffic(traffic *common.TrafficStats) error {
	return nil
}

// DeleteTraffic 删除流量统计
func (m *MockDB) DeleteTraffic(id int64) error {
	return nil
}

// ListTrafficByUserID 获取用户流量统计
func (m *MockDB) ListTrafficByUserID(userID int64) ([]*common.TrafficStats, error) {
	return nil, nil
}

// ListTrafficByProxyID 获取代理流量统计
func (m *MockDB) ListTrafficByProxyID(proxyID int64) ([]*common.TrafficStats, error) {
	return nil, nil
}

// GetTrafficStats 获取流量统计
func (m *MockDB) GetTrafficStats(userID uint) (*model.TrafficStats, error) {
	return &model.TrafficStats{}, nil
}

// CreateTrafficRecord 创建流量记录
func (m *MockDB) CreateTrafficRecord(traffic *model.Traffic) error {
	return nil
}

// CleanupTraffic 清理流量记录
func (m *MockDB) CleanupTraffic(before time.Time) error {
	return nil
}

// CreateProtocol 创建协议
func (m *MockDB) CreateProtocol(protocol *model.Protocol) error {
	return nil
}

// GetProtocol 获取协议
func (m *MockDB) GetProtocol(id int64) (*model.Protocol, error) {
	return &model.Protocol{}, nil
}

// GetProtocolsByUserID 获取用户协议
func (m *MockDB) GetProtocolsByUserID(userID int64) ([]*model.Protocol, error) {
	return nil, nil
}

// UpdateProtocol 更新协议
func (m *MockDB) UpdateProtocol(protocol *model.Protocol) error {
	return nil
}

// DeleteProtocol 删除协议
func (m *MockDB) DeleteProtocol(id int64) error {
	return nil
}

// GetProtocolsByPort 获取指定端口的协议
func (m *MockDB) GetProtocolsByPort(port int) ([]*model.Protocol, error) {
	return nil, nil
}

// ListProtocols 列出协议
func (m *MockDB) ListProtocols(page, pageSize int) ([]*model.Protocol, error) {
	return nil, nil
}

// GetTotalProtocols 获取协议总数
func (m *MockDB) GetTotalProtocols() (int64, error) {
	return 0, nil
}

// SearchProtocols 搜索协议
func (m *MockDB) SearchProtocols(keyword string) ([]*model.Protocol, error) {
	return nil, nil
}

// CreateProtocolStats 创建协议统计
func (m *MockDB) CreateProtocolStats(stats *model.ProtocolStats) error {
	return nil
}

// GetProtocolStats 获取协议统计
func (m *MockDB) GetProtocolStats(id int64) (*model.ProtocolStats, error) {
	return &model.ProtocolStats{}, nil
}

// UpdateProtocolStats 更新协议统计
func (m *MockDB) UpdateProtocolStats(stats *model.ProtocolStats) error {
	return nil
}

// DeleteProtocolStats 删除协议统计
func (m *MockDB) DeleteProtocolStats(id int64) error {
	return nil
}

// ListProtocolStatsByUserID 获取用户协议统计
func (m *MockDB) ListProtocolStatsByUserID(userID int64) ([]*model.ProtocolStats, error) {
	return nil, nil
}

// CreateCertificate 创建证书
func (m *MockDB) CreateCertificate(cert *model.Certificate) error {
	return nil
}

// GetCertificate 获取证书
func (m *MockDB) GetCertificate(id int64) (*model.Certificate, error) {
	return &model.Certificate{}, nil
}

// GetCertificateByDomain 通过域名获取证书
func (m *MockDB) GetCertificateByDomain(domain string) (*model.Certificate, error) {
	return &model.Certificate{}, nil
}

// UpdateCertificate 更新证书
func (m *MockDB) UpdateCertificate(cert *model.Certificate) error {
	return nil
}

// DeleteCertificate 删除证书
func (m *MockDB) DeleteCertificate(id int64) error {
	return nil
}

// ListCertificates 列出证书
func (m *MockDB) ListCertificates() ([]*model.Certificate, error) {
	return nil, nil
}

// GetExpiredCertificates 获取过期证书
func (m *MockDB) GetExpiredCertificates() ([]*model.Certificate, error) {
	return nil, nil
}

// CreateDailyStats 创建每日统计
func (m *MockDB) CreateDailyStats(stats *model.DailyStats) error {
	return nil
}

// GetDailyStats 获取每日统计
func (m *MockDB) GetDailyStats(id int64) (*model.DailyStats, error) {
	return &model.DailyStats{}, nil
}

// GetDailyStatsByUserAndDate 通过用户和日期获取每日统计
func (m *MockDB) GetDailyStatsByUserAndDate(userID int64, date time.Time) (*model.DailyStats, error) {
	return &model.DailyStats{}, nil
}

// UpdateDailyStats 更新每日统计
func (m *MockDB) UpdateDailyStats(stats *model.DailyStats) error {
	return nil
}

// ListDailyStatsByUserID 获取用户每日统计
func (m *MockDB) ListDailyStatsByUserID(userID int64, start, end time.Time) ([]*model.DailyStats, error) {
	return nil, nil
}

// Begin 开始事务
func (m *MockDB) Begin() error {
	return nil
}

// Commit 提交事务
func (m *MockDB) Commit() error {
	return nil
}

// Rollback 回滚事务
func (m *MockDB) Rollback() error {
	return nil
}

// AutoMigrate 自动迁移
func (m *MockDB) AutoMigrate() error {
	return nil
}

// Close 关闭数据库连接
func (m *MockDB) Close() error {
	return nil
}

func init() {
	// Skip initialization to avoid "limit" keyword error
	fmt.Println("Skipping database auto-initialization to avoid SQL syntax errors")
}

// ListAlertRecords returns all alert records
func (db *Database) ListAlertRecords(out *[]*model.AlertRecord) error {
	return db.DB.Find(out).Error
}

// CreateTrafficHistory creates a new traffic history record
func (db *Database) CreateTrafficHistory(history *model.TrafficHistory) error {
	return db.DB.Create(history).Error
}

// ListTrafficHistoryByDateRange returns traffic history records within a date range
func (db *Database) ListTrafficHistoryByDateRange(userID uint, startDate, endDate string, histories *[]*model.TrafficHistory) error {
	return db.DB.Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).Find(histories).Error
}

// GetSettings retrieves a setting value by key
func (db *Database) GetSettings(key string) (string, error) {
	var value string
	err := db.DB.Raw("SELECT value FROM settings WHERE key = ?", key).Scan(&value).Error
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSettings sets a setting value by key
func (db *Database) SetSettings(key, value string) error {
	return db.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value).Error
}
