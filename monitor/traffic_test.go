package monitor

import (
	"testing"
	"time"
	"v/model"
)

type mockTrafficDB struct {
	model.DB
	trafficStats map[int64]*model.TrafficStats
	dailyStats   []*model.DailyStats
}

func (db *mockTrafficDB) DeleteProxy(id int64) error {
	return nil
}

func (db *mockTrafficDB) GetTrafficStats(userID int64) (*model.TrafficStats, error) {
	if stats, ok := db.trafficStats[userID]; ok {
		return stats, nil
	}
	return nil, nil
}

func (db *mockTrafficDB) UpdateTrafficStats(stats *model.TrafficStats) error {
	db.trafficStats[stats.UserID] = stats
	return nil
}

func (db *mockTrafficDB) CreateDailyStats(stats *model.DailyStats) error {
	db.dailyStats = append(db.dailyStats, stats)
	return nil
}

func (db *mockTrafficDB) GetDailyStats(userID int64, start, end time.Time) ([]*model.DailyStats, error) {
	var result []*model.DailyStats
	for _, stats := range db.dailyStats {
		if stats.UserID == userID && stats.Date.After(start) && stats.Date.Before(end) {
			result = append(result, stats)
		}
	}
	return result, nil
}

func TestTrafficMonitor_UpdateTraffic(t *testing.T) {
	db := &mockTrafficDB{
		trafficStats: make(map[int64]*model.TrafficStats),
		dailyStats:   make([]*model.DailyStats, 0),
	}
	monitor := NewTrafficMonitor(db)

	// Test updating traffic for a new user
	err := monitor.UpdateTraffic(1, 1000, 2000)
	if err != nil {
		t.Fatalf("UpdateTraffic failed: %v", err)
	}

	stats, err := monitor.GetTrafficStats(1)
	if err != nil {
		t.Fatalf("GetTrafficStats failed: %v", err)
	}
	if stats.Upload != 1000 {
		t.Errorf("Expected upload to be 1000, got %d", stats.Upload)
	}
	if stats.Download != 2000 {
		t.Errorf("Expected download to be 2000, got %d", stats.Download)
	}

	// Test updating traffic for an existing user
	err = monitor.UpdateTraffic(1, 500, 1000)
	if err != nil {
		t.Fatalf("UpdateTraffic failed: %v", err)
	}

	stats, err = monitor.GetTrafficStats(1)
	if err != nil {
		t.Fatalf("GetTrafficStats failed: %v", err)
	}
	if stats.Upload != 1500 {
		t.Errorf("Expected upload to be 1500, got %d", stats.Upload)
	}
	if stats.Download != 3000 {
		t.Errorf("Expected download to be 3000, got %d", stats.Download)
	}

	// Check daily stats
	if len(db.dailyStats) != 2 {
		t.Errorf("Expected 2 daily stats, got %d", len(db.dailyStats))
	}
}

func TestTrafficMonitor_GetDailyStats(t *testing.T) {
	db := &mockTrafficDB{
		trafficStats: make(map[int64]*model.TrafficStats),
		dailyStats:   make([]*model.DailyStats, 0),
	}
	monitor := NewTrafficMonitor(db)

	// Add some daily stats
	now := time.Now()
	db.dailyStats = append(db.dailyStats, &model.DailyStats{
		UserID:   1,
		Date:     now.Add(-2 * time.Hour),
		Upload:   1000,
		Download: 2000,
		Total:    3000,
	})
	db.dailyStats = append(db.dailyStats, &model.DailyStats{
		UserID:   1,
		Date:     now.Add(-1 * time.Hour),
		Upload:   500,
		Download: 1000,
		Total:    1500,
	})

	// Test getting daily stats
	start := now.Add(-3 * time.Hour)
	end := now
	stats, err := monitor.GetDailyStats(1, start, end)
	if err != nil {
		t.Fatalf("GetDailyStats failed: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected 2 daily stats, got %d", len(stats))
	}
}

func TestTrafficMonitor_ResetTraffic(t *testing.T) {
	db := &mockTrafficDB{
		trafficStats: make(map[int64]*model.TrafficStats),
		dailyStats:   make([]*model.DailyStats, 0),
	}
	monitor := NewTrafficMonitor(db)

	// Add some traffic
	err := monitor.UpdateTraffic(1, 1000, 2000)
	if err != nil {
		t.Fatalf("UpdateTraffic failed: %v", err)
	}

	// Reset traffic
	err = monitor.ResetTraffic(1)
	if err != nil {
		t.Fatalf("ResetTraffic failed: %v", err)
	}

	// Check if traffic is reset
	stats, err := monitor.GetTrafficStats(1)
	if err != nil {
		t.Fatalf("GetTrafficStats failed: %v", err)
	}
	if stats.Upload != 0 {
		t.Errorf("Expected upload to be 0, got %d", stats.Upload)
	}
	if stats.Download != 0 {
		t.Errorf("Expected download to be 0, got %d", stats.Download)
	}
}
