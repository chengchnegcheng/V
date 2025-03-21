package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RotationConfig 日志轮转配置
type RotationConfig struct {
	// MaxSize 单个日志文件最大尺寸（MB）
	MaxSize int `json:"max_size"`
	// MaxAge 日志文件最大保留时间（天）
	MaxAge int `json:"max_age"`
	// MaxBackups 最大备份文件数
	MaxBackups int `json:"max_backups"`
	// LocalTime 使用本地时间
	LocalTime bool `json:"local_time"`
	// Compress 是否压缩
	Compress bool `json:"compress"`
}

// RotateWriter 旋转日志写入器
type RotateWriter struct {
	filename   string
	config     RotationConfig
	size       int64
	file       *os.File
	mu         sync.Mutex
	startTime  time.Time
	lastRotate time.Time
}

// NewRotateWriter 创建旋转日志写入器
func NewRotateWriter(filename string, config RotationConfig) (*RotateWriter, error) {
	// 创建日志目录
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create directory: %v", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// 创建写入器
	w := &RotateWriter{
		filename:   filename,
		config:     config,
		size:       info.Size(),
		file:       file,
		startTime:  time.Now(),
		lastRotate: time.Now(),
	}

	// 清理旧日志
	if err := w.cleanup(); err != nil {
		return nil, err
	}

	return w, nil
}

// Write 实现io.Writer接口
func (w *RotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 写入数据
	n, err = w.file.Write(p)
	if err != nil {
		return n, err
	}

	// 更新大小
	w.size += int64(n)

	// 检查是否需要轮转
	if w.shouldRotate() {
		if err := w.rotate(); err != nil {
			return n, err
		}
	}

	return n, nil
}

// Close 关闭日志文件
func (w *RotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.file.Close()
}

// shouldRotate 检查是否需要轮转
func (w *RotateWriter) shouldRotate() bool {
	// 检查文件大小
	if w.config.MaxSize > 0 && w.size > int64(w.config.MaxSize*1024*1024) {
		return true
	}

	// 检查时间
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	if now.After(midnight) && w.lastRotate.Before(midnight) {
		return true
	}

	return false
}

// rotate 轮转日志
func (w *RotateWriter) rotate() error {
	// 关闭当前文件
	if err := w.file.Close(); err != nil {
		return err
	}

	// 生成新文件名
	now := time.Now()
	newFilename := w.backupName(now)

	// 重命名文件
	if err := os.Rename(w.filename, newFilename); err != nil {
		return err
	}

	// 打开新文件
	file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// 更新状态
	w.file = file
	w.size = 0
	w.lastRotate = now

	// 清理旧日志
	return w.cleanup()
}

// backupName 生成备份文件名
func (w *RotateWriter) backupName(t time.Time) string {
	dir := filepath.Dir(w.filename)
	filename := filepath.Base(w.filename)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	timestamp := t.Format("2006-01-02-150405")

	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

// cleanup 清理旧日志
func (w *RotateWriter) cleanup() error {
	if w.config.MaxBackups == 0 && w.config.MaxAge == 0 {
		return nil
	}

	// 获取所有日志文件
	pattern := w.filename + ".*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// 按修改时间排序
	type backupFile struct {
		path    string
		modTime time.Time
	}
	files := make([]backupFile, 0, len(matches))
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		files = append(files, backupFile{path, info.ModTime()})
	}

	// 按时间降序排序
	sort := func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	}
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if sort(i, j) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// 删除超过最大备份数的文件
	if w.config.MaxBackups > 0 && len(files) > w.config.MaxBackups {
		for i := w.config.MaxBackups; i < len(files); i++ {
			os.Remove(files[i].path)
		}
		files = files[:w.config.MaxBackups]
	}

	// 删除过期文件
	if w.config.MaxAge > 0 {
		cutoff := time.Now().Add(-time.Duration(w.config.MaxAge) * 24 * time.Hour)
		for _, file := range files {
			if file.modTime.Before(cutoff) {
				os.Remove(file.path)
			}
		}
	}

	return nil
}

// MultiWriter 多写入器
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter 创建多写入器
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write 实现io.Writer接口
func (w *MultiWriter) Write(p []byte) (n int, err error) {
	for _, writer := range w.writers {
		n, err := writer.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}
