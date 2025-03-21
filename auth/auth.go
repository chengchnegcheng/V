package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"v/logger"
	"v/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key") // 在实际应用中应该从配置文件读取
var db model.DB

// Init 初始化认证系统
func Init(database model.DB) {
	db = database
}

// Claims 自定义JWT声明
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// Manager handles authentication and authorization
type Manager struct {
	log    *logger.Logger
	tokens map[string]*Token
	db     model.DB
}

// Token represents an authentication token
type Token struct {
	UserID    int64
	Username  string
	IsAdmin   bool
	CreatedAt time.Time
	ExpiresAt time.Time
}

// New creates a new authentication manager
func New(log *logger.Logger, database model.DB) *Manager {
	return &Manager{
		log:    log,
		tokens: make(map[string]*Token),
		db:     database,
	}
}

// Login authenticates a user and returns a token
func (m *Manager) Login(username, password string) (string, error) {
	user, err := m.db.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !CheckPassword(password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return "", errors.New("account is locked")
	}

	// Reset login attempts on successful login
	user.LoginAttempts = 0
	if err := m.db.UpdateUser(user); err != nil {
		return "", err
	}

	// Generate JWT token
	token, err := GenerateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Logout invalidates a token
func (m *Manager) Logout(token string) error {
	// JWT tokens are stateless, so we don't need to do anything here
	return nil
}

// ValidateToken checks if a token is valid and returns the associated user info
func (m *Manager) ValidateToken(token string) (*Claims, error) {
	return ValidateToken(token)
}

// RequireAdmin checks if the user has admin privileges
func (m *Manager) RequireAdmin(token string) error {
	claims, err := m.ValidateToken(token)
	if err != nil {
		return err
	}

	if !claims.IsAdmin {
		return errors.New("admin privileges required")
	}

	return nil
}

// verifyPassword checks if a password matches the stored hash
func (m *Manager) verifyPassword(password, hash, salt string) bool {
	h := sha256.New()
	h.Write([]byte(password + salt))
	return hex.EncodeToString(h.Sum(nil)) == hash
}

// generateToken creates a new token string
func (m *Manager) generateToken(t *Token) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d:%s:%d", t.UserID, t.Username, time.Now().UnixNano())))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateToken 生成JWT令牌
func GenerateToken(user *model.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken 验证JWT令牌
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// HashPassword 对密码进行加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Login 用户登录
func Login(username, password string) (string, error) {
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return "", errors.New("user not found")
	}

	if !CheckPassword(password, user.Password) {
		return "", errors.New("invalid password")
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return "", errors.New("account is locked")
	}

	// 更新最后登录时间
	user.LastLoginAt = &time.Time{}
	*user.LastLoginAt = time.Now()
	if err := db.UpdateUser(user); err != nil {
		return "", err
	}

	token, err := GenerateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Register 用户注册
func Register(username, password, email string) error {
	// 检查用户名是否已存在
	_, err := db.GetUserByUsername(username)
	if err == nil {
		return errors.New("username already exists")
	}

	// 加密密码
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// 创建新用户
	user := &model.User{
		Username:     username,
		Password:     hashedPassword,
		Email:        email,
		IsAdmin:      false,
		TrafficLimit: 1024 * 1024 * 1024, // 默认1GB流量限制
		TrafficUsed:  0,
	}

	return db.CreateUser(user)
}
