package user

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"v/errors"
	"v/logger"
	"v/model"
	"v/settings"
)

// Manager represents a user manager
type Manager struct {
	log      *logger.Logger
	settings *settings.Manager
	db       model.DB
}

// New creates a new user manager
func New(log *logger.Logger, settings *settings.Manager, db model.DB) *Manager {
	return &Manager{
		log:      log,
		settings: settings,
		db:       db,
	}
}

// Create creates a new user
func (m *Manager) Create(username, email, password string) (*model.User, error) {
	// Validate input
	if err := m.validateInput(username, email, password); err != nil {
		return nil, err
	}

	// Check if username exists
	if _, err := m.GetByUsername(username); err == nil {
		return nil, errors.WithMessage(errors.ErrBadRequest, "Username already exists")
	}

	// Check if email exists
	if _, err := m.GetByEmail(email); err == nil {
		return nil, errors.WithMessage(errors.ErrBadRequest, "Email already exists")
	}

	// Generate salt
	salt, err := m.generateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	// Hash password
	hashedPassword, err := m.hashPassword(password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Get settings
	s := m.settings.Get()

	// Create user
	user := &model.User{
		Username:     username,
		Email:        email,
		Password:     hashedPassword,
		Salt:         salt,
		TrafficLimit: s.Traffic.DefaultLimit,
		ExpireAt:     &time.Time{},
		LastLoginAt:  &time.Time{},
	}

	// Save user to database
	if err := m.db.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	m.log.Info("User created", logger.Fields{
		"user_id":  user.ID,
		"username": username,
		"email":    email,
	})

	return user, nil
}

// Get returns a user by ID
func (m *Manager) Get(id int64) (*model.User, error) {
	user, err := m.db.GetUser(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.WithMessage(errors.ErrNotFound, "User not found")
	}
	return user, nil
}

// GetByUsername returns a user by username
func (m *Manager) GetByUsername(username string) (*model.User, error) {
	user, err := m.db.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.WithMessage(errors.ErrNotFound, "User not found")
	}
	return user, nil
}

// GetByEmail returns a user by email
func (m *Manager) GetByEmail(email string) (*model.User, error) {
	user, err := m.db.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.WithMessage(errors.ErrNotFound, "User not found")
	}
	return user, nil
}

// Update updates a user
func (m *Manager) Update(user *model.User) error {
	// Validate input
	if err := m.validateInput(user.Username, user.Email, ""); err != nil {
		return err
	}

	// Check if username exists (excluding current user)
	if existingUser, err := m.GetByUsername(user.Username); err == nil && existingUser.ID != user.ID {
		return errors.WithMessage(errors.ErrBadRequest, "Username already exists")
	}

	// Check if email exists (excluding current user)
	if existingUser, err := m.GetByEmail(user.Email); err == nil && existingUser.ID != user.ID {
		return errors.WithMessage(errors.ErrBadRequest, "Email already exists")
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Update user in database
	if err := m.db.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	m.log.Info("User updated", logger.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	})

	return nil
}

// Delete deletes a user
func (m *Manager) Delete(id int64) error {
	// Delete user from database
	if err := m.db.DeleteUser(id); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	m.log.Info("User deleted", logger.Fields{
		"user_id": id,
	})

	return nil
}

// Authenticate authenticates a user
func (m *Manager) Authenticate(username, password string) (*model.User, error) {
	// Get user
	user, err := m.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	// Check if user is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, errors.WithMessage(errors.ErrUnauthorized, "Account is locked")
	}

	// Verify password
	if err := m.verifyPassword(password, user.Password, user.Salt); err != nil {
		// Increment login attempts
		user.LoginAttempts++
		if user.LoginAttempts >= m.settings.Get().Security.LoginAttempts {
			lockedUntil := time.Now().Add(m.settings.Get().Security.LockoutTime)
			user.LockedUntil = &lockedUntil
		}
		// Update user in database
		if err := m.db.UpdateUser(user); err != nil {
			m.log.Error("Failed to update login attempts", logger.Fields{
				"error": err,
			})
		}
		return nil, errors.WithMessage(errors.ErrUnauthorized, "Invalid password")
	}

	// Reset login attempts
	user.LoginAttempts = 0
	lastLoginAt := time.Now()
	user.LastLoginAt = &lastLoginAt
	// Update user in database
	if err := m.db.UpdateUser(user); err != nil {
		m.log.Error("Failed to update last login", logger.Fields{
			"error": err,
		})
	}

	return user, nil
}

