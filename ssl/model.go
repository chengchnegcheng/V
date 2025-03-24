package ssl

import (
	"time"
)

// CertificateModel represents a certificate in the application model
type CertificateModel struct {
	ID            int64     `json:"id"`
	Domain        string    `json:"domain"`
	CertFile      string    `json:"cert_file"`
	KeyFile       string    `json:"key_file"`
	Status        string    `json:"status"`
	LastCheckedAt time.Time `json:"last_checked_at"`
	LastRenewedAt time.Time `json:"last_renewed_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	AutoRenew     bool      `json:"auto_renew"`
}

// CertificateService defines the operations for certificate management
type CertificateService interface {
	// Self-signed certificate operations
	CreateSelfSigned(domain string, autoRenew bool) (*CertificateModel, error)
	GetSelfSigned(id int64) (*CertificateModel, error)
	GetSelfSignedByDomain(domain string) (*CertificateModel, error)
	UpdateSelfSigned(cert *CertificateModel) error
	DeleteSelfSigned(id int64) error
	RenewSelfSigned(id int64) error

	// ACME certificate operations
	CreateACME(domain string) error
	GetACME(domain string) (*CertificateModel, error)
	RenewACME(domain string) error
	DeleteACME(domain string) error

	// Common operations
	ListCertificates() ([]*CertificateModel, error)
	LoadCertificate(certFile, keyFile string) error
}
