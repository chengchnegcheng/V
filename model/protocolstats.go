package model

import (
	"database/sql"
	"time"
)

// ProtocolStatsManager 协议统计管理器
type ProtocolStatsManager struct {
	db *sql.DB
}

// NewProtocolStatsManager 创建协议统计管理器
func NewProtocolStatsManager(db *sql.DB) *ProtocolStatsManager {
	return &ProtocolStatsManager{db: db}
}

// CreateStats 创建协议统计
func (m *ProtocolStatsManager) CreateStats(stats *ProtocolStats) error {
	query := `INSERT INTO protocol_stats (protocol_id, user_id, upload, download, last_active, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	stats.CreatedAt = now
	stats.UpdatedAt = now

	_, err := m.db.Exec(
		query,
		stats.ProtocolID,
		stats.UserID,
		stats.Upload,
		stats.Download,
		stats.LastActive,
		stats.CreatedAt,
		stats.UpdatedAt,
	)
	return err
}

// UpdateStats 更新协议统计
func (m *ProtocolStatsManager) UpdateStats(stats *ProtocolStats) error {
	query := `UPDATE protocol_stats
              SET upload = ?, download = ?, last_active = ?, updated_at = ?
              WHERE id = ?`

	stats.UpdatedAt = time.Now()

	_, err := m.db.Exec(
		query,
		stats.Upload,
		stats.Download,
		stats.LastActive,
		stats.UpdatedAt,
		stats.ID,
	)
	return err
}

// GetStatsByID 根据ID获取协议统计
func (m *ProtocolStatsManager) GetStatsByID(id int64) (*ProtocolStats, error) {
	query := `SELECT id, protocol_id, user_id, upload, download, last_active, created_at, updated_at
              FROM protocol_stats
              WHERE id = ?`

	stats := &ProtocolStats{}
	var lastActiveStr, createdAtStr, updatedAtStr string

	err := m.db.QueryRow(query, id).Scan(
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
			return nil, ErrNotFound
		}
		return nil, err
	}

	// 解析时间
	stats.LastActive, _ = time.Parse("2006-01-02 15:04:05", lastActiveStr)
	stats.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	stats.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return stats, nil
}

// GetStatsByProtocolID 根据协议ID获取统计
func (m *ProtocolStatsManager) GetStatsByProtocolID(protocolID int64) (*ProtocolStats, error) {
	query := `SELECT id, protocol_id, user_id, upload, download, last_active, created_at, updated_at
              FROM protocol_stats
              WHERE protocol_id = ?`

	stats := &ProtocolStats{}
	var lastActiveStr, createdAtStr, updatedAtStr string

	err := m.db.QueryRow(query, protocolID).Scan(
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
			return nil, ErrNotFound
		}
		return nil, err
	}

	// 解析时间
	stats.LastActive, _ = time.Parse("2006-01-02 15:04:05", lastActiveStr)
	stats.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	stats.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return stats, nil
}

// ListStatsByUserID 获取用户所有协议统计
func (m *ProtocolStatsManager) ListStatsByUserID(userID int64) ([]*ProtocolStats, error) {
	query := `SELECT ps.id, ps.protocol_id, ps.user_id, ps.upload, ps.download, ps.last_active, ps.created_at, ps.updated_at
              FROM protocol_stats ps
              INNER JOIN protocols p ON ps.protocol_id = p.id
              WHERE p.user_id = ?`

	rows, err := m.db.Query(query, userID)
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
