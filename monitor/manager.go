package monitor

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"v/logger"
	"v/model"
	"v/notification"
	"v/settings"
)

// MonitorManager 系统监控管理器
type MonitorManager struct {
	log      *logger.Logger
	alertMgr *AlertManager
	stopCh   chan struct{}
}

// New 创建系统监控管理器
func New(log *logger.Logger, settings *settings.Manager, notifier notification.Notifier, db model.DB) *MonitorManager {
	return &MonitorManager{
		log:      log,
		alertMgr: NewAlertManager(log, settings, notifier, db),
		stopCh:   make(chan struct{}),
	}
}

// Start 启动系统监控
func (m *MonitorManager) Start() error {
	go m.monitorLoop()
	return nil
}

// Stop 停止系统监控
func (m *MonitorManager) Stop() {
	close(m.stopCh)
}

// monitorLoop 监控循环
func (m *MonitorManager) monitorLoop() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			stats, err := m.collectStats()
			if err != nil {
				m.log.Error("Failed to collect system stats", logger.Fields{
					"error": err.Error(),
				})
				continue
			}

			// 记录系统状态
			m.log.Info("System stats collected", logger.Fields{
				"cpu_usage":              stats.CPUUsage,
				"memory_usage":           stats.MemoryUsage,
				"disk_usage":             stats.DiskUsage,
				"network_bytes_sent":     stats.NetworkBytesSent,
				"network_bytes_received": stats.NetworkBytesReceived,
			})

			// 检查告警
			if err := m.alertMgr.CheckSystemStats(stats); err != nil {
				m.log.Error("Failed to check system alerts", logger.Fields{
					"error": err.Error(),
				})
			}
		}
	}
}

// collectStats 收集系统状态
func (m *MonitorManager) collectStats() (*model.SystemStats, error) {
	stats := &model.SystemStats{
		CreatedAt: time.Now(),
	}

	// 获取 CPU 信息
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %v", err)
	}
	if len(cpuPercent) > 0 {
		stats.CPUUsage = cpuPercent[0]
	}
	stats.CPUCount = runtime.NumCPU()

	// 获取负载信息
	times, err := cpu.Times(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU times: %v", err)
	}
	if len(times) > 0 {
		total := times[0].Total()
		idle := times[0].Idle
		load := (total - idle) / total
		stats.LoadAvg = []float64{load, load, load} // 使用当前负载作为1分钟、5分钟和15分钟的负载
	}

	// 获取内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}
	stats.MemoryTotal = memInfo.Total
	stats.MemoryUsed = memInfo.Used
	stats.MemoryFree = memInfo.Free
	stats.MemoryUsage = memInfo.UsedPercent

	// 获取磁盘信息
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %v", err)
	}
	stats.DiskTotal = diskInfo.Total
	stats.DiskUsed = diskInfo.Used
	stats.DiskFree = diskInfo.Free
	stats.DiskUsage = diskInfo.UsedPercent

	// 获取网络信息
	netStats, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get network stats: %v", err)
	}
	if len(netStats) > 0 {
		stats.NetworkBytesSent = netStats[0].BytesSent
		stats.NetworkBytesReceived = netStats[0].BytesRecv
		stats.NetworkPacketsSent = netStats[0].PacketsSent
		stats.NetworkPacketsRecv = netStats[0].PacketsRecv
	}

	// 获取系统运行时间
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %v", err)
	}
	stats.Uptime = time.Duration(hostInfo.Uptime) * time.Second

	return stats, nil
}

// GetStats 获取当前系统状态
func (m *MonitorManager) GetStats() (*model.SystemStats, error) {
	return m.collectStats()
}

// SendTestAlert 发送测试告警
func (m *MonitorManager) SendTestAlert() error {
	return m.alertMgr.SendTestAlert()
}
