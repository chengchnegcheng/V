package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"crypto"

	"github.com/go-acme/lego/v4/registration"
)

// User implements the registration.User interface
type User struct {
	Email        string
	Registration *registration.Resource
	key          *rsa.PrivateKey
}

// NewUser creates a new user for Let's Encrypt registration
func NewUser() (*User, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	return &User{
		Email: "admin@example.com", // TODO: Make configurable
		key:   privateKey,
	}, nil
}

// GetEmail returns the user's email
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the user's registration resource
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the user's private key
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// SavePrivateKey saves the private key to a file
func (u *User) SavePrivateKey(filename string) error {
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(u.key),
	})

	return os.WriteFile(filename, keyPEM, 0600)
}

// LoadPrivateKey loads the private key from a file
func (u *User) LoadPrivateKey(filename string) error {
	keyPEM, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %v", err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return fmt.Errorf("failed to decode private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	u.key = key
	return nil
}
