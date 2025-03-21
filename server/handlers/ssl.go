package handlers

import (
	"net/http"
	"time"

	"v/ssl"

	"github.com/gin-gonic/gin"
)

// SSLHandler handles SSL certificate management
type SSLHandler struct {
	certManager *ssl.CertificateManager
}

// NewSSLHandler creates a new SSL handler
func NewSSLHandler(certManager *ssl.CertificateManager) *SSLHandler {
	return &SSLHandler{
		certManager: certManager,
	}
}

// Certificate represents a certificate in the API
type Certificate struct {
	Domain      string    `json:"domain"`
	ExpireAt    time.Time `json:"expire_at"`
	AutoRenew   bool      `json:"auto_renew"`
	LastRenewed time.Time `json:"last_renewed"`
}

// HandleListCertificates lists all certificates
func (h *SSLHandler) HandleListCertificates(c *gin.Context) {
	certs, err := h.certManager.ListCertificates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list certificates",
		})
		return
	}

	certificates := make([]Certificate, len(certs))
	for i, cert := range certs {
		certificates[i] = Certificate{
			Domain:      cert.Domain,
			ExpireAt:    cert.ExpireAt,
			AutoRenew:   cert.AutoRenew,
			LastRenewed: cert.LastRenewed,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": certificates,
	})
}

// HandleGetCertificate gets a specific certificate
func (h *SSLHandler) HandleGetCertificate(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "domain is required",
		})
		return
	}

	cert, err := h.certManager.GetCertificate(domain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "certificate not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificate": Certificate{
			Domain:      cert.Domain,
			ExpireAt:    cert.ExpireAt,
			AutoRenew:   cert.AutoRenew,
			LastRenewed: cert.LastRenewed,
		},
	})
}

// HandleCreateCertificate creates a new certificate
func (h *SSLHandler) HandleCreateCertificate(c *gin.Context) {
	var req struct {
		Domain    string `json:"domain" binding:"required"`
		AutoRenew bool   `json:"auto_renew"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	cert, err := h.certManager.GetCertificate(req.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create certificate",
		})
		return
	}

	cert.AutoRenew = req.AutoRenew

	c.JSON(http.StatusCreated, gin.H{
		"certificate": Certificate{
			Domain:      cert.Domain,
			ExpireAt:    cert.ExpireAt,
			AutoRenew:   cert.AutoRenew,
			LastRenewed: cert.LastRenewed,
		},
	})
}

// HandleRenewCertificate renews a certificate
func (h *SSLHandler) HandleRenewCertificate(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "domain is required",
		})
		return
	}

	err := h.certManager.RenewCertificate(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to renew certificate",
		})
		return
	}

	cert, err := h.certManager.GetCertificate(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get renewed certificate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificate": Certificate{
			Domain:      cert.Domain,
			ExpireAt:    cert.ExpireAt,
			AutoRenew:   cert.AutoRenew,
			LastRenewed: cert.LastRenewed,
		},
	})
}

// HandleDeleteCertificate deletes a certificate
func (h *SSLHandler) HandleDeleteCertificate(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "domain is required",
		})
		return
	}

	err := h.certManager.DeleteCertificate(domain)
	if err != nil {
		if err.Error() == "certificate not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "certificate not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete certificate",
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
