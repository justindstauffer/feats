package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type ChallengeHandler struct {
	challengeService *services.ChallengeService
}

func NewChallengeHandler(challengeService *services.ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{
		challengeService: challengeService,
	}
}

func (h *ChallengeHandler) ListChallenges(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	includeExpired := c.Query("include_expired") == "true"

	challenges, err := h.challengeService.ListChallenges(groupID, includeExpired)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(challenges))
}

func (h *ChallengeHandler) GetChallenge(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	challengeID := c.Param("id")

	challenge, err := h.challengeService.GetChallengeByID(groupID, challengeID)
	if err != nil {
		if err == services.ErrChallengeNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Challenge not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(challenge))
}

func (h *ChallengeHandler) CreateChallenge(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)

	var input services.CreateChallengeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	challenge, err := h.challengeService.CreateChallenge(groupID, input, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(challenge))
}

func (h *ChallengeHandler) JoinChallenge(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	challengeID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)

	err := h.challengeService.JoinChallenge(groupID, challengeID, userID)
	if err != nil {
		switch err {
		case services.ErrChallengeNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Challenge not found",
			))
		case services.ErrAlreadyParticipating:
			c.JSON(http.StatusConflict, models.ErrorResponse(
				models.ErrCodeConflict,
				"Already participating in this challenge",
			))
		case services.ErrChallengeEnded:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Challenge has ended",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Joined challenge",
	}))
}

func (h *ChallengeHandler) LeaveChallenge(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	challengeID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)

	err := h.challengeService.LeaveChallenge(groupID, challengeID, userID)
	if err != nil {
		if err == services.ErrNotParticipating {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Not participating in this challenge",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Left challenge",
	}))
}

func (h *ChallengeHandler) DeleteChallenge(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	challengeID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	err := h.challengeService.DeleteChallenge(groupID, challengeID, userID, isAdmin)
	if err != nil {
		switch err {
		case services.ErrChallengeNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Challenge not found",
			))
		case services.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not authorized to delete this challenge",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Challenge deleted",
	}))
}
