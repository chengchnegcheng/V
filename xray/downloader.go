package xray

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// AutoDownloader handles automatic downloading of Xray versions
type AutoDownloader struct {
	version     string
	downloadURL string
	mirrors     []string
}

// NewAutoDownloader creates a new AutoDownloader instance
func NewAutoDownloader(version string) *AutoDownloader {
	// Base GitHub release URL
	baseURL := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/%s", version)

	// Determine platform-specific file name
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map architecture names
	switch arch {
	case "amd64":
		arch = "64"
	case "386":
		arch = "32"
	case "arm":
		arch = "arm32-v7a"
	case "arm64":
		arch = "arm64-v8a"
	}

	// Map OS names
	var osNameInFile string
	switch osName {
	case "windows":
		osNameInFile = "windows"
	case "darwin":
		osNameInFile = "macos"
	default:
		osNameInFile = "linux"
	}

	// Determine file name based on version
	var fileName string
	if strings.HasPrefix(version, "v1.") {
		// Old version format (v1.x.x)
		fileName = fmt.Sprintf("Xray-%s-%s.zip", osNameInFile, arch)
	} else {
		// New version format (v24.x.x, v25.x.x)
		fileName = fmt.Sprintf("Xray-%s-%s.zip", arch, osNameInFile)
	}

	// Create download URL
	downloadURL := fmt.Sprintf("%s/%s", baseURL, fileName)

	// Define mirror sites
	mirrors := []string{
		downloadURL,
		fmt.Sprintf("https://download.fastgit.org/XTLS/Xray-core/releases/download/%s/%s", version, fileName),
		fmt.Sprintf("https://ghproxy.com/https://github.com/XTLS/Xray-core/releases/download/%s/%s", version, fileName),
	}

	return &AutoDownloader{
		version:     version,
		downloadURL: downloadURL,
		mirrors:     mirrors,
	}
}

// DownloadAndInstall downloads and installs the specified version
func (d *AutoDownloader) DownloadAndInstall() error {
	// Create temporary directory for downloads
	tempDir := filepath.Join("xray", "downloads")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Temporary zip file path
	zipFile := filepath.Join(tempDir, fmt.Sprintf("xray-%s.zip", d.version))

	// Try each mirror until successful
	var lastError error
	for _, mirror := range d.mirrors {
		fmt.Printf("Trying download from: %s\n", mirror)

		// Download the file
		if err := downloadFile(mirror, zipFile); err != nil {
			fmt.Printf("Download failed from %s: %v\n", mirror, err)
			lastError = err
			continue
		}

		// If download successful, break the loop
		lastError = nil
		break
	}

	// If all downloads failed, return the last error
	if lastError != nil {
		return fmt.Errorf("all download attempts failed: %v", lastError)
	}

	// Create version directory
	versionDir := filepath.Join("xray", "bin", d.version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %v", err)
	}

	// Extract the zip file
	if err := unzip(zipFile, versionDir); err != nil {
		return fmt.Errorf("failed to extract zip file: %v", err)
	}

	// Clean up the temporary zip file
	os.Remove(zipFile)

	return nil
}

// hasToolkit checks if Node.js toolkit is available
func (d *AutoDownloader) hasToolkit() bool {
	toolkitPath := filepath.Join("tools", "node_modules", ".bin", "extract-zip")
	_, err := os.Stat(toolkitPath)
	return err == nil
}

// runToolkit runs the Node.js toolkit for downloading
func (d *AutoDownloader) runToolkit() error {
	// Implementation for Node.js toolkit
	// This is a fallback method and can be implemented later if needed
	return fmt.Errorf("Node.js toolkit not implemented")
}
