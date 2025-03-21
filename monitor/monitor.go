package monitor

import (
	"fmt"
	"runtime"
	"time"

	"v/logger"
	"v/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemStats represents system statistics
type SystemStats struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Network struct {
		Upload   int64 `json:"upload"`
		Download int64 `json:"download"`
	} `json:"network"`
	Disk struct {
		Used  int64 `json:"used"`
		Total int64 `json:"total"`
	} `json:"disk"`
	Uptime time.Duration `json:"uptime"`
}

// Monitor handles system monitoring
type Monitor struct {
	log     *logger.Logger
	stats   *SystemStats
	stopCh  chan struct{}
	startAt time.Time
}

// NewMonitor creates a new monitor instance
func NewMonitor(log *logger.Logger) *Monitor {
	return &Monitor{
		log:     log,
		stats:   &SystemStats{},
		stopCh:  make(chan struct{}),
		startAt: time.Now(),
	}
}

// Start begins monitoring
func (m *Monitor) Start() error {
	m.log.Info("Starting system monitor", logger.Fields{
		"start_time": m.startAt,
	})
	go m.monitorLoop()
	return nil
}

// Stop stops monitoring
func (m *Monitor) Stop() error {
	m.log.Info("Stopping system monitor", logger.Fields{
		"uptime": time.Since(m.startAt),
	})
	close(m.stopCh)
	return nil
}

// GetStats returns current system statistics
func (m *Monitor) GetStats() *SystemStats {
	return m.stats
}

// monitorLoop runs the monitoring loop
func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if err := m.collectStats(); err != nil {
				m.log.Error("Failed to collect system stats", logger.Fields{
					"error": err.Error(),
				})
			}
		}
	}
}

// collectStats collects system statistics
func (m *Monitor) collectStats() error {
	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Errorf("failed to get CPU stats: %v", err)
	}
	if len(cpuPercent) > 0 {
		m.stats.CPU = cpuPercent[0]
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory stats: %v", err)
	}
	m.stats.Memory = memInfo.UsedPercent

	// Network usage
	netStats, err := net.IOCounters(false)
	if err != nil {
		return fmt.Errorf("failed to get network stats: %v", err)
	}
	if len(netStats) > 0 {
		m.stats.Network.Upload = int64(netStats[0].BytesSent)
		m.stats.Network.Download = int64(netStats[0].BytesRecv)
	}

	// Disk usage
	diskStats, err := disk.Usage("/")
	if err != nil {
		return fmt.Errorf("failed to get disk stats: %v", err)
	}
	m.stats.Disk.Used = int64(diskStats.Used)
	m.stats.Disk.Total = int64(diskStats.Total)

	// Uptime
	m.stats.Uptime = time.Since(m.startAt)

	return nil
}

// GetGoroutines returns the number of goroutines
func (m *Monitor) GetGoroutines() int {
	return runtime.NumGoroutine()
}

// GetMemoryUsage returns current memory usage in bytes
func (m *Monitor) GetMemoryUsage() int64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return int64(mem.Alloc)
}

// Manager 系统监控管理器
type Manager struct {
	lastNetworkBytesSent     uint64
	lastNetworkBytesReceived uint64
	lastNetworkPacketsSent   uint64
	lastNetworkPacketsRecv   uint64
	lastCollectTime          time.Time
}

// NewManager creates a new system monitor manager
func NewManager() *Manager {
	return &Manager{
		lastCollectTime: time.Now(),
	}
}

// Collect 收集系统状态
func (m *Manager) Collect() (*model.SystemStats, error) {
	stats := &model.SystemStats{}

	// 收集 CPU 信息
	if err := m.collectCPUInfo(stats); err != nil {
		return nil, err
	}

	// 收集内存信息
	if err := m.collectMemoryInfo(stats); err != nil {
		return nil, err
	}

	// 收集磁盘信息
	if err := m.collectDiskInfo(stats); err != nil {
		return nil, err
	}

	// 收集网络信息
	if err := m.collectNetworkInfo(stats); err != nil {
		return nil, err
	}

	// 收集系统运行时间
	if err := m.collectUptime(stats); err != nil {
		return nil, err
	}

	m.lastCollectTime = time.Now()
	return stats, nil
}

// collectCPUInfo 收集 CPU 信息
func (m *Manager) collectCPUInfo(stats *model.SystemStats) error {
	// 获取 CPU 使用率
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	if len(percent) > 0 {
		stats.CPUUsage = percent[0]
	}

	// 获取 CPU 核心数
	stats.CPUCount = runtime.NumCPU()

	// 获取系统负载
	stats.LoadAvg = []float64{0, 0, 0} // 暂时使用空值，后续可以通过其他方式获取

	return nil
}

// collectMemoryInfo 收集内存信息
func (m *Manager) collectMemoryInfo(stats *model.SystemStats) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	stats.MemoryTotal = v.Total
	stats.MemoryUsed = v.Used
	stats.MemoryFree = v.Free
	stats.MemoryUsage = v.UsedPercent

	return nil
}

// collectDiskInfo 收集磁盘信息
func (m *Manager) collectDiskInfo(stats *model.SystemStats) error {
	parts, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	var total, used, free uint64
	for _, part := range parts {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}
		total += usage.Total
		used += usage.Used
		free += usage.Free
	}

	stats.DiskTotal = total
	stats.DiskUsed = used
	stats.DiskFree = free
	if total > 0 {
		stats.DiskUsage = float64(used) / float64(total) * 100
	}

	return nil
}

// collectNetworkInfo 收集网络信息
func (m *Manager) collectNetworkInfo(stats *model.SystemStats) error {
	netStats, err := net.IOCounters(false)
	if err != nil {
		return err
	}

	if len(netStats) > 0 {
		stats.NetworkBytesSent = netStats[0].BytesSent
		stats.NetworkBytesReceived = netStats[0].BytesRecv
		stats.NetworkPacketsSent = netStats[0].PacketsSent
		stats.NetworkPacketsRecv = netStats[0].PacketsRecv
	}

	return nil
}

// collectUptime 收集系统运行时间
func (m *Manager) collectUptime(stats *model.SystemStats) error {
	info, err := host.Info()
	if err != nil {
		return err
	}

	stats.Uptime = time.Duration(info.Uptime) * time.Second
	return nil
}
