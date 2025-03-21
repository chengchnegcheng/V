package migration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"v/logger"
)

// SQLMigration represents a SQL migration
type SQLMigration struct {
	Version int
	UpSQL   string
	DownSQL string
}

// SQLManager handles SQL migrations
type SQLManager struct {
	log        *logger.Logger
	db         *sql.DB
	migrations []SQLMigration
	currentVer int
	targetVer  int
}

// NewSQLManager creates a new SQL migration manager
func NewSQLManager(log *logger.Logger, db *sql.DB) *SQLManager {
	return &SQLManager{
		log:        log,
		db:         db,
		migrations: make([]SQLMigration, 0),
	}
}

// LoadMigrations loads SQL migrations from a directory
func (m *SQLManager) LoadMigrations(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Load migrations
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Parse version from filename
		var version int
		_, err := fmt.Sscanf(file.Name(), "%d_", &version)
		if err != nil {
			return fmt.Errorf("failed to parse version from filename %s: %v", file.Name(), err)
		}

		// Read migration file
		content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file.Name(), err)
		}

		// Split into up and down migrations
		parts := strings.Split(string(content), "-- Down migration")
		if len(parts) != 2 {
			return fmt.Errorf("invalid migration file %s: missing down migration", file.Name())
		}

		upSQL := strings.TrimSpace(parts[0])
		downSQL := strings.TrimSpace(parts[1])

		m.migrations = append(m.migrations, SQLMigration{
			Version: version,
			UpSQL:   upSQL,
			DownSQL: downSQL,
		})
	}

	return nil
}

// GetCurrentVersion returns the current database version
func (m *SQLManager) GetCurrentVersion() (int, error) {
	// Create migrations table if it doesn't exist
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return 0, fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get current version
	var version int
	err = m.db.QueryRow("SELECT version FROM migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %v", err)
	}

	return version, nil
}

// Migrate performs database migrations
func (m *SQLManager) Migrate(targetVer int) error {
	// Get current version
	currentVer, err := m.GetCurrentVersion()
	if err != nil {
		return err
	}
	m.currentVer = currentVer
	m.targetVer = targetVer

	// Determine migration direction
	if targetVer > currentVer {
		return m.migrateUp()
	} else if targetVer < currentVer {
		return m.migrateDown()
	}

	return nil
}

// migrateUp performs upward migrations
func (m *SQLManager) migrateUp() error {
	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Apply migrations
	for _, migration := range m.migrations {
		if migration.Version <= m.currentVer {
			continue
		}
		if migration.Version > m.targetVer {
			break
		}

		m.log.Info("Applying migration", logger.Fields{"version": migration.Version})
		if _, err := tx.Exec(migration.UpSQL); err != nil {
			return fmt.Errorf("failed to apply migration %d: %v", migration.Version, err)
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO migrations (version, applied_at) VALUES ($1, $2)",
			migration.Version, time.Now())
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %v", migration.Version, err)
		}

		m.log.Info("Applied migration", logger.Fields{"version": migration.Version})
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// migrateDown performs downward migrations
func (m *SQLManager) migrateDown() error {
	// Sort migrations by version in descending order
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version > m.migrations[j].Version
	})

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Apply migrations
	for _, migration := range m.migrations {
		if migration.Version > m.currentVer {
			continue
		}
		if migration.Version <= m.targetVer {
			break
		}

		m.log.Info("Rolling back migration", logger.Fields{"version": migration.Version})
		if _, err := tx.Exec(migration.DownSQL); err != nil {
			return fmt.Errorf("failed to roll back migration %d: %v", migration.Version, err)
		}

		// Remove migration record
		_, err = tx.Exec("DELETE FROM migrations WHERE version = $1", migration.Version)
		if err != nil {
			return fmt.Errorf("failed to remove migration record %d: %v", migration.Version, err)
		}

		m.log.Info("Rolled back migration", logger.Fields{"version": migration.Version})
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
