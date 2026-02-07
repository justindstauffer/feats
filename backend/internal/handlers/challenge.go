package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
	"github.com/jstauff/feats-api/internal/websocket"
)

type ChallengeHandler struct {
	challengeService *services.ChallengeService
	wsHub            *websocket.Hub
}

func NewChallengeHandler(challengeService *services.ChallengeService, wsHub *websocket.Hub) *ChallengeHandler {
	return &ChallengeHandler{
		challengeService: challengeService,
		wsHub:            wsHub,
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

	// Broadcast challenge.created event via WebSocket
	if h.wsHub != nil {
		user, _ := middleware.GetCurrentUser(c)
		activityName := ""
		if challenge.ActivityType != nil {
			activityName = challenge.ActivityType.Name
		}
		payload := websocket.ChallengeCreatedPayload{
			ChallengeID:  challenge.ID,
			Name:         challenge.Title,
			CreatorID:    userID,
			CreatorName:  user.Name,
			ActivityName: activityName,
			TargetCount:  challenge.TargetCount,
		}
		if event, err := websocket.NewEvent(websocket.EventChallengeCreated, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
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

	// Broadcast challenge.joined event via WebSocket
	if h.wsHub != nil {
		user, _ := middleware.GetCurrentUser(c)
		challenge, _ := h.challengeService.GetChallengeByID(groupID, challengeID)
		challengeName := ""
		if challenge != nil {
			challengeName = challenge.Title
		}
		payload := websocket.ChallengeJoinedPayload{
			ChallengeID:   challengeID,
			ChallengeName: challengeName,
			UserID:        userID,
			UserName:      user.Name,
		}
		if event, err := websocket.NewEvent(websocket.EventChallengeJoined, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Joined challenge",
	}))
}

func (h *ChallengeHandler) LeaveChallenge(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	challengeID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)

	// Get challenge info before leaving for the event
	challenge, _ := h.challengeService.GetChallengeByID(groupID, challengeID)

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

	// Broadcast challenge.left event via WebSocket
	if h.wsHub != nil && challenge != nil {
		payload := websocket.ChallengeLeftPayload{
			ChallengeID:   challengeID,
			ChallengeName: challenge.Title,
			UserID:        userID,
		}
		if event, err := websocket.NewEvent(websocket.EventChallengeLeft, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
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
