package models

import (
	"time"
)

type DeviceToken struct {
	ID        string    `gorm:"type:text;primaryKey" json:"id"`
	UserID    string    `gorm:"type:text;not null;index" json:"user_id"`
	Token     string    `gorm:"type:text;uniqueIndex;not null" json:"-"`
	Platform  string    `gorm:"type:text;not null;default:'ios'" json:"platform"`
	CreatedAt time.Time `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime;not null" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (DeviceToken) TableName() string {
	return "device_tokens"
}
