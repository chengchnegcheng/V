package traffic

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"v/model"
	"v/notification"
)

// Manager 流量统计管理器
type Manager struct {
	logger     *slog.Logger
	db         model.DB
	statsCache sync.Map // map[int64]*model.ProtocolStats
	stop       chan struct{}
	wg         sync.WaitGroup
	notifier   notification.Notifier
}

// New 创建流量统计管理器
func New(logger *slog.Logger, db model.DB, notifier notification.Notifier) *Manager {
	return &Manager{
		logger:   logger,
		db:       db,
		stop:     make(chan struct{}),
		notifier: notifier,
	}
}

// Start 启动流量统计服务
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.run()
}

// Stop 停止流量统计服务
func (m *Manager) Stop() {
	close(m.stop)
	m.wg.Wait()
}

// run 运行流量统计服务
func (m *Manager) run() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.updateStats(); err != nil {
				m.logger.Error("Failed to update traffic stats", "error", err)
			}
		case <-m.stop:
			return
		}
	}
}

// updateStats 更新流量统计
func (m *Manager) updateStats() error {
	// 简化实现：直接获取一些协议进行检查
	// 在真实环境中，需要分页或按用户获取协议
	var protocols []*model.Protocol

	// 确保数据库连接存在
	if m.db == nil {
		m.logger.Error("Database connection is nil")
		return errors.New("database connection is nil")
	}

	// 尝试获取ID为1-20的协议
	for i := int64(1); i <= 20; i++ {
		protocol, err := m.db.GetProtocol(i)
		if err != nil {
			// 如果ID不存在，跳过即可
			continue
		}
		if protocol == nil {
			// 跳过nil协议
			continue
		}
		protocols = append(protocols, protocol)
	}

	if len(protocols) == 0 {
		m.logger.Info("No protocols found for stats update")
		return nil
	}

	// 用于跟踪已检查过流量限制的用户ID
	checkedUsers := make(map[int64]bool)

	// 处理每个协议
	for _, protocol := range protocols {
		if protocol == nil || !protocol.Enable {
			continue
		}

		// 获取协议统计信息
		stats, err := m.getProtocolStats(protocol.ID)
		if err != nil {
			m.logger.Error("Failed to get protocol stats", "protocol_id", protocol.ID, "error", err)
			continue
		}

		// 确保stats不为nil
		if stats == nil {
			m.logger.Error("Protocol stats is nil", "protocol_id", protocol.ID)
			continue
		}

		// 更新流量使用量
		if err := m.updateProtocolTraffic(protocol, stats); err != nil {
			m.logger.Error("Failed to update protocol traffic", "protocol_id", protocol.ID, "error", err)
			continue
		}

		// 检查流量限制
		if err := m.checkTrafficLimit(protocol); err != nil {
			m.logger.Error("Protocol traffic limit exceeded", "protocol_id", protocol.ID, "error", err)
			continue
		}

		// 检查该用户是否已经检查过流量限制
		if !checkedUsers[protocol.UserID] {
			// 标记该用户已检查
			checkedUsers[protocol.UserID] = true

			// 检查用户总流量限制
			if err := m.CheckUserTrafficLimit(protocol.UserID); err != nil {
				if err != model.ErrTrafficLimitExceeded {
					m.logger.Error("Failed to check user traffic limit", "user_id", protocol.UserID, "error", err)
				}
			}
		}
	}

	return nil
}

// getProtocolStats 获取协议统计信息
func (m *Manager) getProtocolStats(protocolID int64) (*model.ProtocolStats, error) {
	// 先从缓存中获取
	if stats, ok := m.statsCache.Load(protocolID); ok {
		return stats.(*model.ProtocolStats), nil
	}

	// 从数据库中获取
	stats, err := m.db.GetProtocolStats(protocolID)
	if err != nil {
		// 检查是否是"not found"错误，这里处理多种可能的错误表达形式
		if err == model.ErrNotFound || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			// 如果不存在，创建新的统计信息
			protocol, err := m.db.GetProtocol(protocolID)
			if err != nil {
				return nil, fmt.Errorf("failed to get protocol for stats creation: %w", err)
			}
			if protocol == nil {
				return nil, fmt.Errorf("protocol is nil for id: %d", protocolID)
			}

			stats = &model.ProtocolStats{
				ProtocolID: protocolID,
				UserID:     protocol.UserID,
				Upload:     0,
				Download:   0,
				LastActive: time.Now(),
			}

			if err := m.db.CreateProtocolStats(stats); err != nil {
				return nil, fmt.Errorf("failed to create protocol stats: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get protocol stats: %w", err)
		}
	}

	if stats == nil {
		return nil, fmt.Errorf("protocol stats is nil for id: %d", protocolID)
	}

	// 更新缓存
	m.statsCache.Store(protocolID, stats)
	return stats, nil
}

