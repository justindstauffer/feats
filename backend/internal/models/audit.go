package models

import (
	"time"
)

type AuditAction string

const (
	AuditActionLogin              AuditAction = "LOGIN"
	AuditActionLoginFailed        AuditAction = "LOGIN_FAILED"
	AuditActionLogout             AuditAction = "LOGOUT"
	AuditActionPasswordChange     AuditAction = "PASSWORD_CHANGE"
	AuditActionPasswordReset      AuditAction = "PASSWORD_RESET"
	AuditActionPasswordResetReq   AuditAction = "PASSWORD_RESET_REQUEST"
	AuditActionAccountLocked      AuditAction = "ACCOUNT_LOCKED"
	AuditActionUserCreated        AuditAction = "USER_CREATED"
	AuditActionUserDeleted        AuditAction = "USER_DELETED"
	AuditActionPostCreated        AuditAction = "POST_CREATED"
	AuditActionPostDeleted        AuditAction = "POST_DELETED"
	AuditActionAuthorizationFail  AuditAction = "AUTHORIZATION_FAILED"
	AuditActionRateLimitExceeded  AuditAction = "RATE_LIMIT_EXCEEDED"
	AuditActionTokenRefresh       AuditAction = "TOKEN_REFRESH"
	AuditActionTokenRevoked       AuditAction = "TOKEN_REVOKED"
)

type AuditLog struct {
	ID           string      `gorm:"type:text;primaryKey" json:"id"`
	Timestamp    time.Time   `gorm:"type:datetime;not null;index" json:"timestamp"`
	UserID       *string     `gorm:"type:text;index" json:"user_id,omitempty"`
	Action       AuditAction `gorm:"type:text;not null;index" json:"action"`
	ResourceType *string     `gorm:"type:text" json:"resource_type,omitempty"`
	ResourceID   *string     `gorm:"type:text" json:"resource_id,omitempty"`
	IPHash       *string     `gorm:"type:text" json:"ip_hash,omitempty"`
	UserAgent    *string     `gorm:"type:text" json:"user_agent,omitempty"`
	Details      *string     `gorm:"type:text" json:"details,omitempty"` // JSON string
	Success      bool        `gorm:"not null" json:"success"`
	CreatedAt    time.Time   `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// RateLimit stores rate limiting state (optional, can use in-memory instead)
type RateLimit struct {
	ID         string    `gorm:"type:text;primaryKey" json:"id"`
	Key        string    `gorm:"type:text;uniqueIndex;not null" json:"key"`
	Tokens     int       `gorm:"not null" json:"tokens"`
	LastRefill time.Time `gorm:"type:datetime;not null" json:"last_refill"`
}

func (RateLimit) TableName() string {
	return "rate_limits"
}
