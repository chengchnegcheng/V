package cert

import (
	"os"
	"path/filepath"
	"testing"
	"v/model"
)

type mockDB struct {
	model.DB
	certificates map[string]*model.Certificate
}

func (db *mockDB) GetCertificateByDomain(domain string) (*model.Certificate, error) {
	if cert, ok := db.certificates[domain]; ok {
		return cert, nil
	}
	return nil, nil
}

func (db *mockDB) CreateCertificate(cert *model.Certificate) error {
	db.certificates[cert.Domain] = cert
	return nil
}

func (db *mockDB) DeleteCertificate(id int64) error {
	for domain, cert := range db.certificates {
		if cert.ID == id {
			delete(db.certificates, domain)
			return nil
		}
	}
	return nil
}

func TestCertificateManager_CreateCertificate(t *testing.T) {
	db := &mockDB{
		certificates: make(map[string]*model.Certificate),
	}
	manager := New(db)

	// Test creating a new certificate
	err := manager.CreateCertificate("example.com")
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	// Check if certificate was created in database
	cert, err := manager.GetCertificate("example.com")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}
	if cert == nil {
		t.Fatal("Certificate was not created")
	}
	if cert.Domain != "example.com" {
		t.Errorf("Expected domain to be example.com, got %s", cert.Domain)
	}

	// Check if certificate files exist
	if _, err := os.Stat(cert.CertFile); os.IsNotExist(err) {
		t.Errorf("Certificate file does not exist: %s", cert.CertFile)
	}
	if _, err := os.Stat(cert.KeyFile); os.IsNotExist(err) {
		t.Errorf("Private key file does not exist: %s", cert.KeyFile)
	}

	// Clean up
	err = manager.DeleteCertificate("example.com")
	if err != nil {
		t.Fatalf("DeleteCertificate failed: %v", err)
	}
}

func TestCertificateManager_GetCertificate(t *testing.T) {
	db := &mockDB{
		certificates: make(map[string]*model.Certificate),
	}
	manager := New(db)

	// Test getting non-existent certificate
	cert, err := manager.GetCertificate("nonexistent.com")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}
	if cert != nil {
		t.Error("Expected certificate to be nil")
	}

	// Create a certificate
	err = manager.CreateCertificate("example.com")
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	// Test getting existing certificate
	cert, err = manager.GetCertificate("example.com")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}
	if cert == nil {
		t.Fatal("Certificate was not found")
	}
	if cert.Domain != "example.com" {
		t.Errorf("Expected domain to be example.com, got %s", cert.Domain)
	}

	// Clean up
	err = manager.DeleteCertificate("example.com")
	if err != nil {
		t.Fatalf("DeleteCertificate failed: %v", err)
	}
}

func TestCertificateManager_DeleteCertificate(t *testing.T) {
	db := &mockDB{
		certificates: make(map[string]*model.Certificate),
	}
	manager := New(db)

	// Create a certificate
	err := manager.CreateCertificate("example.com")
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	// Test deleting certificate
	err = manager.DeleteCertificate("example.com")
	if err != nil {
		t.Fatalf("DeleteCertificate failed: %v", err)
	}

	// Check if certificate was deleted from database
	cert, err := manager.GetCertificate("example.com")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}
	if cert != nil {
		t.Error("Certificate was not deleted")
	}

	// Check if certificate files were deleted
	if _, err := os.Stat(filepath.Join("certificates", "example.com")); !os.IsNotExist(err) {
		t.Error("Certificate directory was not deleted")
	}
}

func TestCertificateManager_LoadCertificate(t *testing.T) {
	db := &mockDB{
		certificates: make(map[string]*model.Certificate),
	}
	manager := New(db)

	// Create a certificate
	err := manager.CreateCertificate("example.com")
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	// Test loading certificate
	tlsCert, err := manager.LoadCertificate("example.com")
	if err != nil {
		t.Fatalf("LoadCertificate failed: %v", err)
	}
	if tlsCert == nil {
		t.Fatal("TLS certificate was not loaded")
	}

	// Clean up
	err = manager.DeleteCertificate("example.com")
	if err != nil {
		t.Fatalf("DeleteCertificate failed: %v", err)
	}
}

func TestCertificateManager_RenewCertificate(t *testing.T) {
	db := &mockDB{
		certificates: make(map[string]*model.Certificate),
	}
	manager := New(db)

	// Create a certificate
	err := manager.CreateCertificate("example.com")
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	// Get original certificate
	originalCert, err := manager.GetCertificate("example.com")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}

	// Test renewing certificate
	err = manager.RenewCertificate("example.com")
	if err != nil {
		t.Fatalf("RenewCertificate failed: %v", err)
	}

	// Get renewed certificate
	renewedCert, err := manager.GetCertificate("example.com")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}

	// Check if certificate was renewed
	if renewedCert.IssuedAt.Before(originalCert.IssuedAt) {
		t.Error("Certificate was not renewed")
	}

	// Clean up
	err = manager.DeleteCertificate("example.com")
	if err != nil {
		t.Fatalf("DeleteCertificate failed: %v", err)
	}
}
