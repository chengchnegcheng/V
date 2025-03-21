package model

import "gorm.io/gorm"

// Setting 系统设置
type Setting struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null" json:"key"` // 设置键名
	Value string `gorm:"type:text;not null" json:"value"` // 设置值
	Note  string `gorm:"type:text" json:"note,omitempty"` // 设置说明
}
