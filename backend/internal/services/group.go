package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrGroupNotFound      = errors.New("group not found")
	ErrAlreadyMember      = errors.New("already a member of this group")
	ErrNotGroupMember     = errors.New("not a member of this group")
	ErrNotGroupAdmin      = errors.New("not an admin of this group")
	ErrCannotLeaveAsAdmin = errors.New("cannot leave group as the only admin")
	ErrCannotRemoveSelf   = errors.New("cannot remove yourself from the group")
	ErrInviteNotFound     = errors.New("invite not found")
	ErrInviteExpired      = errors.New("invite has expired")
	ErrInviteMaxUsed      = errors.New("invite has reached maximum uses")
	ErrInvalidInviteCode  = errors.New("invalid invite code")
)

type GroupService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewGroupService(db *gorm.DB, cfg *config.Config) *GroupService {
	return &GroupService{
		db:  db,
		cfg: cfg,
	}
}

// Input types

type CreateGroupInput struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description"`
}

type UpdateGroupInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateInviteInput struct {
	MaxUses   int `json:"max_uses"`   // 0 = unlimited
	ExpiresIn int `json:"expires_in"` // Hours until expiration, 0 = default (168 = 7 days)
}

// Group CRUD

// CreateGroup creates a new group with the user as admin
func (s *GroupService) CreateGroup(input CreateGroupInput, userID string) (*models.Group, error) {
	name := strings.TrimSpace(input.Name)
	if len(name) > 100 {
		name = name[:100]
	}

	var description *string
	if input.Description != nil {
		desc := strings.TrimSpace(*input.Description)
		if len(desc) > 500 {
			desc = desc[:500]
		}
		if desc != "" {
			description = &desc
		}
	}

	now := time.Now()
	group := models.Group{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}

		// Add creator as admin
		member := models.GroupMember{
			ID:       uuid.New().String(),
			GroupID:  group.ID,
			UserID:   userID,
			Role:     models.GroupRoleAdmin,
			JoinedAt: now,
		}
		return tx.Create(&member).Error
	})

	if err != nil {
		return nil, err
	}

	return s.GetGroupByID(group.ID)
}

// GetGroupByID retrieves a group by ID with members
func (s *GroupService) GetGroupByID(groupID string) (*models.Group, error) {
	var group models.Group
	if err := s.db.
		Preload("Creator").
		Preload("Members", "left_at IS NULL").
		Preload("Members.User").
		First(&group, "id = ?", groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return &group, nil
}

// UpdateGroup updates a group's details (admin only - checked in handler)
func (s *GroupService) UpdateGroup(groupID string, input UpdateGroupInput) (*models.Group, error) {
	var group models.Group
	if err := s.db.First(&group, "id = ?", groupID).Error; err != nil {
		return nil, ErrGroupNotFound
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if len(name) > 100 {
			name = name[:100]
		}
		if name != "" {
			group.Name = name
		}
	}

	if input.Description != nil {
		desc := strings.TrimSpace(*input.Description)
		if len(desc) > 500 {
			desc = desc[:500]
		}
		if desc == "" {
			group.Description = nil
		} else {
			group.Description = &desc
		}
	}

	group.UpdatedAt = time.Now()

	if err := s.db.Save(&group).Error; err != nil {
		return nil, err
	}

	return s.GetGroupByID(group.ID)
}

// DeleteGroup deletes a group and all related data (admin only - checked in handler)
func (s *GroupService) DeleteGroup(groupID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete group invites
		if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupInvite{}).Error; err != nil {
			return err
		}

		// Delete group members
		if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}

		// Delete the group
		if err := tx.Delete(&models.Group{}, "id = ?", groupID).Error; err != nil {
			return err
		}

		return nil
	})
}

// ListUserGroups returns all groups the user is an active member of
func (s *GroupService) ListUserGroups(userID string) ([]models.Group, error) {
	var memberships []models.GroupMember
	if err := s.db.
		Where("user_id = ? AND left_at IS NULL", userID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return []models.Group{}, nil
	}

	groupIDs := make([]string, len(memberships))
	for i, m := range memberships {
		groupIDs[i] = m.GroupID
	}

	var groups []models.Group
	if err := s.db.
		Preload("Creator").
		Preload("Members", "left_at IS NULL").
		Preload("Members.User").
		Where("id IN ?", groupIDs).
		Order("created_at DESC").
		Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

// Membership

// GetMembership returns a user's membership in a group
func (s *GroupService) GetMembership(groupID, userID string) (*models.GroupMember, error) {
	var member models.GroupMember
	if err := s.db.
		Where("group_id = ? AND user_id = ? AND left_at IS NULL", groupID, userID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotGroupMember
		}
		return nil, err
	}
	return &member, nil
}

// IsGroupMember checks if user is an active member of the group
func (s *GroupService) IsGroupMember(groupID, userID string) bool {
	var count int64
	s.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND left_at IS NULL", groupID, userID).
		Count(&count)
	return count > 0
}

// IsGroupAdmin checks if user is an admin of the group
func (s *GroupService) IsGroupAdmin(groupID, userID string) bool {
	var count int64
	s.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND role = ? AND left_at IS NULL", groupID, userID, models.GroupRoleAdmin).
		Count(&count)
	return count > 0
}

// LeaveGroup removes a user from a group (soft delete)
func (s *GroupService) LeaveGroup(groupID, userID string) error {
	member, err := s.GetMembership(groupID, userID)
	if err != nil {
		return err
	}

	// Check if user is the only admin
	if member.Role == models.GroupRoleAdmin {
		var adminCount int64
		s.db.Model(&models.GroupMember{}).
			Where("group_id = ? AND role = ? AND left_at IS NULL", groupID, models.GroupRoleAdmin).
			Count(&adminCount)

		if adminCount <= 1 {
			return ErrCannotLeaveAsAdmin
		}
	}

	now := time.Now()
	member.LeftAt = &now
	return s.db.Save(member).Error
}

