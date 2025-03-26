package xray

import (
	"archive/zip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
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

	// 获取平台信息并记录日志
	downloadOS, downloadArch := getPlatformInfo()
	m.log.Info("Detected platform", logger.Fields{
		"os":      downloadOS,
		"arch":    downloadArch,
		"version": version,
	})

	// 首先检查本地是否已有下载好的zip文件
	localZipPath := filepath.Join("xray", "downloads", fmt.Sprintf("Xray-%s-%s.zip", downloadOS, downloadArch))
	if _, err := os.Stat(localZipPath); err == nil {
		m.log.Info("Found local Xray package, using it instead of downloading", logger.Fields{
			"path": localZipPath,
		})

		// 发布进度事件 - 30%
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "progress",
			Message: "发现本地安装包，开始解压",
			Percent: 30,
		})

		// 解压缩本地文件
		if err := unzip(localZipPath, versionDir); err != nil {
			m.log.Error("Failed to extract local xray package", logger.Fields{
				"error": err,
				"path":  localZipPath,
			})
			// 发布错误事件
			m.PublishEvent(XrayEvent{
				Type:    "download",
				Version: version,
				Status:  "error",
				Message: fmt.Sprintf("解压本地安装包失败: %v", err),
				Percent: 30,
			})
			return fmt.Errorf("failed to extract local xray package: %v", err)
		}

		// 发布进度事件 - 80%
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "progress",
			Message: "解压完成，设置可执行权限",
			Percent: 80,
		})

		// 设置可执行权限
		execPath := m.GetExecutablePath(version)
		if runtime.GOOS != "windows" {
			if err := os.Chmod(execPath, 0755); err != nil {
				m.log.Error("Failed to set executable permission", logger.Fields{
					"error": err,
					"path":  execPath,
				})
				// 发布错误事件
				m.PublishEvent(XrayEvent{
					Type:    "download",
					Version: version,
					Status:  "error",
					Message: fmt.Sprintf("设置可执行权限失败: %v", err),
					Percent: 80,
				})
				return fmt.Errorf("failed to set executable permission: %v", err)
			}
		}

		m.log.Info("Installed xray from local package successfully", logger.Fields{
			"version": version,
			"path":    execPath,
		})

		// 发布完成事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "completed",
			Message: "安装成功",
			Percent: 100,
		})

		return nil
	}

	// 发布进度事件 - 20%
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "progress",
		Message: "本地无安装包，开始从网络下载",
		Percent: 20,
	})

	// 构建下载URL
	// 原始GitHub URL
	githubUrl := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
		version, downloadOS, downloadArch)

	// 备用镜像URL列表
	mirrorUrls := []string{
		// 高速镜像（优先尝试）
		fmt.Sprintf("https://mirror.ghproxy.com/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://ghproxy.com/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://gh-proxy.com/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),

		// 国内镜像
		fmt.Sprintf("https://download.fastgit.org/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://github.mirrorz.org/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://ghproxy.net/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://hub.gitmirror.com/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://gh2.yanqishui.work/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://kgithub.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),

		// 新增镜像
		fmt.Sprintf("https://gitclone.com/github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://proxy.zyun.vip/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://gh.ddlc.top/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://gh.idayer.com/https://github.com/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://hub.fgit.ml/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
		fmt.Sprintf("https://hub.nuaa.cf/XTLS/Xray-core/releases/download/%s/Xray-%s-%s.zip",
			version, downloadOS, downloadArch),
	}

	// 记录所有URL用于调试
	allUrls := append([]string{githubUrl}, mirrorUrls...)

	m.log.Info("Download URLs prepared", logger.Fields{
		"urls":    allUrls,
		"os":      downloadOS,
		"arch":    downloadArch,
		"version": version,
	})

	// 创建下载目录（如果不存在）
	if err := os.MkdirAll(filepath.Join("xray", "downloads"), 0755); err != nil {
		m.log.Warn("Failed to create downloads directory", logger.Fields{
			"error": err,
		})
	}

	// 下载到临时文件
	tempFile := filepath.Join(versionDir, "xray.zip")

	// 尝试从所有镜像下载
	var lastError error
	var downloaded bool = false
	var attemptErrors = make(map[string]string) // 记录每个镜像的错误

	// 首先尝试所有镜像
	for i, url := range mirrorUrls {
		m.log.Info("Trying mirror", logger.Fields{
			"url":     url,
			"attempt": i + 1,
			"total":   len(mirrorUrls),
		})

		// 发布进度事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "progress",
			Message: fmt.Sprintf("尝试从镜像 %d/%d 下载 Xray v%s", i+1, len(mirrorUrls), version),
			Percent: 30 + (i * 2), // 减小每个镜像的百分比增量，避免百分比过快增长
			Details: map[string]string{
				"url":     url,
				"os":      downloadOS,
				"arch":    downloadArch,
				"attempt": fmt.Sprintf("%d/%d", i+1, len(mirrorUrls)),
			},
		})

		// 记录开始时间，用于计算下载耗时
		startTime := time.Now()

		err := downloadFile(url, tempFile)

		// 计算耗时
		elapsed := time.Since(startTime)

		if err == nil {
			downloaded = true
			m.log.Info("Successfully downloaded from mirror", logger.Fields{
				"url":      url,
				"attempt":  i + 1,
				"duration": elapsed.String(),
				"size":     getFileSize(tempFile),
			})

			// 发布进度事件 - 下载成功
			m.PublishEvent(XrayEvent{
				Type:    "download",
				Version: version,
				Status:  "progress",
				Message: fmt.Sprintf("从镜像 %d 下载成功，耗时 %s", i+1, elapsed.Round(time.Second).String()),
				Percent: 50,
				Details: map[string]string{
					"url":      url,
					"duration": elapsed.String(),
					"success":  "true",
				},
			})

			break
		}

		lastError = err
		attemptErrors[url] = err.Error()

		m.log.Warn("Failed to download from mirror, trying next", logger.Fields{
			"error":    err,
			"url":      url,
			"attempt":  i + 1,
			"duration": elapsed.String(),
		})

		// 发布进度事件 - 当前镜像失败
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "progress",
			Message: fmt.Sprintf("镜像 %d 下载失败: %v", i+1, err),
			Percent: 30 + (i * 2),
			Details: map[string]string{
				"url":     url,
				"error":   err.Error(),
				"success": "false",
			},
		})
	}

	// 如果所有镜像都失败，尝试GitHub直接下载
	if !downloaded {
		m.log.Warn("All mirrors failed, trying GitHub directly", logger.Fields{
			"error":        lastError,
			"mirror_count": len(mirrorUrls),
			"errors":       attemptErrors,
		})

		// 发布进度事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "progress",
			Message: "所有镜像下载失败，尝试从GitHub直接下载",
			Percent: 45,
			Details: map[string]interface{}{
				"mirror_errors": attemptErrors,
			},
		})

		// 记录开始时间
		startTime := time.Now()

		err := downloadFile(githubUrl, tempFile)

		// 计算耗时
		elapsed := time.Since(startTime)

		if err == nil {
			downloaded = true
			m.log.Info("Successfully downloaded from GitHub", logger.Fields{
				"url":      githubUrl,
				"duration": elapsed.String(),
				"size":     getFileSize(tempFile),
			})
		} else {
			lastError = err
			attemptErrors[githubUrl] = err.Error()

			m.log.Error("Failed to download from GitHub", logger.Fields{
				"error":    err,
				"url":      githubUrl,
				"duration": elapsed.String(),
			})
		}
	}

	if !downloaded {
		m.log.Error("Failed to download xray from all sources", logger.Fields{
			"error":        lastError,
			"version":      version,
			"os":           downloadOS,
			"arch":         downloadArch,
			"mirror_count": len(mirrorUrls) + 1, // +1 for GitHub
			"errors":       attemptErrors,
		})
		// 发布错误事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "error",
			Message: fmt.Sprintf("所有下载源均失败: %v", lastError),
			Percent: 45,
			Details: map[string]interface{}{
				"error":         lastError.Error(),
				"version":       version,
				"os":            downloadOS,
				"arch":          downloadArch,
				"all_errors":    attemptErrors,
				"download_urls": append(mirrorUrls, githubUrl),
			},
		})
		return fmt.Errorf("failed to download xray from all sources: %v", lastError)
	}

	// 发布进度事件 - 60%
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "progress",
		Message: "下载完成，开始解压",
		Percent: 60,
	})

	// 解压缩
	if err := unzip(tempFile, versionDir); err != nil {
		m.log.Error("Failed to extract xray", logger.Fields{
			"error": err,
			"file":  tempFile,
		})
		// 发布错误事件
		m.PublishEvent(XrayEvent{
			Type:    "download",
			Version: version,
			Status:  "error",
			Message: fmt.Sprintf("解压失败: %v", err),
			Percent: 60,
		})
		return fmt.Errorf("failed to extract xray: %v", err)
	}

	// 发布进度事件 - 80%
	m.PublishEvent(XrayEvent{
		Type:    "download",
		Version: version,
		Status:  "progress",
		Message: "解压完成，清理临时文件",
		Percent: 80,
	})

	// 删除临时文件
	os.Remove(tempFile)

	// 设置可执行权限
	execPath := m.GetExecutablePath(version)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(execPath, 0755); err != nil {
			m.log.Error("Failed to set executable permission", logger.Fields{
				"error": err,
				"path":  execPath,
			})
			// 发布错误事件
			m.PublishEvent(XrayEvent{
				Type:    "download",
				Version: version,
				Status:  "error",
				Message: fmt.Sprintf("设置可执行权限失败: %v", err),
				Percent: 80,
			})
			return fmt.Errorf("failed to set executable permission: %v", err)
		}
	}

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

