package xray

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// AutoDownloader 用于自动下载和安装 Xray 的工具
type AutoDownloader struct {
	version       string
	downloadPath  string
	outputPath    string
	githubBaseURL string
}

// NewAutoDownloader 创建一个新的自动下载器
func NewAutoDownloader(version string) *AutoDownloader {
	// 创建下载目录
	downloadPath := filepath.Join("xray", "downloads")
	os.MkdirAll(downloadPath, 0755)

	// 输出目录 (bin/版本号)
	outputPath := filepath.Join("xray", "bin", version)

	return &AutoDownloader{
		version:       version,
		downloadPath:  downloadPath,
		outputPath:    outputPath,
		githubBaseURL: "https://github.com/XTLS/Xray-core/releases/download",
	}
}

// DownloadAndInstall 下载并安装 Xray
func (d *AutoDownloader) DownloadAndInstall() error {
	// 1. 获取系统和架构信息
	osName, osArch := getPlatformInfo()

	// 2. 构建下载 URL
	fileName := d.getFileName(osName, osArch)
	if fileName == "" {
		return fmt.Errorf("unsupported platform: %s/%s", osName, osArch)
	}

	downloadURL := fmt.Sprintf("%s/%s/%s", d.githubBaseURL, d.version, fileName)

	// 3. 下载文件
	downloadFilePath := filepath.Join(d.downloadPath, fileName)
	fmt.Printf("开始下载 Xray: %s\n", downloadURL)

	// 检查文件是否已存在
	if _, err := os.Stat(downloadFilePath); err == nil {
		fileInfo, _ := os.Stat(downloadFilePath)
		if fileInfo.Size() > 0 {
			fmt.Printf("使用已下载的文件: %s (%s)\n",
				downloadFilePath, getFileSize(downloadFilePath))
		} else {
			// 如果文件为空，删除并重新下载
			os.Remove(downloadFilePath)
		}
	}

	// 确保输出目录存在
	os.MkdirAll(d.outputPath, 0755)

	// 如果文件不存在，开始下载
	if _, err := os.Stat(downloadFilePath); os.IsNotExist(err) {
		if err := downloadFile(downloadURL, downloadFilePath); err != nil {
			// 尝试备用下载链接
			altURL := d.getAlternativeURL(fileName)
			if altURL != "" {
				fmt.Printf("主下载链接失败，尝试备用链接: %s\n", altURL)
				err = downloadFile(altURL, downloadFilePath)
				if err != nil {
					return fmt.Errorf("all download attempts failed: %v", err)
				}
			} else {
				return fmt.Errorf("download failed: %v", err)
			}
		}
	}

	// 4. 解压文件
	fmt.Printf("解压文件到: %s\n", d.outputPath)
	if err := unzip(downloadFilePath, d.outputPath); err != nil {
		return fmt.Errorf("failed to extract: %v", err)
	}

	// 5. 设置可执行权限
	executableName := "xray"
	if runtime.GOOS == "windows" {
		executableName = "xray.exe"
	}

	execPath := filepath.Join(d.outputPath, executableName)
	fmt.Printf("设置可执行权限: %s\n", execPath)

	if runtime.GOOS != "windows" {
		cmd := exec.Command("chmod", "+x", execPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("设置可执行权限失败: %v\n", err)
			// 继续执行，不返回错误
		}
	}

	// 验证文件是否有效
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found after installation: %s", execPath)
	}

	fmt.Printf("Xray 安装成功: %s\n", execPath)
	return nil
}

// getFileName 根据操作系统和架构获取下载文件名
func (d *AutoDownloader) getFileName(osName, osArch string) string {
	// 构建文件名
	var fileName string

	// 如果是旧版本格式 (v1.x.x)
	if strings.HasPrefix(d.version, "v1.") {
		if osName == "windows" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-windows-64.zip"
			case "386", "x86":
				fileName = "Xray-windows-32.zip"
			case "arm64":
				fileName = "Xray-windows-arm64-v8a.zip"
			}
		} else if osName == "darwin" || osName == "macos" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-macos-64.zip"
			case "arm64":
				fileName = "Xray-macos-arm64-v8a.zip"
			}
		} else if osName == "linux" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-linux-64.zip"
			case "386", "x86":
				fileName = "Xray-linux-32.zip"
			case "arm64", "aarch64":
				fileName = "Xray-linux-arm64-v8a.zip"
			case "arm":
				fileName = "Xray-linux-arm32-v7a.zip"
			}
		}
	} else if strings.HasPrefix(d.version, "v24.") || strings.HasPrefix(d.version, "v25.") {
		// v24.x.x 和 v25.x.x 版本格式（例如 Xray-windows-x64.zip）
		if osName == "windows" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-windows-x64.zip"
			case "386", "x86":
				fileName = "Xray-windows-x86.zip"
			case "arm64":
				fileName = "Xray-windows-arm64.zip"
			}
		} else if osName == "darwin" || osName == "macos" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-macos-x64.zip"
			case "arm64":
				fileName = "Xray-macos-arm64.zip"
			}
		} else if osName == "linux" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-linux-x64.zip"
			case "386", "x86":
				fileName = "Xray-linux-x86.zip"
			case "arm64", "aarch64":
				fileName = "Xray-linux-arm64.zip"
			case "arm":
				fileName = "Xray-linux-arm.zip"
			}
		}
	} else {
		// 未知版本格式，尝试使用新格式
		if osName == "windows" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-windows-x64.zip"
			case "386", "x86":
				fileName = "Xray-windows-x86.zip"
			case "arm64":
				fileName = "Xray-windows-arm64.zip"
			}
		} else if osName == "darwin" || osName == "macos" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-macos-x64.zip"
			case "arm64":
				fileName = "Xray-macos-arm64.zip"
			}
		} else if osName == "linux" {
			switch osArch {
			case "amd64", "x64", "x86_64":
				fileName = "Xray-linux-x64.zip"
			case "386", "x86":
				fileName = "Xray-linux-x86.zip"
			case "arm64", "aarch64":
				fileName = "Xray-linux-arm64.zip"
			case "arm":
				fileName = "Xray-linux-arm.zip"
			}
		}
	}

	return fileName
}

