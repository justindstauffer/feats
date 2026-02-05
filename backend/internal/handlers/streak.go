package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type StreakHandler struct {
	streakService *services.StreakService
}

func NewStreakHandler(streakService *services.StreakService) *StreakHandler {
	return &StreakHandler{
		streakService: streakService,
	}
}

func (h *StreakHandler) GetUserStreak(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	userID := c.Param("id")

	streak, err := h.streakService.GetUserStreak(groupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(streak))
}

func (h *StreakHandler) GetLeaderboard(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)

	streaks, err := h.streakService.GetLeaderboard(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(streaks))
}
