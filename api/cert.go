package api

import (
	"net/http"

	"v/cert"

	"github.com/gin-gonic/gin"
)

// CertificateHandler SSL证书处理器
type CertificateHandler struct {
	certManager *cert.CertManager
}

// NewCertificateHandler 创建SSL证书处理器
func NewCertificateHandler(certManager *cert.CertManager) *CertificateHandler {
	return &CertificateHandler{
		certManager: certManager,
	}
}

// RegisterRoutes 注册路由
func (h *CertificateHandler) RegisterRoutes(router *gin.RouterGroup) {
	certRouter := router.Group("/certificates")
	{
		certRouter.GET("", h.ListCertificates)
		certRouter.GET("/:domain", h.GetCertificate)
		certRouter.POST("", h.CreateCertificate)
		certRouter.DELETE("/:domain", h.DeleteCertificate)
		certRouter.POST("/:domain/renew", h.RenewCertificate)
	}
}

// ListCertificates 获取所有证书
func (h *CertificateHandler) ListCertificates(c *gin.Context) {
	certs := h.certManager.ListCertificates()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    certs,
	})
}

// GetCertificate 获取证书
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	domain := c.Param("domain")
	cert, err := h.certManager.GetCertificate(domain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Certificate not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cert,
	})
}

// CertificateRequest 证书请求
type CertificateRequest struct {
	Domain string `json:"domain" binding:"required"`
}

// CreateCertificate 创建证书
func (h *CertificateHandler) CreateCertificate(c *gin.Context) {
	var req CertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	cert, err := h.certManager.CreateCertificate(req.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cert,
	})
}

// DeleteCertificate 删除证书
func (h *CertificateHandler) DeleteCertificate(c *gin.Context) {
	domain := c.Param("domain")
	if err := h.certManager.DeleteACMECertificate(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Certificate deleted",
	})
}

// RenewCertificate 更新证书
func (h *CertificateHandler) RenewCertificate(c *gin.Context) {
	domain := c.Param("domain")
	_, err := h.certManager.GetCertificate(domain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Certificate not found",
		})
		return
	}

	if err := h.certManager.RenewCertificate(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Certificate renewed",
	})
}
