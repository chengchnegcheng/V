package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"v/logger"
)

// Event represents an audit event
type Event struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

// Auditor represents an audit logger
type Auditor struct {
	log      *logger.Logger
	file     *os.File
	filePath string
}

// New creates a new auditor
func New(log *logger.Logger, filePath string) (*Auditor, error) {
	// Create audit directory if not exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %v", err)
	}

	// Open audit file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %v", err)
	}

	return &Auditor{
		log:      log,
		file:     file,
		filePath: filePath,
	}, nil
}

// Close closes the auditor
func (a *Auditor) Close() error {
	if a.file != nil {
		return a.file.Close()
	}
	return nil
}

// Log logs an audit event
func (a *Auditor) Log(event *Event) error {
	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	// Write to file
	if _, err := a.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %v", err)
	}

	// Log to logger
	a.log.Info("Audit event", logger.Fields{
		"user_id":    event.UserID,
		"action":     event.Action,
		"resource":   event.Resource,
		"details":    event.Details,
		"ip":         event.IP,
		"user_agent": event.UserAgent,
	})

	return nil
}

// Query queries audit events
func (a *Auditor) Query(userID int64, startTime, endTime time.Time) ([]*Event, error) {
	// Read all events
	events, err := a.readEvents()
	if err != nil {
		return nil, err
	}

	// Filter events
	var filtered []*Event
	for _, event := range events {
		if userID != 0 && event.UserID != userID {
			continue
		}
		if !startTime.IsZero() && event.Timestamp.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && event.Timestamp.After(endTime) {
			continue
		}
		filtered = append(filtered, event)
	}

	return filtered, nil
}

// readEvents reads all events from the audit file
func (a *Auditor) readEvents() ([]*Event, error) {
	// Read file
	data, err := os.ReadFile(a.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read audit file: %v", err)
	}

	// Parse events
	var events []*Event
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %v", err)
		}
		events = append(events, &event)
	}

	return events, nil
}

// Common audit actions
const (
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionLogin   = "login"
	ActionLogout  = "logout"
	ActionEnable  = "enable"
	ActionDisable = "disable"
	ActionRenew   = "renew"
	ActionReset   = "reset"
)

// Common audit resources
const (
	ResourceUser        = "user"
	ResourceProxy       = "proxy"
	ResourceCertificate = "certificate"
	ResourceSystem      = "system"
	ResourceTraffic     = "traffic"
)
