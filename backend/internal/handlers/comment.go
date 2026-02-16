package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
	"github.com/jstauff/feats-api/internal/websocket"
)

type CommentHandler struct {
	commentService *services.CommentService
	pushService    *services.PushService
	wsHub          *websocket.Hub
}

func NewCommentHandler(commentService *services.CommentService, pushService *services.PushService, wsHub *websocket.Hub) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		pushService:    pushService,
		wsHub:          wsHub,
	}
}

func (h *CommentHandler) GetComments(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	postID := c.Param("id")

	comments, err := h.commentService.GetComments(groupID, postID)
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

	c.JSON(http.StatusOK, models.SuccessResponse(comments))
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	postID := c.Param("id")

	var input services.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	comment, err := h.commentService.CreateComment(groupID, postID, userID, input)
	if err != nil {
		switch err {
		case services.ErrPostNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
		case services.ErrCommentTooLong:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Comment exceeds maximum length",
			))
		default:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				err.Error(),
			))
		}
		return
	}

	// Broadcast comment.created event via WebSocket
	if h.wsHub != nil {
		user, _ := middleware.GetCurrentUser(c)
		payload := websocket.CommentCreatedPayload{
			CommentID: comment.ID,
			PostID:    postID,
			UserID:    userID,
			UserName:  user.Name,
			Content:   comment.Content,
			ParentID:  comment.ParentID,
		}
		if event, err := websocket.NewEvent(websocket.EventCommentCreated, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
	}

	// Send push notification to post owner (if not self)
	if h.pushService != nil {
		user, _ := middleware.GetCurrentUser(c)
		postOwnerID, err := h.commentService.GetPostOwnerID(comment.ID)
		if err == nil && postOwnerID != userID {
			go h.pushService.NotifyNewComment(postOwnerID, user.Name, comment.Content, postID)
		}
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(comment))
}

func (h *CommentHandler) UpdateComment(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	commentID := c.Param("id")

	var input services.UpdateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	comment, err := h.commentService.UpdateComment(groupID, commentID, userID, isAdmin, input)
	if err != nil {
		switch err {
		case services.ErrCommentNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Comment not found",
			))
		case services.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not authorized to update this comment",
			))
		case services.ErrCommentTooLong:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Comment exceeds maximum length",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(comment))
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	commentID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	// Get the comment first to know the postID for the event
	comment, _ := h.commentService.GetCommentByID(commentID)

	err := h.commentService.DeleteComment(groupID, commentID, userID, isAdmin)
	if err != nil {
		switch err {
		case services.ErrCommentNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Comment not found",
			))
		case services.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not authorized to delete this comment",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	// Broadcast comment.deleted event via WebSocket
	if h.wsHub != nil && comment != nil {
		payload := websocket.CommentDeletedPayload{
			CommentID: commentID,
			PostID:    comment.PostID,
		}
		if event, err := websocket.NewEvent(websocket.EventCommentDeleted, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Comment deleted",
	}))
}
