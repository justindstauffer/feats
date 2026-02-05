package models

import (
	"strings"
	"time"
)

// BetaInvite represents an invite code for beta app registration
type BetaInvite struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text"`
	Code      string    `json:"code" gorm:"uniqueIndex;type:text;not null"`
	CreatedBy string    `json:"created_by" gorm:"type:text;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxUses   int       `json:"max_uses" gorm:"default:1"`
	UseCount  int       `json:"use_count" gorm:"default:0"`
	Note      string    `json:"note,omitempty" gorm:"type:text"` // Optional note about who this is for
	CreatedAt time.Time `json:"created_at"`

	// Relations
	Creator *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName specifies the table name for BetaInvite
func (BetaInvite) TableName() string {
	return "beta_invites"
}

// IsValid checks if the invite can still be used
func (i *BetaInvite) IsValid() bool {
	return !i.IsExpired() && i.HasUsesRemaining()
}

// IsExpired checks if the invite has expired
func (i *BetaInvite) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// HasUsesRemaining checks if the invite has uses left
func (i *BetaInvite) HasUsesRemaining() bool {
	// MaxUses of 0 means unlimited
	return i.MaxUses == 0 || i.UseCount < i.MaxUses
}

// NormalizeCode removes dashes and converts to uppercase for comparison
func NormalizeBetaCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(code, "-", ""))
}

// Request/Response types

type CreateBetaInviteRequest struct {
	MaxUses   int    `json:"max_uses"`             // 0 = unlimited, default 1
	ExpiresIn int    `json:"expires_in,omitempty"` // Hours until expiration, default 168 (7 days)
	Note      string `json:"note,omitempty"`       // Optional note
}

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email,max=255"`
	Password   string `json:"password" binding:"required,min=12,max=128"`
	Name       string `json:"name" binding:"required,min=1,max=100"`
	InviteCode string `json:"invite_code" binding:"required"`
}
