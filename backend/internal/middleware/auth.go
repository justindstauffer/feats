package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type AuthMiddleware struct {
	authService *services.AuthService
	cfg         *config.Config
}

func NewAuthMiddleware(authService *services.AuthService, cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		cfg:         cfg,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"Authorization header required",
			))
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid authorization header format",
			))
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeTokenInvalid,
				"Invalid or expired token",
			))
			c.Abort()
			return
		}

		// Get user from database
		user, err := m.authService.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"User not found",
			))
			c.Abort()
			return
		}

		// Check if user is locked
		if user.IsLocked() {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeAccountLocked,
				"Account is locked",
			))
			c.Abort()
			return
		}

		// Check if password change is required
		if user.ForcePasswordChange {
			// Allow only password change endpoint
			if c.Request.URL.Path != "/api/v1/auth/password/change" {
				c.JSON(http.StatusForbidden, models.ErrorResponse(
					models.ErrCodeForbidden,
					"Password change required",
				))
				c.Abort()
				return
			}
		}

		// Store user in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", string(user.Role))

		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != string(models.RoleAdmin) {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Admin access required",
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetCurrentUser extracts the current user from context
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	return user.(*models.User), true
}

// GetCurrentUserID extracts the current user ID from context
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return role == string(models.RoleAdmin)
}
