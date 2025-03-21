package model

import (
	"crypto"
	"runtime"
	"time"

	"v/common"

	"github.com/go-acme/lego/v4/registration"
)

// Base 基础模型
type Base struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// User 用户
type User struct {
	Base
	Username      string                 `json:"username" db:"username"`
	Password      string                 `json:"-" db:"password"`
	Salt          string                 `json:"-" db:"salt"`
	Email         string                 `json:"email" db:"email"`
	Key           crypto.PrivateKey      `json:"-"`
	Registration  *registration.Resource `json:"-"`
	Role          string                 `json:"role" db:"role"`
	Status        string                 `json:"status" db:"status"`
	LastLoginAt   *time.Time             `json:"last_login_at" db:"last_login_at"`
	LoginAttempts int                    `json:"-" db:"login_attempts"`
	LockedUntil   *time.Time             `json:"locked_until" db:"locked_until"`
	IsAdmin       bool                   `json:"is_admin" db:"is_admin"`
	TrafficLimit  int64                  `json:"traffic_limit" db:"traffic_limit"`
	TrafficUsed   int64                  `json:"traffic_used" db:"traffic_used"`
	ExpireAt      *time.Time             `json:"expire_at" db:"expire_at"`
}

// GetEmail 获取用户邮箱
func (u *User) GetEmail() string {
	return u.Email
}

// GetPrivateKey 获取私钥
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

// GetRegistration 获取注册信息
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// CertificateConfig 证书配置
type CertificateConfig struct {
	Email string `json:"email"`
}

// Protocol 协议
type Protocol struct {
	Base
	UserID       int64     `json:"user_id" db:"user_id"`
	Type         string    `json:"type" db:"type"`
	Name         string    `json:"name" db:"name"`
	Settings     []byte    `json:"settings" db:"settings"`
	Status       string    `json:"status" db:"status"`
	Port         int       `json:"port" db:"port"`
	TrafficLimit int64     `json:"traffic_limit" db:"traffic_limit"`
	TrafficUsed  int64     `json:"traffic_used" db:"traffic_used"`
	ExpireAt     time.Time `json:"expire_at" db:"expire_at"`
	Enable       bool      `json:"enable" db:"enable"`
	Tags         []string  `json:"tags" db:"tags"`
	LastActive   time.Time `json:"last_active" db:"last_active"`
}

// ProtocolStats 协议流量统计
type ProtocolStats struct {
	Base
	ProtocolID int64     `json:"protocol_id" db:"protocol_id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Upload     int64     `json:"upload" db:"upload"`
	Download   int64     `json:"download" db:"download"`
	LastActive time.Time `json:"last_active" db:"last_active"`
}

// Certificate SSL证书信息
type Certificate struct {
	Base
	Domain        string    `json:"domain" db:"domain"`
	CertFile      string    `json:"cert_file" db:"cert_file"`
	KeyFile       string    `json:"key_file" db:"key_file"`
	Status        string    `json:"status" db:"status"`
	LastCheckedAt time.Time `json:"last_checked_at" db:"last_checked_at"`
	LastRenewedAt time.Time `json:"last_renewed_at" db:"last_renewed_at"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
}

// Traffic 流量统计
type Traffic struct {
	Base
	UserID  int64 `json:"user_id" db:"user_id"`
	ProxyID int64 `json:"proxy_id" db:"proxy_id"`
	Up      int64 `json:"up" db:"up"`     // 上传流量（字节）
	Down    int64 `json:"down" db:"down"` // 下载流量（字节）
}

// TrafficStats 流量统计
type TrafficStats struct {
	Base
	UserID      int64     `json:"user_id" db:"user_id"`
	Upload      int64     `json:"upload" db:"upload"`
	Download    int64     `json:"download" db:"download"`
	Total       int64     `json:"total" db:"total"`
	Limit       int64     `json:"limit" db:"limit"`
	ExpireAt    time.Time `json:"expire_at" db:"expire_at"`
	LastResetAt time.Time `json:"last_reset_at" db:"last_reset_at"`
	UpSpeed     float64   `json:"up_speed"`
	DownSpeed   float64   `json:"down_speed"`
}

