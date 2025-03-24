package monitor

import (
	"fmt"
	"time"

	"v/logger"
	"v/model"
	"v/notification"
	"v/settings"
)

// MonitorManager is the interface that handlers/monitor.go expects
type MonitorManager struct {
	log          *logger.Logger
	settingsMgr  *settings.Manager
	notifier     notification.Notifier
	db           model.DB
	stopCh       chan struct{}
	monitor      *Monitor
	isRunning    bool
	collectTimer *time.Timer
}

// New creates a new monitor manager that implements the expected interface
func New(log *logger.Logger, settingsMgr *settings.Manager, notifier notification.Notifier, db model.DB) *MonitorManager {
	return &MonitorManager{
		log:         log,
		settingsMgr: settingsMgr,
		notifier:    notifier,
		db:          db,
		stopCh:      make(chan struct{}),
		monitor:     NewMonitor(log),
		isRunning:   false,
	}
}

// Start begins the monitoring process
func (m *MonitorManager) Start() error {
	if m.isRunning {
		return nil
	}

	m.isRunning = true
	m.log.Info("Starting system monitor", nil)

	// Start the underlying monitor
	err := m.monitor.Start()
	if err != nil {
		return err
	}

	// Start collection interval
	m.collectTimer = time.NewTimer(1 * time.Minute)
	go m.collectLoop()

	return nil
}

// Stop ends the monitoring process
func (m *MonitorManager) Stop() error {
	if !m.isRunning {
		return nil
	}

	m.isRunning = false
	if m.collectTimer != nil {
		m.collectTimer.Stop()
	}

	close(m.stopCh)
	m.log.Info("Stopped system monitor", nil)

	return m.monitor.Stop()
}

// GetStats returns the current system statistics
func (m *MonitorManager) GetStats() (*SystemStats, error) {
	stats := m.monitor.GetStats()
	return stats, nil
}

// collectLoop periodically collects system statistics and stores them
func (m *MonitorManager) collectLoop() {
	for {
		select {
		case <-m.stopCh:
			return
		case <-m.collectTimer.C:
			stats, err := m.GetStats()
			if err != nil {
				m.log.Error("Failed to collect system stats", logger.Fields{
					"error": err.Error(),
				})
			} else {
				// Store stats in database if needed
				if m.db != nil {
					// Implementation depends on your database schema
					// m.db.SaveSystemStats(stats)
				}

				// Check for alerts
				m.checkAlerts(stats)
			}
			m.collectTimer.Reset(1 * time.Minute)
		}
	}
}

// checkAlerts checks system stats against thresholds and sends notifications if needed
func (m *MonitorManager) checkAlerts(stats *SystemStats) {
	// Example: Alert on high CPU usage
	if stats.CPU > 90 {
		m.log.Warn("High CPU usage detected", logger.Fields{
			"cpu_usage": stats.CPU,
		})

		if m.notifier != nil {
			m.sendNotification("System Alert", fmt.Sprintf("High CPU usage detected: %.2f%%", stats.CPU))
		}
	}

	// Example: Alert on low disk space
	if stats.Disk.Total > 0 {
		diskUsagePercent := (float64(stats.Disk.Used) / float64(stats.Disk.Total)) * 100
		if diskUsagePercent > 90 {
			m.log.Warn("Low disk space detected", logger.Fields{
				"disk_usage": diskUsagePercent,
			})

			if m.notifier != nil {
				m.sendNotification("System Alert", fmt.Sprintf("Low disk space detected: %.2f%%", diskUsagePercent))
			}
		}
	}
}

// sendNotification sends a notification using the appropriate notifier method
func (m *MonitorManager) sendNotification(title, message string) {
	// This would call the appropriate method based on notification.Notifier interface
	// For now, just log the notification
	m.log.Info("Notification", logger.Fields{
		"title":   title,
		"message": message,
	})
}

// SendTestAlert sends a test alert notification
func (m *MonitorManager) SendTestAlert() error {
	m.sendNotification("Test Alert", "This is a test alert from the monitoring system")
	return nil
}
