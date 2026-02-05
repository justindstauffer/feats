package models

import (
	"time"
)

// GroupRole represents a user's role within a group
type GroupRole string

const (
	GroupRoleAdmin  GroupRole = "admin"
	GroupRoleMember GroupRole = "member"
)

// Group represents a collection of users (family, friends, team, etc.)
type Group struct {
	ID          string    `gorm:"type:text;primaryKey" json:"id"`
	Name        string    `gorm:"type:text;not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	CreatedBy   string    `gorm:"type:text;not null" json:"created_by"`
	CreatedAt   time.Time `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:datetime;not null" json:"updated_at"`

	// Relationships
	Creator User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Members []GroupMember `gorm:"foreignKey:GroupID" json:"members,omitempty"`
}

func (Group) TableName() string {
	return "groups"
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	ID       string     `gorm:"type:text;primaryKey" json:"id"`
	GroupID  string     `gorm:"type:text;not null;uniqueIndex:idx_group_user_active,where:left_at IS NULL" json:"group_id"`
	UserID   string     `gorm:"type:text;not null;uniqueIndex:idx_group_user_active,where:left_at IS NULL" json:"user_id"`
	Role     GroupRole  `gorm:"type:text;not null;default:member" json:"role"`
	JoinedAt time.Time  `gorm:"type:datetime;not null" json:"joined_at"`
	LeftAt   *time.Time `gorm:"type:datetime" json:"left_at,omitempty"`

	// Relationships
	Group Group `gorm:"foreignKey:GroupID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

// IsActive returns true if the member hasn't left the group
func (gm *GroupMember) IsActive() bool {
	return gm.LeftAt == nil
}

// IsAdmin returns true if the member has admin role
func (gm *GroupMember) IsAdmin() bool {
	return gm.Role == GroupRoleAdmin
}

// GroupInvite represents an invitation code to join a group
type GroupInvite struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	GroupID   string    `gorm:"type:text;not null;index" json:"group_id"`
	Code      string    `gorm:"type:text;not null;uniqueIndex" json:"code"`
	CreatedBy string    `gorm:"type:text;not null" json:"created_by"`
	ExpiresAt time.Time `gorm:"type:datetime;not null" json:"expires_at"`
	MaxUses   int       `gorm:"not null;default:1" json:"max_uses"`
	UseCount  int       `gorm:"not null;default:0" json:"use_count"`
	CreatedAt time.Time `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	Group   Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Creator User  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (GroupInvite) TableName() string {
	return "group_invites"
}

// IsValid returns true if the invite can still be used
func (gi *GroupInvite) IsValid() bool {
	return time.Now().Before(gi.ExpiresAt) && gi.UseCount < gi.MaxUses
}

// IsExpired returns true if the invite has expired
func (gi *GroupInvite) IsExpired() bool {
	return time.Now().After(gi.ExpiresAt)
}

// HasUsesRemaining returns true if the invite hasn't reached max uses
func (gi *GroupInvite) HasUsesRemaining() bool {
	return gi.UseCount < gi.MaxUses
}
