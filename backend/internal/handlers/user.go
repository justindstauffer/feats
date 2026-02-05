package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type UserHandler struct {
	userService  *services.UserService
	auditService *services.AuditService
	authService  *services.AuthService
}

func NewUserHandler(userService *services.UserService, auditService *services.AuditService, authService *services.AuthService) *UserHandler {
	return &UserHandler{
		userService:  userService,
		auditService: auditService,
		authService:  authService,
	}
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	user, _ := middleware.GetCurrentUser(c)
	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"User not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	var input services.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	user, err := h.userService.UpdateUser(userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(user))
}

// Admin handlers

func (h *UserHandler) CreateUser(c *gin.Context) {
	var input services.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	adminID, _ := middleware.GetCurrentUserID(c)

	user, err := h.userService.CreateUser(input, h.authService)
	if err != nil {
		switch err {
		case services.ErrEmailExists:
			c.JSON(http.StatusConflict, models.ErrorResponse(
				models.ErrCodeConflict,
				"Email already exists",
			))
		case services.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invalid role",
			))
		case services.ErrPasswordTooWeak:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				err.Error(),
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	h.auditService.LogUserCreated(adminID, user.ID, user.Email)

	c.JSON(http.StatusCreated, models.SuccessResponse(user))
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	includeDeleted := c.Query("include_deleted") == "true"

	users, err := h.userService.ListUsers(includeDeleted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(users))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	adminID, _ := middleware.GetCurrentUserID(c)

	// Prevent self-deletion
	if userID == adminID {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Cannot delete your own account",
		))
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"User not found",
			))
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			err.Error(),
		))
		return
	}

	h.auditService.LogUserDeleted(adminID, userID)

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "User deleted",
	}))
}

func (h *UserHandler) GetAuditLogs(c *gin.Context) {
	page := getPageParam(c)
	perPage := getPerPageParam(c)

	logs, total, err := h.auditService.GetLogs(page, perPage, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedSuccessResponse(logs, page, perPage, int(total)))
}
