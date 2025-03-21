package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"v/errors"
	"v/logger"
	"v/settings"
)

// Certificate represents an SSL certificate
type Certificate struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	CertFile  string    `json:"cert_file"`
	KeyFile   string    `json:"key_file"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	AutoRenew bool      `json:"auto_renew"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Manager represents an SSL certificate manager
type Manager struct {
	log      *logger.Logger
	settings *settings.Manager
	certs    map[int64]*Certificate
	mu       sync.RWMutex
}

// New creates a new SSL certificate manager
func New(log *logger.Logger, settings *settings.Manager) *Manager {
	return &Manager{
		log:      log,
		settings: settings,
		certs:    make(map[int64]*Certificate),
	}
}

// Create creates a new SSL certificate
func (m *Manager) Create(domain string, autoRenew bool) (*Certificate, error) {
	// Validate input
	if err := m.validateInput(domain); err != nil {
		return nil, err
	}

	// Get settings
	s := m.settings.Get()

	// Create certificate
	cert := &Certificate{
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

	m.log.Info("SSL certificate created", logger.Fields{
		"domain":     domain,
		"auto_renew": autoRenew,
	})

	return cert, nil
}

// Get returns a certificate by ID
func (m *Manager) Get(id int64) (*Certificate, error) {
	m.mu.RLock()
	cert, ok := m.certs[id]
	m.mu.RUnlock()

	if !ok {
		return nil, errors.New(errors.ErrNotFound, "Certificate not found", nil)
	}

	return cert, nil
}

// GetByDomain returns a certificate by domain
func (m *Manager) GetByDomain(domain string) (*Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, cert := range m.certs {
		if cert.Domain == domain {
			return cert, nil
		}
	}

	return nil, errors.New(errors.ErrNotFound, "Certificate not found", nil)
}

// Update updates a certificate
func (m *Manager) Update(cert *Certificate) error {
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

	m.log.Info("SSL certificate updated", logger.Fields{
		"cert_id":    cert.ID,
		"domain":     cert.Domain,
		"auto_renew": cert.AutoRenew,
	})

	return nil
}

// Delete deletes a certificate
func (m *Manager) Delete(id int64) error {
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

	m.log.Info("SSL certificate deleted", logger.Fields{
		"cert_id": id,
		"domain":  cert.Domain,
	})

	return nil
}

// Renew renews a certificate
func (m *Manager) Renew(id int64) error {
	// Get certificate
	cert, err := m.Get(id)
	if err != nil {
		return err
	}

	// Generate new certificate
	if err := m.generateCertificate(cert); err != nil {
		return err
	}

	m.log.Info("SSL certificate renewed", logger.Fields{
		"cert_id": id,
		"domain":  cert.Domain,
	})

	return nil
}

// validateInput validates certificate input
func (m *Manager) validateInput(domain string) error {
	if domain == "" {
		return errors.New(errors.ErrBadRequest, "Domain is required", nil)
	}

	// TODO: Implement domain validation

	return nil
}

// generateCertificate generates a new SSL certificate
func (m *Manager) generateCertificate(cert *Certificate) error {
	// Get settings
	s := m.settings.Get()

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Generate certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"V Proxy"},
			CommonName:   cert.Domain,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(0, 0, s.SSL.RenewDays),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Create certificate directory
	certDir := filepath.Join(s.SSL.CertDir, cert.Domain)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %v", err)
	}

	// Save certificate
	certFile := filepath.Join(certDir, "cert.pem")
	keyFile := filepath.Join(certDir, "key.pem")

	// Write certificate
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %v", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		certOut.Close()
		return fmt.Errorf("failed to write certificate: %v", err)
	}
	certOut.Close()

	// Write private key
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		keyOut.Close()
		return fmt.Errorf("failed to write private key: %v", err)
	}
	keyOut.Close()

	// Update certificate
	cert.CertFile = certFile
	cert.KeyFile = keyFile
	cert.IssuedAt = time.Now()
	cert.ExpiresAt = time.Now().AddDate(0, 0, s.SSL.RenewDays)
	cert.UpdatedAt = time.Now()

	return nil
}

// deleteCertificateFiles deletes certificate files
func (m *Manager) deleteCertificateFiles(cert *Certificate) error {
	if cert.CertFile != "" {
		if err := os.Remove(cert.CertFile); err != nil {
			return fmt.Errorf("failed to delete certificate file: %v", err)
		}
	}

	if cert.KeyFile != "" {
		if err := os.Remove(cert.KeyFile); err != nil {
			return fmt.Errorf("failed to delete key file: %v", err)
		}
	}

	return nil
}

// LoadCertificate loads a certificate from files
func (m *Manager) LoadCertificate(certFile, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}
	return &cert, nil
}
