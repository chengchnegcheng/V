package errors

import (
	"fmt"
	"net/http"
)

// Error represents an API error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error returns the error message
func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new error
func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// 常见错误
var (
	// 通用错误
	ErrInternalServerError = NewError(http.StatusInternalServerError, "内部服务器错误")
	ErrBadRequest          = NewError(http.StatusBadRequest, "无效的请求")
	ErrUnauthorized        = NewError(http.StatusUnauthorized, "未经授权的访问")
	ErrForbidden           = NewError(http.StatusForbidden, "禁止访问")
	ErrNotFound            = NewError(http.StatusNotFound, "资源不存在")
	ErrMethodNotAllowed    = NewError(http.StatusMethodNotAllowed, "方法不允许")
	ErrConflict            = NewError(http.StatusConflict, "资源冲突")
	ErrResourceGone        = NewError(http.StatusGone, "资源不可用")
	ErrTooManyRequests     = NewError(http.StatusTooManyRequests, "请求过多")

	// 通用API错误
	ErrInvalidRequestBody  = NewError(http.StatusBadRequest, "无效的请求体")
	ErrMissingParameter    = NewError(http.StatusBadRequest, "缺少必要参数")
	ErrInvalidParameter    = NewError(http.StatusBadRequest, "无效的参数")
	ErrResourceNotFound    = NewError(http.StatusNotFound, "请求的资源不存在")
	ErrResourceExists      = NewError(http.StatusConflict, "资源已存在")
	ErrResourceUnavailable = NewError(http.StatusServiceUnavailable, "资源暂时不可用")

	// 认证错误
	ErrInvalidCredentials = NewError(http.StatusUnauthorized, "无效的凭据")
	ErrTokenExpired       = NewError(http.StatusUnauthorized, "令牌已过期")
	ErrInvalidToken       = NewError(http.StatusUnauthorized, "无效的令牌")
	ErrAccessDenied       = NewError(http.StatusForbidden, "拒绝访问")

	// 数据库错误
	ErrDatabaseConnection = NewError(http.StatusInternalServerError, "数据库连接错误")
	ErrDatabaseQuery      = NewError(http.StatusInternalServerError, "数据库查询错误")
	ErrDatabaseInsert     = NewError(http.StatusInternalServerError, "数据库插入错误")
	ErrDatabaseUpdate     = NewError(http.StatusInternalServerError, "数据库更新错误")
	ErrDatabaseDelete     = NewError(http.StatusInternalServerError, "数据库删除错误")

	// Xray错误
	ErrXrayVersionNotFound = NewError(http.StatusNotFound, "指定的Xray版本不存在")
	ErrXrayDownloadFailed  = NewError(http.StatusInternalServerError, "下载Xray版本失败")
	ErrXrayStartFailed     = NewError(http.StatusInternalServerError, "启动Xray失败")
	ErrXrayStopFailed      = NewError(http.StatusInternalServerError, "停止Xray失败")
	ErrXrayAlreadyRunning  = NewError(http.StatusConflict, "Xray已经在运行")
	ErrXrayNotRunning      = NewError(http.StatusConflict, "Xray未在运行")
)

// WithMessage returns a new error with the given message
func WithMessage(err *Error, message string) *Error {
	return NewError(err.Code, message)
}

// WithFormat returns a new error with a formatted message
func WithFormat(err *Error, format string, args ...interface{}) *Error {
	return NewError(err.Code, fmt.Sprintf(format, args...))
}
