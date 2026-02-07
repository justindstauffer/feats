package models

import (
	"strconv"
	"time"
)

type ReactionType int

const (
	ReactionLike   ReactionType = 1
	ReactionLove   ReactionType = 2
	ReactionFire   ReactionType = 3
	ReactionStrong ReactionType = 4
	ReactionClap   ReactionType = 5
)

// String returns the string representation of the reaction type
func (rt ReactionType) String() string {
	return strconv.Itoa(int(rt))
}

var ReactionEmojis = map[ReactionType]string{
	ReactionLike:   "👍",
	ReactionLove:   "❤️",
	ReactionFire:   "🔥",
	ReactionStrong: "💪",
	ReactionClap:   "👏",
}

type Reaction struct {
	ID           string       `gorm:"type:text;primaryKey" json:"id"`
	UserID       string       `gorm:"type:text;not null;index" json:"user_id"`
	PostID       string       `gorm:"type:text;not null;index" json:"post_id"`
	ReactionType ReactionType `gorm:"not null" json:"reaction_type"`
	CreatedAt    time.Time    `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post Post `gorm:"foreignKey:PostID" json:"-"`
}

func (Reaction) TableName() string {
	return "reactions"
}

func (r *Reaction) Emoji() string {
	if emoji, ok := ReactionEmojis[r.ReactionType]; ok {
		return emoji
	}
	return ""
}

// IsValidReactionType checks if the reaction type is valid
func IsValidReactionType(rt ReactionType) bool {
	return rt >= ReactionLike && rt <= ReactionClap
}

// ReactionSummary represents aggregated reaction counts
type ReactionSummary struct {
	Type  ReactionType `json:"type"`
	Emoji string       `json:"emoji"`
	Count int          `json:"count"`
}
