package models

import (
	"time"
)

type Streak struct {
	ID               string     `gorm:"type:text;primaryKey" json:"id"`
	UserID           string     `gorm:"type:text;uniqueIndex;not null" json:"user_id"`
	CurrentStreak    int        `gorm:"not null;default:0" json:"current_streak"`
	LongestStreak    int        `gorm:"not null;default:0" json:"longest_streak"`
	LastActivityDate *time.Time `gorm:"type:date" json:"last_activity_date,omitempty"`
	UpdatedAt        time.Time  `gorm:"type:datetime;not null" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Streak) TableName() string {
	return "streaks"
}

// UpdateForNewActivity updates the streak based on a new activity
func (s *Streak) UpdateForNewActivity(activityDate time.Time, timezone *time.Location) {
	today := time.Now().In(timezone).Truncate(24 * time.Hour)
	activityDay := activityDate.In(timezone).Truncate(24 * time.Hour)

	// If activity is in the future, ignore
	if activityDay.After(today) {
		return
	}

	// First activity ever
	if s.LastActivityDate == nil {
		s.CurrentStreak = 1
		s.LongestStreak = 1
		s.LastActivityDate = &activityDay
		return
	}

	lastDay := s.LastActivityDate.In(timezone).Truncate(24 * time.Hour)

	// Same day - no change to streak
	if activityDay.Equal(lastDay) {
		return
	}

	// Consecutive day - increment streak
	if activityDay.Equal(lastDay.Add(24 * time.Hour)) {
		s.CurrentStreak++
		if s.CurrentStreak > s.LongestStreak {
			s.LongestStreak = s.CurrentStreak
		}
		s.LastActivityDate = &activityDay
		return
	}

	// Gap in days - reset streak
	if activityDay.After(lastDay) {
		s.CurrentStreak = 1
		s.LastActivityDate = &activityDay
	}
}

// CheckAndResetIfNeeded checks if streak should be reset due to missed day
func (s *Streak) CheckAndResetIfNeeded(timezone *time.Location) bool {
	if s.LastActivityDate == nil {
		return false
	}

	today := time.Now().In(timezone).Truncate(24 * time.Hour)
	lastDay := s.LastActivityDate.In(timezone).Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	// If last activity was before yesterday, streak is broken
	if lastDay.Before(yesterday) {
		s.CurrentStreak = 0
		return true
	}

	return false
}