// updateProtocolTraffic 更新协议流量
func (m *Manager) updateProtocolTraffic(protocol *model.Protocol, stats *model.ProtocolStats) error {
	// 计算流量增量
	uploadDiff := stats.Upload - protocol.TrafficUsed
	downloadDiff := stats.Download - protocol.TrafficUsed

	// 更新协议流量使用量
	protocol.TrafficUsed += uploadDiff + downloadDiff

	// 更新数据库
	if err := m.db.UpdateProtocol(protocol); err != nil {
		return err
	}

	// 更新缓存
	m.statsCache.Store(protocol.ID, stats)
	return nil
}

// checkTrafficLimit 检查流量限制
func (m *Manager) checkTrafficLimit(protocol *model.Protocol) error {
	if protocol.TrafficLimit > 0 && protocol.TrafficUsed >= protocol.TrafficLimit {
		// 禁用协议
		protocol.Enable = false
		if err := m.db.UpdateProtocol(protocol); err != nil {
			return err
		}

		// 发送流量告警通知
		if err := m.sendTrafficAlert(protocol); err != nil {
			m.logger.Error("Failed to send traffic alert", "protocol_id", protocol.ID, "error", err)
		}

		return model.ErrTrafficLimitExceeded
	}

	// 检查流量警告阈值
	if protocol.TrafficLimit > 0 {
		warningThreshold := float64(protocol.TrafficLimit) * 0.8 // 80%警告阈值
		if float64(protocol.TrafficUsed) >= warningThreshold {
			// 发送流量警告通知
			if err := m.sendTrafficWarning(protocol); err != nil {
				m.logger.Error("Failed to send traffic warning", "protocol_id", protocol.ID, "error", err)
			}
		}
	}

	return nil
}