// 辅助函数获取系统平台信息
func getPlatformInfo() (string, string) {
	// 确定操作系统
	osName := runtime.GOOS
	if osName == "darwin" {
		osName = "macos"
	}

	// 确定系统架构
	arch := runtime.GOARCH
	// 特殊情况处理
	if arch == "amd64" {
		arch = "64"
	} else if arch == "386" {
		arch = "32"
	} else if arch == "arm64" {
		if osName == "macos" {
			// 苹果M系列处理器
			arch = "arm64-v8a"
		} else {
			arch = "arm64-v8a"
		}
	} else if arch == "arm" {
		arch = "arm32-v7a"
	}

	return osName, arch
}

// SwitchVersion 切换xray版本
func (m *Manager) SwitchVersion(version string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 发布切换开始事件
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "start",
		Message: fmt.Sprintf("开始切换到版本 %s", version),
		Percent: 0,
	})

	// 检查版本是否支持
	found := false
	for _, v := range SupportedVersions {
		if v == version {
			found = true
			break
		}
	}

	if !found {
		// 发布错误事件
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "error",
			Message: fmt.Sprintf("不支持的版本: %s", version),
			Percent: 0,
		})
		return fmt.Errorf("unsupported version: %s", version)
	}

	// 发布进度事件 - 10%
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "progress",
		Message: "检查版本文件",
		Percent: 10,
	})

	// 检查版本是否已下载，如果没有则下载
	if !m.VersionExists(version) {
		m.log.Info("Version not found, downloading", logger.Fields{
			"version": version,
		})

		// 发布进度事件 - 20%
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "progress",
			Message: "版本文件不存在，开始下载（这可能需要几分钟，请耐心等待）",
			Percent: 20,
			Details: map[string]interface{}{
				"timeout_info": "下载可能需要5-10分钟，请不要关闭窗口",
				"download_url": fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/%s/", version),
			},
		})

		// 创建一个goroutine定期更新下载进度
		ticker := time.NewTicker(10 * time.Second)
		done := make(chan bool)

		go func() {
			count := 0
			messages := []string{
				"正在下载中，请耐心等待...",
				"下载进行中，这可能需要几分钟时间...",
				"下载速度可能较慢，请稍候...",
				"正在尝试多个下载源，请稍后...",
				"如果下载失败，可以尝试手动下载并放入xray/bin/%s/目录",
				"国内网络环境下载可能较慢，请继续等待...",
				"正在连接境外服务器，可能需要较长时间...",
			}

			percent := 22

			for {
				select {
				case <-done:
					ticker.Stop()
					return
				case <-ticker.C:
					message := messages[count%len(messages)]
					// 如果是特定消息，替换版本号
					if strings.Contains(message, "%s") {
						message = fmt.Sprintf(message, version)
					}

					// 每次递增一点进度，最高到48%
					if percent < 48 {
						percent += 2
					}

					m.PublishEvent(XrayEvent{
						Type:    "switch",
						Version: version,
						Status:  "progress",
						Message: message,
						Percent: percent,
					})

					count++
				}
			}
		}()

		// 执行下载
		err := m.DownloadVersion(version)
		close(done) // 停止进度更新goroutine

		if err != nil {
			// 添加更多的错误信息和建议
			errorMsg := fmt.Sprintf("下载失败: %v", err)
			suggestions := []string{
				"请检查您的网络连接",
				"尝试使用VPN或代理",
				"手动下载Xray到xray/downloads目录",
				fmt.Sprintf("检查版本 %s 是否存在", version),
			}

			errorDetails := map[string]interface{}{
				"error":       err.Error(),
				"suggestions": suggestions,
				"manual_url":  fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/tag/%s", version),
			}

			// 下载失败事件
			m.PublishEvent(XrayEvent{
				Type:    "switch",
				Version: version,
				Status:  "error",
				Message: errorMsg,
				Percent: 20,
				Details: errorDetails,
			})
			return fmt.Errorf("failed to download version %s: %v", version, err)
		}

		// 下载成功，继续切换
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "progress",
			Message: "下载完成，准备切换",
			Percent: 50,
		})
	} else {
		// 版本已存在，跳过下载
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "progress",
			Message: "版本文件已存在，准备切换",
			Percent: 50,
		})
	}

	// 保存旧版本状态，用于后续恢复
	wasRunning := m.running
	oldVersion := m.currentVersion

	// 停止当前运行的xray
	if m.running {
		m.log.Info("Stopping current Xray version", logger.Fields{
			"version": m.currentVersion,
		})

		// 发布进度事件 - 60%
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "progress",
			Message: "停止当前运行的版本",
			Percent: 60,
		})

		if err := m.Stop(); err != nil {
			// 发布错误事件
			m.PublishEvent(XrayEvent{
				Type:    "switch",
				Version: version,
				Status:  "error",
				Message: fmt.Sprintf("停止当前版本失败: %v", err),
				Percent: 60,
			})
			return fmt.Errorf("failed to stop current xray: %v", err)
		}

		// 确保进程完全停止
		time.Sleep(500 * time.Millisecond)
	}

	// 发布进度事件 - 70%
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "progress",
		Message: "验证新版本文件",
		Percent: 70,
	})

	// 先检查新版本的可执行文件是否存在
	newExecPath := m.GetExecutablePath(version)
	if _, err := os.Stat(newExecPath); os.IsNotExist(err) {
		// 如果新版本不存在，尝试回退到之前的版本
		m.log.Error("New version executable not found, trying to revert", logger.Fields{
			"version": version,
			"path":    newExecPath,
		})

		// 发布错误事件
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "error",
			Message: fmt.Sprintf("新版本可执行文件不存在: %s", newExecPath),
			Percent: 70,
		})

		if oldVersion != "" && wasRunning {
			m.currentVersion = oldVersion
			m.Start()

			// 发布回退事件
			m.PublishEvent(XrayEvent{
				Type:    "switch",
				Version: oldVersion,
				Status:  "reverted",
				Message: fmt.Sprintf("回退到版本 %s", oldVersion),
				Percent: 100,
			})
		}

		return fmt.Errorf("xray executable for version %s not found at path: %s", version, newExecPath)
	}

	// 更新当前版本
	m.currentVersion = version

	// 发布进度事件 - 80%
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "progress",
		Message: "更新设置",
		Percent: 80,
	})

	// 更新设置
	settings := m.settings.Get()
	settings.Xray.Version = version
	if err := m.settings.Save(); err != nil {
		m.log.Error("Failed to save settings, but continuing", logger.Fields{
			"error": err,
		})
		// 不中断流程，继续执行
	}

	m.log.Info("Switched xray version", logger.Fields{
		"version":  version,
		"execPath": newExecPath,
	})

	// 发布进度事件 - 90%
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "progress",
		Message: "版本切换成功",
		Percent: 90,
	})

	// 如果之前在运行，则启动新版本
	if wasRunning {
		m.log.Info("Restarting Xray with new version", logger.Fields{
			"version": version,
		})

		// 发布重启事件
		m.PublishEvent(XrayEvent{
			Type:    "switch",
			Version: version,
			Status:  "progress",
			Message: "重启服务",
			Percent: 95,
		})

		if err := m.Start(); err != nil {
			// 如果新版本启动失败，尝试回退到旧版本
			m.log.Error("Failed to start new xray version, trying to revert", logger.Fields{
				"error":   err,
				"version": version,
			})

			// 发布错误事件
			m.PublishEvent(XrayEvent{
				Type:    "switch",
				Version: version,
				Status:  "error",
				Message: fmt.Sprintf("启动新版本失败: %v", err),
				Percent: 95,
			})

			if oldVersion != "" {
				m.currentVersion = oldVersion
				m.Start()

				// 发布回退事件
				m.PublishEvent(XrayEvent{
					Type:    "switch",
					Version: oldVersion,
					Status:  "reverted",
					Message: fmt.Sprintf("回退到版本 %s", oldVersion),
					Percent: 100,
				})
			}

			return fmt.Errorf("failed to start new xray version: %v", err)
		}
	}

	// 发布完成事件
	m.PublishEvent(XrayEvent{
		Type:    "switch",
		Version: version,
		Status:  "completed",
		Message: "版本切换完成",
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

// 辅助函数

// downloadFile 下载文件到指定路径
func downloadFile(url, filepath string) error {
	// 根据系统选择下载方法
	if runtime.GOOS == "windows" {
		return downloadFileWindows(url, filepath)
	} else {
		return downloadFileUnix(url, filepath)
	}
}

// downloadFileUnix 在Linux/Unix系统上使用curl命令下载文件
func downloadFileUnix(url, filepath string) error {
	// 确保目录存在
	if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	fmt.Printf("使用curl下载: %s\n", url)

	// 创建curl命令，使用-L跟随重定向，-o指定输出文件
	cmd := exec.Command("curl", "-L", "--connect-timeout", "30",
		"--retry", "5", "--retry-delay", "2", "--retry-max-time", "120",
		"-H", "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"-o", filepath, url)

	// 获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("curl download failed: %v, output: %s", err, string(output))
	}

	// 验证下载文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("download completed but file not found")
	}

	// 检查文件是否为空
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("downloaded file is empty")
	}

	return nil
}

