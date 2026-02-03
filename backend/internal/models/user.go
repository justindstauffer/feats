package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID                  string         `gorm:"type:text;primaryKey" json:"id"`
	Email               string         `gorm:"type:text;uniqueIndex;not null" json:"email"`
	PasswordHash        string         `gorm:"type:text;not null" json:"-"`
	Name                string         `gorm:"type:text;not null" json:"name"`
	ProfilePicture      *string        `gorm:"type:text" json:"profile_picture,omitempty"`
	Bio                 *string        `gorm:"type:text" json:"bio,omitempty"`
	Role                UserRole       `gorm:"type:text;not null;default:'user'" json:"role"`
	FailedLoginAttempts int            `gorm:"not null;default:0" json:"-"`
	LockedUntil         *time.Time     `gorm:"type:datetime" json:"-"`
	PasswordChangedAt   time.Time      `gorm:"type:datetime;not null" json:"-"`
	ForcePasswordChange bool           `gorm:"not null;default:false" json:"-"`
	LastLoginAt         *time.Time     `gorm:"type:datetime" json:"-"`
	LastLoginIPHash     *string        `gorm:"type:text" json:"-"`
	CreatedAt           time.Time      `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"type:datetime;not null" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"type:datetime;index" json:"-"`

	// Relationships
	Posts       []Post       `gorm:"foreignKey:UserID" json:"-"`
	Comments    []Comment    `gorm:"foreignKey:UserID" json:"-"`
	Reactions   []Reaction   `gorm:"foreignKey:UserID" json:"-"`
	Goals       []Goal       `gorm:"foreignKey:UserID" json:"-"`
	Streak      *Streak      `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
