package api

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up all API routes
func SetupRoutes(r *gin.Engine) {
	// API group
	api := r.Group("/api")
	{
		// Proxy sharing/QR code routes
		api.GET("/proxy/:id/link", func(c *gin.Context) {
			id := c.Param("id")
			if id == "" {
				c.JSON(400, gin.H{"error": "Missing parameter"})
				return
			}

			// 根据代理ID生成分享链接
			var link string
			switch id {
			case "1", "2":
				// 标准格式：trojan://密码@服务器:端口?security=tls&sni=域名&allowInsecure=false#备注
				link = "trojan://" + url.QueryEscape("password123") + "@example.com:443?security=tls&sni=example.com&allowInsecure=false#" + url.QueryEscape("trojan"+id)
			case "3":
				link = "trojan://" + url.QueryEscape("password123") + "@example.com:60606?security=tls&sni=example.com&allowInsecure=false#" + url.QueryEscape("trojan")
			case "4":
				link = "trojan://" + url.QueryEscape("password123") + "@example.com:60605?security=tls&sni=example.com&allowInsecure=false#" + url.QueryEscape("trojan")
			default:
				c.JSON(404, gin.H{"error": "Proxy not found"})
				return
			}

			c.JSON(200, gin.H{"link": link})
		})

		api.GET("/proxy/:id/qrcode", func(c *gin.Context) {
			// 简化实现，返回链接以在前端生成QR码
			id := c.Param("id")
			if id == "" {
				c.JSON(400, gin.H{"error": "Missing parameter"})
				return
			}

			// 先获取分享链接
			var link string
			switch id {
			case "1", "2":
				// 标准格式：trojan://密码@服务器:端口?security=tls&sni=域名&allowInsecure=false#备注
				link = "trojan://" + url.QueryEscape("password123") + "@example.com:443?security=tls&sni=example.com&allowInsecure=false#" + url.QueryEscape("trojan"+id)
			case "3":
				link = "trojan://" + url.QueryEscape("password123") + "@example.com:60606?security=tls&sni=example.com&allowInsecure=false#" + url.QueryEscape("trojan")
			case "4":
				link = "trojan://" + url.QueryEscape("password123") + "@example.com:60605?security=tls&sni=example.com&allowInsecure=false#" + url.QueryEscape("trojan")
			default:
				c.JSON(404, gin.H{"error": "Proxy not found"})
				return
			}

			c.JSON(200, gin.H{
				"qrcode": link,
				"link":   link,
				"size":   256,
			})
		})

		// 新增入站链接API
		api.GET("/inbounds/:id/link", func(c *gin.Context) {
			id := c.Param("id")
			if id == "" {
				c.JSON(400, gin.H{"error": "Missing parameter"})
				return
			}

			// 根据入站ID生成分享链接
			var link string

			// 模拟数据
			mockedInbounds := map[string]gin.H{
				"1": {
					"protocol": "vmess",
					"port":     10086,
					"remark":   "VIP用户",
				},
				"2": {
					"protocol": "vless",
					"port":     20086,
					"remark":   "免费用户",
				},
				"3": {
					"protocol": "trojan",
					"port":     30086,
					"remark":   "测试节点",
					"settings": map[string]interface{}{
						"password": "password123",
						"sni":      "example.com",
					},
				},
			}

			inbound, exists := mockedInbounds[id]
			if !exists {
				c.JSON(404, gin.H{"error": "请求的资源不存在"})
				return
			}

			protocol, _ := inbound["protocol"].(string)
			port, _ := inbound["port"].(int)
			remark, _ := inbound["remark"].(string)

			switch protocol {
			case "trojan":
				settings, _ := inbound["settings"].(map[string]interface{})
				password := "password123"
				sni := "example.com"

				if settings != nil {
					if p, ok := settings["password"].(string); ok && p != "" {
						password = p
					}
					if s, ok := settings["sni"].(string); ok && s != "" {
						sni = s
					}
				}

				// 生成标准格式链接
				link = fmt.Sprintf("trojan://%s@example.com:%d?security=tls&sni=%s#%s",
					url.QueryEscape(password),
					port,
					url.QueryEscape(sni),
					url.QueryEscape(remark))
			case "vmess":
				// 简化的VMess链接
				link = fmt.Sprintf("vmess://eyJhZGQiOiJleGFtcGxlLmNvbSIsInBvcnQiOiIlZCIsImlkIjoiOGFkMzg4ZmYtOGQ4Mi00MThjLTljNDQtZmJiM2E1ODBjMWZiIiwibmV0IjoidGNwIiwicHMiOiIlcyJ9",
					port, url.QueryEscape(remark))
			case "vless":
				// 简化的VLESS链接
				link = fmt.Sprintf("vless://8ad388ff-8d82-418c-9c44-fbb3a580c1fb@example.com:%d?encryption=none&security=tls&type=tcp#%s",
					port, url.QueryEscape(remark))
			default:
				c.JSON(400, gin.H{"error": "不支持的协议类型"})
				return
			}

			c.JSON(200, gin.H{"link": link})
		})
	}
}
