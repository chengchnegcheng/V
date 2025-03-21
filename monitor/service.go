package monitor

import (
	"sync"
	"time"

	"v/logger"
	"v/model"
)

// Service 系统监控服务
type Service struct {
	manager *Manager
	history *StatsHistory
	logger  *logger.Logger
	stop    chan struct{}
	wg      sync.WaitGroup
}

// NewService 创建系统监控服务
func NewService(logger *logger.Logger) *Service {
	return &Service{
		manager: NewManager(),
		history: NewStatsHistory(3600), // 保存1小时的数据
		logger:  logger,
		stop:    make(chan struct{}),
	}
}

// Start 启动系统监控服务
func (s *Service) Start() {
	s.wg.Add(1)
	go s.run()
}

// Stop 停止系统监控服务
func (s *Service) Stop() {
	close(s.stop)
	s.wg.Wait()
}

// run 运行系统监控服务
func (s *Service) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			stats, err := s.manager.Collect()
			if err != nil {
				s.logger.Error("Failed to collect system stats", logger.Fields{
					"error": err.Error(),
				})
				continue
			}

			s.history.Add(stats)
		}
	}
}

// GetLatestStats 获取最新的系统状态数据
func (s *Service) GetLatestStats() *model.SystemStats {
	return s.history.GetLatest()
}

// GetStatsHistory 获取指定时间范围内的系统状态数据
func (s *Service) GetStatsHistory(start, end time.Time) []*model.SystemStats {
	return s.history.Get(start, end)
}

// ClearHistory 清空历史数据
func (s *Service) ClearHistory() {
	s.history.Clear()
}
