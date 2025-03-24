package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"v/database"
	"v/logger"
	"v/model"
	"v/protocol"
	"v/settings"

	"github.com/gin-gonic/gin"
)

var (
	protocolLogger *logger.Logger
	protocolMgr    *protocol.Manager
)

func init() {
	db := database.GetWrappedDB()
	protocolLogger = logger.NewLogger()
	settingsMgr := settings.New(protocolLogger)
	protocolMgr = protocol.New(protocolLogger, settingsMgr, db)
}

// HandleCreateProtocol handles the creation of a new protocol
func HandleCreateProtocol(c *gin.Context) {
	var req CreateProtocolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Parse the settings string into an interface{}
	var settings interface{}
	if err := json.Unmarshal([]byte(req.Settings), &settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings format"})
		return
	}

	userID := c.GetInt64("user_id")

	// Create a new Protocol object
	protocol := &model.Protocol{
		UserID: userID,
		Type:   req.Type,
		Name:   req.Name,
		Port:   req.Port,
	}

	// Set the settings as JSON
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize settings"})
		return
	}
	protocol.Settings = settingsJSON

	// Create the protocol
	if err := protocolMgr.Create(protocol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, protocol)
}

// HandleGetProtocol handles getting a protocol by ID
func HandleGetProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protocol ID"})
		return
	}

	protocol, err := protocolMgr.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Protocol not found"})
		return
	}

	c.JSON(http.StatusOK, protocol)
}

// HandleListProtocols handles listing all protocols
func HandleListProtocols(c *gin.Context) {
	userID := c.GetInt64("user_id")
	protocols, err := protocolMgr.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list protocols"})
		return
	}

	c.JSON(http.StatusOK, protocols)
}

// HandleUpdateProtocol handles updating a protocol
func HandleUpdateProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protocol ID"})
		return
	}

	var req UpdateProtocolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	protocol, err := protocolMgr.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Protocol not found"})
		return
	}

	// Update protocol fields if provided
	if req.Name != nil {
		protocol.Name = *req.Name
	}
	if req.Port != nil {
		protocol.Port = *req.Port
	}
	if req.Settings != nil {
		protocol.Settings = []byte(*req.Settings)
	}
	if req.Enable != nil {
		protocol.Enable = *req.Enable
	}
	if req.TrafficLimit != nil {
		protocol.TrafficLimit = *req.TrafficLimit
	}

	if err := protocolMgr.Update(protocol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, protocol)
}

// HandleDeleteProtocol handles deleting a protocol
func HandleDeleteProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protocol ID"})
		return
	}

	if err := protocolMgr.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Protocol deleted successfully"})
}

// HandleEnableProtocol handles enabling a protocol
func HandleEnableProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protocol ID"})
		return
	}

	if err := protocolMgr.Enable(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Protocol enabled successfully"})
}

// HandleDisableProtocol handles disabling a protocol
func HandleDisableProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protocol ID"})
		return
	}

	protocol, err := protocolMgr.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Protocol not found"})
		return
	}

	protocol.Enable = false
	if err := protocolMgr.Update(protocol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Protocol disabled successfully"})
}