// RemoveMember removes another user from the group (admin only)
func (s *GroupService) RemoveMember(groupID, targetUserID, adminUserID string) error {
	if targetUserID == adminUserID {
		return ErrCannotRemoveSelf
	}

	member, err := s.GetMembership(groupID, targetUserID)
	if err != nil {
		return err
	}

	now := time.Now()
	member.LeftAt = &now
	return s.db.Save(member).Error
}

// UpdateMemberRole changes a member's role (admin only)
func (s *GroupService) UpdateMemberRole(groupID, targetUserID string, role models.GroupRole) error {
	member, err := s.GetMembership(groupID, targetUserID)
	if err != nil {
		return err
	}

	// If demoting from admin, ensure there's at least one other admin
	if member.Role == models.GroupRoleAdmin && role == models.GroupRoleMember {
		var adminCount int64
		s.db.Model(&models.GroupMember{}).
			Where("group_id = ? AND role = ? AND left_at IS NULL", groupID, models.GroupRoleAdmin).
			Count(&adminCount)

		if adminCount <= 1 {
			return errors.New("cannot demote the only admin")
		}
	}

	member.Role = role
	return s.db.Save(member).Error
}

// ListMembers returns all active members of a group
func (s *GroupService) ListMembers(groupID string) ([]models.GroupMember, error) {
	var members []models.GroupMember
	if err := s.db.
		Preload("User").
		Where("group_id = ? AND left_at IS NULL", groupID).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// Invites

// CreateInvite creates a new invite code for a group
func (s *GroupService) CreateInvite(groupID, userID string, input CreateInviteInput) (*models.GroupInvite, error) {
	// Defaults
	maxUses := input.MaxUses
	if maxUses <= 0 {
		maxUses = 1 // Default to single use
	}

	expiresInHours := input.ExpiresIn
	if expiresInHours <= 0 {
		expiresInHours = 7 * 24 // Default 7 days (168 hours)
	}

	invite := models.GroupInvite{
		ID:        uuid.New().String(),
		GroupID:   groupID,
		Code:      GenerateInviteCode(),
		CreatedBy: userID,
		ExpiresAt: time.Now().Add(time.Duration(expiresInHours) * time.Hour),
		MaxUses:   maxUses,
		UseCount:  0,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&invite).Error; err != nil {
		return nil, err
	}

	// Reload with relationships
	if err := s.db.Preload("Group").Preload("Creator").First(&invite, "id = ?", invite.ID).Error; err != nil {
		return nil, err
	}

	return &invite, nil
}

// RedeemInvite adds a user to a group using an invite code
func (s *GroupService) RedeemInvite(code, userID string) (*models.Group, error) {
	// Normalize code (remove dashes, uppercase)
	code = strings.ToUpper(strings.ReplaceAll(code, "-", ""))
	if len(code) != 12 {
		return nil, ErrInvalidInviteCode
	}
	formattedCode := fmt.Sprintf("%s-%s-%s", code[0:4], code[4:8], code[8:12])

	var invite models.GroupInvite
	if err := s.db.Where("code = ?", formattedCode).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidInviteCode
		}
		return nil, err
	}

	if invite.IsExpired() {
		return nil, ErrInviteExpired
	}

	if !invite.HasUsesRemaining() {
		return nil, ErrInviteMaxUsed
	}

	// Check if already a member
	if s.IsGroupMember(invite.GroupID, userID) {
		return nil, ErrAlreadyMember
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Add user as member
		member := models.GroupMember{
			ID:       uuid.New().String(),
			GroupID:  invite.GroupID,
			UserID:   userID,
			Role:     models.GroupRoleMember,
			JoinedAt: time.Now(),
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		// Increment use count
		invite.UseCount++
		return tx.Save(&invite).Error
	})

	if err != nil {
		return nil, err
	}

	return s.GetGroupByID(invite.GroupID)
}

// ListInvites returns all invites for a group
func (s *GroupService) ListInvites(groupID string) ([]models.GroupInvite, error) {
	var invites []models.GroupInvite
	if err := s.db.
		Preload("Creator").
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

// RevokeInvite deletes an invite
func (s *GroupService) RevokeInvite(inviteID, groupID string) error {
	result := s.db.Where("id = ? AND group_id = ?", inviteID, groupID).Delete(&models.GroupInvite{})
	if result.RowsAffected == 0 {
		return ErrInviteNotFound
	}
	return result.Error
}

// GetMemberUserIDs returns all active member user IDs for a group
func (s *GroupService) GetMemberUserIDs(groupID string) ([]string, error) {
	var members []models.GroupMember
	if err := s.db.
		Where("group_id = ? AND left_at IS NULL", groupID).
		Find(&members).Error; err != nil {
		return nil, err
	}

	userIDs := make([]string, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}
	return userIDs, nil
}

// GenerateInviteCode creates a random invite code in XXXX-XXXX-XXXX format
func GenerateInviteCode() string {
	// Characters: A-Z, 2-9 (excluding 0/O, 1/I/L for clarity)
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to less secure but still functional
		for i := range bytes {
			bytes[i] = charset[i%len(charset)]
		}
	}

	code := make([]byte, 12)
	for i := range code {
		code[i] = charset[int(bytes[i])%len(charset)]
	}

	return fmt.Sprintf("%s-%s-%s", string(code[0:4]), string(code[4:8]), string(code[8:12]))
}
