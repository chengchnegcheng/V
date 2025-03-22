package cert

import (
	"crypto"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"v/logger"
	"v/model"
	"v/notification"
	"v/settings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"
)

// CertificateStatus 证书状态
type CertificateStatus string

const (
	// CertificateStatusValid 有效
	CertificateStatusValid CertificateStatus = "valid"
	// CertificateStatusExpiringSoon 即将过期
	CertificateStatusExpiringSoon CertificateStatus = "expiring_soon"
	// CertificateStatusExpired 已过期
	CertificateStatusExpired CertificateStatus = "expired"
	// CertificateStatusError 错误
	CertificateStatusError CertificateStatus = "error"
	// CertificateStatusUnknown 未知
	CertificateStatusUnknown CertificateStatus = "unknown"
)

// CertManager SSL证书管理器，改名以避免与 cert.go 中的 Manager 重名
type CertManager struct {
	log      *logger.Logger
	settings *settings.Manager
	notifier notification.Notifier
	db       model.DB
	certs    map[string]*model.Certificate
	mu       sync.RWMutex
	stopCh   chan struct{}
	webRoot  string
}

// NewCertManager 创建SSL证书管理器
func NewCertManager(log *logger.Logger, settings *settings.Manager, notifier notification.Notifier, db model.DB, webRoot string) *CertManager {
	return &CertManager{
		log:      log,
		settings: settings,
		notifier: notifier,
		db:       db,
		certs:    make(map[string]*model.Certificate),
		stopCh:   make(chan struct{}),
		webRoot:  webRoot,
	}
}

// Start 启动SSL证书管理器
func (m *CertManager) Start() error {
	// 加载所有证书信息
	if err := m.loadCertificates(); err != nil {
		return fmt.Errorf("failed to load certificates: %v", err)
	}

	// 启动证书检查循环
	go m.checkLoop()

	// 启动证书续期循环
	go m.renewLoop()

	return nil
}

// Stop 停止SSL证书管理器
func (m *CertManager) Stop() {
	close(m.stopCh)
}

// loadCertificates 加载所有证书信息
func (m *CertManager) loadCertificates() error {
	certs, err := m.db.ListCertificates()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cert := range certs {
		m.certs[cert.Domain] = cert
	}

	return nil
}

// checkLoop 证书检查循环
func (m *CertManager) checkLoop() {
	s := m.settings.Get()
	ticker := time.NewTicker(s.SSL.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if err := m.checkAllCertificates(); err != nil {
				m.log.Error("Failed to check certificates", logger.Fields{
					"error": err.Error(),
				})
			}
		}
	}
}

// renewLoop 证书续期循环
func (m *CertManager) renewLoop() {
	s := m.settings.Get()
	ticker := time.NewTicker(s.SSL.RenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if err := m.renewExpiringCertificates(); err != nil {
				m.log.Error("Failed to renew certificates", logger.Fields{
					"error": err.Error(),
				})
			}
		}
	}
}

// checkAllCertificates 检查所有证书
func (m *CertManager) checkAllCertificates() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for domain, cert := range m.certs {
		status, err := m.checkCertificate(domain, cert)
		if err != nil {
			m.log.Error("Failed to check certificate", logger.Fields{
				"domain": domain,
				"error":  err.Error(),
			})
			continue
		}

		// 更新证书状态
		cert.Status = string(status)
		cert.LastCheckedAt = time.Now()
		if err := m.db.UpdateCertificate(cert); err != nil {
			m.log.Error("Failed to update certificate status", logger.Fields{
				"domain": domain,
				"error":  err.Error(),
			})
		}

		// 证书即将过期，发送通知
		if status == CertificateStatusExpiringSoon {
			if err := m.notifyCertificateExpiring(domain, cert); err != nil {
				m.log.Error("Failed to send certificate expiring notification", logger.Fields{
					"domain": domain,
					"error":  err.Error(),
				})
			}
		}

		// 证书已过期，发送通知
		if status == CertificateStatusExpired {
			if err := m.notifyCertificateExpired(domain, cert); err != nil {
				m.log.Error("Failed to send certificate expired notification", logger.Fields{
					"domain": domain,
					"error":  err.Error(),
				})
			}
		}
	}

	return nil
}

