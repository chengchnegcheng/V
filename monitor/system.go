package monitor

import (
	"runtime"
	"time"
	"v/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemStatsMonitor handles system monitoring
type SystemStatsMonitor struct {
	db model.DB
}

// NewSystemStatsMonitor creates a new system monitor
func NewSystemStatsMonitor(db model.DB) *SystemStatsMonitor {
	return &SystemStatsMonitor{
		db: db,
	}
}

// GetSystemStats returns current system statistics
func (m *SystemStatsMonitor) GetSystemStats() (*model.SystemStats, error) {
	stats := &model.SystemStats{}

	// Get CPU stats
	if err := m.getCPUStats(stats); err != nil {
		return nil, err
	}

	// Get memory stats
	if err := m.getMemoryStats(stats); err != nil {
		return nil, err
	}

	// Get disk stats
	if err := m.getDiskStats(stats); err != nil {
		return nil, err
	}

	// Get network stats
	if err := m.getNetworkStats(stats); err != nil {
		return nil, err
	}

	// Get uptime
	stats.Uptime = time.Since(time.Now())

	return stats, nil
}

// getCPUStats gets CPU statistics
func (m *SystemStatsMonitor) getCPUStats(stats *model.SystemStats) error {
	// Get CPU usage
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	if len(percent) > 0 {
		stats.CPUUsage = percent[0]
	}

	// Get CPU count
	stats.CPUCount = runtime.NumCPU()

	// Get load average
	times, err := cpu.Times(false)
	if err != nil {
		return err
	}
	if len(times) > 0 {
		total := times[0].Total()
		idle := times[0].Idle
		load := (total - idle) / total
		stats.LoadAvg = []float64{load, load, load} // 使用当前负载作为1分钟、5分钟和15分钟的负载
	}

	return nil
}

// getMemoryStats gets memory statistics
func (m *SystemStatsMonitor) getMemoryStats(stats *model.SystemStats) error {
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

// getDiskStats gets disk statistics
func (m *SystemStatsMonitor) getDiskStats(stats *model.SystemStats) error {
	d, err := disk.Usage("/")
	if err != nil {
		return err
	}

	stats.DiskTotal = d.Total
	stats.DiskUsed = d.Used
	stats.DiskFree = d.Free
	stats.DiskUsage = d.UsedPercent

	return nil
}

// getNetworkStats gets network statistics
func (m *SystemStatsMonitor) getNetworkStats(stats *model.SystemStats) error {
	n, err := net.IOCounters(false)
	if err != nil {
		return err
	}

	if len(n) > 0 {
		stats.NetworkBytesSent = n[0].BytesSent
		stats.NetworkBytesReceived = n[0].BytesRecv
		stats.NetworkPacketsSent = n[0].PacketsSent
		stats.NetworkPacketsRecv = n[0].PacketsRecv
	}

	return nil
}
