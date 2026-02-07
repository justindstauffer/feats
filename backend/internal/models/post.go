package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID             string         `gorm:"type:text;primaryKey" json:"id"`
	GroupID        string         `gorm:"type:text;not null;index" json:"group_id"`
	UserID         string         `gorm:"type:text;not null;index" json:"user_id"`
	ActivityTypeID string         `gorm:"type:text;not null" json:"activity_type_id"`
	Description    *string        `gorm:"type:text" json:"description,omitempty"`
	CreatedAt      time.Time      `gorm:"type:datetime;not null;index" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:datetime;not null" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"type:datetime;index" json:"-"`

	// Computed fields (not stored in DB)
	CommentCount int `gorm:"-" json:"comment_count"`

	// Relationships
	Group        Group        `gorm:"foreignKey:GroupID" json:"-"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ActivityType ActivityType `gorm:"foreignKey:ActivityTypeID" json:"activity_type,omitempty"`
	Images       []PostImage  `gorm:"foreignKey:PostID" json:"images,omitempty"`
	Reactions    []Reaction   `gorm:"foreignKey:PostID" json:"reactions,omitempty"`
	Comments     []Comment    `gorm:"foreignKey:PostID" json:"comments,omitempty"`
}

func (Post) TableName() string {
	return "posts"
}

type PostImage struct {
	ID           string    `gorm:"type:text;primaryKey" json:"id"`
	PostID       string    `gorm:"type:text;not null;index" json:"post_id"`
	FilePath     string    `gorm:"type:text;not null" json:"-"`
	DisplayOrder int       `gorm:"not null" json:"display_order"`
	CreatedAt    time.Time `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	Post Post `gorm:"foreignKey:PostID" json:"-"`
}

func (PostImage) TableName() string {
	return "post_images"
}

// ImageURL returns the public URL for the image
func (p *PostImage) ImageURL(baseURL string) string {
	return baseURL + "/images/" + p.ID
}
