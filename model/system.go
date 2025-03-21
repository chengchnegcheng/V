package model

import (
	"time"
)

// System 系统设置
type System struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
	CPUs      int    `json:"cpus"`
	GoVersion string `json:"go_version"`
}

// SystemStats 系统统计信息
type SystemStats struct {
	CPU     float64      `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Disk    DiskStats    `json:"disk"`
	Load    []float64    `json:"load"`
	Network NetworkStats `json:"network"`
	Time    int64        `json:"time"`

	// 添加 model.go 中定义的字段，以便兼容已有代码
	CPUUsage             float64       `json:"cpu_usage"`
	CPUCount             int           `json:"cpu_count"`
	LoadAvg              []float64     `json:"load_avg"`
	MemoryTotal          uint64        `json:"memory_total"`
	MemoryUsed           uint64        `json:"memory_used"`
	MemoryFree           uint64        `json:"memory_free"`
	MemoryUsage          float64       `json:"memory_usage"`
	DiskTotal            uint64        `json:"disk_total"`
	DiskUsed             uint64        `json:"disk_used"`
	DiskFree             uint64        `json:"disk_free"`
	DiskUsage            float64       `json:"disk_usage"`
	NetworkBytesSent     uint64        `json:"network_bytes_sent"`
	NetworkBytesReceived uint64        `json:"network_bytes_received"`
	NetworkPacketsSent   uint64        `json:"network_packets_sent"`
	NetworkPacketsRecv   uint64        `json:"network_packets_recv"`
	Uptime               time.Duration `json:"uptime"`
	CreatedAt            time.Time     `json:"created_at"`
}

// MemoryStats 内存统计信息
type MemoryStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskStats 磁盘统计信息
type DiskStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkStats 网络统计信息
type NetworkStats struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// GetSystemValue 获取系统设置值
func GetSystemValue(db DB, key string) (string, error) {
	var value string
	var err error

	// 从数据库中获取系统设置
	if impl, ok := db.(*SQLiteDB); ok {
		value, err = impl.getSystemValue(key)
	}

	return value, err
}

// SetSystemValue 设置系统设置值
func SetSystemValue(db DB, key, value string) error {
	var err error

	// 更新系统设置到数据库
	if impl, ok := db.(*SQLiteDB); ok {
		err = impl.setSystemValue(key, value)
	}

	return err
}
