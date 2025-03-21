package monitor

import (
	"fmt"
	"time"

	"v/logger"
	"v/model"
	"v/notification"
	"v/settings"
)

// AlertType 告警类型
type AlertType string

const (
	// AlertCPUUsage CPU使用率告警
	AlertCPUUsage AlertType = "cpu_usage"
	// AlertMemoryUsage 内存使用率告警
	AlertMemoryUsage AlertType = "memory_usage"
	// AlertDiskUsage 磁盘使用率告警
	AlertDiskUsage AlertType = "disk_usage"
	// AlertTrafficUsage 流量使用率告警
	AlertTrafficUsage AlertType = "traffic_usage"
)

// Alert 告警信息
type Alert struct {
	Type      AlertType
	Value     float64
	Threshold float64
	Message   string
	Timestamp time.Time
}

// AlertManager 告警管理器
type AlertManager struct {
	log       *logger.Logger
	settings  *settings.Manager
	notifier  notification.Notifier
	lastAlert map[AlertType]time.Time
	db        model.DB
}

// NewAlertManager 创建告警管理器
func NewAlertManager(log *logger.Logger, settings *settings.Manager, notifier notification.Notifier, db model.DB) *AlertManager {
	return &AlertManager{
		log:       log,
		settings:  settings,
		notifier:  notifier,
		lastAlert: make(map[AlertType]time.Time),
		db:        db,
	}
}

// CheckSystemStats 检查系统状态是否触发告警
func (m *AlertManager) CheckSystemStats(stats *model.SystemStats) error {
	s := m.settings.Get()

	// 检查CPU使用率
	if s.Monitor.EnableCPUAlert && stats.CPUUsage >= s.Monitor.CPUThreshold {
		if err := m.sendAlert(AlertCPUUsage, stats.CPUUsage, s.Monitor.CPUThreshold,
			fmt.Sprintf("CPU使用率过高: %.2f%%", stats.CPUUsage)); err != nil {
			m.log.Error("Failed to send CPU usage alert", logger.Fields{
				"error": err.Error(),
			})
		}
	}

	// 检查内存使用率
	if s.Monitor.EnableMemoryAlert && stats.MemoryUsage >= s.Monitor.MemoryThreshold {
		if err := m.sendAlert(AlertMemoryUsage, stats.MemoryUsage, s.Monitor.MemoryThreshold,
			fmt.Sprintf("内存使用率过高: %.2f%%", stats.MemoryUsage)); err != nil {
			m.log.Error("Failed to send memory usage alert", logger.Fields{
				"error": err.Error(),
			})
		}
	}

	// 检查磁盘使用率
	if s.Monitor.EnableDiskAlert && stats.DiskUsage >= s.Monitor.DiskThreshold {
		if err := m.sendAlert(AlertDiskUsage, stats.DiskUsage, s.Monitor.DiskThreshold,
			fmt.Sprintf("磁盘使用率过高: %.2f%%", stats.DiskUsage)); err != nil {
			m.log.Error("Failed to send disk usage alert", logger.Fields{
				"error": err.Error(),
			})
		}
	}

	return nil
}

// sendAlert 发送告警通知
func (m *AlertManager) sendAlert(alertType AlertType, value, threshold float64, message string) error {
	s := m.settings.Get()

	// 检查告警间隔
	if lastTime, ok := m.lastAlert[alertType]; ok {
		if time.Since(lastTime) < time.Duration(s.Monitor.AlertInterval)*time.Minute {
			return nil
		}
	}

	// 更新最后告警时间
	m.lastAlert[alertType] = time.Now()

	// 创建告警记录
	alert := &model.AlertRecord{
		Type:      string(alertType),
		Value:     value,
		Threshold: threshold,
		Message:   message,
	}

	// 保存告警记录
	if err := m.db.CreateAlert(alert); err != nil {
		return fmt.Errorf("failed to save alert record: %v", err)
	}

	// 发送告警通知
	notification := &notification.Notification{
		To:      []string{s.Admin.Email},
		Subject: fmt.Sprintf("系统告警: %s", alertType),
		Body: fmt.Sprintf(`
			<p>尊敬的管理员：</p>
			<p>系统触发了%s告警。</p>
			<p>%s</p>
			<p>当前值：%.2f%%</p>
			<p>阈值：%.2f%%</p>
			<p>时间：%s</p>
			<p>请及时处理！</p>
		`, alertType, message, value, threshold, time.Now().Format("2006-01-02 15:04:05")),
		Type: "system_alert",
	}

	return m.notifier.Send(notification)
}

// SendTestAlert 发送测试告警
func (m *AlertManager) SendTestAlert() error {
	s := m.settings.Get()

	// 创建测试告警通知
	notification := &notification.Notification{
		To:      []string{s.Admin.Email},
		Subject: "系统告警测试",
		Body: `
			<p>这是一条测试告警通知。</p>
			<p>如果您收到此邮件，则表示系统告警功能运行正常。</p>
			<p>此邮件由系统自动发送，请勿回复。</p>
		`,
		Type: "test_alert",
	}

	// 发送测试告警
	if err := m.notifier.Send(notification); err != nil {
		return fmt.Errorf("发送测试告警失败: %v", err)
	}

	return nil
}