// getAlternativeURL 获取备用下载链接
func (d *AutoDownloader) getAlternativeURL(fileName string) string {
	// 提供多个备用下载链接
	// 1. 使用 GitHub 镜像站点
	githubMirrors := []string{
		"https://download.fastgit.org/XTLS/Xray-core/releases/download",
		"https://ghproxy.com/https://github.com/XTLS/Xray-core/releases/download",
	}

	// 随机选择一个镜像站点
	if len(githubMirrors) > 0 {
		// 简单随机：使用时间戳的最后几位
		index := int(time.Now().UnixNano() % int64(len(githubMirrors)))
		mirrorURL := githubMirrors[index]
		return fmt.Sprintf("%s/%s/%s", mirrorURL, d.version, fileName)
	}

	return ""
}

// hasToolkit 检查是否有Node.js下载工具包
func (d *AutoDownloader) hasToolkit() bool {
	toolkitPath := filepath.Join("tools", "download_xray.js")
	if _, err := os.Stat(toolkitPath); err == nil {
		return true
	}
	return false
}

// runToolkit 运行Node.js下载工具包
func (d *AutoDownloader) runToolkit() error {
	toolkitPath := filepath.Join("tools", "download_xray.js")

	// 检查文件是否存在
	if _, err := os.Stat(toolkitPath); os.IsNotExist(err) {
		return errors.New("toolkit script not found")
	}

	// 检查node是否可用
	nodeCmd := "node"
	if runtime.GOOS == "windows" {
		// 检查是否是绝对路径的node
		nodePath := filepath.Join("tools", "node_modules", ".bin", "node.exe")
		if _, err := os.Stat(nodePath); err == nil {
			nodeCmd = nodePath
		}
	}

	// 运行脚本
	cmd := exec.Command(nodeCmd, toolkitPath, d.version)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("XRAY_VERSION=%s", d.version),
		fmt.Sprintf("OUTPUT_DIR=%s", d.outputPath),
	)

	// 设置超时
	done := make(chan error)
	go func() {
		output, err := cmd.CombinedOutput()
		if err != nil {
			done <- fmt.Errorf("toolkit error: %v, output: %s", err, string(output))
		} else {
			fmt.Printf("工具包输出: %s\n", string(output))
			done <- nil
		}
	}()

	// 等待完成或超时
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Minute):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return errors.New("toolkit execution timed out")
	}
}

// getPlatformInfo 获取系统和架构信息
func getPlatformInfo() (osName, osArch string) {
	osName = runtime.GOOS
	osArch = runtime.GOARCH

	// 标准化操作系统名称
	switch osName {
	case "darwin":
		osName = "macos"
	}

	// 标准化架构名称
	switch osArch {
	case "amd64":
		osArch = "x64"
	case "386":
		osArch = "x86"
	}

	return
}

// VerifyGithubConnectivity 验证GitHub连接性
func VerifyGithubConnectivity() error {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 忽略SSL证书验证
			// 增加连接超时
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 3 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			Proxy:                 http.ProxyFromEnvironment, // 支持系统代理
		},
	}

	// 尝试多个URL以验证连接性
	urls := []string{
		"https://github.com",
		"https://api.github.com",
		"https://raw.githubusercontent.com",
	}

	var lastErr error
	for _, url := range urls {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil // 任何一个成功就返回成功
		}
		lastErr = err
		fmt.Printf("无法连接到 %s: %v\n", url, err)
	}

	return fmt.Errorf("无法连接到GitHub或相关服务: %v", lastErr)
}
