package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type PushHandler struct {
	pushService *services.PushService
}

func NewPushHandler(pushService *services.PushService) *PushHandler {
	return &PushHandler{
		pushService: pushService,
	}
}

type RegisterTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=ios android"`
}

// RegisterToken registers a device token for push notifications
func (h *PushHandler) RegisterToken(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrCodeUnauthorized,
			"Unauthorized",
		))
		return
	}

	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	if err := h.pushService.RegisterToken(userID, req.Token, req.Platform); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"Failed to register device token",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Device token registered",
	}))
}

// UnregisterToken removes a device token
func (h *PushHandler) UnregisterToken(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.ErrCodeUnauthorized,
			"Unauthorized",
		))
		return
	}

	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	if err := h.pushService.UnregisterToken(userID, req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"Failed to unregister device token",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Device token unregistered",
	}))
}
