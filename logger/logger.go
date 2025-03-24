package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LogLevel 日志级别
type LogLevel int

const (
	// DEBUG 调试级别
	DEBUG LogLevel = iota
	// INFO 信息级别
	INFO
	// WARN 警告级别
	WARN
	// ERROR 错误级别
	ERROR
	// FATAL 致命错误级别
	FATAL
)

// String 返回日志级别字符串
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Configuration 日志配置
type Configuration struct {
	// Level 日志级别
	Level LogLevel `json:"level"`
	// Console 是否输出到控制台
	Console bool `json:"console"`
	// File 是否输出到文件
	File bool `json:"file"`
	// FilePath 日志文件路径
	FilePath string `json:"file_path"`
	// Rotation 日志轮转配置
	Rotation RotationConfig `json:"rotation"`
}

// Fields represents log fields
type Fields map[string]interface{}

// Logger represents a logger instance
type Logger struct {
	logger     *log.Logger
	level      LogLevel
	config     Configuration
	writer     io.Writer
	fileWriter *RotateWriter
}

// NewLogger creates a new logger instance with default configuration
func NewLogger() *Logger {
	return NewLoggerWithConfig(Configuration{
		Level:    INFO,
		Console:  true,
		File:     true,
		FilePath: filepath.Join("logs", "app.log"),
		Rotation: RotationConfig{
			MaxSize:    50,
			MaxAge:     7,
			MaxBackups: 10,
			LocalTime:  true,
			Compress:   true,
		},
	})
}

// New is an alias for NewLogger
func New() *Logger {
	return NewLogger()
}

// NewLoggerWithConfig creates a new logger instance with specified configuration
func NewLoggerWithConfig(config Configuration) *Logger {
	// 创建日志目录
	if config.File && config.FilePath != "" {
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create log directory: %v", err)
		}
	}

	var writer io.Writer
	var fileWriter *RotateWriter

	// 创建日志写入器
	writers := []io.Writer{}

	// 控制台输出
	if config.Console {
		writers = append(writers, os.Stdout)
	}

	// 文件输出
	if config.File && config.FilePath != "" {
		var err error
		fileWriter, err = NewRotateWriter(config.FilePath, config.Rotation)
		if err != nil {
			log.Printf("Failed to create log file: %v", err)
		} else {
			writers = append(writers, fileWriter)
		}
	}

	// 创建多写入器
	if len(writers) > 0 {
		writer = NewMultiWriter(writers...)
	} else {
		writer = os.Stdout
	}

	return &Logger{
		logger:     log.New(writer, "", log.LstdFlags),
		level:      config.Level,
		config:     config,
		writer:     writer,
		fileWriter: fileWriter,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// log 记录日志（内部方法）
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// 获取调用堆栈
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	// 获取函数名
	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn == nil {
		funcName = "???"
	} else {
		funcName = fn.Name()
		// 只保留函数名，不包含包名
		if idx := strings.LastIndex(funcName, "."); idx >= 0 {
			funcName = funcName[idx+1:]
		}
	}
	// 只保留文件名，不包含路径
	filename := filepath.Base(file)

	// 格式化消息
	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	// 记录日志
	l.logger.Printf("[%s] %s:%d %s() %s", level.String(), filename, line, funcName, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// WithFields logs a message with fields
func (l *Logger) WithFields(message string, fields Fields) {
	// 格式化字段
	var fieldString string
	if len(fields) > 0 {
		fieldParts := make([]string, 0, len(fields))
		for k, v := range fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldString = " " + strings.Join(fieldParts, " ")
	}

	// 记录日志
	l.Info("%s%s", message, fieldString)
}

// DebugWithFields logs a debug message with fields
func (l *Logger) DebugWithFields(message string, fields Fields) {
	// 格式化字段
	var fieldString string
	if len(fields) > 0 {
		fieldParts := make([]string, 0, len(fields))
		for k, v := range fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldString = " " + strings.Join(fieldParts, " ")
	}

	// 记录日志
	l.Debug("%s%s", message, fieldString)
}

// ErrorWithFields logs an error message with fields
func (l *Logger) ErrorWithFields(message string, fields Fields) {
	// 格式化字段
	var fieldString string
	if len(fields) > 0 {
		fieldParts := make([]string, 0, len(fields))
		for k, v := range fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldString = " " + strings.Join(fieldParts, " ")
	}

	// 记录日志
	l.Error("%s%s", message, fieldString)
}

// WarnWithFields logs a warning message with fields
func (l *Logger) WarnWithFields(message string, fields Fields) {
	// 格式化字段
	var fieldString string
	if len(fields) > 0 {
		fieldParts := make([]string, 0, len(fields))
		for k, v := range fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldString = " " + strings.Join(fieldParts, " ")
	}

	// 记录日志
	l.Warn("%s%s", message, fieldString)
}

// Close closes the logger
func (l *Logger) Close() error {
	if l.fileWriter != nil {
		return l.fileWriter.Close()
	}
	return nil
}

// Start is a no-op implementation for interface compatibility
func (l *Logger) Start() error {
	return nil
}

// Stop closes the logger resources
func (l *Logger) Stop() error {
	return l.Close()
}