// DailyStats 每日流量统计
type DailyStats struct {
	Base
	UserID   int64     `json:"user_id" db:"user_id"`
	Date     time.Time `json:"date" db:"date"`
	Upload   int64     `json:"upload" db:"upload"`
	Download int64     `json:"download" db:"download"`
	Total    int64     `json:"total" db:"total"`
}

// AlertRecord 告警记录
type AlertRecord struct {
	Base
	Type      string  `json:"type" db:"type"`           // 告警类型：cpu, memory, disk, traffic, etc.
	Value     float64 `json:"value" db:"value"`         // 当前值
	Threshold float64 `json:"threshold" db:"threshold"` // 阈值
	Message   string  `json:"message" db:"message"`     // 告警消息
}

// TrafficHistory 流量历史记录
type TrafficHistory struct {
	Base
	UserID   int64  `json:"user_id" db:"user_id"`
	Protocol string `json:"protocol" db:"protocol"`
	Upload   int64  `json:"upload" db:"upload"`
	Download int64  `json:"download" db:"download"`
	Date     string `json:"date" db:"date"`
}

// DB 数据库接口
type DB interface {
	// 用户相关
	CreateUser(user *User) error
	GetUser(id int64) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int64) error
	ListUsers(page, pageSize int) ([]*User, error)
	GetTotalUsers() (int64, error)
	SearchUsers(keyword string) ([]*User, error)

	// 代理相关
	CreateProxy(proxy *common.Proxy) error
	GetProxy(id int64) (*common.Proxy, error)
	GetProxiesByUserID(userID int64) ([]*common.Proxy, error)
	UpdateProxy(proxy *common.Proxy) error
	DeleteProxy(id int64) error
	GetProxiesByPort(port int) ([]*common.Proxy, error)
	ListProxies(page, pageSize int) ([]*common.Proxy, error)
	GetTotalProxies() (int64, error)
	SearchProxies(keyword string) ([]*common.Proxy, error)

	// 流量统计相关
	CreateTraffic(traffic *common.TrafficStats) error
	GetTraffic(id int64) (*common.TrafficStats, error)
	UpdateTraffic(traffic *common.TrafficStats) error
	DeleteTraffic(id int64) error
	ListTrafficByUserID(userID int64) ([]*common.TrafficStats, error)
	ListTrafficByProxyID(proxyID int64) ([]*common.TrafficStats, error)
	GetTrafficStats(userID uint) (*TrafficStats, error)
	CreateTrafficRecord(traffic *Traffic) error
	CleanupTraffic(before time.Time) error

	// 协议相关
	CreateProtocol(protocol *Protocol) error
	GetProtocol(id int64) (*Protocol, error)
	GetProtocolsByUserID(userID int64) ([]*Protocol, error)
	UpdateProtocol(protocol *Protocol) error
	DeleteProtocol(id int64) error
	GetProtocolsByPort(port int) ([]*Protocol, error)
	ListProtocols(page, pageSize int) ([]*Protocol, error)
	SearchProtocols(keyword string) ([]*Protocol, error)

	// 协议统计相关
	CreateProtocolStats(stats *ProtocolStats) error
	GetProtocolStats(id int64) (*ProtocolStats, error)
	UpdateProtocolStats(stats *ProtocolStats) error
	ListProtocolStatsByUserID(userID int64) ([]*ProtocolStats, error)

	// 证书相关
	CreateCertificate(cert *Certificate) error
	GetCertificate(domain string) (*Certificate, error)
	UpdateCertificate(cert *Certificate) error
	DeleteCertificate(domain string) error
	ListCertificates() ([]*Certificate, error)

	// 告警相关
	CreateAlert(alert *AlertRecord) error
	GetAlert(id int64) (*AlertRecord, error)
	ListAlerts(page, pageSize int) ([]*AlertRecord, error)
	DeleteAlert(id int64) error

	// 事务相关
	Begin() error
	Commit() error
	Rollback() error

	// 日志相关
	CreateLog(log *Log) error
	GetLog(id int64) (*Log, error)
	UpdateLog(log *Log) error
	DeleteLog(id int64) error
	ListLogs(query *LogQuery) ([]*Log, error)
	GetTotalLogs(query *LogQuery) (int64, error)
	DeleteLogsBefore(time.Time) error
	ExportLogs(query *LogQuery) (string, error)

	// 备份相关
	CreateBackup(backup *Backup) error
	GetBackup(id int64) (*Backup, error)
	UpdateBackup(backup *Backup) error
	DeleteBackup(id int64) error
	ListBackups() ([]*Backup, error)
	GetTotalBackups() (int64, error)
	DeleteBackupsBefore(time.Time) error

	// 流量统计相关方法
	CreateDailyStats(stats *DailyStats) error
	DeleteDailyStatsBefore(date time.Time) error
	ListDailyStatsByUserID(userID int64) ([]*DailyStats, error)
	ListProtocolStatsByProtocolID(protocolID int64) ([]*ProtocolStats, error)

	// 告警记录
	CreateAlertRecord(record *AlertRecord) error
	ListAlertRecords(out *[]*AlertRecord) error

	// 流量历史
	CreateTrafficHistory(history *TrafficHistory) error
	ListTrafficHistoryByDateRange(userID uint, startDate, endDate string, histories *[]*TrafficHistory) error

	// 系统设置
	GetSettings(key string) (string, error)
	SetSettings(key, value string) error

	// 关闭数据库
	Close() error
	AutoMigrate() error
}

