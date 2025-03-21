package ssl

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"
	"v/database"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// Manager handles SSL certificate operations
type Manager struct {
	email    string
	client   *lego.Client
	account  *registration.Resource
	renewals map[string]*time.Timer
}

// Certificate represents an SSL certificate
type Certificate struct {
	ID          int64
	Domain      string
	Certificate string
	PrivateKey  string
	CreatedAt   time.Time
	ExpireAt    time.Time
}

// NewManager creates a new SSL certificate manager
func NewManager(email string) (*Manager, error) {
	// Create user
	user := &User{
		Email: email,
	}

	// Create lego config
	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryProduction // Use production Let's Encrypt

	// Create client
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create lego client: %v", err)
	}

	// Use HTTP-01 challenge
	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "80"))
	if err != nil {
		return nil, fmt.Errorf("failed to set HTTP-01 provider: %v", err)
	}

	// Create new account
	account, err := client.Registration.Register(registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register account: %v", err)
	}

	return &Manager{
		email:    email,
		client:   client,
		account:  account,
		renewals: make(map[string]*time.Timer),
	}, nil
}

// ObtainCertificate obtains a new SSL certificate for the given domain
func (m *Manager) ObtainCertificate(domain string) error {
	// Request certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := m.client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %v", err)
	}

	// Parse certificate
	certBlock, _ := pem.Decode(certificates.Certificate)
	if certBlock == nil {
		return fmt.Errorf("failed to decode certificate")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Save to database
	_, err = database.DB.Exec(`
		INSERT INTO ssl_certificates (domain, certificate, private_key, created_at, expire_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, ?)
	`, domain, string(certificates.Certificate), string(certificates.PrivateKey), cert.NotAfter)
	if err != nil {
		return fmt.Errorf("failed to save certificate: %v", err)
	}

	// Schedule renewal
	m.scheduleRenewal(domain, cert.NotAfter)

	log.Printf("Certificate obtained for domain: %s", domain)
	return nil
}

// RenewCertificate renews an existing SSL certificate
func (m *Manager) RenewCertificate(domain string) error {
	// Get existing certificate
	var cert Certificate
	err := database.DB.QueryRow(`
		SELECT id, certificate, private_key, expire_at
		FROM ssl_certificates
		WHERE domain = ?
	`, domain).Scan(&cert.ID, &cert.Certificate, &cert.PrivateKey, &cert.ExpireAt)
	if err != nil {
		return fmt.Errorf("failed to get certificate: %v", err)
	}

	// Parse certificate
	certBlock, _ := pem.Decode([]byte(cert.Certificate))
	if certBlock == nil {
		return fmt.Errorf("failed to decode certificate")
	}
	x509Cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Check if renewal is needed
	if time.Now().Add(30 * 24 * time.Hour).Before(x509Cert.NotAfter) {
		log.Printf("Certificate for %s does not need renewal yet", domain)
		return nil
	}

	// Request new certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := m.client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("failed to renew certificate: %v", err)
	}

	// Parse new certificate
	newCertBlock, _ := pem.Decode(certificates.Certificate)
	if newCertBlock == nil {
		return fmt.Errorf("failed to decode new certificate")
	}
	newCert, err := x509.ParseCertificate(newCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse new certificate: %v", err)
	}

	// Update database
	_, err = database.DB.Exec(`
		UPDATE ssl_certificates
		SET certificate = ?, private_key = ?, expire_at = ?
		WHERE domain = ?
	`, string(certificates.Certificate), string(certificates.PrivateKey), newCert.NotAfter, domain)
	if err != nil {
		return fmt.Errorf("failed to update certificate: %v", err)
	}

	// Schedule next renewal
	m.scheduleRenewal(domain, newCert.NotAfter)

	log.Printf("Certificate renewed for domain: %s", domain)
	return nil
}

// scheduleRenewal schedules automatic renewal of a certificate
func (m *Manager) scheduleRenewal(domain string, expiry time.Time) {
	// Cancel existing renewal timer if any
	if timer, exists := m.renewals[domain]; exists {
		timer.Stop()
	}

	// Calculate renewal time (30 days before expiry)
	renewalTime := expiry.Add(-30 * 24 * time.Hour)
	if renewalTime.Before(time.Now()) {
		renewalTime = time.Now().Add(time.Hour)
	}

	// Schedule renewal
	timer := time.AfterFunc(time.Until(renewalTime), func() {
		if err := m.RenewCertificate(domain); err != nil {
			log.Printf("Failed to renew certificate for %s: %v", domain, err)
		}
	})
	m.renewals[domain] = timer
}

// GetCertificate retrieves a certificate from the database
func (m *Manager) GetCertificate(domain string) (*Certificate, error) {
	var cert Certificate
	err := database.DB.QueryRow(`
		SELECT id, domain, certificate, private_key, created_at, expire_at
		FROM ssl_certificates
		WHERE domain = ?
	`, domain).Scan(
		&cert.ID,
		&cert.Domain,
		&cert.Certificate,
		&cert.PrivateKey,
		&cert.CreatedAt,
		&cert.ExpireAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %v", err)
	}
	return &cert, nil
}

// ListCertificates returns all certificates
func (m *Manager) ListCertificates() ([]Certificate, error) {
	rows, err := database.DB.Query(`
		SELECT id, domain, certificate, private_key, created_at, expire_at
		FROM ssl_certificates
		ORDER BY domain
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %v", err)
	}
	defer rows.Close()

	var certs []Certificate
	for rows.Next() {
		var cert Certificate
		err := rows.Scan(
			&cert.ID,
			&cert.Domain,
			&cert.Certificate,
			&cert.PrivateKey,
			&cert.CreatedAt,
			&cert.ExpireAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan certificate: %v", err)
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

// DeleteCertificate deletes a certificate
func (m *Manager) DeleteCertificate(domain string) error {
	// Cancel renewal timer if exists
	if timer, exists := m.renewals[domain]; exists {
		timer.Stop()
		delete(m.renewals, domain)
	}

	// Delete from database
	result, err := database.DB.Exec("DELETE FROM ssl_certificates WHERE domain = ?", domain)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %v", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("certificate not found")
	}

	log.Printf("Certificate deleted for domain: %s", domain)
	return nil
}
