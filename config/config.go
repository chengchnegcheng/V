package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 程序配置
type Config struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	BackupDir string `json:"backup_dir"` // 备份目录
	Database  struct {
		DSN  string `json:"dsn"`  // 数据库连接字符串
		Type string `json:"type"` // 数据库类型 (sqlite, mysql, etc.)
		Path string `json:"path"` // 数据库文件路径 (对于SQLite)
	} `json:"database"`
	SSL struct {
		Enabled  bool   `json:"enabled"`   // SSL是否启用
		CertFile string `json:"cert_file"` // SSL证书文件
		KeyFile  string `json:"key_file"`  // SSL密钥文件
	} `json:"ssl"`
	System struct {
		DefaultValidityDays int64 `json:"default_validity_days"` // 默认有效期（天）
		DefaultTrafficLimit int64 `json:"default_traffic_limit"` // 默认流量限制（GB）
	} `json:"system"`
}

// Global configuration instance
var GlobalConfig Config

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	// 默认配置
	cfg := &Config{}
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8080
	cfg.System.DefaultValidityDays = 30
	cfg.System.DefaultTrafficLimit = 100 // 100GB default traffic limit

	// 如果文件不存在，使用默认配置并保存
	if _, err := os.Stat(path); os.IsNotExist(err) {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal default config: %w", err)
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			return nil, fmt.Errorf("write default config: %w", err)
		}

		GlobalConfig = *cfg
		return cfg, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Set the global configuration
	GlobalConfig = *cfg

	return cfg, nil
}
