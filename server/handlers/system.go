package handlers

import (
	"net/http"
	"runtime"
	"time"
	"v/database"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

var startTime = time.Now()

// HandleGetSystemStatus returns the current system status
func HandleGetSystemStatus(c *gin.Context) {
	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		cpuPercent = []float64{0}
	}

	// Get memory usage
	memInfo, err := mem.VirtualMemory()
	var memPercent float64
	if err != nil {
		memPercent = 0
	} else {
		memPercent = memInfo.UsedPercent
	}

	// Get load average
	loadInfo, err := load.Avg()
	loadAvg := []float64{0, 0, 0}
	if err == nil {
		loadAvg = []float64{loadInfo.Load1, loadInfo.Load5, loadInfo.Load15}
	}

	// Get uptime
	uptime := time.Since(startTime).Seconds()

	// Get number of goroutines
	numGoroutines := runtime.NumGoroutine()

	// Get user statistics
	var userStats struct {
		TotalUsers    int64 `json:"total_users"`
		ActiveUsers   int64 `json:"active_users"`
		ExpiredUsers  int64 `json:"expired_users"`
		DisabledUsers int64 `json:"disabled_users"`
		TotalTraffic  int64 `json:"total_traffic"`  // In bytes
		TrafficToday  int64 `json:"traffic_today"`  // In bytes
		TrafficWeek   int64 `json:"traffic_week"`   // In bytes
		TrafficMonth  int64 `json:"traffic_month"`  // In bytes
		OnlineUsers   int64 `json:"online_users"`   // Users with traffic in last 5 minutes
		ProxyConfigs  int64 `json:"proxy_configs"`  // Total proxy configurations
		ActiveProxies int64 `json:"active_proxies"` // Proxies with traffic in last 24 hours
	}

	// Get total users and their status
	err = database.DB.QueryRow(`
		SELECT 
			COUNT(*) as total_users,
			SUM(CASE WHEN enabled = 1 AND (expire_at IS NULL OR expire_at > CURRENT_TIMESTAMP) THEN 1 ELSE 0 END) as active_users,
			SUM(CASE WHEN expire_at <= CURRENT_TIMESTAMP THEN 1 ELSE 0 END) as expired_users,
			SUM(CASE WHEN enabled = 0 THEN 1 ELSE 0 END) as disabled_users
		FROM users
	`).Scan(&userStats.TotalUsers, &userStats.ActiveUsers, &userStats.ExpiredUsers, &userStats.DisabledUsers)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user statistics"})
		return
	}

	// Get traffic statistics
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err = database.DB.QueryRow(`
		SELECT 
			COALESCE(SUM(upload + download), 0) as total_traffic,
			COALESCE(SUM(CASE WHEN timestamp >= ? THEN upload + download ELSE 0 END), 0) as traffic_today,
			COALESCE(SUM(CASE WHEN timestamp >= ? THEN upload + download ELSE 0 END), 0) as traffic_week,
			COALESCE(SUM(CASE WHEN timestamp >= ? THEN upload + download ELSE 0 END), 0) as traffic_month
		FROM traffic_logs
	`, startOfDay, startOfWeek, startOfMonth).Scan(
		&userStats.TotalTraffic,
		&userStats.TrafficToday,
		&userStats.TrafficWeek,
		&userStats.TrafficMonth,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get traffic statistics"})
		return
	}

	// Get online users (users with traffic in last 5 minutes)
	fiveMinutesAgo := now.Add(-5 * time.Minute)
	err = database.DB.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM traffic_logs 
		WHERE timestamp >= ?
	`, fiveMinutesAgo).Scan(&userStats.OnlineUsers)

	if err != nil {
		userStats.OnlineUsers = 0
	}

	// Get proxy statistics
	err = database.DB.QueryRow(`
		SELECT 
			COUNT(*) as total_proxies,
			SUM(CASE WHEN EXISTS (
				SELECT 1 FROM traffic_logs 
				WHERE proxy_id = proxy_configs.id 
				AND timestamp >= ?
			) THEN 1 ELSE 0 END) as active_proxies
		FROM proxy_configs
	`, now.Add(-24*time.Hour)).Scan(&userStats.ProxyConfigs, &userStats.ActiveProxies)

	if err != nil {
		userStats.ProxyConfigs = 0
		userStats.ActiveProxies = 0
	}

	status := database.SystemStatus{
		CPU:     cpuPercent[0],
		Memory:  memPercent,
		Uptime:  int64(uptime),
		LoadAvg: loadAvg,
	}

	c.JSON(http.StatusOK, gin.H{
		"system": status,
		"stats":  userStats,
		"runtime": gin.H{
			"goroutines": numGoroutines,
			"version":    runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		},
	})
}