// downloadFileWindows 在Windows系统上使用内置HTTP客户端下载文件
func downloadFileWindows(url, filepath string) error {
	// 创建http客户端，设置超时
	client := &http.Client{
		Timeout: 300 * time.Second, // 增加到300秒 (5分钟)
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 忽略SSL证书验证
			// 增加连接超时
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second, // 增加连接超时
				KeepAlive: 60 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       120 * time.Second,
			TLSHandshakeTimeout:   30 * time.Second,
			ExpectContinueTimeout: 30 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
			Proxy:                 http.ProxyFromEnvironment, // 支持系统代理
			DisableCompression:    false,                     // 允许压缩
			ForceAttemptHTTP2:     true,                      // 尝试使用HTTP/2
		},
	}

	// 最多重试8次
	maxRetries := 8
	var err error
	var resp *http.Response

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		// 设置请求头，模拟浏览器行为
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")

		// 添加 Range 请求头以支持断点续传
		if i > 0 {
			// 检查文件是否已经存在并有内容
			if fi, err := os.Stat(filepath); err == nil && fi.Size() > 0 {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", fi.Size()))
				fmt.Printf("断点续传: 从字节 %d 继续下载\n", fi.Size())
			}
		}

		// 发送请求
		resp, err = client.Do(req)

		// 处理超时错误并提供更具体的消息
		if err != nil {
			errMsg := err.Error()
			if os.IsTimeout(err) || strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
				fmt.Printf("第 %d 次下载超时: %v\n", i+1, err)
			} else {
				fmt.Printf("第 %d 次下载失败: %v\n", i+1, err)
			}

			if i < maxRetries-1 {
				// 指数退避策略，每次重试等待更长时间
				waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
				if waitTime > 30*time.Second {
					waitTime = 30 * time.Second // 最大等待30秒
				}
				fmt.Printf("等待 %v 后重试...\n", waitTime)
				time.Sleep(waitTime)
				continue
			}
			return fmt.Errorf("download failed after %d retries: %v", maxRetries, err)
		}

		// 检查是否支持断点续传
		if resp.StatusCode == http.StatusPartialContent {
			fmt.Println("服务器支持断点续传，继续下载...")
		}

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
			defer resp.Body.Close()

			// 决定是追加还是创建新文件
			var out *os.File
			var fileMode int
			if resp.StatusCode == http.StatusPartialContent && i > 0 {
				fileMode = os.O_WRONLY | os.O_APPEND
				fmt.Println("正在追加到现有文件...")
			} else {
				fileMode = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
				fmt.Println("正在创建新文件...")
			}

			// 创建或打开目标文件
			out, err = os.OpenFile(filepath, fileMode, 0644)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			// 使用更大的缓冲区提高复制速度
			buf := make([]byte, 64*1024) // 64KB的缓冲区

			// 设置进度报告
			total := resp.ContentLength
			var downloaded int64 = 0
			lastReportTime := time.Now()

			// 创建读取器以计算进度
			reader := &progressReader{
				Reader: resp.Body,
				Total:  total,
				OnProgress: func(n int64) {
					downloaded += int64(n)
					// 每秒最多报告一次进度
					if time.Since(lastReportTime) > time.Second {
						percent := float64(0)
						if total > 0 {
							percent = float64(downloaded) / float64(total) * 100
						}
						fmt.Printf("下载进度: %.2f%% (%d/%d 字节)\n", percent, downloaded, total)
						lastReportTime = time.Now()
					}
				},
			}

			_, err = io.CopyBuffer(out, reader, buf)
			out.Close()

			if err != nil {
				fmt.Printf("复制文件内容失败: %v，将重试...\n", err)
				if i < maxRetries-1 {
					time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
					continue
				}
				return fmt.Errorf("failed to write file after %d retries: %v", maxRetries, err)
			}

			fmt.Println("下载完成!")
			return nil
		}

		// 处理不成功的响应状态
		if resp != nil {
			errBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			err = fmt.Errorf("HTTP %s: %s", resp.Status, string(errBody))

			// 特定状态码的处理，例如重定向
			if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
				if location, err := resp.Location(); err == nil && location != nil {
					// 处理重定向
					redirectURL := location.String()
					fmt.Printf("重定向到: %s\n", redirectURL)
					url = redirectURL // 更新URL以便下次重试
				}
			}
		}

		if i < maxRetries-1 {
			// 指数退避策略，每次重试等待更长时间
			waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
			if waitTime > 30*time.Second {
				waitTime = 30 * time.Second // 最大等待30秒
			}
			fmt.Printf("第 %d 次下载失败, 状态码: %d, 等待 %v 后重试: %v\n",
				i+1, resp.StatusCode, waitTime, err)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("download failed after %d retries, last error: %v", maxRetries, err)
}

