package xray

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// AutoDownloader 提供改进的Xray下载功能
type AutoDownloader struct {
	VersionTag   string
	OS           string
	Arch         string
	DownloadPath string
	InstallPath  string
	mirrors      []string
	mutex        sync.Mutex
}

// NewAutoDownloader 创建新的自动下载器
func NewAutoDownloader(version string) *AutoDownloader {
	// 获取系统信息
	osName, arch := getPlatformInfo()

	// 构建下载和安装路径
	downloadDir := filepath.Join("xray", "downloads")
	installDir := filepath.Join("xray", "bin", version)

	// 确保目录存在
	os.MkdirAll(downloadDir, 0755)
	os.MkdirAll(installDir, 0755)

	return &AutoDownloader{
		VersionTag:   version,
		OS:           osName,
		Arch:         arch,
		DownloadPath: downloadDir,
		InstallPath:  installDir,
		mirrors: []string{
			"https://mirror.ghproxy.com/https://github.com/XTLS/Xray-core/releases/download",
			"https://github.mirrorz.org/XTLS/Xray-core/releases/download",
			"https://ghproxy.com/https://github.com/XTLS/Xray-core/releases/download",
			"https://gh-proxy.com/https://github.com/XTLS/Xray-core/releases/download",
			"https://github.com/XTLS/Xray-core/releases/download",
		},
	}
}

// DownloadAndInstall 下载并安装指定版本的Xray
func (d *AutoDownloader) DownloadAndInstall() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// 检查目标可执行文件是否已存在
	execName := "xray"
	if runtime.GOOS == "windows" {
		execName = "xray.exe"
	}

	execPath := filepath.Join(d.InstallPath, execName)
	if _, err := os.Stat(execPath); err == nil {
		fmt.Printf("Xray %s 已安装在 %s\n", d.VersionTag, execPath)
		return nil
	}

	// 构建下载文件名
	fileName := fmt.Sprintf("Xray-%s-%s.zip", d.OS, d.Arch)
	zipPath := filepath.Join(d.DownloadPath, fileName)

	// 尝试所有镜像下载
	var lastErr error
	success := false

	for i, mirror := range d.mirrors {
		url := fmt.Sprintf("%s/%s/%s", mirror, d.VersionTag, fileName)
		fmt.Printf("尝试从镜像 %d/%d 下载: %s\n", i+1, len(d.mirrors), url)

		err := d.downloadFile(url, zipPath)
		if err != nil {
			fmt.Printf("从镜像 %s 下载失败: %v\n", mirror, err)
			lastErr = err
			continue
		}

		// 验证下载文件
		if !d.verifyDownload(zipPath) {
			lastErr = fmt.Errorf("文件验证失败")
			fmt.Println("下载的文件无效，尝试下一个镜像")
			continue
		}

		success = true
		break
	}

	if !success {
		return fmt.Errorf("从所有镜像下载失败: %v", lastErr)
	}

	// 提取内容
	fmt.Printf("解压 %s 到 %s\n", zipPath, d.InstallPath)
	err := d.extractZip(zipPath, d.InstallPath)
	if err != nil {
		return fmt.Errorf("解压失败: %v", err)
	}

	// 在非Windows系统上设置可执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(execPath, 0755); err != nil {
			return fmt.Errorf("设置可执行权限失败: %v", err)
		}
	}

	fmt.Printf("Xray %s 安装成功!\n", d.VersionTag)
	return nil
}

// downloadFile 下载文件到指定路径
func (d *AutoDownloader) downloadFile(url, filePath string) error {
	// 创建目录（如果不存在）
	os.MkdirAll(filepath.Dir(filePath), 0755)

	// 使用系统命令（curl/wget）或HTTP客户端
	if runtime.GOOS != "windows" && d.commandExists("curl") {
		return d.downloadWithCurl(url, filePath)
	} else if runtime.GOOS != "windows" && d.commandExists("wget") {
		return d.downloadWithWget(url, filePath)
	} else {
		return d.downloadWithHTTP(url, filePath)
	}
}