// checkCertificate 检查证书状态
func (m *CertManager) checkCertificate(domain string, cert *model.Certificate) (CertificateStatus, error) {
	s := m.settings.Get()

	// 如果证书文件不存在，返回错误
	if _, err := os.Stat(cert.CertFile); os.IsNotExist(err) {
		return CertificateStatusError, fmt.Errorf("certificate file not found")
	}

	// 如果密钥文件不存在，返回错误
	if _, err := os.Stat(cert.KeyFile); os.IsNotExist(err) {
		return CertificateStatusError, fmt.Errorf("key file not found")
	}

	// 加载证书
	certData, err := ioutil.ReadFile(cert.CertFile)
	if err != nil {
		return CertificateStatusError, err
	}

	keyData, err := ioutil.ReadFile(cert.KeyFile)
	if err != nil {
		return CertificateStatusError, err
	}

	tlsCert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return CertificateStatusError, err
	}

	// 解析证书
	leaf := tlsCert.Leaf
	if leaf == nil {
		return CertificateStatusUnknown, fmt.Errorf("failed to parse certificate")
	}

	// 检查证书是否过期
	now := time.Now()
	if now.After(leaf.NotAfter) {
		return CertificateStatusExpired, nil
	}

	// 检查证书是否即将过期
	expiringThreshold := s.SSL.ExpiryWarningDays * 24 * time.Hour
	if now.Add(expiringThreshold).After(leaf.NotAfter) {
		return CertificateStatusExpiringSoon, nil
	}

	return CertificateStatusValid, nil
}

// renewExpiringCertificates 更新即将过期的证书
func (m *CertManager) renewExpiringCertificates() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for domain, cert := range m.certs {
		// 检查证书状态
		status, err := m.checkCertificate(domain, cert)
		if err != nil {
			m.log.Error("Failed to check certificate for renewal", logger.Fields{
				"domain": domain,
				"error":  err.Error(),
			})
			continue
		}

		// 如果证书已过期或即将过期，尝试更新
		if status == CertificateStatusExpired || status == CertificateStatusExpiringSoon {
			if err := m.renewCertificate(domain, cert); err != nil {
				m.log.Error("Failed to renew certificate", logger.Fields{
					"domain": domain,
					"error":  err.Error(),
				})
				// 发送证书更新失败通知
				if err := m.notifyCertificateRenewFailed(domain, cert, err); err != nil {
					m.log.Error("Failed to send certificate renewal failure notification", logger.Fields{
						"domain": domain,
						"error":  err.Error(),
					})
				}
			} else {
				m.log.Info("Certificate renewed successfully", logger.Fields{
					"domain": domain,
				})
				// 发送证书更新成功通知
				if err := m.notifyCertificateRenewed(domain, cert); err != nil {
					m.log.Error("Failed to send certificate renewal success notification", logger.Fields{
						"domain": domain,
						"error":  err.Error(),
					})
				}
			}
		}
	}

	return nil
}

// renewCertificate 更新证书
func (m *CertManager) renewCertificate(domain string, cert *model.Certificate) error {
	s := m.settings.Get()

	// 创建临时目录来保存证书
	tempDir, err := ioutil.TempDir("", "ssl-cert-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// 创建 ACME 用户
	privateKey, err := certcrypto.GeneratePrivateKey(certcrypto.RSA2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	user := &User{
		Email: s.Admin.Email,
		Key:   privateKey,
	}

	// 初始化 ACME 客户端
	config := lego.NewConfig(user)
	config.CADirURL = s.SSL.AcmeURL
	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}

	// 注册用户
	reg, err := client.Registration.Register(registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	})
	if err != nil {
		return err
	}
	user.Registration = reg

	// 配置验证方式
	switch s.SSL.ChallengeType {
	case "http-01":
		// 配置HTTP-01验证
		provider, err := webroot.NewHTTPProvider(m.webRoot)
		if err != nil {
			return fmt.Errorf("failed to create HTTP provider: %v", err)
		}
		err = client.Challenge.SetHTTP01Provider(provider)
	case "tls-alpn-01":
		// 配置TLS-ALPN-01验证
		err = client.Challenge.SetTLSALPN01Provider(
			tlsalpn01.NewProviderServer("", ""),
		)
	default:
		return fmt.Errorf("unsupported challenge type: %s", s.SSL.ChallengeType)
	}

	if err != nil {
		return err
	}

	// 申请证书
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	// 保存证书和密钥
	err = ioutil.WriteFile(cert.CertFile, certificates.Certificate, 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cert.KeyFile, certificates.PrivateKey, 0600)
	if err != nil {
		return err
	}

	// 更新证书信息
	cert.Status = string(CertificateStatusValid)
	cert.LastRenewedAt = time.Now()
	return m.db.UpdateCertificate(cert)
}

// GetCertificate 获取证书信息
func (m *CertManager) GetCertificate(domain string) (*model.Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cert, ok := m.certs[domain]
	if !ok {
		return nil, fmt.Errorf("certificate not found for domain %s", domain)
	}

	return cert, nil
}

// CreateCertificate 创建证书
func (m *CertManager) CreateCertificate(domain string) (*model.Certificate, error) {
	s := m.settings.Get()
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查域名是否已存在证书
	if _, ok := m.certs[domain]; ok {
		return nil, fmt.Errorf("certificate already exists for domain %s", domain)
	}

	// 创建证书记录
	cert := &model.Certificate{
		Domain:        domain,
		CertFile:      filepath.Join(s.SSL.CertDir, domain+".crt"),
		KeyFile:       filepath.Join(s.SSL.CertDir, domain+".key"),
		Status:        string(CertificateStatusUnknown),
		LastCheckedAt: time.Now(),
		LastRenewedAt: time.Now(),
		ExpiresAt:     time.Now().AddDate(1, 0, 0),
	}

	// 保存到数据库
	if err := m.db.CreateCertificate(cert); err != nil {
		return nil, err
	}

	// 申请证书
	if err := m.renewCertificate(domain, cert); err != nil {
		// 如果申请失败，仍然保留记录，但更新状态
		cert.Status = string(CertificateStatusError)
		m.db.UpdateCertificate(cert)
		return nil, err
	}

	// 添加到内存中
	m.certs[domain] = cert

	return cert, nil
}

