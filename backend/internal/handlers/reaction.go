package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
	"github.com/jstauff/feats-api/internal/websocket"
)

type ReactionHandler struct {
	reactionService *services.ReactionService
	pushService     *services.PushService
	wsHub           *websocket.Hub
}

func NewReactionHandler(reactionService *services.ReactionService, pushService *services.PushService, wsHub *websocket.Hub) *ReactionHandler {
	return &ReactionHandler{
		reactionService: reactionService,
		pushService:     pushService,
		wsHub:           wsHub,
	}
}

func (h *ReactionHandler) GetReactions(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	postID := c.Param("id")

	summaries, reactions, err := h.reactionService.GetReactions(groupID, postID)
	if err != nil {
		if err == services.ErrPostNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
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
		"summary":   summaries,
		"reactions": reactions,
	}))
}

func (h *ReactionHandler) AddReaction(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	postID := c.Param("id")

	var input services.AddReactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	reaction, isNew, err := h.reactionService.AddReaction(groupID, postID, userID, models.ReactionType(input.ReactionType))
	if err != nil {
		if err == services.ErrInvalidReaction {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invalid reaction type",
			))
			return
		}
		if err == services.ErrPostNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	status := http.StatusOK
	if isNew {
		status = http.StatusCreated
	}

	// Broadcast reaction.added event via WebSocket
	if h.wsHub != nil {
		user, _ := middleware.GetCurrentUser(c)
		payload := websocket.ReactionPayload{
			PostID:       postID,
			UserID:       userID,
			UserName:     user.Name,
			ReactionType: reaction.ReactionType.String(),
		}
		if event, err := websocket.NewEvent(websocket.EventReactionAdded, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
	}

	c.JSON(status, models.SuccessResponse(reaction))
}

func (h *ReactionHandler) RemoveReaction(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	postID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)

	err := h.reactionService.RemoveReaction(groupID, postID, userID)
	if err != nil {
		if err == services.ErrReactionNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Reaction not found",
			))
			return
		}
		if err == services.ErrPostNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	// Broadcast reaction.removed event via WebSocket
	if h.wsHub != nil {
		user, _ := middleware.GetCurrentUser(c)
		payload := websocket.ReactionPayload{
			PostID:   postID,
			UserID:   userID,
			UserName: user.Name,
		}
		if event, err := websocket.NewEvent(websocket.EventReactionRemoved, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Reaction removed",
	}))
}
