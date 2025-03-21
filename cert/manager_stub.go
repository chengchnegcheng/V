package cert

import (
	"time"
	"v/model"
)

// Manager 证书管理器
type Manager struct {
	db model.DB
}

// NewManager 创建一个新的证书管理器
func NewManager(db model.DB) *Manager {
	return &Manager{
		db: db,
	}
}

// ListCertificates 列出所有证书
func (m *Manager) ListCertificates() ([]*model.Certificate, error) {
	return m.db.ListCertificates()
}

// GetCertificate 获取指定域名的证书
func (m *Manager) GetCertificate(domain string) (*model.Certificate, error) {
	return m.db.GetCertificate(domain)
}

// CreateCertificate 创建证书
func (m *Manager) CreateCertificate(domain string) (*model.Certificate, error) {
	// 简单实现，实际应该生成或申请证书
	cert := &model.Certificate{
		Base: model.Base{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Domain:        domain,
		CertFile:      "/certs/" + domain + ".crt",
		KeyFile:       "/certs/" + domain + ".key",
		Status:        "pending",
		LastCheckedAt: time.Now(),
		LastRenewedAt: time.Now(),
		ExpiresAt:     time.Now().Add(90 * 24 * time.Hour), // 90天有效期
	}

	// 保存到数据库
	if err := m.db.CreateCertificate(cert); err != nil {
		return nil, err
	}

	return cert, nil
}

// DeleteCertificate 删除证书
func (m *Manager) DeleteCertificate(domain string) error {
	return m.db.DeleteCertificate(domain)
}

// RenewCertificate 续期证书
func (m *Manager) RenewCertificate(domain string) error {
	cert, err := m.db.GetCertificate(domain)
	if err != nil {
		return err
	}

	if cert == nil {
		return nil
	}

	// 更新证书状态
	cert.LastRenewedAt = time.Now()
	cert.ExpiresAt = time.Now().Add(90 * 24 * time.Hour)
	cert.Base.UpdatedAt = time.Now()
	cert.Status = "valid"

	return m.db.UpdateCertificate(cert)
}
