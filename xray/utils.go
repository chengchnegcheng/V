package xray

import (
	"archive/zip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// progressReader 用于报告下载进度
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

// writeCounter counts bytes written and reports progress
type writeCounter struct {
	total      int64
	written    int64
	onProgress func(written, total int64)
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.written += int64(n)
	if wc.onProgress != nil {
		wc.onProgress(wc.written, wc.total)
	}
	return n, nil
}

// downloadFile downloads a file from the given URL to the specified local path
func downloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer out.Close()

	// Create custom HTTP client with timeout and TLS config
	client := &http.Client{
		Timeout: 300 * time.Second, // 5 minutes timeout
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			DisableKeepAlives: true,
		},
	}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create a progress tracker
	lastProgress := 0

	// Writer that tracks progress
	writer := io.MultiWriter(out, &writeCounter{
		total: resp.ContentLength,
		onProgress: func(written int64, total int64) {
			if total > 0 {
				progress := int(float64(written) / float64(total) * 100)
				if progress > lastProgress {
					fmt.Printf("\rDownloading... %d%%", progress)
					lastProgress = progress
				}
			}
		},
	})

	// Write the body to file
	_, err = io.Copy(writer, resp.Body)
	fmt.Println() // New line after progress
	return err
}

// unzip extracts a zip file to the specified destination
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// getFileSize returns a human-readable file size
func getFileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "unknown size"
	}

	size := float64(info.Size())
	units := []string{"B", "KB", "MB", "GB"}
	unit := 0

	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}

	return fmt.Sprintf("%.2f %s", size, units[unit])
}
