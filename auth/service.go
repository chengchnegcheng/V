package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"time"
	"v/model"
	"v/notification"
	"v/settings"

	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrAccountExpired     = errors.New("account has expired")
	ErrAccountDisabled    = errors.New("account is disabled")
)

// Service handles authentication and authorization
type Service struct {
	db           model.DB
	notification *notification.Manager
	settings     *settings.Manager
}

// NewAuthService creates a new authentication service
func NewAuthService(db model.DB, notification *notification.Manager, settings *settings.Manager) *Service {
	return &Service{
		db:           db,
		notification: notification,
		settings:     settings,
	}
}

// Login authenticates a user and returns a session token
func (s *Service) Login(username, password string) (*model.User, string, error) {
	// Get user by username
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Check if account is locked
	if user.LockedUntil.After(time.Now()) {
		return nil, "", ErrAccountLocked
	}

	// Check if account is expired
	if !user.ExpireAt.IsZero() && user.ExpireAt.Before(time.Now()) {
		return nil, "", ErrAccountExpired
	}

	// Check if account is enabled
	if !user.Enabled {
		return nil, "", ErrAccountDisabled
	}

	// Verify password
	if !s.verifyPassword(password, user.Password, user.Salt) {
		// Increment login attempts
		user.LoginAttempts++
		if user.LoginAttempts >= 5 {
			// Lock account for 30 minutes
			user.LockedUntil = time.Now().Add(30 * time.Minute)
		}
		s.db.UpdateUser(user)
		return nil, "", ErrInvalidCredentials
	}

	// Reset login attempts on successful login
	user.LoginAttempts = 0
	user.LastLoginAt = time.Now()
	s.db.UpdateUser(user)

	// Generate session token
	token := s.generateToken()

	return user, token, nil
}

// Logout invalidates a session token
func (s *Service) Logout(token string) error {
	// TODO: Implement token invalidation
	return nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.db.GetUser(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !s.verifyPassword(oldPassword, user.Password, user.Salt) {
		return ErrInvalidCredentials
	}

	// Generate new salt and hash password
	salt := s.generateSalt()
	hashedPassword := s.hashPassword(newPassword, salt)

	// Update user password
	user.Password = hashedPassword
	user.Salt = salt
	return s.db.UpdateUser(user)
}

// ResetPassword resets a user's password
func (s *Service) ResetPassword(userID int64) error {
	user, err := s.db.GetUser(userID)
	if err != nil {
		return err
	}

	// Generate new password
	newPassword := s.generateRandomPassword()
	salt := s.generateSalt()
	hashedPassword := s.hashPassword(newPassword, salt)

	// Update user password
	user.Password = hashedPassword
	user.Salt = salt
	if err := s.db.UpdateUser(user); err != nil {
		return err
	}

	// Send new password to user's email
	notification := &notification.Notification{
		To:      []string{user.Email},
		Subject: "Password Reset",
		Body: fmt.Sprintf(`
			<p>Dear %s,</p>
			<p>Your password has been reset. Here is your new password:</p>
			<p><strong>%s</strong></p>
			<p>Please change your password after logging in.</p>
			<p>Best regards,<br>%s</p>
		`, user.Username, newPassword, s.settings.Get().Site.Name),
		Type: "password_reset",
	}

	return s.notification.Send(notification)
}

// verifyPassword checks if a password matches the stored hash
func (s *Service) verifyPassword(password, hash, salt string) bool {
	hashedPassword := s.hashPassword(password, salt)
	return hashedPassword == hash
}

// hashPassword hashes a password using PBKDF2
func (s *Service) hashPassword(password, salt string) string {
	key := pbkdf2.Key([]byte(password), []byte(salt), 4096, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(key)
}

// generateSalt generates a random salt
func (s *Service) generateSalt() string {
	salt := make([]byte, 16)
	rand.Read(salt)
	return base64.StdEncoding.EncodeToString(salt)
}

// generateToken generates a random session token
func (s *Service) generateToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return base64.URLEncoding.EncodeToString(token)
}

// generateRandomPassword generates a random password
func (s *Service) generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, 12)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