// ChangePassword changes a user's password
func (m *Manager) ChangePassword(id int64, oldPassword, newPassword string) error {
	// Get user
	user, err := m.Get(id)
	if err != nil {
		return err
	}

	// Verify old password
	if err := m.verifyPassword(oldPassword, user.Password, user.Salt); err != nil {
		return errors.WithMessage(errors.ErrUnauthorized, "Invalid old password")
	}

	// Generate new salt
	salt, err := m.generateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %v", err)
	}

	// Hash new password
	hashedPassword, err := m.hashPassword(newPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update password
	user.Password = hashedPassword
	user.Salt = salt
	user.UpdatedAt = time.Now()

	// Update user in database
	if err := m.db.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	m.log.Info("Password changed", logger.Fields{
		"user_id": id,
	})

	return nil
}

// ResetPassword resets a user's password
func (m *Manager) ResetPassword(id int64) (string, error) {
	// Get user
	user, err := m.Get(id)
	if err != nil {
		return "", err
	}

	// Generate new password
	newPassword, err := m.generatePassword()
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %v", err)
	}

	// Generate new salt
	salt, err := m.generateSalt()
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %v", err)
	}

	// Hash new password
	hashedPassword, err := m.hashPassword(newPassword, salt)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}

	// Update password
	user.Password = hashedPassword
	user.Salt = salt
	user.UpdatedAt = time.Now()

	// Update user in database
	if err := m.db.UpdateUser(user); err != nil {
		return "", fmt.Errorf("failed to update password: %v", err)
	}

	m.log.Info("Password reset", logger.Fields{
		"user_id": id,
	})

	return newPassword, nil
}

// Unlock unlocks a user's account
func (m *Manager) Unlock(id int64) error {
	// Get user
	user, err := m.Get(id)
	if err != nil {
		return err
	}

	// Reset login attempts and locked until
	user.LoginAttempts = 0
	user.LockedUntil = nil
	user.UpdatedAt = time.Now()

	// Update user in database
	if err := m.db.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to unlock user: %v", err)
	}

	m.log.Info("User unlocked", logger.Fields{
		"user_id": id,
	})

	return nil
}

// List returns a list of users
func (m *Manager) List(page, pageSize int) ([]*model.User, error) {
	return m.db.ListUsers((page-1)*pageSize, pageSize)
}

// Search searches for users
func (m *Manager) Search(query string, page, pageSize int) ([]*model.User, error) {
	// TODO: Implement search functionality
	return m.db.ListUsers((page-1)*pageSize, pageSize)
}

// validateInput validates user input
func (m *Manager) validateInput(username, email, password string) error {
	s := m.settings.Get()

	// Validate username
	if len(username) < 3 {
		return errors.WithMessage(errors.ErrBadRequest, "Username must be at least 3 characters")
	}

	// Validate email
	if !m.isValidEmail(email) {
		return errors.WithMessage(errors.ErrBadRequest, "Invalid email address")
	}

	// Validate password
	if password != "" && len(password) < s.Security.MinPasswordLength {
		return errors.WithFormat(errors.ErrBadRequest, "Password must be at least %d characters", s.Security.MinPasswordLength)
	}

	return nil
}

// isValidEmail validates an email address
func (m *Manager) isValidEmail(email string) bool {
	// Basic email validation
	if len(email) < 5 || len(email) > 254 {
		return false
	}

	// Check for @ symbol
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 {
		return false
	}

	// Check local part
	localPart := email[:atIndex]
	if len(localPart) < 1 || len(localPart) > 64 {
		return false
	}

	// Check domain part
	domain := email[atIndex+1:]
	if len(domain) < 1 || len(domain) > 255 {
		return false
	}

	// Check for dots in domain
	if strings.Count(domain, ".") < 1 {
		return false
	}

	return true
}

// generateSalt generates a random salt
func (m *Manager) generateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(salt), nil
}

// generatePassword generates a random password
func (m *Manager) generatePassword() (string, error) {
	password := make([]byte, 12)
	if _, err := rand.Read(password); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(password), nil
}

// hashPassword hashes a password with a salt
func (m *Manager) hashPassword(password, salt string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword verifies a password against a hash and salt
func (m *Manager) verifyPassword(password, hash, salt string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+salt))
}
