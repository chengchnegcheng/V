package logger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// LogAnalyzer 日志分析器
type LogAnalyzer struct {
	logsDir string
}

// NewLogAnalyzer 创建日志分析器
func NewLogAnalyzer(logsDir string) *LogAnalyzer {
	return &LogAnalyzer{
		logsDir: logsDir,
	}
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Function  string    `json:"function"`
	Message   string    `json:"message"`
	RawText   string    `json:"raw_text"`
}

// LogQuery 日志查询条件
type LogQuery struct {
	StartTime time.Time
	EndTime   time.Time
	Level     string
	Message   string
	File      string
	Function  string
	Limit     int
	Offset    int
}

// SearchLogs 搜索日志
func (a *LogAnalyzer) SearchLogs(query *LogQuery) ([]*LogEntry, error) {
	// 获取日志文件列表
	files, err := a.getLogFiles(query.StartTime, query.EndTime)
	if err != nil {
		return nil, err
	}

	// 解析日志文件
	var entries []*LogEntry
	for _, file := range files {
		fileEntries, err := a.parseLogFile(file, query)
		if err != nil {
			continue
		}
		entries = append(entries, fileEntries...)
	}

	// 按时间排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// 应用分页
	if query.Limit > 0 {
		end := query.Offset + query.Limit
		if end > len(entries) {
			end = len(entries)
		}
		if query.Offset < len(entries) {
			entries = entries[query.Offset:end]
		} else {
			entries = []*LogEntry{}
		}
	}

	return entries, nil
}

// getLogFiles 获取日志文件列表
func (a *LogAnalyzer) getLogFiles(startTime, endTime time.Time) ([]string, error) {
	// 获取文件列表
	files, err := filepath.Glob(filepath.Join(a.logsDir, "*.log*"))
	if err != nil {
		return nil, err
	}

	// 如果没有指定时间范围，返回所有文件
	if startTime.IsZero() && endTime.IsZero() {
		return files, nil
	}

	// 过滤文件
	var result []string
	for _, file := range files {
		// 从文件名中提取日期
		base := filepath.Base(file)
		dateStr := strings.TrimSuffix(base, filepath.Ext(base))
		if strings.Contains(dateStr, "-") {
			// 处理带有时间戳的轮转文件
			parts := strings.Split(dateStr, "-")
			if len(parts) >= 3 {
				dateStr = parts[0]
			}
		}

		// 解析日期
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// 如果无法解析日期，假定文件可能包含所需日志
			result = append(result, file)
			continue
		}

		// 检查日期范围
		if !startTime.IsZero() && fileDate.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && fileDate.After(endTime) {
			continue
		}

		result = append(result, file)
	}

	return result, nil
}

