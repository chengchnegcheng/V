package stats

import (
	"fmt"
	"time"

	"v/logger"
	"v/model"
)

// StatsManager 流量统计管理器
type StatsManager struct {
	log    *logger.Logger
	db     model.DB
	stopCh chan struct{}
}

// NewStatsManager 创建流量统计管理器
func NewStatsManager(log *logger.Logger, db model.DB) *StatsManager {
	return &StatsManager{
		log:    log,
		db:     db,
		stopCh: make(chan struct{}),
	}
}

// Start 启动流量统计
func (m *StatsManager) Start() error {
	go m.statsLoop()
	return nil
}

// Stop 停止流量统计
func (m *StatsManager) Stop() {
	close(m.stopCh)
}

// statsLoop 统计循环
func (m *StatsManager) statsLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if err := m.collectDailyStats(); err != nil {
				m.log.Error("Failed to collect daily stats", logger.Fields{
					"error": err.Error(),
				})
				continue
			}

			// 检查并清理过期的统计数据
			if err := m.cleanupOldStats(); err != nil {
				m.log.Error("Failed to cleanup old stats", logger.Fields{
					"error": err.Error(),
				})
			}
		}
	}
}

// collectDailyStats 收集每日流量统计
func (m *StatsManager) collectDailyStats() error {
	// 获取所有协议
	protocols, err := m.db.ListProtocols(0, 0)
	if err != nil {
		return fmt.Errorf("failed to get protocols: %v", err)
	}

	// 获取今天的日期（UTC）
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// 开始事务
	if err := m.db.Begin(); err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// 为每个协议创建每日统计
	for _, protocol := range protocols {
		stats := &model.DailyStats{
			UserID:   protocol.UserID,
			Date:     today,
			Upload:   protocol.TrafficUsed,
			Download: protocol.TrafficUsed,
			Total:    protocol.TrafficUsed * 2,
		}

		if err := m.db.CreateDailyStats(stats); err != nil {
			m.db.Rollback()
			return fmt.Errorf("failed to create daily stats: %v", err)
		}

		// 重置协议流量计数
		protocol.TrafficUsed = 0
		if err := m.db.UpdateProtocol(protocol); err != nil {
			m.db.Rollback()
			return fmt.Errorf("failed to reset protocol traffic: %v", err)
		}
	}

	// 提交事务
	if err := m.db.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// cleanupOldStats 清理过期的统计数据
func (m *StatsManager) cleanupOldStats() error {
	// 删除30天前的统计数据
	cutoff := time.Now().UTC().AddDate(0, 0, -30)
	if err := m.db.DeleteDailyStatsBefore(cutoff); err != nil {
		return fmt.Errorf("failed to cleanup old stats: %v", err)
	}

	return nil
}

// GetUserStats 获取用户流量统计
func (m *StatsManager) GetUserStats(userID int64, startDate, endDate time.Time) ([]*model.DailyStats, error) {
	// 获取用户的每日统计
	stats, err := m.db.ListDailyStatsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %v", err)
	}

	// 过滤日期范围
	var filteredStats []*model.DailyStats
	for _, stat := range stats {
		if stat.Date.After(startDate) && stat.Date.Before(endDate) {
			filteredStats = append(filteredStats, stat)
		}
	}

	return filteredStats, nil
}

// GetProtocolStats 获取协议流量统计
func (m *StatsManager) GetProtocolStats(protocolID int64, startDate, endDate time.Time) ([]*model.ProtocolStats, error) {
	// 获取协议的统计
	stats, err := m.db.ListProtocolStatsByProtocolID(protocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get protocol stats: %v", err)
	}

	// 过滤日期范围
	var filteredStats []*model.ProtocolStats
	for _, stat := range stats {
		if stat.Date.After(startDate) && stat.Date.Before(endDate) {
			filteredStats = append(filteredStats, stat)
		}
	}

	return filteredStats, nil
}

// UpdateProtocolTraffic 更新协议流量
func (m *StatsManager) UpdateProtocolTraffic(protocolID int64, upload, download int64) error {
	// 获取协议
	protocol, err := m.db.GetProtocol(protocolID)
	if err != nil {
		return fmt.Errorf("failed to get protocol: %v", err)
	}

	// 更新流量
	protocol.TrafficUsed += upload + download

	// 检查流量限制
	if protocol.TrafficLimit > 0 && protocol.TrafficUsed > protocol.TrafficLimit {
		// 禁用协议
		protocol.Enable = false
		if err := m.db.UpdateProtocol(protocol); err != nil {
			return fmt.Errorf("failed to update protocol: %v", err)
		}

		// 记录日志
		m.log.Warn("Protocol traffic limit exceeded", logger.Fields{
			"protocol_id":   protocolID,
			"traffic_used":  protocol.TrafficUsed,
			"traffic_limit": protocol.TrafficLimit,
		})

		return model.ErrTrafficLimitExceeded
	}

	// 更新协议
	if err := m.db.UpdateProtocol(protocol); err != nil {
		return fmt.Errorf("failed to update protocol: %v", err)
	}

	return nil
}
