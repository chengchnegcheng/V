package model

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// GetSystemInfo 获取系统信息，返回与前端期望格式匹配的数据
func GetSystemInfo() map[string]interface{} {
	hostname, _ := os.Hostname()

	// 计算运行时间
	uptime := time.Since(startTime)
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	uptimeStr := fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)

	// 获取内核版本 - 针对不同操作系统进行处理
	kernelVersion := runtime.Version() // 默认使用Go版本
	osType := runtime.GOOS

	// 格式化操作系统名称，使其更友好
	var osName string
	switch osType {
	case "windows":
		osName = "Windows"
	case "darwin":
		osName = "macOS"
	case "linux":
		osName = "Linux"
	default:
		osName = osType
	}

	// 获取负载信息（在Windows上使用模拟数据）
	var loadAvg []float64
	if runtime.GOOS == "windows" {
		loadAvg = []float64{0, 0, 0}
	} else if runtime.GOOS == "darwin" {
		// macOS 也可以获取负载，但这里简化处理
		loadAvg = []float64{0, 0, 0}
	} else {
		// 在Linux/Unix系统上尝试获取实际负载
		// 这里简化处理，使用模拟数据
		loadAvg = []float64{0, 0, 0}
	}

	// 确保loadAvg不为nil，以避免前端调用join时出错
	if loadAvg == nil {
		loadAvg = []float64{0, 0, 0}
	}

	// 获取IP地址 - 简单情况下用本地IP，生产中应当获取实际外部IP
	ipAddress := "0.0.0.0" // 默认值

	// 格式化为前端期望的格式
	return map[string]interface{}{
		"os":        osName,
		"kernel":    kernelVersion,
		"hostname":  hostname,
		"uptime":    uptimeStr,
		"load":      loadAvg,
		"ipAddress": ipAddress,

		// 保留原始字段，以便向后兼容
		"platform":   osType,
		"arch":       runtime.GOARCH,
		"cpus":       runtime.NumCPU(),
		"go_version": runtime.Version(),
	}
}

// GetSystemStats 获取系统统计信息
func GetSystemStats() (*SystemStats, error) {
	// 获取CPU使用率
	cpuUsage, err := getCPUUsage()
	if err != nil {
		return nil, err
	}

	// 获取内存使用情况
	memUsage, err := getMemoryUsage()
	if err != nil {
		return nil, err
	}

	// 获取磁盘使用情况
	diskUsage, err := getDiskUsage()
	if err != nil {
		return nil, err
	}

	// 获取系统负载
	loadAvg, err := getLoadAverage()
	if err != nil {
		return nil, err
	}

	// 获取网络流量
	netIO, err := getNetworkIO()
	if err != nil {
		return nil, err
	}

	return &SystemStats{
		CPU:     cpuUsage,
		Memory:  memUsage,
		Disk:    diskUsage,
		Load:    loadAvg,
		Network: netIO,
		Time:    time.Now().Unix(),

		// 添加兼容字段
		CPUUsage:    cpuUsage,
		CPUCount:    runtime.NumCPU(),
		LoadAvg:     loadAvg,
		MemoryTotal: memUsage.Total,
		MemoryUsed:  memUsage.Used,
		MemoryFree:  memUsage.Free,
		MemoryUsage: memUsage.UsedPercent,
		DiskTotal:   diskUsage.Total,
		DiskUsed:    diskUsage.Used,
		DiskFree:    diskUsage.Free,
		DiskUsage:   diskUsage.UsedPercent,
		Uptime:      time.Since(startTime),
		CreatedAt:   time.Now(),
	}, nil
}

// 以下是辅助函数，简化实现，实际应使用系统API获取真实数据

// getCPUUsage 获取CPU使用率
func getCPUUsage() (float64, error) {
	// 示例实现，生产环境应使用系统API获取真实数据
	return 30.0, nil // 返回30%的CPU使用率
}

// getMemoryUsage 获取内存使用情况
func getMemoryUsage() (MemoryStats, error) {
	// 示例实现，生产环境应使用系统API获取真实数据
	totalMem := uint64(8 * 1024 * 1024 * 1024) // 8GB
	usedMem := uint64(3 * 1024 * 1024 * 1024)  // 3GB
	return MemoryStats{
		Total:       totalMem,
		Used:        usedMem,
		Free:        totalMem - usedMem,
		UsedPercent: float64(usedMem) / float64(totalMem) * 100,
	}, nil
}

// getDiskUsage 获取磁盘使用情况
func getDiskUsage() (DiskStats, error) {
	// 示例实现，生产环境应使用系统API获取真实数据
	totalDisk := uint64(500 * 1024 * 1024 * 1024) // 500GB
	usedDisk := uint64(200 * 1024 * 1024 * 1024)  // 200GB
	return DiskStats{
		Total:       totalDisk,
		Used:        usedDisk,
		Free:        totalDisk - usedDisk,
		UsedPercent: float64(usedDisk) / float64(totalDisk) * 100,
	}, nil
}

// getLoadAverage 获取系统负载
func getLoadAverage() ([]float64, error) {
	// 示例实现，生产环境应使用系统API获取真实数据
	return []float64{0.8, 1.0, 1.2}, nil
}

// getNetworkIO 获取网络IO
func getNetworkIO() (NetworkStats, error) {
	// 示例实现，生产环境应使用系统API获取真实数据
	return NetworkStats{
		BytesSent:   1024 * 1024 * 10, // 10MB
		BytesRecv:   1024 * 1024 * 50, // 50MB
		PacketsSent: 5000,
		PacketsRecv: 8000,
	}, nil
}