// 进度读取器，用于报告下载进度
type progressReader struct {
	io.Reader
	Total      int64
	OnProgress func(n int64)
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if n > 0 && r.OnProgress != nil {
		r.OnProgress(int64(n))
	}
	return
}

// unzip 解压zip文件到指定目录
func unzip(src, dest string) error {
	// 先检查源文件是否存在和是否有效
	fileInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat zip file: %v", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("zip file is empty")
	}

	fmt.Printf("解压文件: %s (大小: %d 字节)\n", src, fileInfo.Size())

	// 尝试使用系统命令解压(Linux/Unix)
	if runtime.GOOS != "windows" {
		// 确保目标目录存在
		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %v", err)
		}

		// 在Linux上使用unzip命令
		cmd := exec.Command("unzip", "-o", src, "-d", dest)
		output, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Printf("使用系统unzip命令解压成功\n")
			return nil
		}

		fmt.Printf("系统unzip命令失败: %v, 输出: %s\n尝试使用Go内置解压方法...\n", err, string(output))
	}

	// 使用原生 Go 实现，避免在 Windows 上使用 PowerShell 命令行
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	// 创建目标目录（如果不存在）
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	fmt.Printf("ZIP文件信息: 共包含 %d 个文件\n", len(r.File))

	// 遍历 zip 文件中的所有文件/目录
	for _, f := range r.File {
		// 构建解压后的路径
		fpath := filepath.Join(dest, f.Name)

		// 检查文件路径是否在目标目录内（安全检查）
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		// 打印出处理的文件名，便于调试
		fmt.Printf("解压文件: %s\n", f.Name)

		// 如果是目录，则创建
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			continue
		}

		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directory: %v", err)
		}

		// 打开 zip 中的文件
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}

		// 创建目标文件
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create output file: %v", err)
		}

		// 复制内容
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file content: %v", err)
		}
	}

	fmt.Printf("ZIP文件成功解压到: %s\n", dest)
	return nil
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

// getFileSize 返回文件大小的字符串表示
func getFileSize(filePath string) string {
	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "Unknown"
	}

	sizeInBytes := fileInfo.Size()

	// 转换为友好格式
	if sizeInBytes < 1024 {
		return fmt.Sprintf("%d B", sizeInBytes)
	} else if sizeInBytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(sizeInBytes)/1024)
	} else if sizeInBytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(sizeInBytes)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(sizeInBytes)/(1024*1024*1024))
	}
}
