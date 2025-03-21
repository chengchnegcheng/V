package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"v/logger"
	"v/model"
	"v/settings"
)

// SelfSignedCertManager 自签名证书管理器
type SelfSignedCertManager struct {
	logger   *logger.Logger
	settings *settings.Manager
	db       model.DB
}

// NewSelfSignedCertManager 创建自签名证书管理器
func NewSelfSignedCertManager(logger *logger.Logger, settings *settings.Manager, db model.DB) *SelfSignedCertManager {
	return &SelfSignedCertManager{
		logger:   logger,
		settings: settings,
		db:       db,
	}
}

// CreateSelfSignedCertificate 创建自签名证书
func (m *SelfSignedCertManager) CreateSelfSignedCertificate(domain string) (*model.Certificate, error) {
	// 检查是否已存在证书
	cert, err := m.db.GetCertificate(domain)
	if err == nil && cert != nil {
		return cert, nil
	}

	// 创建证书目录
	s := m.settings.Get()
	certDir := s.SSL.CertDir
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, err
	}

	// 生成密钥对
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// 定义证书模板
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"V Proxy"},
			OrganizationalUnit: []string{"Self-Signed Certificates"},
			CommonName:         domain,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 添加域名作为DNS和IP（如果是IP）
	if ip := net.ParseIP(domain); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, domain)
		// 添加www子域
		if !strings.HasPrefix(domain, "www.") {
			template.DNSNames = append(template.DNSNames, "www."+domain)
		}
	}

	// 创建自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	// 保存证书到文件
	certFile := filepath.Join(certDir, domain+".crt")
	certOut, err := os.Create(certFile)
	if err != nil {
		return nil, err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, err
	}

	// 保存私钥到文件
	keyFile := filepath.Join(certDir, domain+".key")
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer keyOut.Close()

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return nil, err
	}

	// 创建证书记录
	certificate := &model.Certificate{
		Domain:        domain,
		CertFile:      certFile,
		KeyFile:       keyFile,
		Status:        "valid",
		LastCheckedAt: time.Now(),
		LastRenewedAt: time.Now(),
		ExpiresAt:     notAfter,
	}

	// 保存证书信息到数据库
	if err := m.db.CreateCertificate(certificate); err != nil {
		// 删除已创建的证书文件
		os.Remove(certFile)
		os.Remove(keyFile)
		return nil, err
	}

	return certificate, nil
}

// GetSelfSignedCertificate 获取自签名证书
func (m *SelfSignedCertManager) GetSelfSignedCertificate(domain string) (*model.Certificate, error) {
	return m.db.GetCertificate(domain)
}

// DeleteSelfSignedCertificate 删除自签名证书
func (m *SelfSignedCertManager) DeleteSelfSignedCertificate(domain string) error {
	cert, err := m.db.GetCertificate(domain)
	if err != nil {
		return err
	}

	// 删除证书文件
	if err := os.Remove(cert.CertFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 删除私钥文件
	if err := os.Remove(cert.KeyFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 从数据库删除
	return m.db.DeleteCertificate(domain)
}

// LoadSelfSignedCertificate 加载自签名证书
func (m *SelfSignedCertManager) LoadSelfSignedCertificate(domain string) (*tls.Certificate, error) {
	cert, err := m.db.GetCertificate(domain)
	if err != nil {
		return nil, err
	}

	// 加载证书
	tlsCert, err := tls.LoadX509KeyPair(cert.CertFile, cert.KeyFile)
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}

// RenewSelfSignedCertificate 更新自签名证书
func (m *SelfSignedCertManager) RenewSelfSignedCertificate(domain string) (*model.Certificate, error) {
	// 删除旧证书
	if err := m.DeleteSelfSignedCertificate(domain); err != nil {
		return nil, err
	}

	// 创建新证书
	return m.CreateSelfSignedCertificate(domain)
}
