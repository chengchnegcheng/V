package proxy

import (
	"fmt"
	"sync"
	"v/database"
)

// TrafficLimiter manages traffic limits for proxy servers
type TrafficLimiter struct {
	sync.RWMutex
	limits map[int64]int64
}

// NewTrafficLimiter creates a new traffic limiter
func NewTrafficLimiter() *TrafficLimiter {
	return &TrafficLimiter{
		limits: make(map[int64]int64),
	}
}

// CheckTrafficLimit checks if the user has exceeded their traffic limit
func (l *TrafficLimiter) CheckTrafficLimit(userID int64) (bool, error) {
	var trafficLimit, usedTraffic int64
	err := database.DB.QueryRow(`
		SELECT traffic_limit, used_traffic 
		FROM users 
		WHERE id = ?
	`, userID).Scan(&trafficLimit, &usedTraffic)
	if err != nil {
		return false, fmt.Errorf("failed to get user traffic info: %v", err)
	}

	// If traffic limit is 0, it means unlimited
	if trafficLimit == 0 {
		return true, nil
	}

	return usedTraffic < trafficLimit, nil
}

// GetUserTrafficInfo returns the traffic limit and usage for a user
func (l *TrafficLimiter) GetUserTrafficInfo(userID int64) (int64, int64, error) {
	var trafficLimit, usedTraffic int64
	err := database.DB.QueryRow(`
		SELECT traffic_limit, used_traffic 
		FROM users 
		WHERE id = ?
	`, userID).Scan(&trafficLimit, &usedTraffic)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user traffic info: %v", err)
	}

	return trafficLimit, usedTraffic, nil
}

// UpdateTrafficUsage updates the traffic usage for a user
func (l *TrafficLimiter) UpdateTrafficUsage(userID int64, upload, download int64) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Get current traffic limit and usage
	trafficLimit, usedTraffic, err := l.GetUserTrafficInfo(userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Check if traffic limit is exceeded
	newTraffic := usedTraffic + upload + download
	if trafficLimit > 0 && newTraffic > trafficLimit {
		tx.Rollback()
		return fmt.Errorf("traffic limit exceeded")
	}

	// Update user's traffic usage
	_, err = tx.Exec(`
		UPDATE users 
		SET used_traffic = used_traffic + ?
		WHERE id = ?
	`, upload+download, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user traffic: %v", err)
	}

	// Insert traffic log
	_, err = tx.Exec(`
		INSERT INTO traffic_logs (user_id, upload, download, timestamp)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, userID, upload, download)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert traffic log: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// ResetTrafficUsage resets the traffic usage for a user
func (l *TrafficLimiter) ResetTrafficUsage(userID int64) error {
	_, err := database.DB.Exec(`
		UPDATE users 
		SET used_traffic = 0 
		WHERE id = ?
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to reset user traffic: %v", err)
	}

	return nil
}

// SetTrafficLimit sets the traffic limit for a user
func (l *TrafficLimiter) SetTrafficLimit(userID int64, limit int64) error {
	_, err := database.DB.Exec(`
		UPDATE users 
		SET traffic_limit = ? 
		WHERE id = ?
	`, limit, userID)
	if err != nil {
		return fmt.Errorf("failed to set user traffic limit: %v", err)
	}

	l.Lock()
	l.limits[userID] = limit
	l.Unlock()

	return nil
}

// GetTrafficLimit returns the traffic limit for a user
func (l *TrafficLimiter) GetTrafficLimit(userID int64) (int64, error) {
	l.RLock()
	limit, ok := l.limits[userID]
	l.RUnlock()

	if ok {
		return limit, nil
	}

	var trafficLimit int64
	err := database.DB.QueryRow(`
		SELECT traffic_limit 
		FROM users 
		WHERE id = ?
	`, userID).Scan(&trafficLimit)
	if err != nil {
		return 0, fmt.Errorf("failed to get user traffic limit: %v", err)
	}

	l.Lock()
	l.limits[userID] = trafficLimit
	l.Unlock()

	return trafficLimit, nil
}
