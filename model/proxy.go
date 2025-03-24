package model

import (
	"time"
)

// ProxyGormModel represents a proxy configuration in the database
type ProxyGormModel struct {
	ID           int64     `gorm:"primaryKey"`
	UserID       int64     `gorm:"index"`
	Type         string    `gorm:"size:32;not null"`
	Port         int       `gorm:"not null"`
	Settings     string    `gorm:"type:text;not null"`
	Enabled      bool      `gorm:"not null;default:true"`
	Upload       int64     `gorm:"not null;default:0"`
	Download     int64     `gorm:"not null;default:0"`
	LastActiveAt time.Time `gorm:"index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
