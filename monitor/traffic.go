package monitor

import (
	"sync"
	"time"
	"v/model"
)

// TrafficMonitor handles traffic monitoring
type TrafficMonitor struct {
	db model.DB
	mu sync.RWMutex
}

// NewTrafficMonitor creates a new traffic monitor
func NewTrafficMonitor(db model.DB) *TrafficMonitor {
	return &TrafficMonitor{
		db: db,
	}
}

// UpdateTraffic updates traffic statistics for a user
func (m *TrafficMonitor) UpdateTraffic(userID int64, upload, download int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get current traffic stats for user
	traffics, err := m.db.ListTrafficByUserID(userID)
	if err != nil {
		return err
	}

	// 创建或更新流量统计记录
	var totalUpload, totalDownload int64
	if len(traffics) == 0 {
		// 如果不存在，以当前流量为统计值
		totalUpload = upload
		totalDownload = download
	} else {
		// 累计流量
		for _, t := range traffics {
			totalUpload += t.Upload
			totalDownload += t.Download
		}
		totalUpload += upload
		totalDownload += download
	}

	// 创建新的流量记录
	traffic := &model.Traffic{
		UserID: userID,
		Up:     upload,
		Down:   download,
	}

	// 保存到数据库
	if err := m.db.CreateTrafficRecord(traffic); err != nil {
		return err
	}

	// 创建每日统计
	today := time.Now().Truncate(24 * time.Hour)
	dailyStats := &model.DailyStats{
		UserID:   userID,
		Date:     today,
		Upload:   upload,
		Download: download,
		Total:    upload + download,
	}

	// 保存每日统计
	return m.db.CreateDailyStats(dailyStats)
}

// GetTrafficStats returns traffic statistics for a user
func (m *TrafficMonitor) GetTrafficStats(userID int64) (*model.TrafficStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 获取用户所有流量统计
	traffics, err := m.db.ListTrafficByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 汇总流量数据
	stats := &model.TrafficStats{
		UserID:   userID,
		Upload:   0,
		Download: 0,
		Total:    0,
	}

	for _, t := range traffics {
		stats.Upload += t.Upload
		stats.Download += t.Download
		stats.Total += t.Upload + t.Download
	}

	return stats, nil
}

// GetDailyStats returns daily traffic statistics for a user
func (m *TrafficMonitor) GetDailyStats(userID int64, start, end time.Time) ([]*model.DailyStats, error) {
	return m.db.ListDailyStatsByUserID(userID)
}

// ResetTraffic resets traffic statistics for a user
func (m *TrafficMonitor) ResetTraffic(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清理用户的流量记录
	return m.db.CleanupTraffic(time.Now())
}
