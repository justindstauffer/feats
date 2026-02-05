package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type BetaInviteHandler struct {
	betaInviteService *services.BetaInviteService
}

func NewBetaInviteHandler(betaInviteService *services.BetaInviteService) *BetaInviteHandler {
	return &BetaInviteHandler{
		betaInviteService: betaInviteService,
	}
}

// CreateBetaInvite creates a new beta invite code (admin only)
func (h *BetaInviteHandler) CreateBetaInvite(c *gin.Context) {
	var req models.CreateBetaInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body with defaults
		req = models.CreateBetaInviteRequest{}
	}

	userID, _ := middleware.GetCurrentUserID(c)

	invite, err := h.betaInviteService.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create beta invite",
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(invite))
}

// ListBetaInvites returns all beta invites (admin only)
func (h *BetaInviteHandler) ListBetaInvites(c *gin.Context) {
	invites, err := h.betaInviteService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list beta invites",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(invites))
}

// GetBetaInvite returns a specific beta invite (admin only)
func (h *BetaInviteHandler) GetBetaInvite(c *gin.Context) {
	inviteID := c.Param("id")

	invite, err := h.betaInviteService.GetByID(inviteID)
	if err != nil {
		if err == services.ErrBetaInviteNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Beta invite not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get beta invite",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(invite))
}

// DeleteBetaInvite deletes a beta invite (admin only)
func (h *BetaInviteHandler) DeleteBetaInvite(c *gin.Context) {
	inviteID := c.Param("id")

	if err := h.betaInviteService.Delete(inviteID); err != nil {
		if err == services.ErrBetaInviteNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Beta invite not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"Failed to delete beta invite",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Beta invite deleted",
	}))
}
