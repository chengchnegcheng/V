package handlers

import (
	"net/http"
	"strconv"

	"v/certificate"

	"github.com/gin-gonic/gin"
)

var certMgr = certificate.New()

// CreateCertificateRequest 创建证书请求
type CreateCertificateRequest struct {
	Domain     string `json:"domain" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	AutoRenew  bool   `json:"auto_renew"`
	Validation string `json:"validation" binding:"required,oneof=http dns"`
}

// HandleCreateCertificate 处理创建证书的请求
func HandleCreateCertificate(c *gin.Context) {
	var req CreateCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	// 创建证书
	cert, err := certMgr.Create(req.Domain, req.AutoRenew)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, cert)
}

// HandleGetCertificate 处理获取证书的请求
func HandleGetCertificate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid certificate ID",
		})
		return
	}

	// 获取证书
	cert, err := certMgr.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Certificate not found",
		})
		return
	}

	c.JSON(http.StatusOK, cert)
}

// HandleListCertificates 处理获取证书列表的请求
func HandleListCertificates(c *gin.Context) {
	certs, err := certMgr.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, certs)
}

// HandleDeleteCertificate 处理删除证书的请求
func HandleDeleteCertificate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid certificate ID",
		})
		return
	}

	// 删除证书
	if err := certMgr.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate deleted successfully",
	})
}

// HandleRenewCertificate 处理续期证书的请求
func HandleRenewCertificate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid certificate ID",
		})
		return
	}

	// 续期证书
	if err := certMgr.Renew(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate renewed successfully",
	})
}

// HandleValidateCertificate 处理验证证书的请求
func HandleValidateCertificate(c *gin.Context) {
	domain := c.Param("domain")
	if err := certMgr.Validate(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate validated successfully",
	})
}
