package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type CommentHandler struct {
	commentService *services.CommentService
}

func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
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

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Comment deleted",
	}))
}
