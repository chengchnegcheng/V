package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseBodyWriter 是一个自定义的响应体写入器
type ResponseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法以捕获响应体
func (r ResponseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// WriteString 重写WriteString方法以捕获响应体
func (r ResponseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

// Logger 是一个增强的日志中间件，记录请求和响应的详细信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 捕获请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应体缓冲区
		responseBodyWriter := &ResponseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseBodyWriter

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 状态码
		statusCode := c.Writer.Status()

		// 记录请求和响应信息
		logDetails := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   c.ClientIP(),
		}

		// 如果是/api/auth/login路径，记录请求和响应详情
		if c.Request.URL.Path == "/api/auth/login" {
			var reqJSON map[string]interface{}
			if json.Unmarshal(requestBody, &reqJSON) == nil {
				if username, ok := reqJSON["username"].(string); ok {
					logDetails["username"] = username
					// 不记录密码
					if _, hasPassword := reqJSON["password"]; hasPassword {
						logDetails["has_password"] = true
					}
				}
			}

			var respJSON map[string]interface{}
			responseBody := responseBodyWriter.body.Bytes()
			if json.Unmarshal(responseBody, &respJSON) == nil {
				// 不记录敏感信息如token
				hasToken := false
				if _, ok := respJSON["token"]; ok {
					hasToken = true
				}
				logDetails["response_has_token"] = hasToken

				// 记录是否有user字段
				hasUser := false
				if _, ok := respJSON["user"]; ok {
					hasUser = true
				}
				logDetails["response_has_user"] = hasUser

				// 记录错误信息
				if errMsg, ok := respJSON["error"].(string); ok {
					logDetails["error"] = errMsg
				}
			} else {
				// 如果不是有效的JSON，记录原始响应
				logDetails["raw_response"] = string(responseBody)
			}
		}

		// 使用log.Printf记录日志
		logJSON, _ := json.Marshal(logDetails)
		log.Printf("REQUEST LOG: %s", string(logJSON))
	}
}
