package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
	"github.com/jstauff/feats-api/internal/websocket"
)

type GroupHandler struct {
	groupService *services.GroupService
	auditService *services.AuditService
	wsHub        *websocket.Hub
}

func NewGroupHandler(groupService *services.GroupService, auditService *services.AuditService, wsHub *websocket.Hub) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		auditService: auditService,
		wsHub:        wsHub,
	}
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var input services.CreateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	group, err := h.groupService.CreateGroup(input, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(group))
}

// ListGroups returns all groups the user is a member of
func (h *GroupHandler) ListGroups(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	groups, err := h.groupService.ListUserGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(groups))
}

// GetGroup returns a single group
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)

	group, err := h.groupService.GetGroupByID(groupID)
	if err != nil {
		if err == services.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Group not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(group))
}

// UpdateGroup updates a group's details (admin only)
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)

	var input services.UpdateGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	group, err := h.groupService.UpdateGroup(groupID, input)
	if err != nil {
		if err == services.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Group not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(group))
}

// DeleteGroup deletes a group (admin only)
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)

	if err := h.groupService.DeleteGroup(groupID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Group deleted",
	}))
}

// LeaveGroup removes the current user from the group
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	userID, _ := middleware.GetCurrentUserID(c)
	user, _ := middleware.GetCurrentUser(c)

	if err := h.groupService.LeaveGroup(groupID, userID); err != nil {
		switch err {
		case services.ErrNotGroupMember:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Not a member of this group",
			))
		case services.ErrCannotLeaveAsAdmin:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Cannot leave as the only admin. Promote another member first.",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	// Broadcast member.left event via WebSocket
	if h.wsHub != nil {
		payload := websocket.MemberLeftPayload{
			UserID:   userID,
			UserName: user.Name,
		}
		if event, err := websocket.NewEvent(websocket.EventMemberLeft, groupID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Left group",
	}))
}

// ListMembers returns all members of a group
func (h *GroupHandler) ListMembers(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)

	members, err := h.groupService.ListMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(members))
}

// UpdateMember updates a member's role (admin only)
func (h *GroupHandler) UpdateMember(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	targetUserID := c.Param("uid")

	var input struct {
		Role models.GroupRole `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	if input.Role != models.GroupRoleAdmin && input.Role != models.GroupRoleMember {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid role. Must be 'admin' or 'member'",
		))
		return
	}

	if err := h.groupService.UpdateMemberRole(groupID, targetUserID, input.Role); err != nil {
		switch err {
		case services.ErrNotGroupMember:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"User is not a member of this group",
			))
		default:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				err.Error(),
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Member role updated",
	}))
}

// RemoveMember removes a member from the group (admin only)
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	targetUserID := c.Param("uid")
	adminUserID, _ := middleware.GetCurrentUserID(c)

	if err := h.groupService.RemoveMember(groupID, targetUserID, adminUserID); err != nil {
		switch err {
		case services.ErrNotGroupMember:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"User is not a member of this group",
			))
		case services.ErrCannotRemoveSelf:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Cannot remove yourself. Use leave instead.",
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
		"message": "Member removed",
	}))
}

// CreateInvite creates a new invite code (admin only)
func (h *GroupHandler) CreateInvite(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	userID, _ := middleware.GetCurrentUserID(c)

	var input services.CreateInviteInput
	// Don't require body - use defaults if not provided
	c.ShouldBindJSON(&input)

	invite, err := h.groupService.CreateInvite(groupID, userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(invite))
}

// ListInvites returns all invites for a group (admin only)
func (h *GroupHandler) ListInvites(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)

	invites, err := h.groupService.ListInvites(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(invites))
}

// RevokeInvite deletes an invite (admin only)
func (h *GroupHandler) RevokeInvite(c *gin.Context) {
	groupID, _ := middleware.GetCurrentGroupID(c)
	inviteID := c.Param("iid")

	if err := h.groupService.RevokeInvite(inviteID, groupID); err != nil {
		if err == services.ErrInviteNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Invite not found",
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
		"message": "Invite revoked",
	}))
}

// RedeemInvite joins a group using an invite code
func (h *GroupHandler) RedeemInvite(c *gin.Context) {
	userID, _ := middleware.GetCurrentUserID(c)

	var input struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invite code is required",
		))
		return
	}

	// Validate code format (should be 12 chars, possibly with dashes)
	code := input.Code
	codeLen := len(code) - 2 // Account for potential dashes
	if len(code) != 12 && codeLen != 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid invite code format",
		))
		return
	}

	group, err := h.groupService.RedeemInvite(code, userID)
	if err != nil {
		switch err {
		case services.ErrInvalidInviteCode:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Invalid invite code",
			))
		case services.ErrInviteExpired:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invite has expired",
			))
		case services.ErrInviteMaxUsed:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invite has reached maximum uses",
			))
		case services.ErrAlreadyMember:
			c.JSON(http.StatusConflict, models.ErrorResponse(
				models.ErrCodeConflict,
				"Already a member of this group",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	// Broadcast member.joined event via WebSocket
	if h.wsHub != nil {
		user, _ := middleware.GetCurrentUser(c)
		payload := websocket.MemberJoinedPayload{
			UserID:   userID,
			UserName: user.Name,
		}
		if event, err := websocket.NewEvent(websocket.EventMemberJoined, group.ID, userID, payload); err == nil {
			h.wsHub.BroadcastToGroup(event)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(group))
}
