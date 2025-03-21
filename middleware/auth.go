package middleware

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims JWT claims
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// generateToken 生成JWT令牌
func generateToken(userID int64, username string, isAdmin bool, expiration time.Duration, secret string) (string, error) {
	// 设置JWT claims
	claims := Claims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并返回JWT令牌字符串
	return token.SignedString([]byte(secret))
}

// validateToken 验证JWT令牌
func validateToken(tokenString string, secret string) (*Claims, error) {
	// 解析JWT令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌并提取claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// refreshToken 刷新JWT令牌
func refreshToken(oldTokenString string, expiration time.Duration, secret string) (string, error) {
	// 验证旧令牌
	claims, err := validateToken(oldTokenString, secret)
	if err != nil {
		return "", err
	}

	// 生成新令牌
	return generateToken(claims.UserID, claims.Username, claims.IsAdmin, expiration, secret)
}
