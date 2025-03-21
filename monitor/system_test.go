package monitor

import (
	"testing"
	"v/model"
)

type mockSystemDB struct {
	model.DB
}

func (db *mockSystemDB) DeleteProxy(id int64) error {
	return nil
}

func TestSystemMonitor_GetSystemStats(t *testing.T) {
	monitor := NewSystemMonitor(&mockSystemDB{})

	stats, err := monitor.GetSystemStats()
	if err != nil {
		t.Fatalf("GetSystemStats failed: %v", err)
	}

	// Check CPU stats
	if stats.CPUUsage < 0 || stats.CPUUsage > 100 {
		t.Errorf("Invalid CPU usage: %f", stats.CPUUsage)
	}
	if stats.CPUCount <= 0 {
		t.Errorf("Invalid CPU count: %d", stats.CPUCount)
	}
	if len(stats.LoadAvg) != 3 {
		t.Errorf("Invalid load average length: %d", len(stats.LoadAvg))
	}

	// Check memory stats
	if stats.MemoryTotal <= 0 {
		t.Errorf("Invalid memory total: %d", stats.MemoryTotal)
	}
	if stats.MemoryUsed <= 0 {
		t.Errorf("Invalid memory used: %d", stats.MemoryUsed)
	}
	if stats.MemoryFree <= 0 {
		t.Errorf("Invalid memory free: %d", stats.MemoryFree)
	}
	if stats.MemoryUsage < 0 || stats.MemoryUsage > 100 {
		t.Errorf("Invalid memory usage: %f", stats.MemoryUsage)
	}

	// Check disk stats
	if stats.DiskTotal <= 0 {
		t.Errorf("Invalid disk total: %d", stats.DiskTotal)
	}
	if stats.DiskUsed <= 0 {
		t.Errorf("Invalid disk used: %d", stats.DiskUsed)
	}
	if stats.DiskFree <= 0 {
		t.Errorf("Invalid disk free: %d", stats.DiskFree)
	}
	if stats.DiskUsage < 0 || stats.DiskUsage > 100 {
		t.Errorf("Invalid disk usage: %f", stats.DiskUsage)
	}

	// Check network stats
	if stats.NetworkBytesSent < 0 {
		t.Errorf("Invalid network bytes sent: %d", stats.NetworkBytesSent)
	}
	if stats.NetworkBytesReceived < 0 {
		t.Errorf("Invalid network bytes received: %d", stats.NetworkBytesReceived)
	}
	if stats.NetworkPacketsSent < 0 {
		t.Errorf("Invalid network packets sent: %d", stats.NetworkPacketsSent)
	}
	if stats.NetworkPacketsRecv < 0 {
		t.Errorf("Invalid network packets received: %d", stats.NetworkPacketsRecv)
	}

	// Check uptime
	if stats.Uptime < 0 {
		t.Errorf("Invalid uptime: %v", stats.Uptime)
	}
}

func TestSystemMonitor_GetCPUStats(t *testing.T) {
	monitor := NewSystemMonitor(&mockSystemDB{})

	stats := &model.SystemStats{}
	err := monitor.getCPUStats(stats)
	if err != nil {
		t.Fatalf("getCPUStats failed: %v", err)
	}

	if stats.CPUUsage < 0 || stats.CPUUsage > 100 {
		t.Errorf("Invalid CPU usage: %f", stats.CPUUsage)
	}
	if stats.CPUCount <= 0 {
		t.Errorf("Invalid CPU count: %d", stats.CPUCount)
	}
	if len(stats.LoadAvg) != 3 {
		t.Errorf("Invalid load average length: %d", len(stats.LoadAvg))
	}
}

func TestSystemMonitor_GetMemoryStats(t *testing.T) {
	monitor := NewSystemMonitor(&mockSystemDB{})

	stats := &model.SystemStats{}
	err := monitor.getMemoryStats(stats)
	if err != nil {
		t.Fatalf("getMemoryStats failed: %v", err)
	}

	if stats.MemoryTotal <= 0 {
		t.Errorf("Invalid memory total: %d", stats.MemoryTotal)
	}
	if stats.MemoryUsed <= 0 {
		t.Errorf("Invalid memory used: %d", stats.MemoryUsed)
	}
	if stats.MemoryFree <= 0 {
		t.Errorf("Invalid memory free: %d", stats.MemoryFree)
	}
	if stats.MemoryUsage < 0 || stats.MemoryUsage > 100 {
		t.Errorf("Invalid memory usage: %f", stats.MemoryUsage)
	}
}

func TestSystemMonitor_GetDiskStats(t *testing.T) {
	monitor := NewSystemMonitor(&mockSystemDB{})

	stats := &model.SystemStats{}
	err := monitor.getDiskStats(stats)
	if err != nil {
		t.Fatalf("getDiskStats failed: %v", err)
	}

	if stats.DiskTotal <= 0 {
		t.Errorf("Invalid disk total: %d", stats.DiskTotal)
	}
	if stats.DiskUsed <= 0 {
		t.Errorf("Invalid disk used: %d", stats.DiskUsed)
	}
	if stats.DiskFree <= 0 {
		t.Errorf("Invalid disk free: %d", stats.DiskFree)
	}
	if stats.DiskUsage < 0 || stats.DiskUsage > 100 {
		t.Errorf("Invalid disk usage: %f", stats.DiskUsage)
	}
}

func TestSystemMonitor_GetNetworkStats(t *testing.T) {
	monitor := NewSystemMonitor(&mockSystemDB{})

	stats := &model.SystemStats{}
	err := monitor.getNetworkStats(stats)
	if err != nil {
		t.Fatalf("getNetworkStats failed: %v", err)
	}

	if stats.NetworkBytesSent < 0 {
		t.Errorf("Invalid network bytes sent: %d", stats.NetworkBytesSent)
	}
	if stats.NetworkBytesReceived < 0 {
		t.Errorf("Invalid network bytes received: %d", stats.NetworkBytesReceived)
	}
	if stats.NetworkPacketsSent < 0 {
		t.Errorf("Invalid network packets sent: %d", stats.NetworkPacketsSent)
	}
	if stats.NetworkPacketsRecv < 0 {
		t.Errorf("Invalid network packets received: %d", stats.NetworkPacketsRecv)
	}
}
