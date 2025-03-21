package middleware

import (
	"net/http"
	"time"

	"v/logger"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

// Middleware 定义HTTP中间件函数类型
type Middleware func(http.Handler) http.Handler

// ToMuxMiddleware 将标准中间件转换为mux中间件
func ToMuxMiddleware(middleware Middleware) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return middleware(next)
	}
}

// Logging 日志中间件
func Logging(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 包装ResponseWriter以捕获状态码
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// 处理请求
			next.ServeHTTP(wrapped, r)

			// 记录请求信息
			duration := time.Since(start)
			log.Info("Request completed", logger.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     wrapped.statusCode,
				"duration":   duration,
				"client_ip":  r.RemoteAddr,
				"user_agent": r.UserAgent(),
			})
		})
	}
}

// Recovery 异常恢复中间件
func Recovery(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("Panic recovered", logger.Fields{
						"method": r.Method,
						"path":   r.URL.Path,
						"error":  err,
					})
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS 跨域资源共享中间件
func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// 速率限制器
var limiter = rate.NewLimiter(1, 5) // 默认1个请求/秒，突发最多5个请求

// RateLimit 速率限制中间件
func RateLimit() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter 是对http.ResponseWriter的包装，用于捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 重写WriteHeader方法以捕获状态码
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