// Event represents an audit event model
type Event struct {
	Base
	UserID    int64  `json:"user_id" db:"user_id"`
	Username  string `json:"username" db:"username"`
	Action    string `json:"action" db:"action"`
	Resource  string `json:"resource" db:"resource"`
	Details   string `json:"details" db:"details"`
	IP        string `json:"ip" db:"ip"`
	UserAgent string `json:"user_agent" db:"user_agent"`
}

// Backup represents a backup model
type Backup struct {
	Base
	Path      string    `json:"path" db:"path"`
	Size      int64     `json:"size" db:"size"`
	Status    string    `json:"status" db:"status"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// Log 日志
type Log struct {
	Base
	Level     string `json:"level" db:"level"`
	Module    string `json:"module" db:"module"`
	Message   string `json:"message" db:"message"`
	Details   string `json:"details" db:"details"`
	IP        string `json:"ip" db:"ip"`
	UserAgent string `json:"user_agent" db:"user_agent"`
	UserID    int64  `json:"user_id" db:"user_id"`
	Username  string `json:"username" db:"username"`
}

// LogQuery 日志查询参数
type LogQuery struct {
	Level     string    `json:"level"`
	Module    string    `json:"module"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	UserID    int64     `json:"user_id"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() *SystemInfo {
	return &SystemInfo{
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		CPUs:      runtime.NumCPU(),
		GoVersion: runtime.Version(),
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
	}, nil
}

// 以下是辅助函数，实际实现中应根据不同操作系统提供具体实现
func getCPUUsage() (float64, error) {
	// 示例实现，实际应根据操作系统获取
	return 0.0, nil
}

func getMemoryUsage() (MemoryStats, error) {
	// 示例实现，实际应根据操作系统获取
	return MemoryStats{}, nil
}

func getDiskUsage() (DiskStats, error) {
	// 示例实现，实际应根据操作系统获取
	return DiskStats{}, nil
}

func getLoadAverage() ([]float64, error) {
	// 示例实现，实际应根据操作系统获取
	return []float64{0.0, 0.0, 0.0}, nil
}

func getNetworkIO() (NetworkStats, error) {
	// 示例实现，实际应根据操作系统获取
	return NetworkStats{}, nil
}
