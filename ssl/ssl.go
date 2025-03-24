package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"v/logger"
	"v/settings"
)

// SelfSignedCert represents a self-signed SSL certificate
type SelfSignedCert struct {
	ID          int64     `json:"id"`
	Domain      string    `json:"domain"`
	CertFile    string    `json:"cert_file"`
	KeyFile     string    `json:"key_file"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ExpireAt    time.Time `json:"expire_at"` // Alias for ExpireAt to match the API
	AutoRenew   bool      `json:"auto_renew"`
	LastRenewed time.Time `json:"last_renewed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LocalCertManager represents a manager for self-signed certificates
type LocalCertManager struct {
	log      *logger.Logger
	settings *settings.Manager
	certs    map[int64]*SelfSignedCert
	mu       sync.RWMutex
}

// NewLocalManager creates a new self-signed certificate manager
func NewLocalManager(log *logger.Logger, settings *settings.Manager) *LocalCertManager {
	return &LocalCertManager{
		log:      log,
		settings: settings,
		certs:    make(map[int64]*SelfSignedCert),
	}
}

// Create creates a new self-signed SSL certificate
func (m *LocalCertManager) Create(domain string, autoRenew bool) (*SelfSignedCert, error) {
	// Validate input
	if err := m.validateInput(domain); err != nil {
		return nil, err
	}

	// Create certificate
	cert := &SelfSignedCert{
		Domain:    domain,
		AutoRenew: autoRenew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate certificate
	if err := m.generateCertificate(cert); err != nil {
		return nil, err
	}

	// Save certificate
	m.mu.Lock()
	m.certs[cert.ID] = cert
	m.mu.Unlock()

	m.log.Info("Self-signed SSL certificate created", logger.Fields{
		"domain":     domain,
		"auto_renew": autoRenew,
	})

	return cert, nil
}

// Get returns a certificate by ID
func (m *LocalCertManager) Get(id int64) (*SelfSignedCert, error) {
	m.mu.RLock()
	cert, ok := m.certs[id]
	m.mu.RUnlock()

	if !ok {
		return nil, errors.New("certificate not found")
	}

	return cert, nil
}

// GetByDomain returns a certificate by domain
func (m *LocalCertManager) GetByDomain(domain string) (*SelfSignedCert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, cert := range m.certs {
		if cert.Domain == domain {
			return cert, nil
		}
	}

	return nil, errors.New("certificate not found")
}

// Update updates a certificate
func (m *LocalCertManager) Update(cert *SelfSignedCert) error {
	// Validate input
	if err := m.validateInput(cert.Domain); err != nil {
		return err
	}

	// Update timestamp
	cert.UpdatedAt = time.Now()

	// Save certificate
	m.mu.Lock()
	m.certs[cert.ID] = cert
	m.mu.Unlock()

	m.log.Info("Self-signed SSL certificate updated", logger.Fields{
		"cert_id":    cert.ID,
		"domain":     cert.Domain,
		"auto_renew": cert.AutoRenew,
	})

	return nil
}

// Delete deletes a certificate
func (m *LocalCertManager) Delete(id int64) error {
	// Get certificate
	cert, err := m.Get(id)
	if err != nil {
		return err
	}

	// Delete certificate files
	if err := m.deleteCertificateFiles(cert); err != nil {
		return err
	}

	// Delete certificate
	m.mu.Lock()
	delete(m.certs, id)
	m.mu.Unlock()

	m.log.Info("Self-signed SSL certificate deleted", logger.Fields{
		"cert_id": id,
		"domain":  cert.Domain,
	})

	return nil
}

// Renew renews a certificate
func (m *LocalCertManager) Renew(id int64) error {
	// Get certificate
	cert, err := m.Get(id)
	if err != nil {
		return err
	}

	// Generate new certificate
	if err := m.generateCertificate(cert); err != nil {
		return err
	}

	m.log.Info("Self-signed SSL certificate renewed", logger.Fields{
		"cert_id": id,
		"domain":  cert.Domain,
	})

	return nil
}

// ListCertificates returns a list of all certificates
func (m *LocalCertManager) ListCertificates() ([]*SelfSignedCert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certs := make([]*SelfSignedCert, 0, len(m.certs))
	for _, cert := range m.certs {
		certs = append(certs, cert)
	}

	m.log.Info("Listed all self-signed SSL certificates", logger.Fields{
		"count": len(certs),
	})

	return certs, nil
}

// validateInput validates certificate input
func (m *LocalCertManager) validateInput(domain string) error {
	if domain == "" {
		return errors.New("domain is required")
	}

	// Domain validation
	// Check if the domain has a valid format
	if !isValidDomain(domain) {
		return errors.New("invalid domain format")
	}

	return nil
}

// isValidDomain checks if a domain name is valid
func isValidDomain(domain string) bool {
	// Check for basic domain format
	// - Contains at least one dot
	// - No spaces
	// - No special characters except dots and hyphens
	// - Does not start or end with a hyphen
	// - Each part is 1-63 characters

	if len(domain) > 253 {
		return false
	}

	// Domain must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	// Split the domain into parts
	parts := strings.Split(domain, ".")

	// Check each part
	for _, part := range parts {
		// Empty parts are not allowed
		if len(part) == 0 {
			return false
		}

		// Parts must be 1-63 characters
		if len(part) > 63 {
			return false
		}

		// Parts must not start or end with a hyphen
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}

		// Parts must only contain alphanumeric characters and hyphens
		for _, r := range part {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == '-') {
				return false
			}
		}
	}

	return true
}

// generateCertificate generates a new self-signed certificate
func (m *LocalCertManager) generateCertificate(cert *SelfSignedCert) error {
	// Generate key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // 1 year

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Self Signed Certificate"},
			CommonName:   cert.Domain,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{cert.Domain},
	}

	// Add www subdomain
	if !strings.HasPrefix(cert.Domain, "www.") {
		template.DNSNames = append(template.DNSNames, "www."+cert.Domain)
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Create cert directory if it doesn't exist
	certDir := filepath.Join("certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %v", err)
	}

	// Write certificate to file
	certOut, err := os.Create(filepath.Join(certDir, cert.Domain+".crt"))
	if err != nil {
		return fmt.Errorf("failed to open certificate file for writing: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate to file: %v", err)
	}

	// Write private key to file
	keyOut, err := os.OpenFile(filepath.Join(certDir, cert.Domain+".key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open private key file for writing: %v", err)
	}
	defer keyOut.Close()

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write private key to file: %v", err)
	}

	// Update certificate paths
	cert.CertFile = filepath.Join(certDir, cert.Domain+".crt")
	cert.KeyFile = filepath.Join(certDir, cert.Domain+".key")
	cert.IssuedAt = notBefore
	cert.ExpiresAt = notAfter
	cert.UpdatedAt = time.Now()

	return nil
}

// deleteCertificateFiles deletes certificate files
func (m *LocalCertManager) deleteCertificateFiles(cert *SelfSignedCert) error {
	// Delete certificate file
	if cert.CertFile != "" {
		if err := os.Remove(cert.CertFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete certificate file: %v", err)
		}
	}

	// Delete key file
	if cert.KeyFile != "" {
		if err := os.Remove(cert.KeyFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete key file: %v", err)
		}
	}

	return nil
}

// LoadCertificate loads a certificate from files
func (m *LocalCertManager) LoadCertificate(certFile, keyFile string) (*tls.Certificate, error) {
	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}

	return &cert, nil
}

// Manager is an alias for ACMEManager for backward compatibility
type Manager = ACMEManager

// CertificateManager is an alias for LocalCertManager for backward compatibility
type CertificateManager = LocalCertManager
