package models

import (
	"time"
)

type RefreshToken struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	UserID    string    `gorm:"type:text;not null;index" json:"user_id"`
	TokenHash string    `gorm:"type:text;not null" json:"-"`
	ExpiresAt time.Time `gorm:"type:datetime;not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

type PasswordHistory struct {
	ID           string    `gorm:"type:text;primaryKey" json:"id"`
	UserID       string    `gorm:"type:text;not null;index" json:"user_id"`
	PasswordHash string    `gorm:"type:text;not null" json:"-"`
	CreatedAt    time.Time `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (PasswordHistory) TableName() string {
	return "password_history"
}

type PasswordResetToken struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	UserID    string    `gorm:"type:text;not null;index" json:"user_id"`
	TokenHash string    `gorm:"type:text;not null" json:"-"`
	ExpiresAt time.Time `gorm:"type:datetime;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"type:datetime" json:"used_at,omitempty"`
	CreatedAt time.Time `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

func (p *PasswordResetToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func (p *PasswordResetToken) IsUsed() bool {
	return p.UsedAt != nil
}
