package traffic

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

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

	// 尝试获取ID为1-20的协议
	for i := int64(1); i <= 20; i++ {
		protocol, err := m.db.GetProtocol(i)
		if err != nil {
			// 如果ID不存在，跳过即可
			continue
		}
		protocols = append(protocols, protocol)
	}

	if len(protocols) == 0 {
		m.logger.Info("No protocols found for traffic check")
		return nil
	}

	// 用于跟踪已检查过流量限制的用户ID
	checkedUsers := make(map[int64]bool)

	// 处理每个协议
	for _, protocol := range protocols {
		if !protocol.Enable {
			continue
		}

		// 获取协议统计信息
		stats, err := m.getProtocolStats(protocol.ID)
		if err != nil {
			m.logger.Error("Failed to get protocol stats", "protocol_id", protocol.ID, "error", err)
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
		if err == model.ErrNotFound {
			// 如果不存在，创建新的统计信息
			protocol, err := m.db.GetProtocol(protocolID)
			if err != nil {
				return nil, err
			}

			stats = &model.ProtocolStats{
				ProtocolID: protocolID,
				UserID:     protocol.UserID,
				Upload:     0,
				Download:   0,
				LastActive: time.Now(),
			}

			if err := m.db.CreateProtocolStats(stats); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
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
