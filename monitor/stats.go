package monitor

import (
	"sync"
	"time"

	"v/model"
)

// StatsHistory 系统监控历史数据
type StatsHistory struct {
	mu    sync.RWMutex
	stats []*model.SystemStats
	max   int
}

// NewStatsHistory 创建系统监控历史数据存储
func NewStatsHistory(max int) *StatsHistory {
	return &StatsHistory{
		stats: make([]*model.SystemStats, 0, max),
		max:   max,
	}
}

// Add 添加系统状态数据
func (h *StatsHistory) Add(stats *model.SystemStats) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stats = append(h.stats, stats)
	if len(h.stats) > h.max {
		h.stats = h.stats[1:]
	}
}

// Get 获取指定时间范围内的系统状态数据
func (h *StatsHistory) Get(start, end time.Time) []*model.SystemStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*model.SystemStats
	for _, stat := range h.stats {
		if stat.CreatedAt.After(start) && stat.CreatedAt.Before(end) {
			result = append(result, stat)
		}
	}
	return result
}

// GetLatest 获取最新的系统状态数据
func (h *StatsHistory) GetLatest() *model.SystemStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.stats) == 0 {
		return nil
	}
	return h.stats[len(h.stats)-1]
}

// Clear 清空历史数据
func (h *StatsHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stats = make([]*model.SystemStats, 0, h.max)
}
