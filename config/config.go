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
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	// 默认配置
	cfg := &Config{}
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8080

	// 如果文件不存在，使用默认配置并保存
	if _, err := os.Stat(path); os.IsNotExist(err) {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal default config: %w", err)
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			return nil, fmt.Errorf("write default config: %w", err)
		}

		return cfg, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// 解析配置
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
