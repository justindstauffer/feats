package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type ReactionHandler struct {
	reactionService *services.ReactionService
}

func NewReactionHandler(reactionService *services.ReactionService) *ReactionHandler {
	return &ReactionHandler{
		reactionService: reactionService,
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

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Reaction removed",
	}))
}