// parseLogFile 解析日志文件
func (a *LogAnalyzer) parseLogFile(filePath string, query *LogQuery) ([]*LogEntry, error) {
	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 定义日志行的正则表达式
	// 格式: [时间] [级别] 文件:行 函数() 消息
	logRegex := regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \[([A-Z]+)\] ([^:]+):(\d+) ([^(]+)\(\) (.+)$`)

	// 解析日志行
	var entries []*LogEntry
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// 匹配日志格式
		matches := logRegex.FindStringSubmatch(line)
		if len(matches) < 7 {
			continue
		}

		// 解析时间
		timestamp, err := time.Parse("2006/01/02 15:04:05", matches[1])
		if err != nil {
			continue
		}

		// 应用时间过滤
		if !query.StartTime.IsZero() && timestamp.Before(query.StartTime) {
			continue
		}
		if !query.EndTime.IsZero() && timestamp.After(query.EndTime) {
			continue
		}

		// 解析其他字段
		level := matches[2]
		file := matches[3]
		lineNum := 0
		fmt.Sscanf(matches[4], "%d", &lineNum)
		function := matches[5]
		message := matches[6]

		// 应用其他过滤条件
		if query.Level != "" && !strings.EqualFold(level, query.Level) {
			continue
		}
		if query.File != "" && !strings.Contains(file, query.File) {
			continue
		}
		if query.Function != "" && !strings.Contains(function, query.Function) {
			continue
		}
		if query.Message != "" && !strings.Contains(message, query.Message) {
			continue
		}

		// 创建日志条目
		entry := &LogEntry{
			Timestamp: timestamp,
			Level:     level,
			File:      file,
			Line:      lineNum,
			Function:  function,
			Message:   message,
			RawText:   line,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// GetLogStats 获取日志统计信息
func (a *LogAnalyzer) GetLogStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	// 查询日志
	query := &LogQuery{
		StartTime: startTime,
		EndTime:   endTime,
	}
	logs, err := a.SearchLogs(query)
	if err != nil {
		return nil, err
	}

	// 统计不同级别的日志数量
	levelCounts := make(map[string]int)
	fileCounts := make(map[string]int)
	functionCounts := make(map[string]int)
	hourlyDistribution := make(map[int]int)

	// 遍历日志进行统计
	for _, log := range logs {
		// 级别统计
		levelCounts[log.Level]++

		// 文件统计
		fileCounts[log.File]++

		// 函数统计
		functionCounts[log.Function]++

		// 时间分布统计
		hour := log.Timestamp.Hour()
		hourlyDistribution[hour]++
	}

	// 排序文件和函数，只保留前10个
	type CountItem struct {
		Name  string
		Count int
	}

	// 文件统计排序
	fileStats := make([]CountItem, 0, len(fileCounts))
	for file, count := range fileCounts {
		fileStats = append(fileStats, CountItem{file, count})
	}
	sort.Slice(fileStats, func(i, j int) bool {
		return fileStats[i].Count > fileStats[j].Count
	})
	if len(fileStats) > 10 {
		fileStats = fileStats[:10]
	}

	// 函数统计排序
	functionStats := make([]CountItem, 0, len(functionCounts))
	for function, count := range functionCounts {
		functionStats = append(functionStats, CountItem{function, count})
	}
	sort.Slice(functionStats, func(i, j int) bool {
		return functionStats[i].Count > functionStats[j].Count
	})
	if len(functionStats) > 10 {
		functionStats = functionStats[:10]
	}

	// 每小时分布
	hourlyStats := make([]CountItem, 0, 24)
	for hour := 0; hour < 24; hour++ {
		hourlyStats = append(hourlyStats, CountItem{fmt.Sprintf("%02d:00", hour), hourlyDistribution[hour]})
	}

	// 构建返回结果
	result := map[string]interface{}{
		"total_logs":          len(logs),
		"level_distribution":  levelCounts,
		"top_files":           fileStats,
		"top_functions":       functionStats,
		"hourly_distribution": hourlyStats,
		"start_time":          startTime,
		"end_time":            endTime,
	}

	return result, nil
}

// GetErrorLogs 获取错误日志
func (a *LogAnalyzer) GetErrorLogs(days int, limit int) ([]*LogEntry, error) {
	// 计算起始时间
	startTime := time.Now().AddDate(0, 0, -days)

	// 查询错误和致命错误日志
	query := &LogQuery{
		StartTime: startTime,
		Level:     "ERROR",
		Limit:     limit,
	}
	errorLogs, err := a.SearchLogs(query)
	if err != nil {
		return nil, err
	}

	query.Level = "FATAL"
	fatalLogs, err := a.SearchLogs(query)
	if err != nil {
		return nil, err
	}

	// 合并结果
	logs := append(errorLogs, fatalLogs...)

	// 按时间排序
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	// 应用限制
	if limit > 0 && len(logs) > limit {
		logs = logs[:limit]
	}

	return logs, nil
}

// TruncateLogs 清理指定日期之前的日志
func (a *LogAnalyzer) TruncateLogs(before time.Time) error {
	// 获取日志文件列表
	files, err := filepath.Glob(filepath.Join(a.logsDir, "*.log*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		// 获取文件信息
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		// 检查修改时间
		if info.ModTime().Before(before) {
			os.Remove(file)
		}
	}

	return nil
}
