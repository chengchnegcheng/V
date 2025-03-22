package monitor

// SystemMonitor 是监控接口
type SystemMonitor interface {
	GetSystemStats() interface{}
}

// DummyMonitor 是一个空实现的监控器
type DummyMonitor struct{}

// GetSystemStats 获取系统统计信息
func (m *DummyMonitor) GetSystemStats() interface{} {
	return map[string]interface{}{
		"cpu":    0,
		"memory": 0,
		"disk":   0,
	}
}
