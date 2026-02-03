package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        string         `gorm:"type:text;primaryKey" json:"id"`
	PostID    string         `gorm:"type:text;not null;index" json:"post_id"`
	UserID    string         `gorm:"type:text;not null" json:"user_id"`
	ParentID  *string        `gorm:"type:text;index" json:"parent_id,omitempty"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;index" json:"-"`

	// Relationships
	Post    Post      `gorm:"foreignKey:PostID" json:"-"`
	User    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID" json:"-"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}

func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}