// downloadWithCurl 使用curl下载
func (d *AutoDownloader) downloadWithCurl(url, filePath string) error {
	cmd := exec.Command("curl", "-L", "-o", filePath,
		"--connect-timeout", "30",
		"--retry", "3",
		"--retry-delay", "5",
		"-H", "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		url)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("curl 下载失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// downloadWithWget 使用wget下载
func (d *AutoDownloader) downloadWithWget(url, filePath string) error {
	cmd := exec.Command("wget", "-O", filePath,
		"--timeout=30",
		"--tries=3",
		"--user-agent=Mozilla/5.0",
		url)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wget 下载失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// downloadWithHTTP 使用HTTP客户端下载
func (d *AutoDownloader) downloadWithHTTP(url, filePath string) error {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("执行HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP响应错误: %s", resp.Status)
	}

	// 创建文件
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	// 复制内容
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// verifyDownload 验证下载的文件是否有效
func (d *AutoDownloader) verifyDownload(filePath string) bool {
	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("文件不存在: %v\n", err)
		return false
	}

	// 检查文件大小
	if info.Size() < 1000 { // 至少1KB
		fmt.Printf("文件太小 (%d 字节)\n", info.Size())
		return false
	}

	return true
}

// extractZip 解压ZIP文件
func (d *AutoDownloader) extractZip(zipPath, destDir string) error {
	// 根据系统选择解压方法
	if runtime.GOOS == "windows" {
		return d.extractZipWindows(zipPath, destDir)
	} else {
		return d.extractZipLinux(zipPath, destDir)
	}
}

// extractZipWindows 在Windows上解压
func (d *AutoDownloader) extractZipWindows(zipPath, destDir string) error {
	// 尝试使用PowerShell
	if d.commandExists("powershell") {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", zipPath, destDir))
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		fmt.Printf("PowerShell解压失败: %v, 输出: %s\n", err, string(output))
	}

	// 回退到内置解压
	return d.extractWithGo(zipPath, destDir)
}

// extractZipLinux 在Linux上解压
func (d *AutoDownloader) extractZipLinux(zipPath, destDir string) error {
	// 尝试使用unzip命令
	if d.commandExists("unzip") {
		cmd := exec.Command("unzip", "-o", zipPath, "-d", destDir)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		fmt.Printf("unzip命令失败: %v, 输出: %s\n", err, string(output))
	}

	// 回退到内置解压
	return d.extractWithGo(zipPath, destDir)
}

// extractWithGo 使用Go的zip包解压
func (d *AutoDownloader) extractWithGo(zipPath, destDir string) error {
	// 调用修改后的unzip函数
	return unzip(zipPath, destDir)
}

// commandExists 检查系统命令是否存在
func (d *AutoDownloader) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// hasToolkit 检查是否具有Node.js工具包
func (d *AutoDownloader) hasToolkit() bool {
	toolPath := filepath.Join("tools", "download_xray.js")
	_, err := os.Stat(toolPath)
	return err == nil
}

// runToolkit 运行Node.js工具包
func (d *AutoDownloader) runToolkit() error {
	// 检查Node.js是否存在
	if !d.commandExists("node") {
		return fmt.Errorf("未安装Node.js")
	}

	// 检查工具是否存在
	toolPath := filepath.Join("tools", "download_xray.js")
	if _, err := os.Stat(toolPath); err != nil {
		return fmt.Errorf("未找到工具: %v", err)
	}

	// 执行工具
	cmd := exec.Command("node", toolPath, d.VersionTag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("工具执行失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// GetAvailableVersions 获取可用的Xray版本
func (d *AutoDownloader) GetAvailableVersions() ([]string, error) {
	// 构建URL
	url := "https://mirror.ghproxy.com/https://api.github.com/repos/XTLS/Xray-core/releases"

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析JSON获取版本列表
	content := string(body)
	versions := []string{}

	// 解析版本号标签
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "tag_name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				version := strings.TrimSpace(parts[1])
				version = strings.Trim(version, "\",")
				if version != "" {
					versions = append(versions, version)
				}
			}
		}
	}

	// 如果未找到版本，使用硬编码的默认值
	if len(versions) == 0 {
		versions = SupportedVersions
	}

	return versions, nil
}

// getPlatformInfo 获取系统平台信息
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
