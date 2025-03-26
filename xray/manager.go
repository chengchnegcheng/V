package xray

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"v/logger"
	"v/settings"
)

// 支持的版本列表
var SupportedVersions = []string{
	// v1系列
	"v1.8.23", "v1.8.24", "v1.8.21", "v1.8.20", "v1.8.19", "v1.8.18", "v1.8.17", "v1.8.16", "v1.8.15", "v1.8.13",
	// v24系列
	"v24.12.31", "v24.12.28", "v24.12.18", "v24.12.15", "v24.11.30", "v24.11.21", "v24.11.11", "v24.11.5", "v24.10.31", "v24.10.16",
	"v24.9.30", "v24.9.19", "v24.9.16", "v24.9.7",
	// v25系列
	"v25.3.6", "v25.3.3", "v25.2.21", "v25.2.18", "v25.1.30", "v25.1.1",
}

// Manager 是xray版本管理器
type Manager struct {
	log            *logger.Logger
	settings       *settings.Manager
	binPath        string      // xray可执行文件目录
	process        *os.Process // 当前运行的xray进程
	mutex          sync.Mutex
	running        bool
	currentVersion string
	// 事件通知相关
	eventsMutex      sync.RWMutex
	eventSubscribers map[chan XrayEvent]bool
}

// XrayEvent 表示Xray事件
type XrayEvent struct {
	Type    string      `json:"type"`
	Version string      `json:"version"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Percent int         `json:"percent"`
	Details interface{} `json:"details,omitempty"`
}

// New 创建一个新的xray版本管理器
func New(log *logger.Logger, settingsManager *settings.Manager) *Manager {
	binPath := filepath.Join("xray", "bin")

	// 确保二进制目录存在
	os.MkdirAll(binPath, 0755)

	return &Manager{
		log:              log,
		settings:         settingsManager,
		binPath:          binPath,
		running:          false,
		eventSubscribers: make(map[chan XrayEvent]bool),
	}
}

// Initialize 初始化xray版本管理器
func (m *Manager) Initialize() error {
	// 获取当前保存的版本设置
	settings := m.settings.Get()
	currentVersion := settings.Xray.Version

	// 如果没有设置版本或版本不存在，使用默认版本
	if currentVersion == "" || !m.VersionExists(currentVersion) {
		// 使用第一个支持的版本作为默认版本
		if len(SupportedVersions) > 0 {
			currentVersion = SupportedVersions[0]
			m.log.Info("Using default version", logger.Fields{
				"version": currentVersion,
			})
		} else {
			// 如果没有支持的版本列表，设置一个固定的默认版本
			currentVersion = "v1.8.24"
			m.log.Warn("No supported versions found, using hardcoded default", logger.Fields{
				"version": currentVersion,
			})
		}

		// 更新设置
		settings.Xray.Version = currentVersion
		if err := m.settings.Save(); err != nil {
			m.log.Error("Failed to save settings", logger.Fields{
				"error": err,
			})
		}
	}

	// 设置当前版本
	m.currentVersion = currentVersion
	m.log.Info("Current Xray version", logger.Fields{
		"version": m.currentVersion,
	})

	// 检查xray二进制文件是否存在，如果不存在则下载
	if !m.VersionExists(currentVersion) {
		m.log.Info("Downloading xray", logger.Fields{
			"version": currentVersion,
		})

		if err := m.DownloadVersion(currentVersion); err != nil {
			return fmt.Errorf("failed to download xray: %v", err)
		}
	}

	// 检查配置文件是否存在，如果不存在则创建默认配置
	configPath := m.GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		m.log.Info("Config file not found, creating default configuration", logger.Fields{
			"path": configPath,
		})

		// 生成默认配置
		defaultConfig, err := m.GenerateConfig()
		if err != nil {
			return fmt.Errorf("failed to generate default config: %v", err)
		}

		// 更新配置文件
		if err := m.UpdateConfig(defaultConfig); err != nil {
			return fmt.Errorf("failed to update config: %v", err)
		}
	}

	m.log.Info("Initialized xray manager", logger.Fields{
		"version": currentVersion,
	})

	return nil
}

// VersionExists 检查指定版本的xray是否已下载
func (m *Manager) VersionExists(version string) bool {
	execPath := m.GetExecutablePath(version)
	_, err := os.Stat(execPath)
	return err == nil
}

// GetExecutablePath 获取指定版本xray的可执行文件路径
func (m *Manager) GetExecutablePath(version string) string {
	filename := "xray"
	if runtime.GOOS == "windows" {
		filename = "xray.exe"
	}

	return filepath.Join(m.binPath, version, filename)
}

// GetConfigPath 获取xray配置文件路径
func (m *Manager) GetConfigPath() string {
	return filepath.Join("xray", "config.json")
}

// DownloadVersion 下载指定版本的xray
func (m *Manager) DownloadVersion(version string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 发布下载开始事件
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "start",
		Message: fmt.Sprintf("开始下载 Xray 版本 %s", version),
		Percent: 0,
	})

	// 创建版本目录
	versionDir := filepath.Join(m.binPath, version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		// 发布错误事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "error",
			Message: fmt.Sprintf("创建目录失败: %v", err),
			Percent: 0,
		})
		return fmt.Errorf("failed to create version directory: %v", err)
	}

	// 发布进度事件 - 10%
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "progress",
		Message: "准备下载环境",
		Percent: 10,
	})

	// 使用新的自动下载器
	downloader := NewAutoDownloader(version)

	// 发布进度事件 - 20%
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "progress",
		Message: "启动自动下载器",
		Percent: 20,
	})

	// 尝试下载和安装
	execPath := m.GetExecutablePath(version)

	// 检查文件是否已经存在
	if _, err := os.Stat(execPath); err == nil {
		m.log.Info("Xray executable already exists", logger.Fields{
			"path":    execPath,
			"version": version,
		})

		// 发布完成事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "completed",
			Message: "可执行文件已存在，跳过下载",
			Percent: 100,
		})

		return nil
	}

	// 接下来尝试执行下载
	m.log.Info("Downloading Xray using auto downloader", logger.Fields{
		"version": version,
	})

	// 发布进度事件 - 30%
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "progress",
		Message: "开始下载Xray...",
		Percent: 30,
	})

	// 启动下载
	err := downloader.DownloadAndInstall()
	if err != nil {
		m.log.Error("Failed to download and install", logger.Fields{
			"error":   err,
			"version": version,
		})

		// 尝试使用Node.js工具作为后备
		if downloader.hasToolkit() {
			m.log.Info("Trying Node.js toolkit as fallback", logger.Fields{
				"version": version,
			})

			// 发布进度事件 - 备用方法
			m.PublishEvent(XrayEvent{
				Type:    "download",
				Version: version,
				Status:  "progress",
				Message: "自动下载失败，尝试使用Node.js工具作为备用方案",
				Percent: 50,
			})

			// 执行Node.js工具
			if err := downloader.runToolkit(); err != nil {
				m.log.Error("Node.js toolkit failed", logger.Fields{
					"error": err,
				})

				// 发布错误事件
				m.PublishEvent(XrayEvent{
					Type:    "download",
					Version: version,
					Status:  "error",
					Message: fmt.Sprintf("所有下载方法均失败: %v", err),
					Percent: 50,
				})

				return fmt.Errorf("all download methods failed: %v", err)
			}
		} else {
			// 发布错误事件
			m.PublishEvent(XrayEvent{
				Type:    "download",
				Version: version,
				Status:  "error",
				Message: fmt.Sprintf("下载失败: %v", err),
				Percent: 50,
			})

			return fmt.Errorf("download failed: %v", err)
		}
	}

	// 发布进度事件 - 80%
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "progress",
		Message: "下载完成，验证文件",
		Percent: 80,
	})

	// 验证可执行文件是否存在且有效
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		m.log.Error("Xray executable not found after download", logger.Fields{
			"path":    execPath,
			"version": version,
		})
		// 发布错误事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "error",
			Message: fmt.Sprintf("下载完成但可执行文件不存在: %s", execPath),
			Percent: 90,
		})
		return fmt.Errorf("xray executable not found after download: %s", execPath)
	}

	m.log.Info("Downloaded xray successfully", logger.Fields{
		"version": version,
		"path":    execPath,
	})

	// 发布完成事件
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "completed",
		Message: "下载安装成功",
		Percent: 100,
	})

	return nil
}

// SwitchVersion 切换到指定版本的xray
func (m *Manager) SwitchVersion(version string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查版本是否支持
	supported := false
	for _, v := range SupportedVersions {
		if v == version {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("unsupported version: %s", version)
	}

	// 发布切换开始事件
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "start",
		Message: fmt.Sprintf("开始切换到版本 %s", version),
		Percent: 0,
	})

	// 如果版本不存在，先下载
	if !m.VersionExists(version) {
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "progress",
			Message: fmt.Sprintf("版本 %s 不存在，开始下载", version),
			Percent: 10,
		})

		if err := m.DownloadVersion(version); err != nil {
			m.PublishEvent(XrayEvent{
				Type:    "switch",
				Version: version,
				Status:  "error",
				Message: fmt.Sprintf("下载失败: %v", err),
				Percent: 0,
			})
			return fmt.Errorf("failed to download version %s: %v", version, err)
		}
	}

	// 如果当前有实例在运行，先停止
	if m.running {
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "progress",
			Message: "停止当前运行的实例",
			Percent: 50,
		})

		if err := m.Stop(); err != nil {
			m.PublishEvent(XrayEvent{
				Type:    "switch",
				Version: version,
				Status:  "error",
				Message: fmt.Sprintf("停止当前实例失败: %v", err),
				Percent: 50,
			})
			return fmt.Errorf("failed to stop current instance: %v", err)
		}
	}

	// 更新当前版本
	m.currentVersion = version

	// 更新设置
	settings := m.settings.Get()
	settings.Xray.Version = version
	if err := m.settings.Save(); err != nil {
		m.log.Error("Failed to save settings", logger.Fields{
			"error": err,
		})
	}

	// 发布完成事件
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "completed",
		Message: fmt.Sprintf("已切换到版本 %s", version),
		Percent: 100,
	})

	return nil
}

// GetCurrentVersion 获取当前使用的xray版本
func (m *Manager) GetCurrentVersion() string {
	if m.currentVersion == "" {
		return "未知"
	}
	return m.currentVersion
}

// GetSupportedVersions 获取所有支持的xray版本
func (m *Manager) GetSupportedVersions() []string {
	return SupportedVersions
}

// Start 启动xray
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		m.log.Warn("Attempt to start Xray when it's already running")
		return fmt.Errorf("xray is already running")
	}

	// 获取系统信息
	osName, osArch := getPlatformInfo()
	m.log.Info("Starting Xray on platform", logger.Fields{
		"os":      osName,
		"arch":    osArch,
		"version": m.currentVersion,
	})

	// 获取可执行文件路径
	execPath := m.GetExecutablePath(m.currentVersion)

	// 在 Windows 上转换为绝对路径
	if runtime.GOOS == "windows" {
		absPath, err := filepath.Abs(execPath)
		if err != nil {
			m.log.Warn("Failed to get absolute path, using original", logger.Fields{
				"path":  execPath,
				"error": err,
			})
		} else {
			execPath = absPath
			m.log.Info("Using absolute path for Windows", logger.Fields{
				"path": execPath,
			})
		}
	}

	if _, err := os.Stat(execPath); err != nil {
		m.log.Error("Xray executable not found", logger.Fields{
			"path":    execPath,
			"version": m.currentVersion,
			"error":   err,
		})

		// 如果找不到当前版本，尝试下载
		m.log.Info("Trying to download missing Xray version", logger.Fields{
			"version": m.currentVersion,
		})
		if err := m.DownloadVersion(m.currentVersion); err != nil {
			m.log.Error("Failed to download Xray", logger.Fields{
				"error":   err,
				"version": m.currentVersion,
			})
			return fmt.Errorf("xray executable not found and download failed: %v", err)
		}

		// 重新检查可执行文件
		if _, err := os.Stat(execPath); err != nil {
			m.log.Error("Xray executable still not found after download", logger.Fields{
				"path":    execPath,
				"version": m.currentVersion,
			})
			return fmt.Errorf("xray executable not found even after download: %v", err)
		}
	}

	// 获取配置文件路径
	configPath := m.GetConfigPath()

	// 在 Windows 上转换为绝对路径
	if runtime.GOOS == "windows" {
		absPath, err := filepath.Abs(configPath)
		if err == nil {
			configPath = absPath
		}
	}

	// 检查是否使用自定义配置
	settings := m.settings.Get()
	if settings.Xray.CustomConfig && settings.Xray.ConfigPath != "" {
		// 使用自定义配置文件路径
		configPath = settings.Xray.ConfigPath
		m.log.Info("Using custom Xray config for startup", logger.Fields{
			"path": configPath,
		})

		// 验证自定义配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			m.log.Error("Custom config file does not exist", logger.Fields{
				"path":  configPath,
				"error": err,
			})
			return fmt.Errorf("custom config file does not exist: %v", err)
		}

		// 验证自定义配置文件是否为有效的JSON
		configData, err := os.ReadFile(configPath)
		if err != nil {
			m.log.Error("Failed to read custom config file", logger.Fields{
				"path":  configPath,
				"error": err,
			})
			return fmt.Errorf("failed to read custom config file: %v", err)
		}

		var configJSON map[string]interface{}
		if err := json.Unmarshal(configData, &configJSON); err != nil {
			m.log.Error("Custom config file is not valid JSON", logger.Fields{
				"path":  configPath,
				"error": err,
			})
			return fmt.Errorf("custom config file is not valid JSON: %v", err)
		}
	} else {
		// 验证默认配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			m.log.Warn("Default config file does not exist, creating", logger.Fields{
				"path": configPath,
			})

			// 生成默认配置
			defaultConfig, err := m.GenerateConfig()
			if err != nil {
				m.log.Error("Failed to generate default config", logger.Fields{
					"error": err,
				})
				return fmt.Errorf("failed to generate default config: %v", err)
			}

			// 更新配置文件
			if err := m.UpdateConfig(defaultConfig); err != nil {
				m.log.Error("Failed to update config", logger.Fields{
					"error": err,
				})
				return fmt.Errorf("failed to update config: %v", err)
			}
		}
	}

	// 确保日志目录存在
	logDir := filepath.Join("logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		m.log.Error("Failed to create logs directory", logger.Fields{
			"dir":   logDir,
			"error": err,
		})
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// 启动xray进程
	cmd := exec.Command(execPath, "-config", configPath)

	// 设置进程属性
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	// 设置输出
	stdout, err := os.Create(filepath.Join(logDir, "xray_stdout.log"))
	if err != nil {
		m.log.Error("Failed to create stdout log", logger.Fields{
			"error": err,
		})
		return fmt.Errorf("failed to create stdout log: %v", err)
	}

	stderr, err := os.Create(filepath.Join(logDir, "xray_stderr.log"))
	if err != nil {
		stdout.Close()
		m.log.Error("Failed to create stderr log", logger.Fields{
			"error": err,
		})
		return fmt.Errorf("failed to create stderr log: %v", err)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// 启动进程
	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		m.log.Error("Failed to start Xray process", logger.Fields{
			"error":   err,
			"path":    execPath,
			"version": m.currentVersion,
		})
		return fmt.Errorf("failed to start xray: %v", err)
	}

	m.process = cmd.Process
	m.running = true

	m.log.Info("Started Xray successfully", logger.Fields{
		"version": m.currentVersion,
		"pid":     m.process.Pid,
		"config":  configPath,
		"os":      osName,
		"arch":    osArch,
		"path":    execPath,
	})

	// 异步等待进程结束
	go func() {
		err := cmd.Wait()
		m.mutex.Lock()
		m.running = false
		m.process = nil
		m.mutex.Unlock()

		if err != nil {
			m.log.Error("Xray process exited with error", logger.Fields{
				"error":   err,
				"version": m.currentVersion,
			})
		} else {
			m.log.Info("Xray process exited normally", logger.Fields{
				"version": m.currentVersion,
			})
		}

		stdout.Close()
		stderr.Close()
	}()

	return nil
}

// Stop 停止xray
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running || m.process == nil {
		m.log.Info("No running Xray process to stop")
		return nil
	}

	pid := m.process.Pid
	m.log.Info("Stopping Xray process", logger.Fields{
		"pid":     pid,
		"version": m.currentVersion,
	})

	// 在Windows上使用taskkill确保彻底终止进程
	if runtime.GOOS == "windows" {
		// 首先尝试正常终止
		m.log.Info("Attempting to gracefully terminate Xray on Windows", logger.Fields{
			"pid": pid,
		})

		// 创建并执行taskkill命令
		cmd := exec.Command("taskkill", "/PID", fmt.Sprint(pid))
		if err := cmd.Run(); err != nil {
			m.log.Warn("Failed to gracefully terminate process, forcing termination", logger.Fields{
				"pid":   pid,
				"error": err,
			})

			// 强制终止
			forceCmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprint(pid))
			if err := forceCmd.Run(); err != nil {
				m.log.Error("Failed to forcefully terminate process", logger.Fields{
					"pid":   pid,
					"error": err,
				})
				// 即使命令失败，我们也继续，因为进程可能已经被终止
			}
		}

		// 等待短暂时间确保进程被终止
		time.Sleep(time.Second)

		// 验证进程是否已终止
		if processExists(pid) {
			m.log.Warn("Process still exists after termination attempt", logger.Fields{
				"pid": pid,
			})
			// 再次尝试强制终止
			exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(pid)).Run()
			time.Sleep(500 * time.Millisecond)
		}
	} else {
		// 在类Unix系统上发送SIGTERM信号
		m.log.Info("Sending SIGTERM to Xray process", logger.Fields{
			"pid": pid,
		})

		if err := m.process.Signal(os.Interrupt); err != nil {
			m.log.Warn("Failed to send SIGTERM", logger.Fields{
				"error": err,
			})
		}

		// 等待一段时间让进程优雅退出
		time.Sleep(time.Second)

		// 如果进程还在运行，强制终止
		if processExists(pid) {
			m.log.Info("Process still running after SIGTERM, sending SIGKILL", logger.Fields{
				"pid": pid,
			})

			if err := m.process.Kill(); err != nil {
				m.log.Error("Failed to kill process", logger.Fields{
					"error": err,
				})
			}

			time.Sleep(500 * time.Millisecond)
		}
	}

	// 标记为未运行，无论终止命令是否成功
	m.running = false
	m.process = nil

	m.log.Info("Stopped Xray process", logger.Fields{
		"version": m.currentVersion,
		"pid":     pid,
	})

	return nil
}

// processExists 检查进程是否存在
func processExists(pid int) bool {
	if runtime.GOOS == "windows" {
		// 在 Windows 上使用 tasklist 检查进程是否存在
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), fmt.Sprintf("%d", pid))
	} else {
		// 在类 Unix 系统上尝试发送信号0检查进程是否存在
		process, err := os.FindProcess(pid)
		if err != nil {
			return false
		}
		err = process.Signal(syscall.Signal(0))
		return err == nil
	}
}

// IsRunning 检查xray是否在运行
func (m *Manager) IsRunning() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.running
}

// UpdateConfig 更新xray配置文件
func (m *Manager) UpdateConfig(config map[string]interface{}) error {
	// 获取当前设置
	settings := m.settings.Get()
	configPath := m.GetConfigPath()

	// 检查是否使用自定义配置
	if settings.Xray.CustomConfig && settings.Xray.ConfigPath != "" {
		// 如果使用自定义配置，则使用自定义配置路径
		configPath = settings.Xray.ConfigPath
		m.log.Info("Using custom Xray config", logger.Fields{
			"path": configPath,
		})

		// 验证自定义配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("custom config file does not exist: %v", err)
		}
	} else {
		// 将配置转换为JSON
		jsonData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		// 写入配置文件
		if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}

		m.log.Info("Updated xray config", logger.Fields{
			"path": configPath,
		})
	}

	// 如果xray正在运行，重启它以应用新配置
	if m.running {
		if err := m.Stop(); err != nil {
			return fmt.Errorf("failed to stop xray: %v", err)
		}

		// 短暂延迟，确保进程完全停止
		time.Sleep(500 * time.Millisecond)

		if err := m.Start(); err != nil {
			return fmt.Errorf("failed to restart xray: %v", err)
		}
	}

	return nil
}

// GenerateConfig 生成完整的Xray配置
func (m *Manager) GenerateConfig() (map[string]interface{}, error) {
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"access":   "none",
			"error":    filepath.Join("logs", "xray.log"),
			"loglevel": "warning",
		},
		"inbounds": []map[string]interface{}{},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"tag":      "direct",
				"settings": map[string]interface{}{
					"domainStrategy": "UseIP",
				},
				"mux": map[string]interface{}{
					"enabled":     true,
					"concurrency": 8,
				},
			},
			{
				"protocol": "blackhole",
				"tag":      "blocked",
				"settings": map[string]interface{}{},
			},
		},
		"routing": map[string]interface{}{
			"domainStrategy": "AsIs",
			"rules": []map[string]interface{}{
				{
					"type":        "field",
					"outboundTag": "blocked",
					"ip":          []string{"geoip:private"},
				},
				{
					"type":        "field",
					"outboundTag": "direct",
					"domain":      []string{"geosite:cn"},
				},
				{
					"type":        "field",
					"outboundTag": "direct",
					"ip":          []string{"geoip:cn"},
				},
			},
		},
		"dns": map[string]interface{}{
			"servers": []string{
				"1.1.1.1",
				"8.8.8.8",
				"localhost",
			},
		},
		"policy": map[string]interface{}{
			"levels": map[string]interface{}{
				"0": map[string]interface{}{
					"handshake":         4,
					"connIdle":          300,
					"uplinkOnly":        2,
					"downlinkOnly":      5,
					"statsUserUplink":   true,
					"statsUserDownlink": true,
				},
			},
			"system": map[string]interface{}{
				"statsInboundUplink":    true,
				"statsInboundDownlink":  true,
				"statsOutboundUplink":   true,
				"statsOutboundDownlink": true,
			},
		},
	}

	// 获取当前设置
	// 注意: 当前未使用此设置，保留以便未来扩展
	// settings := m.settings.Get()

	// 添加API入站
	apiPort := 62789 // 默认API端口
	apiInbound := map[string]interface{}{
		"listen":   "127.0.0.1",
		"port":     apiPort,
		"protocol": "dokodemo-door",
		"settings": map[string]interface{}{
			"address": "127.0.0.1",
		},
		"tag": "api",
	}

	// 添加到入站列表
	inbounds := config["inbounds"].([]map[string]interface{})
	inbounds = append(inbounds, apiInbound)
	config["inbounds"] = inbounds

	// 添加API路由规则
	rules := config["routing"].(map[string]interface{})["rules"].([]map[string]interface{})
	rules = append(rules, map[string]interface{}{
		"type":        "field",
		"inboundTag":  []string{"api"},
		"outboundTag": "api",
	})
	config["routing"].(map[string]interface{})["rules"] = rules

	m.log.Info("Generated Xray config", logger.Fields{
		"version": m.currentVersion,
	})

	return config, nil
}

// SubscribeEvents 订阅Xray事件
func (m *Manager) SubscribeEvents() chan XrayEvent {
	m.eventsMutex.Lock()
	defer m.eventsMutex.Unlock()

	ch := make(chan XrayEvent, 10) // 缓冲通道，避免阻塞
	m.eventSubscribers[ch] = true

	return ch
}

// UnsubscribeEvents 取消订阅Xray事件
func (m *Manager) UnsubscribeEvents(ch chan XrayEvent) {
	m.eventsMutex.Lock()
	defer m.eventsMutex.Unlock()

	if _, exists := m.eventSubscribers[ch]; exists {
		delete(m.eventSubscribers, ch)
		close(ch)
	}
}

// PublishEvent 发布Xray事件
func (m *Manager) PublishEvent(event XrayEvent) {
	m.eventsMutex.RLock()
	defer m.eventsMutex.RUnlock()

	for ch := range m.eventSubscribers {
		// 非阻塞发送，如果通道已满则跳过
		select {
		case ch <- event:
		default:
			m.log.Warn("Event channel full, skipping event", logger.Fields{
				"event": event,
			})
		}
	}
}
