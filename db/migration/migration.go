package migration

import (
	"database/sql"
	"fmt"
	"time"

	"v/logger"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Up      func(*sql.Tx) error
	Down    func(*sql.Tx) error
}

// Manager handles database migrations
type Manager struct {
	log        *logger.Logger
	db         *sql.DB
	migrations []Migration
	currentVer int
	targetVer  int
}

// New creates a new migration manager
func New(log *logger.Logger, db *sql.DB) *Manager {
	return &Manager{
		log:        log,
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Add adds a new migration
func (m *Manager) Add(version int, up, down func(*sql.Tx) error) {
	m.migrations = append(m.migrations, Migration{
		Version: version,
		Up:      up,
		Down:    down,
	})
}

// GetCurrentVersion returns the current database version
func (m *Manager) GetCurrentVersion() (int, error) {
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
func (m *Manager) Migrate(targetVer int) error {
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
func (m *Manager) migrateUp() error {
	// Sort migrations by version
	sortMigrations(m.migrations)

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
		if err := migration.Up(tx); err != nil {
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
func (m *Manager) migrateDown() error {
	// Sort migrations by version in descending order
	sortMigrationsDesc(m.migrations)

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
		if err := migration.Down(tx); err != nil {
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

// sortMigrations sorts migrations by version in ascending order
func sortMigrations(migrations []Migration) {
	for i := 0; i < len(migrations)-1; i++ {
		for j := i + 1; j < len(migrations); j++ {
			if migrations[i].Version > migrations[j].Version {
				migrations[i], migrations[j] = migrations[j], migrations[i]
			}
		}
	}
}

// sortMigrationsDesc sorts migrations by version in descending order
func sortMigrationsDesc(migrations []Migration) {
	for i := 0; i < len(migrations)-1; i++ {
		for j := i + 1; j < len(migrations); j++ {
			if migrations[i].Version < migrations[j].Version {
				migrations[i], migrations[j] = migrations[j], migrations[i]
			}
		}
	}
}