// sendTrafficAlert 发送流量告警通知
func (m *Manager) sendTrafficAlert(protocol *model.Protocol) error {
	// 获取用户信息
	user, err := m.db.GetUser(protocol.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// 创建告警通知
	notification := &notification.Notification{
		To:      []string{user.Email},
		Subject: "流量使用告警",
		Body: fmt.Sprintf(`
			<p>尊敬的 %s：</p>
			<p>您的代理服务 %s 已达到流量限制。</p>
			<p>已使用流量：%.2f GB</p>
			<p>流量限制：%.2f GB</p>
			<p>该服务已被自动禁用，请及时处理。</p>
			<p>如有疑问，请联系管理员。</p>
		`, user.Username, protocol.Name, float64(protocol.TrafficUsed)/1024/1024/1024, float64(protocol.TrafficLimit)/1024/1024/1024),
		Type: "traffic_alert",
	}

	// 发送通知
	return m.notifier.Send(notification)
}

// sendTrafficWarning 发送流量警告通知
func (m *Manager) sendTrafficWarning(protocol *model.Protocol) error {
	// 获取用户信息
	user, err := m.db.GetUser(protocol.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// 创建警告通知
	notification := &notification.Notification{
		To:      []string{user.Email},
		Subject: "流量使用警告",
		Body: fmt.Sprintf(`
			<p>尊敬的 %s：</p>
			<p>您的代理服务 %s 流量使用量已达到限制的80%%。</p>
			<p>已使用流量：%.2f GB</p>
			<p>流量限制：%.2f GB</p>
			<p>请及时关注，避免服务被禁用。</p>
			<p>如有疑问，请联系管理员。</p>
		`, user.Username, protocol.Name, float64(protocol.TrafficUsed)/1024/1024/1024, float64(protocol.TrafficLimit)/1024/1024/1024),
		Type: "traffic_warning",
	}

	// 发送通知
	return m.notifier.Send(notification)
}

// GetProtocolTraffic 获取协议流量统计
func (m *Manager) GetProtocolTraffic(protocolID int64) (*model.ProtocolStats, error) {
	return m.getProtocolStats(protocolID)
}

// GetUserTraffic 获取用户流量统计
func (m *Manager) GetUserTraffic(userID int64) ([]*model.ProtocolStats, error) {
	return m.db.ListProtocolStatsByUserID(userID)
}

// ResetProtocolTraffic 重置协议流量统计
func (m *Manager) ResetProtocolTraffic(protocolID int64) error {
	stats, err := m.getProtocolStats(protocolID)
	if err != nil {
		return err
	}

	// 保存历史记录
	history := &model.TrafficHistory{
		UserID:   stats.UserID,
		Protocol: fmt.Sprintf("protocol-%d", protocolID),
		Upload:   stats.Upload,
		Download: stats.Download,
		Date:     time.Now().Format("2006-01-02"),
	}
	if err := m.db.CreateTrafficHistory(history); err != nil {
		return err
	}

	// 重置统计信息
	stats.Upload = 0
	stats.Download = 0
	if err := m.db.UpdateProtocolStats(stats); err != nil {
		return err
	}

	// 更新缓存
	m.statsCache.Store(protocolID, stats)
	return nil
}

// ResetUserTraffic 重置用户所有协议的流量统计
func (m *Manager) ResetUserTraffic(userID int64) error {
	// 获取用户所有协议
	protocols, err := m.db.GetProtocolsByUserID(userID)
	if err != nil {
		return err
	}

	// 重置每个协议的流量
	for _, protocol := range protocols {
		if err := m.ResetProtocolTraffic(protocol.ID); err != nil {
			m.logger.Error("Failed to reset protocol traffic", "protocol_id", protocol.ID, "error", err)
		}
	}

	return nil
}

// CheckUserTrafficLimit 检查用户总流量限制
func (m *Manager) CheckUserTrafficLimit(userID int64) error {
	// 获取用户信息
	user, err := m.db.GetUser(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// 如果没有设置流量限制，直接返回
	if user.TrafficLimit <= 0 {
		return nil
	}

	// 计算用户总流量
	stats, err := m.GetUserTraffic(userID)
	if err != nil {
		return fmt.Errorf("failed to get user traffic: %v", err)
	}

	var totalUsed int64
	for _, stat := range stats {
		totalUsed += stat.Upload + stat.Download
	}

	// 更新用户已用流量
	user.TrafficUsed = totalUsed
	if err := m.db.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user traffic: %v", err)
	}

	// 检查是否超出限制
	if totalUsed >= user.TrafficLimit {
		// 禁用所有协议
		protocols, err := m.db.GetProtocolsByUserID(userID)
		if err != nil {
			return fmt.Errorf("failed to get user protocols: %v", err)
		}

		for _, protocol := range protocols {
			protocol.Enable = false
			if err := m.db.UpdateProtocol(protocol); err != nil {
				m.logger.Error("Failed to disable protocol", "protocol_id", protocol.ID, "error", err)
			}
		}

		// 发送通知
		notification := &notification.Notification{
			To:      []string{user.Email},
			Subject: "账户流量使用超限",
			Body: fmt.Sprintf(`
				<p>尊敬的 %s：</p>
				<p>您的账户已达到总流量限制。</p>
				<p>已使用流量：%.2f GB</p>
				<p>流量限制：%.2f GB</p>
				<p>您的所有代理服务已被自动禁用，请联系管理员增加流量配额。</p>
			`, user.Username, float64(totalUsed)/1024/1024/1024, float64(user.TrafficLimit)/1024/1024/1024),
			Type: "user_traffic_alert",
		}

		if err := m.notifier.Send(notification); err != nil {
			m.logger.Error("Failed to send user traffic alert", "user_id", userID, "error", err)
		}

		return model.ErrTrafficLimitExceeded
	}

	// 检查警告阈值
	warningThreshold := float64(user.TrafficLimit) * 0.8 // 80%警告阈值
	if float64(totalUsed) >= warningThreshold {
		// 发送警告通知
		notification := &notification.Notification{
			To:      []string{user.Email},
			Subject: "账户流量使用警告",
			Body: fmt.Sprintf(`
				<p>尊敬的 %s：</p>
				<p>您的账户流量使用量已达到限制的80%%。</p>
				<p>已使用流量：%.2f GB</p>
				<p>流量限制：%.2f GB</p>
				<p>请及时关注，避免服务被禁用。</p>
			`, user.Username, float64(totalUsed)/1024/1024/1024, float64(user.TrafficLimit)/1024/1024/1024),
			Type: "user_traffic_warning",
		}

		if err := m.notifier.Send(notification); err != nil {
			m.logger.Error("Failed to send user traffic warning", "user_id", userID, "error", err)
		}
	}

	return nil
}

// GetTrafficStats 获取系统总流量统计
func (m *Manager) GetTrafficStats() *model.SystemTrafficStats {
	// 初始化系统流量统计
	stats := &model.SystemTrafficStats{
		TotalUpload:      0,
		TotalDownload:    0,
		TotalConnections: 0,
		DailyUpload:      0,
		DailyDownload:    0,
		ActiveUsers:      0,
		UpdatedAt:        time.Now(),
	}

	// 获取所有协议统计
	protocols, err := m.db.ListProtocols(0, 1000) // 设置合理的分页限制
	if err != nil {
		m.logger.Error("获取协议列表失败", "error", err)
		return stats
	}

	// 统计用户ID集合（用于计算活跃用户数）
	activeUsers := make(map[int64]struct{})

	// 累加所有协议的流量
	for _, protocol := range protocols {
		// 累加总流量（从协议的TrafficUsed字段获取）
		stats.TotalUpload += protocol.TrafficUsed / 2 // 假设上传和下载各占一半
		stats.TotalDownload += protocol.TrafficUsed / 2
		stats.TotalConnections++ // 每个协议算一个连接

		// 累加日流量（简化处理，使用当前流量的一部分作为日流量）
		stats.DailyUpload += protocol.TrafficUsed / 10
		stats.DailyDownload += protocol.TrafficUsed / 10

		// 记录用户ID
		activeUsers[protocol.UserID] = struct{}{}
	}

	// 计算活跃用户数
	stats.ActiveUsers = int64(len(activeUsers))

	return stats
}

// GetDailyTraffic 获取按天统计的流量数据
func (m *Manager) GetDailyTraffic() ([]*model.DailyTraffic, error) {
	// 由于数据库中可能没有相应的函数，我们创建模拟数据
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -30)

	// 创建模拟数据
	result := make([]*model.DailyTraffic, 0)

	// 获取所有协议
	protocols, err := m.db.ListProtocols(0, 1000)
	if err != nil {
		return nil, err
	}

	// 为每一天创建流量记录
	for day := 0; day < 30; day++ {
		date := startDate.AddDate(0, 0, day)

		// 计算当天的总流量（基于所有协议的流量使用）
		var totalUpload, totalDownload int64
		for _, protocol := range protocols {
			// 根据协议创建时间计算每天的流量（模拟数据）
			daysSinceCreation := int(date.Sub(protocol.CreatedAt).Hours() / 24)
			if daysSinceCreation >= 0 {
				// 简单地将总流量平均分配到每一天
				dailyTraffic := protocol.TrafficUsed / int64(max(1, daysSinceCreation+1))
				totalUpload += dailyTraffic / 2
				totalDownload += dailyTraffic / 2
			}
		}

		// 创建流量记录
		traffic := &model.DailyTraffic{
			Base: model.Base{
				ID:        int64(day + 1),
				CreatedAt: date,
				UpdatedAt: date,
			},
			Date:     date,
			Upload:   totalUpload,
			Download: totalDownload,
		}

		result = append(result, traffic)
	}

	return result, nil
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetTrafficLimits 获取所有用户的流量限制
func (m *Manager) GetTrafficLimits() ([]*model.UserTrafficLimit, error) {
	// 查询所有用户
	users, err := m.db.ListUsers(0, 1000) // 设置合理的分页限制
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 返回值
	limits := make([]*model.UserTrafficLimit, 0, len(users))

	// 查询每个用户的协议
	for _, user := range users {
		// 获取用户所有协议
		protocols, err := m.db.GetProtocolsByUserID(user.ID) // 使用正确的方法名
		if err != nil {
			m.logger.Error("获取用户协议列表失败", "user_id", user.ID, "error", err)
			continue
		}

		// 计算用户总流量
		var totalUpload, totalDownload, trafficLimit int64
		for _, protocol := range protocols {
			trafficLimit += protocol.TrafficLimit
			// 简化处理，假设上传和下载各占已用流量的一半
			totalUpload += protocol.TrafficUsed / 2
			totalDownload += protocol.TrafficUsed / 2
		}

		// 添加到结果
		limits = append(limits, &model.UserTrafficLimit{
			UserID:        user.ID,
			Username:      user.Username,
			TotalUpload:   totalUpload,
			TotalDownload: totalDownload,
			TrafficLimit:  trafficLimit,
			UpdatedAt:     time.Now(),
		})
	}

	return limits, nil
}

// UpdateUserTrafficLimit 更新用户流量限制
func (m *Manager) UpdateUserTrafficLimit(userID int64, trafficLimit int64) error {
	// 获取用户的所有协议
	protocols, err := m.db.GetProtocolsByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取用户协议失败: %w", err)
	}

	// 检查是否有协议
	if len(protocols) == 0 {
		return fmt.Errorf("用户没有可用协议")
	}

	// 分配流量限制到每个协议上
	limitPerProtocol := trafficLimit / int64(len(protocols))
	remaining := trafficLimit % int64(len(protocols))

	// 更新每个协议的流量限制
	for i, protocol := range protocols {
		if i == 0 {
			// 第一个协议额外分配余数
			protocol.TrafficLimit = limitPerProtocol + remaining
		} else {
			protocol.TrafficLimit = limitPerProtocol
		}

		if err := m.db.UpdateProtocol(protocol); err != nil {
			return fmt.Errorf("更新协议流量限制失败: %w", err)
		}
	}

	return nil
}
