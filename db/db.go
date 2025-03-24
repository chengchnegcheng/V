package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"v/common"
	verrors "v/errors"
	"v/logger"
	"v/model"
)

// DB represents a database implementation
type DB struct {
	log    *logger.Logger
	db     *sql.DB
	tx     *sql.Tx
	models *model.DB
}

// New creates a new database instance
func New(log *logger.Logger, dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{
		log: log,
		db:  db,
	}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.db.Close()
}

// Begin starts a transaction
func (d *DB) Begin() (*sql.Tx, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	d.tx = tx
	return tx, nil
}

// Commit commits a transaction
func (d *DB) Commit(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	d.tx = nil
	return nil
}

// Rollback rolls back a transaction
func (d *DB) Rollback(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %v", err)
	}
	d.tx = nil
	return nil
}

// CreateUser creates a new user
func (d *DB) CreateUser(user *model.User) error {
	query := `
		INSERT INTO users (
			username, email, password, salt, is_admin,
			traffic_limit, traffic_used, expire_at,
			last_login_at, login_attempts, locked_until,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id`

	var id int64
	err := d.db.QueryRow(
		query,
		user.Username, user.Email, user.Password, user.Salt,
		user.IsAdmin, user.TrafficLimit, user.TrafficUsed,
		user.ExpireAt, user.LastLoginAt, user.LoginAttempts,
		user.LockedUntil, user.CreatedAt, user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	user.ID = id
	return nil
}

// GetUser returns a user by ID
func (d *DB) GetUser(id int64) (*model.User, error) {
	query := `
		SELECT id, username, email, password, salt, is_admin,
			traffic_limit, traffic_used, expire_at,
			last_login_at, login_attempts, locked_until,
			created_at, updated_at
		FROM users WHERE id = $1`

	user := &model.User{}
	err := d.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Salt, &user.IsAdmin, &user.TrafficLimit,
		&user.TrafficUsed, &user.ExpireAt, &user.LastLoginAt,
		&user.LoginAttempts, &user.LockedUntil, &user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetUserByUsername returns a user by username
func (d *DB) GetUserByUsername(username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, salt, is_admin,
			traffic_limit, traffic_used, expire_at,
			last_login_at, login_attempts, locked_until,
			created_at, updated_at
		FROM users WHERE username = $1`

	user := &model.User{}
	err := d.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Salt, &user.IsAdmin, &user.TrafficLimit,
		&user.TrafficUsed, &user.ExpireAt, &user.LastLoginAt,
		&user.LoginAttempts, &user.LockedUntil, &user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetUserByEmail returns a user by email
func (d *DB) GetUserByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, salt, is_admin,
			traffic_limit, traffic_used, expire_at,
			last_login_at, login_attempts, locked_until,
			created_at, updated_at
		FROM users WHERE email = $1`

	user := &model.User{}
	err := d.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Salt, &user.IsAdmin, &user.TrafficLimit,
		&user.TrafficUsed, &user.ExpireAt, &user.LastLoginAt,
		&user.LoginAttempts, &user.LockedUntil, &user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// UpdateUser updates a user
func (d *DB) UpdateUser(user *model.User) error {
	query := `
		UPDATE users SET
			username = $1, email = $2, password = $3,
			salt = $4, is_admin = $5, traffic_limit = $6,
			traffic_used = $7, expire_at = $8,
			last_login_at = $9, login_attempts = $10,
			locked_until = $11, updated_at = $12
		WHERE id = $13`

	result, err := d.db.Exec(
		query,
		user.Username, user.Email, user.Password,
		user.Salt, user.IsAdmin, user.TrafficLimit,
		user.TrafficUsed, user.ExpireAt, user.LastLoginAt,
		user.LoginAttempts, user.LockedUntil, user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}

// DeleteUser deletes a user
func (d *DB) DeleteUser(id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}

// ListUsers returns a list of users
func (d *DB) ListUsers(offset, limit int) ([]*model.User, error) {
	query := `
		SELECT id, username, email, password, salt, is_admin,
			traffic_limit, traffic_used, expire_at,
			last_login_at, login_attempts, locked_until,
			created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.Salt, &user.IsAdmin, &user.TrafficLimit,
			&user.TrafficUsed, &user.ExpireAt, &user.LastLoginAt,
			&user.LoginAttempts, &user.LockedUntil, &user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %v", err)
	}

	return users, nil
}

// CreateProxy 鍒涘缓浠ｇ悊
func (d *DB) CreateProxy(proxy *common.Proxy) error {
	query := `
		INSERT INTO proxies (
			user_id, protocol, port, config, settings,
			listen_addr, remote_addr, enabled,
			created_at, updated_at, expire_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id`

	var id int64
	err := d.db.QueryRow(
		query,
		proxy.UserID, proxy.Protocol, proxy.Port,
		proxy.Config, proxy.Settings,
		proxy.ListenAddr, proxy.RemoteAddr,
		proxy.Enabled,
		proxy.CreatedAt, proxy.UpdatedAt,
		proxy.ExpireAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	proxy.ID = id
	return nil
}

// GetProxy 鑾峰彇浠ｇ悊
func (d *DB) GetProxy(id int64) (*common.Proxy, error) {
	query := `
		SELECT id, user_id, protocol, port, config, settings,
			listen_addr, remote_addr, enabled,
			upload, download, last_active_at,
			created_at, updated_at, expire_at
		FROM proxies WHERE id = $1`

	proxy := &common.Proxy{}
	err := d.db.QueryRow(query, id).Scan(
		&proxy.ID, &proxy.UserID, &proxy.Protocol,
		&proxy.Port, &proxy.Config, &proxy.Settings,
		&proxy.ListenAddr, &proxy.RemoteAddr,
		&proxy.Enabled,
		&proxy.Upload, &proxy.Download,
		&proxy.LastActiveAt,
		&proxy.CreatedAt, &proxy.UpdatedAt,
		&proxy.ExpireAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy: %v", err)
	}

	return proxy, nil
}

// GetProxiesByUser returns all proxies for a user
func (d *DB) GetProxiesByUser(userID int64) ([]*common.Proxy, error) {
	query := `
		SELECT id, user_id, port, protocol, listen_addr,
			remote_addr, enabled, last_active_at,
			created_at, updated_at
		FROM proxies WHERE user_id = $1`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list proxies: %v", err)
	}
	defer rows.Close()

	var proxies []*common.Proxy
	for rows.Next() {
		proxy := &common.Proxy{}
		err := rows.Scan(
			&proxy.ID, &proxy.UserID, &proxy.Port,
			&proxy.Protocol, &proxy.ListenAddr, &proxy.RemoteAddr,
			&proxy.Enabled, &proxy.LastActiveAt, &proxy.CreatedAt,
			&proxy.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan proxy: %v", err)
		}

		proxies = append(proxies, proxy)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate proxies: %v", err)
	}

	return proxies, nil
}

// UpdateProxy 鏇存柊浠ｇ悊
func (d *DB) UpdateProxy(proxy *common.Proxy) error {
	query := `
		UPDATE proxies SET
			user_id = $1, protocol = $2, port = $3,
			config = $4, settings = $5,
			listen_addr = $6, remote_addr = $7,
			enabled = $8,
			upload = $9, download = $10,
			last_active_at = $11,
			updated_at = $12, expire_at = $13
		WHERE id = $14`

	result, err := d.db.Exec(
		query,
		proxy.UserID, proxy.Protocol, proxy.Port,
		proxy.Config, proxy.Settings,
		proxy.ListenAddr, proxy.RemoteAddr,
		proxy.Enabled,
		proxy.Upload, proxy.Download,
		proxy.LastActiveAt,
		proxy.UpdatedAt, proxy.ExpireAt,
		proxy.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update proxy: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}

// DeleteProxy 鍒犻櫎浠ｇ悊
func (d *DB) DeleteProxy(id int64) error {
	query := `DELETE FROM proxies WHERE id = $1`
	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete proxy: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}

// ListProxies 鍒楀嚭浠ｇ悊
func (d *DB) ListProxies(offset, limit int) ([]*common.Proxy, error) {
	query := `
		SELECT id, user_id, protocol, port, config, settings,
			listen_addr, remote_addr, enabled,
			upload, download, last_active_at,
			created_at, updated_at, expire_at
		FROM proxies
		ORDER BY id DESC
		LIMIT $1 OFFSET $2`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list proxies: %v", err)
	}
	defer rows.Close()

	var proxies []*common.Proxy
	for rows.Next() {
		proxy := &common.Proxy{}
		err := rows.Scan(
			&proxy.ID, &proxy.UserID, &proxy.Protocol,
			&proxy.Port, &proxy.Config, &proxy.Settings,
			&proxy.ListenAddr, &proxy.RemoteAddr,
			&proxy.Enabled,
			&proxy.Upload, &proxy.Download,
			&proxy.LastActiveAt,
			&proxy.CreatedAt, &proxy.UpdatedAt,
			&proxy.ExpireAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proxy: %v", err)
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

// CreateCertificate creates a new SSL certificate
func (d *DB) CreateCertificate(cert *model.Certificate) error {
	query := `
		INSERT INTO certificates (
			domain, cert_file, key_file, last_renewed_at,
			expires_at, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id`

	var id int64
	err := d.db.QueryRow(
		query,
		cert.Domain, cert.CertFile, cert.KeyFile,
		cert.LastRenewedAt, cert.ExpiresAt, cert.Status,
		cert.CreatedAt, cert.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	cert.ID = id
	return nil
}

// GetCertificate returns a certificate by ID
func (d *DB) GetCertificate(id int64) (*model.Certificate, error) {
	query := `
		SELECT id, domain, cert_file, key_file,
			last_renewed_at, expires_at, status,
			created_at, updated_at
		FROM certificates WHERE id = $1`

	cert := &model.Certificate{}
	err := d.db.QueryRow(query, id).Scan(
		&cert.ID, &cert.Domain, &cert.CertFile,
		&cert.KeyFile, &cert.LastRenewedAt, &cert.ExpiresAt,
		&cert.Status, &cert.CreatedAt, &cert.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %v", err)
	}

	return cert, nil
}

// GetCertificateByDomain returns a certificate by domain
func (d *DB) GetCertificateByDomain(domain string) (*model.Certificate, error) {
	query := `
		SELECT id, domain, cert_file, key_file,
			last_renewed_at, expires_at, status,
			created_at, updated_at
		FROM certificates WHERE domain = $1`

	cert := &model.Certificate{}
	err := d.db.QueryRow(query, domain).Scan(
		&cert.ID, &cert.Domain, &cert.CertFile,
		&cert.KeyFile, &cert.LastRenewedAt, &cert.ExpiresAt,
		&cert.Status, &cert.CreatedAt, &cert.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %v", err)
	}

	return cert, nil
}

// UpdateCertificate updates a certificate
func (d *DB) UpdateCertificate(cert *model.Certificate) error {
	query := `
		UPDATE certificates SET
			domain = $1, cert_file = $2, key_file = $3,
			last_renewed_at = $4, expires_at = $5, status = $6,
			updated_at = $7
		WHERE id = $8`

	result, err := d.db.Exec(
		query,
		cert.Domain, cert.CertFile, cert.KeyFile,
		cert.LastRenewedAt, cert.ExpiresAt, cert.Status,
		cert.UpdatedAt, cert.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update certificate: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}

// DeleteCertificate deletes a certificate
func (d *DB) DeleteCertificate(id int64) error {
	query := `DELETE FROM certificates WHERE id = $1`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}

// ListCertificates returns a list of certificates
func (d *DB) ListCertificates(offset, limit int) ([]*model.Certificate, error) {
	query := `
		SELECT id, domain, cert_file, key_file,
			last_renewed_at, expires_at, status,
			created_at, updated_at
		FROM certificates
		ORDER BY id
		LIMIT $1 OFFSET $2`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %v", err)
	}
	defer rows.Close()

	var certificates []*model.Certificate
	for rows.Next() {
		cert := &model.Certificate{}
		err := rows.Scan(
			&cert.ID, &cert.Domain, &cert.CertFile,
			&cert.KeyFile, &cert.LastRenewedAt, &cert.ExpiresAt,
			&cert.Status, &cert.CreatedAt, &cert.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan certificate: %v", err)
		}

		certificates = append(certificates, cert)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate certificates: %v", err)
	}

	return certificates, nil
}

// CreateTrafficStats creates traffic statistics
func (d *DB) CreateTrafficStats(stats *model.TrafficStats) error {
	query := `
		INSERT INTO traffic_stats (
			user_id, upload, download, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING id`

	var id int64
	err := d.db.QueryRow(
		query,
		stats.UserID, stats.Upload, stats.Download,
		stats.CreatedAt, stats.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create traffic stats: %v", err)
	}

	stats.ID = id
	return nil
}

// GetTrafficStats returns traffic statistics for a user
func (d *DB) GetTrafficStats(userID int64) (*model.TrafficStats, error) {
	query := `
		SELECT id, user_id, upload, download,
			created_at, updated_at
		FROM traffic_stats WHERE user_id = $1`

	stats := &model.TrafficStats{}
	err := d.db.QueryRow(query, userID).Scan(
		&stats.ID, &stats.UserID, &stats.Upload,
		&stats.Download, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic stats: %v", err)
	}

	return stats, nil
}

// UpdateTrafficStats updates traffic statistics
func (d *DB) UpdateTrafficStats(stats *model.TrafficStats) error {
	query := `
		UPDATE traffic_stats SET
			user_id = $1, upload = $2, download = $3,
			updated_at = $4
		WHERE id = $5`

	result, err := d.db.Exec(
		query,
		stats.UserID, stats.Upload, stats.Download,
		stats.UpdatedAt, stats.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update traffic stats: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}

// CreateDailyStats creates daily traffic statistics
func (d *DB) CreateDailyStats(stats *model.DailyStats) error {
	query := `
		INSERT INTO daily_stats (
			user_id, date, upload, download, total,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id`

	var id int64
	err := d.db.QueryRow(
		query,
		stats.UserID, stats.Date, stats.Upload,
		stats.Download, stats.Total, stats.CreatedAt,
		stats.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create daily stats: %v", err)
	}

	stats.ID = id
	return nil
}

// GetDailyStats returns daily traffic statistics for a user
func (d *DB) GetDailyStats(userID int64, start, end time.Time) ([]*model.DailyStats, error) {
	query := `
		SELECT id, user_id, date, upload, download,
			total, created_at, updated_at
		FROM daily_stats
		WHERE user_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date`

	rows, err := d.db.Query(query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list daily stats: %v", err)
	}
	defer rows.Close()

	var stats []*model.DailyStats
	for rows.Next() {
		s := &model.DailyStats{}
		err := rows.Scan(
			&s.ID, &s.UserID, &s.Date, &s.Upload,
			&s.Download, &s.Total, &s.CreatedAt,
			&s.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %v", err)
		}

		stats = append(stats, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate daily stats: %v", err)
	}

	return stats, nil
}

// CreateEvent creates an audit event
func (d *DB) CreateEvent(event *model.Event) error {
	query := `
		INSERT INTO events (
			user_id, username, action, resource,
			details, ip, user_agent, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id`

	var id int64
	err := d.db.QueryRow(
		query,
		event.UserID, event.Username, event.Action,
		event.Resource, event.Details, event.IP,
		event.UserAgent, event.CreatedAt, event.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create event: %v", err)
	}

	event.ID = id
	return nil
}

// GetEvent returns an event by ID
func (d *DB) GetEvent(id int64) (*model.Event, error) {
	query := `
		SELECT id, user_id, username, action, resource,
			details, ip, user_agent, created_at, updated_at
		FROM events WHERE id = $1`

	event := &model.Event{}
	err := d.db.QueryRow(query, id).Scan(
		&event.ID, &event.UserID, &event.Username,
		&event.Action, &event.Resource, &event.Details,
		&event.IP, &event.UserAgent, &event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %v", err)
	}

	return event, nil
}

// ListEvents returns a list of events
func (d *DB) ListEvents(userID int64, start, end time.Time) ([]*model.Event, error) {
	query := `
		SELECT id, user_id, username, action, resource,
			details, ip, user_agent, created_at, updated_at
		FROM events
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC`

	rows, err := d.db.Query(query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %v", err)
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		event := &model.Event{}
		err := rows.Scan(
			&event.ID, &event.UserID, &event.Username,
			&event.Action, &event.Resource, &event.Details,
			&event.IP, &event.UserAgent, &event.CreatedAt,
			&event.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %v", err)
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate events: %v", err)
	}

	return events, nil
}

// CreateBackup creates a backup record
func (d *DB) CreateBackup(backup *model.Backup) error {
	query := `
		INSERT INTO backups (
			path, size, status, timestamp,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id`

	var id int64
	err := d.db.QueryRow(
		query,
		backup.Path, backup.Size, backup.Status,
		backup.Timestamp, backup.CreatedAt, backup.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	backup.ID = id
	return nil
}

// GetBackup returns a backup by ID
func (d *DB) GetBackup(id int64) (*model.Backup, error) {
	query := `
		SELECT id, path, size, status, timestamp,
			created_at, updated_at
		FROM backups WHERE id = $1`

	backup := &model.Backup{}
	err := d.db.QueryRow(query, id).Scan(
		&backup.ID, &backup.Path, &backup.Size,
		&backup.Status, &backup.Timestamp, &backup.CreatedAt,
		&backup.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, verrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %v", err)
	}

	return backup, nil
}

// ListBackups returns a list of backups
func (d *DB) ListBackups(offset, limit int) ([]*model.Backup, error) {
	query := `
		SELECT id, path, size, status, timestamp,
			created_at, updated_at
		FROM backups
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %v", err)
	}
	defer rows.Close()

	var backups []*model.Backup
	for rows.Next() {
		backup := &model.Backup{}
		err := rows.Scan(
			&backup.ID, &backup.Path, &backup.Size,
			&backup.Status, &backup.Timestamp, &backup.CreatedAt,
			&backup.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan backup: %v", err)
		}

		backups = append(backups, backup)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate backups: %v", err)
	}

	return backups, nil
}

// DeleteBackup deletes a backup
func (d *DB) DeleteBackup(id int64) error {
	query := `DELETE FROM backups WHERE id = $1`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return verrors.ErrNotFound
	}

	return nil
}
