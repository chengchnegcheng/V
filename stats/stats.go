package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"v/logger"
	"v/notification"
	"v/settings"
)

// TrafficStats represents traffic statistics
type TrafficStats struct {
	UserID    int64     `json:"user_id"`
	Upload    int64     `json:"upload"`
	Download  int64     `json:"download"`
	Timestamp time.Time `json:"timestamp"`
}

// DailyStats represents daily traffic statistics
type DailyStats struct {
	UserID   int64     `json:"user_id"`
	Date     time.Time `json:"date"`
	Upload   int64     `json:"upload"`
	Download int64     `json:"download"`
	Total    int64     `json:"total"`
}

// Manager represents a statistics manager
type Manager struct {
	log       *logger.Logger
	settings  *settings.Manager
	notifier  *notification.Manager
	statsPath string
	stats     map[int64]*TrafficStats
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// New creates a new statistics manager
func New(log *logger.Logger, settings *settings.Manager, notifier *notification.Manager) *Manager {
	return &Manager{
		log:       log,
		settings:  settings,
		notifier:  notifier,
		statsPath: filepath.Join("stats"),
		stats:     make(map[int64]*TrafficStats),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the statistics manager
func (m *Manager) Start() error {
	s := m.settings.Get()

	// Create stats directory
	if err := os.MkdirAll(m.statsPath, 0755); err != nil {
		return fmt.Errorf("failed to create stats directory: %v", err)
	}

	// Load existing stats
	if err := m.loadStats(); err != nil {
		m.log.Error("Failed to load stats", logger.Fields{
			"error": err,
		})
	}

	// Start stats routine
	go m.statsRoutine()

	m.log.Info("Statistics manager started", logger.Fields{
		"stats_path": m.statsPath,
		"interval":   s.Traffic.StatsInterval,
	})

	return nil
}

// Stop stops the statistics manager
func (m *Manager) Stop() {
	close(m.stopChan)
}

// statsRoutine runs the statistics routine
func (m *Manager) statsRoutine() {
	s := m.settings.Get()
	ticker := time.NewTicker(s.Traffic.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			if err := m.saveStats(); err != nil {
				m.log.Error("Failed to save stats", logger.Fields{
					"error": err,
				})
			}

			if err := m.generateDailyStats(); err != nil {
				m.log.Error("Failed to generate daily stats", logger.Fields{
					"error": err,
				})
			}
		}
	}
}

// AddTraffic adds traffic statistics
func (m *Manager) AddTraffic(userID int64, upload, download int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create stats for user
	stats, exists := m.stats[userID]
	if !exists {
		stats = &TrafficStats{
			UserID:    userID,
			Timestamp: time.Now(),
		}
		m.stats[userID] = stats
	}

	// Update stats
	stats.Upload += upload
	stats.Download += download
	stats.Timestamp = time.Now()

	return nil
}

// GetTraffic returns traffic statistics for a user
func (m *Manager) GetTraffic(userID int64) (*TrafficStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, exists := m.stats[userID]
	if !exists {
		return nil, fmt.Errorf("no stats found for user: %d", userID)
	}

	return stats, nil
}

// GetDailyStats returns daily traffic statistics for a user
func (m *Manager) GetDailyStats(userID int64, start, end time.Time) ([]*DailyStats, error) {
	statsFile := filepath.Join(m.statsPath, fmt.Sprintf("daily_%d.json", userID))

	// Read daily stats file
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read daily stats file: %v", err)
	}

	var stats []*DailyStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal daily stats: %v", err)
	}

	// Filter stats by date range
	var filtered []*DailyStats
	for _, s := range stats {
		if s.Date.After(start) && s.Date.Before(end) {
			filtered = append(filtered, s)
		}
	}

	return filtered, nil
}

// loadStats loads existing statistics
func (m *Manager) loadStats() error {
	statsFile := filepath.Join(m.statsPath, "stats.json")

	// Read stats file
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %v", err)
	}

	// Unmarshal stats
	if err := json.Unmarshal(data, &m.stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %v", err)
	}

	return nil
}

// saveStats saves current statistics
func (m *Manager) saveStats() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statsFile := filepath.Join(m.statsPath, "stats.json")

	// Marshal stats
	data, err := json.MarshalIndent(m.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %v", err)
	}

	// Write stats file
	if err := os.WriteFile(statsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write stats file: %v", err)
	}

	return nil
}

// generateDailyStats generates daily traffic statistics
func (m *Manager) generateDailyStats() error {
	s := m.settings.Get()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Generate daily stats for each user
	for userID, stats := range m.stats {
		// Skip if stats are from today
		if stats.Timestamp.After(today) {
			continue
		}

		// Create daily stats
		dailyStats := &DailyStats{
			UserID:   userID,
			Date:     today,
			Upload:   stats.Upload,
			Download: stats.Download,
			Total:    stats.Upload + stats.Download,
		}

		// Save daily stats
		statsFile := filepath.Join(m.statsPath, fmt.Sprintf("daily_%d.json", userID))
		if err := m.saveDailyStats(statsFile, dailyStats); err != nil {
			m.log.Error("Failed to save daily stats", logger.Fields{
				"user_id": userID,
				"error":   err,
			})
			continue
		}

		// Check traffic limit
		if dailyStats.Total > s.Traffic.DefaultLimit {
			// Send warning notification
			if err := m.notifier.SendTrafficWarning(userID, "", dailyStats.Total, s.Traffic.DefaultLimit); err != nil {
				m.log.Error("Failed to send traffic warning", logger.Fields{
					"user_id": userID,
					"error":   err,
				})
			}
		}

		// Reset stats
		stats.Upload = 0
		stats.Download = 0
		stats.Timestamp = now
	}

	return nil
}

// saveDailyStats saves daily traffic statistics
func (m *Manager) saveDailyStats(statsFile string, stats *DailyStats) error {
	// Read existing daily stats
	var dailyStats []*DailyStats
	if data, err := os.ReadFile(statsFile); err == nil {
		if err := json.Unmarshal(data, &dailyStats); err != nil {
			return fmt.Errorf("failed to unmarshal daily stats: %v", err)
		}
	}

	// Add new daily stats
	dailyStats = append(dailyStats, stats)

	// Marshal daily stats
	data, err := json.MarshalIndent(dailyStats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal daily stats: %v", err)
	}

	// Write daily stats file
	if err := os.WriteFile(statsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write daily stats file: %v", err)
	}

	return nil
}
