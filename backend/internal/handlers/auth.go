package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type AuthHandler struct {
	authService       *services.AuthService
	userService       *services.UserService
	betaInviteService *services.BetaInviteService
	auditService      *services.AuditService
	cfg               *config.Config
}

func NewAuthHandler(
	authService *services.AuthService,
	userService *services.UserService,
	betaInviteService *services.BetaInviteService,
	auditService *services.AuditService,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		authService:       authService,
		userService:       userService,
		betaInviteService: betaInviteService,
		auditService:      auditService,
		cfg:               cfg,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type PasswordResetRequestInput struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

type RegisterDeviceInput struct {
	DeviceToken string `json:"device_token" binding:"required"`
}

// Register handles user self-registration with a beta invite code
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	// Validate and consume the beta invite code
	_, err := h.betaInviteService.ValidateAndConsume(input.InviteCode)
	if err != nil {
		switch err {
		case services.ErrBetaInviteInvalid:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invalid invite code",
			))
		case services.ErrBetaInviteExpired:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invite code has expired",
			))
		case services.ErrBetaInviteMaxUses:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invite code has reached maximum uses",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	// Create the user
	user, err := h.userService.RegisterUser(input.Email, input.Password, input.Name, h.authService)
	if err != nil {
		switch err {
		case services.ErrEmailExists:
			c.JSON(http.StatusConflict, models.ErrorResponse(
				models.ErrCodeConflict,
				"Email already registered",
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

	ipHash := services.HashIP(c.ClientIP())
	userAgent := c.GetHeader("User-Agent")

	// Log the user in automatically after registration
	tokens, _, err := h.authService.Login(input.Email, input.Password, ipHash, userAgent)
	if err != nil {
		// User created but login failed - they can login manually
		c.JSON(http.StatusCreated, models.SuccessResponse(gin.H{
			"user":    user,
			"message": "Account created. Please login.",
		}))
		return
	}

	h.auditService.LogLogin(&user.ID, ipHash, userAgent, true, "registered")

	c.JSON(http.StatusCreated, models.SuccessResponse(gin.H{
		"tokens": tokens,
		"user":   user,
	}))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	ipHash := services.HashIP(c.ClientIP())
	userAgent := c.GetHeader("User-Agent")

	tokens, user, err := h.authService.Login(input.Email, input.Password, ipHash, userAgent)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			h.auditService.LogLoginFailed(input.Email, ipHash, userAgent, "invalid_credentials")
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeInvalidCredentials,
				"Invalid email or password",
			))
		case services.ErrAccountLocked:
			h.auditService.LogLoginFailed(input.Email, ipHash, userAgent, "account_locked")
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeAccountLocked,
				"Account is locked. Please try again later.",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	h.auditService.LogLogin(&user.ID, ipHash, userAgent, true, "success")

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"tokens": tokens,
		"user":   user,
	}))
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input RefreshRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	tokens, err := h.authService.RefreshToken(input.RefreshToken)
	if err != nil {
		switch err {
		case services.ErrTokenExpired:
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeTokenExpired,
				"Refresh token has expired",
			))
		case services.ErrTokenInvalid:
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeTokenInvalid,
				"Invalid refresh token",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(tokens))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	if err := h.authService.Logout(userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	h.auditService.LogLogout(userID)

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Logged out successfully",
	}))
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var input PasswordResetRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	// Always return success to prevent email enumeration
	_, _ = h.authService.CreatePasswordResetToken(input.Email)

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "If the email exists, a password reset has been initiated",
	}))
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input PasswordResetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	err := h.authService.ResetPassword(input.Token, input.NewPassword)
	if err != nil {
		switch err {
		case services.ErrTokenExpired, services.ErrTokenInvalid:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeTokenInvalid,
				"Invalid or expired reset token",
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

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Password reset successfully",
	}))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	err := h.authService.ChangePassword(userID, input.CurrentPassword, input.NewPassword)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeInvalidCredentials,
				"Current password is incorrect",
			))
		case services.ErrPasswordTooWeak:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				err.Error(),
			))
		case services.ErrPasswordReused:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Cannot reuse a recent password",
			))
		default:
			h.auditService.LogPasswordChange(userID, false)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	h.auditService.LogPasswordChange(userID, true)

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Password changed successfully",
	}))
}

func (h *AuthHandler) RegisterDevice(c *gin.Context) {
	var input RegisterDeviceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	// TODO: Implement device token registration for push notifications
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Device registered",
	}))
}

func (h *AuthHandler) UnregisterDevice(c *gin.Context) {
	// TODO: Implement device token unregistration
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Device unregistered",
	}))
}
