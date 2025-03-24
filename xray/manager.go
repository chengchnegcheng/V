package xray

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
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
}

// New 创建一个新的xray版本管理器
func New(log *logger.Logger, settingsManager *settings.Manager) *Manager {
	binPath := filepath.Join("xray", "bin")

	// 确保二进制目录存在
	os.MkdirAll(binPath, 0755)

	return &Manager{
		log:      log,
		settings: settingsManager,
		binPath:  binPath,
		running:  false,
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

	// 创建版本目录
	versionDir := filepath.Join(m.binPath, version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %v", err)
	}

	// 首先检查本地是否已有下载好的zip文件
	localZipPath := filepath.Join("xray", "downloads", fmt.Sprintf("Xray-%s-%s.zip", runtime.GOOS, getArchString()))
	if _, err := os.Stat(localZipPath); err == nil {
		m.log.Info("Found local Xray package, using it instead of downloading", logger.Fields{
			"path": localZipPath,
		})

		// 解压缩本地文件
		if err := unzip(localZipPath, versionDir); err != nil {
			return fmt.Errorf("failed to extract local xray package: %v", err)
		}

		// 设置可执行权限
		execPath := m.GetExecutablePath(version)
		if runtime.GOOS != "windows" {
			if err := os.Chmod(execPath, 0755); err != nil {
				return fmt.Errorf("failed to set executable permission: %v", err)
			}
		}

		m.log.Info("Installed xray from local package successfully", logger.Fields{
			"version": version,
			"path":    execPath,
		})

		return nil
	}

	// 确定系统架构
	arch := getArchString()

	// 构建下载URL
	url := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
		version, runtime.GOOS, arch)

	// 创建下载目录（如果不存在）
	if err := os.MkdirAll(filepath.Join("xray", "downloads"), 0755); err != nil {
		m.log.Warn("Failed to create downloads directory", logger.Fields{
			"error": err,
		})
	}

	// 下载到临时文件
	tempFile := filepath.Join(versionDir, "xray.zip")
	m.log.Info("Downloading Xray", logger.Fields{
		"url": url,
		"to":  tempFile,
	})

	if err := downloadFile(url, tempFile); err != nil {
		return fmt.Errorf("failed to download xray: %v", err)
	}

	// 解压缩
	if err := unzip(tempFile, versionDir); err != nil {
		return fmt.Errorf("failed to extract xray: %v", err)
	}

	// 删除临时文件
	os.Remove(tempFile)

	// 设置可执行权限
	execPath := m.GetExecutablePath(version)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(execPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %v", err)
		}
	}

	m.log.Info("Downloaded xray successfully", logger.Fields{
		"version": version,
		"path":    execPath,
	})

	return nil
}

// 辅助函数获取架构字符串
func getArchString() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "64"
	} else if arch == "386" {
		arch = "32"
	} else if arch == "arm" {
		arch = "arm"
	} else if arch == "arm64" {
		arch = "arm64-v8a"
	}

	// 确定操作系统
	osName := runtime.GOOS
	if osName == "darwin" {
		osName = "macos"
	}

	return arch
}

// SwitchVersion 切换xray版本
func (m *Manager) SwitchVersion(version string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查版本是否支持
	found := false
	for _, v := range SupportedVersions {
		if v == version {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unsupported version: %s", version)
	}

	// 检查版本是否已下载，如果没有则下载
	if !m.VersionExists(version) {
		if err := m.DownloadVersion(version); err != nil {
			return fmt.Errorf("failed to download version %s: %v", version, err)
		}
	}

	// 停止当前运行的xray
	if m.running {
		if err := m.Stop(); err != nil {
			return fmt.Errorf("failed to stop current xray: %v", err)
		}
	}

	// 更新当前版本
	m.currentVersion = version

	// 更新设置
	settings := m.settings.Get()
	settings.Xray.Version = version
	if err := m.settings.Save(); err != nil {
		return fmt.Errorf("failed to save settings: %v", err)
	}

	m.log.Info("Switched xray version", logger.Fields{
		"version": version,
	})

	// 如果之前在运行，则启动新版本
	if m.running {
		if err := m.Start(); err != nil {
			return fmt.Errorf("failed to start new xray version: %v", err)
		}
	}

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
		return fmt.Errorf("xray is already running")
	}

	// 获取可执行文件路径
	execPath := m.GetExecutablePath(m.currentVersion)
	if _, err := os.Stat(execPath); err != nil {
		return fmt.Errorf("xray executable not found: %v", err)
	}

	// 获取配置文件路径
	configPath := m.GetConfigPath()

	// 启动xray进程
	cmd := exec.Command(execPath, "-config", configPath)

	// 设置输出
	stdout, err := os.Create(filepath.Join("logs", "xray_stdout.log"))
	if err != nil {
		return fmt.Errorf("failed to create stdout log: %v", err)
	}

	stderr, err := os.Create(filepath.Join("logs", "xray_stderr.log"))
	if err != nil {
		return fmt.Errorf("failed to create stderr log: %v", err)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xray: %v", err)
	}

	m.process = cmd.Process
	m.running = true

	m.log.Info("Started xray", logger.Fields{
		"version": m.currentVersion,
		"pid":     m.process.Pid,
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
				"error": err,
			})
		} else {
			m.log.Info("Xray process exited normally")
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
		return nil
	}

	// 在Windows上使用taskkill确保彻底终止进程
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(m.process.Pid)).Run()
	} else {
		// 在类Unix系统上发送SIGTERM信号
		m.process.Signal(os.Interrupt)

		// 等待一段时间让进程优雅退出
		time.Sleep(time.Second)

		// 如果进程还在运行，强制终止
		if isProcessRunning(m.process.Pid) {
			m.process.Kill()
		}
	}

	m.running = false
	m.process = nil

	m.log.Info("Stopped xray", logger.Fields{
		"version": m.currentVersion,
	})

	return nil
}

// IsRunning 检查xray是否在运行
func (m *Manager) IsRunning() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.running
}

// 辅助函数

// downloadFile 下载文件到指定路径
func downloadFile(url, filepath string) error {
	// 创建http请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// 创建目标文件
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 写入响应正文到文件
	_, err = io.Copy(out, resp.Body)
	return err
}

// unzip 解压zip文件到指定目录
func unzip(src, dest string) error {
	// 注意：这里简化了实现，实际项目中应该用archive/zip包实现
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// 在Windows上使用PowerShell解压
		cmd = exec.Command("powershell", "-command",
			fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", src, dest))
	} else {
		// 在类Unix系统上使用unzip命令
		cmd = exec.Command("unzip", "-o", src, "-d", dest)
	}

	return cmd.Run()
}

// isProcessRunning 检查进程是否在运行
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 在Windows上，FindProcess总是成功，所以需要额外检查
	if runtime.GOOS == "windows" {
		// 发送信号0测试进程是否存在
		err = process.Signal(os.Kill)
		return err == nil
	}

	return true
}

// UpdateConfig 更新xray配置文件
func (m *Manager) UpdateConfig(config map[string]interface{}) error {
	configPath := m.GetConfigPath()

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

	// 如果xray正在运行，重启它以应用新配置
	if m.running {
		if err := m.Stop(); err != nil {
			return fmt.Errorf("failed to stop xray: %v", err)
		}

		if err := m.Start(); err != nil {
			return fmt.Errorf("failed to restart xray: %v", err)
		}
	}

	return nil
}
