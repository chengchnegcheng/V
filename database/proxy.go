package database

import (
	"time"
	"v/model"
)

// GetProxy 获取代理
func (db *DB) GetProxy(id uint) (*model.Proxy, error) {
	var proxy model.Proxy
	if err := db.First(&proxy, id).Error; err != nil {
		return nil, err
	}
	return &proxy, nil
}

// GetProxiesByUser 获取用户的所有代理
func (db *DB) GetProxiesByUser(userID uint) ([]*model.Proxy, error) {
	var proxies []*model.Proxy
	if err := db.Where("user_id = ?", userID).Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// CreateProxy 创建代理
func (db *DB) CreateProxy(proxy *model.Proxy) error {
	return db.Create(proxy).Error
}

// UpdateProxy 更新代理
func (db *DB) UpdateProxy(proxy *model.Proxy) error {
	return db.Save(proxy).Error
}

// DeleteProxy 删除代理
func (db *DB) DeleteProxy(id uint) error {
	return db.Delete(&model.Proxy{}, id).Error
}

// ListProxies 获取代理列表
func (db *DB) ListProxies(offset, limit int) ([]*model.Proxy, error) {
	var proxies []*model.Proxy
	if err := db.Offset(offset).Limit(limit).Find(&proxies).Error; err != nil {
		return nil, err
	}
	return proxies, nil
}

// UpdateTraffic 更新流量统计
func (db *DB) UpdateTraffic(id uint, upload, download int64) error {
	return db.Model(&model.Proxy{}).Where("id = ?", id).Updates(map[string]interface{}{
		"upload":   upload,
		"download": download,
	}).Error
}

// Enable 启用代理
func (db *DB) Enable(id uint) error {
	return db.Model(&model.Proxy{}).Where("id = ?", id).Update("enabled", true).Error
}

// Disable 禁用代理
func (db *DB) Disable(id uint) error {
	return db.Model(&model.Proxy{}).Where("id = ?", id).Update("enabled", false).Error
}

// UpdateLastActive 更新最后活动时间
func (db *DB) UpdateLastActive(id uint) error {
	return db.Model(&model.Proxy{}).Where("id = ?", id).Update("last_active_at", time.Now()).Error
}
