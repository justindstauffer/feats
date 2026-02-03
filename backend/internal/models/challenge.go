package models

import (
	"time"
)

type Challenge struct {
	ID             string     `gorm:"type:text;primaryKey" json:"id"`
	CreatedBy      string     `gorm:"type:text;not null" json:"created_by"`
	Title          string     `gorm:"type:text;not null" json:"title"`
	Description    *string    `gorm:"type:text" json:"description,omitempty"`
	ActivityTypeID *string    `gorm:"type:text" json:"activity_type_id,omitempty"`
	TargetCount    int        `gorm:"not null" json:"target_count"`
	StartDate      *time.Time `gorm:"type:date" json:"start_date,omitempty"`
	EndDate        *time.Time `gorm:"type:date" json:"end_date,omitempty"`
	CreatedAt      time.Time  `gorm:"type:datetime;not null" json:"created_at"`

	// Relationships
	Creator      User                   `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	ActivityType *ActivityType          `gorm:"foreignKey:ActivityTypeID" json:"activity_type,omitempty"`
	Participants []ChallengeParticipant `gorm:"foreignKey:ChallengeID" json:"participants,omitempty"`
}

func (Challenge) TableName() string {
	return "challenges"
}

// IsActive checks if the challenge is currently active
func (c *Challenge) IsActive() bool {
	now := time.Now()

	// If no dates set, always active
	if c.StartDate == nil && c.EndDate == nil {
		return true
	}

	// Check start date
	if c.StartDate != nil && now.Before(*c.StartDate) {
		return false
	}

	// Check end date
	if c.EndDate != nil && now.After(*c.EndDate) {
		return false
	}

	return true
}

// IsTimeBound returns true if the challenge has date constraints
func (c *Challenge) IsTimeBound() bool {
	return c.StartDate != nil || c.EndDate != nil
}

type ChallengeParticipant struct {
	ID          string     `gorm:"type:text;primaryKey" json:"id"`
	ChallengeID string     `gorm:"type:text;not null;index" json:"challenge_id"`
	UserID      string     `gorm:"type:text;not null" json:"user_id"`
	Progress    int        `gorm:"not null;default:0" json:"progress"`
	CompletedAt *time.Time `gorm:"type:datetime" json:"completed_at,omitempty"`
	JoinedAt    time.Time  `gorm:"type:datetime;not null" json:"joined_at"`

	// Relationships
	Challenge Challenge `gorm:"foreignKey:ChallengeID" json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ChallengeParticipant) TableName() string {
	return "challenge_participants"
}

// IsCompleted checks if the participant has completed the challenge
func (cp *ChallengeParticipant) IsCompleted() bool {
	return cp.CompletedAt != nil
}

// IncrementProgress increases progress and marks completed if target reached
func (cp *ChallengeParticipant) IncrementProgress(targetCount int) bool {
	if cp.IsCompleted() {
		return false
	}

	cp.Progress++

	if cp.Progress >= targetCount {
		now := time.Now()
		cp.CompletedAt = &now
		return true // Just completed
	}

	return false
}
