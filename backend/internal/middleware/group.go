package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type GroupMiddleware struct {
	groupService *services.GroupService
}

func NewGroupMiddleware(groupService *services.GroupService) *GroupMiddleware {
	return &GroupMiddleware{
		groupService: groupService,
	}
}

// RequireGroupMember validates that the current user is an active member of the group
func (m *GroupMiddleware) RequireGroupMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.Param("gid")
		if groupID == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Group ID is required",
			))
			c.Abort()
			return
		}

		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"Authentication required",
			))
			c.Abort()
			return
		}

		if !m.groupService.IsGroupMember(groupID, userID) {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not a member of this group",
			))
			c.Abort()
			return
		}

		// Store group ID in context for handlers
		c.Set("group_id", groupID)
		c.Next()
	}
}

// RequireGroupAdmin validates that the current user is an admin of the group
func (m *GroupMiddleware) RequireGroupAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.Param("gid")
		if groupID == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Group ID is required",
			))
			c.Abort()
			return
		}

		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.ErrCodeUnauthorized,
				"Authentication required",
			))
			c.Abort()
			return
		}

		if !m.groupService.IsGroupAdmin(groupID, userID) {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Admin access required for this group",
			))
			c.Abort()
			return
		}

		// Store group ID in context for handlers
		c.Set("group_id", groupID)
		c.Next()
	}
}

// GetCurrentGroupID extracts the group ID from the context
func GetCurrentGroupID(c *gin.Context) (string, bool) {
	groupID, exists := c.Get("group_id")
	if !exists {
		return "", false
	}
	return groupID.(string), true
}