// DeleteCertificate 删除证书
func (m *CertManager) DeleteACMECertificate(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cert, ok := m.certs[domain]
	if !ok {
		return fmt.Errorf("certificate not found for domain %s", domain)
	}

	// 删除证书文件
	if err := os.Remove(cert.CertFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 删除密钥文件
	if err := os.Remove(cert.KeyFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 从数据库中删除
	if err := m.db.DeleteCertificate(domain); err != nil {
		return err
	}

	// 从内存中删除
	delete(m.certs, domain)

	return nil
}

// ListCertificates 列出所有证书
func (m *CertManager) ListCertificates() []*model.Certificate {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certs := make([]*model.Certificate, 0, len(m.certs))
	for _, cert := range m.certs {
		certs = append(certs, cert)
	}

	return certs
}

// notifyCertificateExpiring 发送证书即将过期通知
func (m *CertManager) notifyCertificateExpiring(domain string, cert *model.Certificate) error {
	s := m.settings.Get()
	notification := &notification.Notification{
		To:      []string{s.Admin.Email},
		Subject: fmt.Sprintf("SSL证书即将过期提醒: %s", domain),
		Body: fmt.Sprintf(`
			<p>管理员您好：</p>
			<p>域名 %s 的SSL证书即将过期。</p>
			<p>请及时续期！</p>
			<p>此邮件由系统自动发送，请勿回复。</p>
		`, domain),
		Type: "certificate_expiring",
	}

	return m.notifier.Send(notification)
}

// notifyCertificateExpired 发送证书已过期通知
func (m *CertManager) notifyCertificateExpired(domain string, cert *model.Certificate) error {
	s := m.settings.Get()
	notification := &notification.Notification{
		To:      []string{s.Admin.Email},
		Subject: fmt.Sprintf("SSL证书已过期: %s", domain),
		Body: fmt.Sprintf(`
			<p>管理员您好：</p>
			<p>域名 %s 的SSL证书已过期。</p>
			<p>请立即续期！</p>
			<p>此邮件由系统自动发送，请勿回复。</p>
		`, domain),
		Type: "certificate_expired",
	}

	return m.notifier.Send(notification)
}

// notifyCertificateRenewed 发送证书续期成功通知
func (m *CertManager) notifyCertificateRenewed(domain string, cert *model.Certificate) error {
	s := m.settings.Get()
	notification := &notification.Notification{
		To:      []string{s.Admin.Email},
		Subject: fmt.Sprintf("SSL证书续期成功: %s", domain),
		Body: fmt.Sprintf(`
			<p>管理员您好：</p>
			<p>域名 %s 的SSL证书已成功续期。</p>
			<p>此邮件由系统自动发送，请勿回复。</p>
		`, domain),
		Type: "certificate_renewed",
	}

	return m.notifier.Send(notification)
}

// notifyCertificateRenewFailed 发送证书续期失败通知
func (m *CertManager) notifyCertificateRenewFailed(domain string, cert *model.Certificate, renewErr error) error {
	s := m.settings.Get()
	notification := &notification.Notification{
		To:      []string{s.Admin.Email},
		Subject: fmt.Sprintf("SSL证书续期失败: %s", domain),
		Body: fmt.Sprintf(`
			<p>管理员您好：</p>
			<p>域名 %s 的SSL证书续期失败。</p>
			<p>错误信息: %s</p>
			<p>请手动处理！</p>
			<p>此邮件由系统自动发送，请勿回复。</p>
		`, domain, renewErr.Error()),
		Type: "certificate_renew_failed",
	}

	return m.notifier.Send(notification)
}

// User ACME用户
type User struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

// GetEmail 获取邮箱
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration 获取注册信息
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey 获取私钥
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

// RenewCertificate 公开的更新证书方法
func (m *CertManager) RenewCertificate(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cert, ok := m.certs[domain]
	if !ok {
		return fmt.Errorf("certificate not found for domain %s", domain)
	}

	return m.renewCertificate(domain, cert)
}

// CertificateManager 是证书管理器接口 (替换原有的Manager接口)
type CertificateManager interface {
	GetCertificate(domain string) (*model.Certificate, error)
	CreateCertificate(domain string) (*model.Certificate, error)
	RenewCertificate(domain string) error
	DeleteCertificate(domain string) error
	ListCertificates() []*model.Certificate
}

// DummyManager 是一个空实现的证书管理器
type DummyManager struct{}
