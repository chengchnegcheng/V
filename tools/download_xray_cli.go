package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"v/xray"
)

func main() {
	// 解析命令行参数
	version := flag.String("version", "v1.8.24", "要下载的Xray版本")
	listVersions := flag.Bool("list", false, "列出所有可用版本")
	flag.Parse()

	// 创建自动下载器
	downloader := xray.NewAutoDownloader(*version)

	// 如果请求列出所有版本
	if *listVersions {
		fmt.Println("获取可用Xray版本列表...")
		versions, err := downloader.GetAvailableVersions()
		if err != nil {
			fmt.Printf("获取版本列表失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("可用版本:")
		for i, ver := range versions {
			fmt.Printf("%3d. %s\n", i+1, ver)
		}
		return
	}

	// 下载Xray
	fmt.Printf("开始下载 Xray %s...\n", *version)
	fmt.Printf("系统平台: %s-%s\n", downloader.OS, downloader.Arch)
	fmt.Printf("下载目录: %s\n", downloader.DownloadPath)
	fmt.Printf("安装目录: %s\n", downloader.InstallPath)

	err := downloader.DownloadAndInstall()
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		os.Exit(1)
	}

	// 检查安装结果
	execName := "xray"
	if runtime.GOOS == "windows" {
		execName = "xray.exe"
	}

	execPath := filepath.Join(downloader.InstallPath, execName)
	_, err = os.Stat(execPath)
	if err != nil {
		fmt.Printf("安装验证失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Xray %s 已成功安装到 %s\n", *version, execPath)
}

// 同时下载多个版本的示例代码
func downloadMultipleVersions(versions []string) {
	var wg sync.WaitGroup
	for _, version := range versions {
		wg.Add(1)
		go func(ver string) {
			defer wg.Done()

			downloader := xray.NewAutoDownloader(ver)
			fmt.Printf("开始下载 Xray %s...\n", ver)

			err := downloader.DownloadAndInstall()
			if err != nil {
				fmt.Printf("下载 %s 失败: %v\n", ver, err)
				return
			}

			fmt.Printf("Xray %s 下载完成!\n", ver)
		}(version)
	}

	wg.Wait()
	fmt.Println("所有版本下载完成!")
}
