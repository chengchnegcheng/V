package model

import "errors"

// 系统错误
var (
	// ErrNotFound 未找到记录
	ErrNotFound = errors.New("record not found")

	// ErrInvalidCredentials 无效的凭证
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserExists 用户已存在
	ErrUserExists = errors.New("user already exists")

	// ErrEmailExists 邮箱已存在
	ErrEmailExists = errors.New("email already exists")

	// ErrAccountLocked 账号已锁定
	ErrAccountLocked = errors.New("account is locked")

	// ErrInvalidToken 无效的令牌
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken 令牌已过期
	ErrExpiredToken = errors.New("token has expired")

	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInvalidData 无效的数据
	ErrInvalidData = errors.New("invalid data")

	// ErrPortInUse 端口已占用
	ErrPortInUse = errors.New("port already in use")

	// ErrDomainExists 域名已存在
	ErrDomainExists = errors.New("domain already exists")

	// ErrCertificateFailed 证书申请失败
	ErrCertificateFailed = errors.New("certificate request failed")

	// ErrTrafficLimitExceeded 流量限制超过
	ErrTrafficLimitExceeded = errors.New("traffic limit exceeded")

	// ErrAccountExpired 账号已过期
	ErrAccountExpired = errors.New("account has expired")

	// ErrUnsupportedProtocol 不支持的协议
	ErrUnsupportedProtocol = errors.New("unsupported protocol")
)
