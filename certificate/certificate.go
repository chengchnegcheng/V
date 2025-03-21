package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"v/logger"
	"v/model"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// Manager 证书管理器
type Manager struct {
	log    *logger.Logger
	db     model.DB
	config *model.CertificateConfig
	user   *model.User
	client *lego.Client
	stopCh chan struct{}
}

// NewManager 创建证书管理器
func NewManager(log *logger.Logger, db model.DB, config *model.CertificateConfig) (*Manager, error) {
	// 创建用户
	key, err := certcrypto.GeneratePrivateKey(certcrypto.EC256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	user := &model.User{
		Email: config.Email,
		Key:   key,
	}

	// 创建 ACME 客户端配置
	legoConfig := lego.NewConfig(user)
	legoConfig.Certificate.KeyType = certcrypto.EC256

	// 创建 ACME 客户端
	client, err := lego.NewClient(legoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ACME client: %v", err)
	}

	return &Manager{
		log:    log,
		db:     db,
		config: config,
		user:   user,
		client: client,
		stopCh: make(chan struct{}),
	}, nil
}

// Start 启动证书管理器
func (m *Manager) Start() error {
	m.log.Info("Starting certificate manager", logger.Fields{
		"email": m.config.Email,
	})

	// 注册用户
	reg, err := m.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return fmt.Errorf("failed to register user: %v", err)
	}
	m.user.Registration = reg

	// 启动证书监控
	go m.monitorLoop()

	return nil
}

// Stop 停止证书管理器
func (m *Manager) Stop() error {
	m.log.Info("Stopping certificate manager", logger.Fields{})
	close(m.stopCh)
	return nil
}

// monitorLoop 证书监控循环
func (m *Manager) monitorLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if err := m.checkAndRenewCertificates(); err != nil {
				m.log.Error("Failed to check and renew certificates", logger.Fields{
					"error": err.Error(),
				})
			}
		}
	}
}

// checkAndRenewCertificates 检查并续期证书
func (m *Manager) checkAndRenewCertificates() error {
	certs, err := m.db.ListCertificates()
	if err != nil {
		return fmt.Errorf("failed to list certificates: %v", err)
	}

	for _, cert := range certs {
		if err := m.CheckAndRenewCertificate(cert); err != nil {
			m.log.Error("Failed to check and renew certificate", logger.Fields{
				"domain": cert.Domain,
				"error":  err.Error(),
			})
		}
	}

	return nil
}

// CheckAndRenewCertificate 检查并续期单个证书
func (m *Manager) CheckAndRenewCertificate(cert *model.Certificate) error {
	// 检查证书是否需要续期
	if !m.shouldRenew(cert) {
		return nil
	}

	// 申请新证书
	request := certificate.ObtainRequest{
		Domains: []string{cert.Domain},
		Bundle:  true,
	}

	certificates, err := m.client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %v", err)
	}

	// 保存证书
	if err := m.saveCertificate(cert, certificates); err != nil {
		return fmt.Errorf("failed to save certificate: %v", err)
	}

	// 更新数据库
	cert.Certificate = certificates.Certificate
	cert.PrivateKey = certificates.PrivateKey
	cert.ExpireAt = time.Now().Add(90 * 24 * time.Hour) // Let's Encrypt 证书有效期为 90 天

	if err := m.db.UpdateCertificate(cert); err != nil {
		return fmt.Errorf("failed to update certificate in database: %v", err)
	}

	m.log.Info("Certificate renewed successfully", logger.Fields{
		"domain": cert.Domain,
	})

	return nil
}

// shouldRenew 检查证书是否需要续期
func (m *Manager) shouldRenew(cert *model.Certificate) bool {
	// 如果证书将在 30 天内过期，则需要续期
	return time.Until(cert.ExpireAt) < 30*24*time.Hour
}

// saveCertificate 保存证书文件
func (m *Manager) saveCertificate(cert *model.Certificate, certificates *certificate.Resource) error {
	// 创建证书目录
	certDir := filepath.Join("certificates", cert.Domain)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %v", err)
	}

	// 保存证书文件
	if err := ioutil.WriteFile(filepath.Join(certDir, "cert.pem"), certificates.Certificate, 0644); err != nil {
		return fmt.Errorf("failed to save certificate file: %v", err)
	}

	// 保存私钥文件
	if err := ioutil.WriteFile(filepath.Join(certDir, "key.pem"), certificates.PrivateKey, 0600); err != nil {
		return fmt.Errorf("failed to save private key file: %v", err)
	}

	return nil
}

// GetCertificate 获取证书
func (m *Manager) GetCertificate(domain string) (*tls.Certificate, error) {
	cert, err := m.db.GetCertificate(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %v", err)
	}

	// 解析证书
	tlsCert, err := tls.X509KeyPair(cert.Certificate, cert.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return &tlsCert, nil
}

// ValidateCertificate 验证证书
func (m *Manager) ValidateCertificate(cert *model.Certificate) error {
	// 解析证书
	block, _ := pem.Decode(cert.Certificate)
	if block == nil {
		return fmt.Errorf("failed to decode certificate")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// 检查证书是否过期
	if time.Now().After(x509Cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}

	// 检查证书是否即将过期
	if time.Until(x509Cert.NotAfter) < 30*24*time.Hour {
		return fmt.Errorf("certificate will expire soon")
	}

	return nil
}
