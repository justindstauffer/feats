package models

import (
	"time"
)

type Streak struct {
	ID               string     `gorm:"type:text;primaryKey" json:"id"`
	GroupID          string     `gorm:"type:text;not null;uniqueIndex:idx_streak_group_user" json:"group_id"`
	UserID           string     `gorm:"type:text;not null;uniqueIndex:idx_streak_group_user" json:"user_id"`
	CurrentStreak    int        `gorm:"not null;default:0" json:"current_streak"`
	LongestStreak    int        `gorm:"not null;default:0" json:"longest_streak"`
	LastActivityDate *time.Time `gorm:"type:date" json:"last_activity_date,omitempty"`
	UpdatedAt        time.Time  `gorm:"type:datetime;not null" json:"updated_at"`

	// Relationships
	Group Group `gorm:"foreignKey:GroupID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Streak) TableName() string {
	return "streaks"
}

// calendarDay returns t's calendar date in tz, encoded as midnight UTC. Encoding
// as UTC midnight (rather than local midnight) keeps day arithmetic DST-safe and
// round-trips cleanly through the DATE column. NOTE: do not use
// time.Truncate(24*time.Hour) for this — Truncate operates on the absolute
// instant and always snaps to UTC midnight regardless of Location, which made
// "calendar day" mean the UTC day instead of the user's local day.
func calendarDay(t time.Time, tz *time.Location) time.Time {
	y, m, d := t.In(tz).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// storedDay normalizes a persisted LastActivityDate (already stored as UTC
// midnight of a calendar date) back to UTC midnight for comparison.
func storedDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// UpdateForNewActivity updates the streak based on a new activity.
func (s *Streak) UpdateForNewActivity(activityDate time.Time, timezone *time.Location) {
	s.updateForNewActivity(activityDate, time.Now(), timezone)
}

func (s *Streak) updateForNewActivity(activityDate, now time.Time, tz *time.Location) {
	today := calendarDay(now, tz)
	activityDay := calendarDay(activityDate, tz)

	// Ignore activity dated in the future.
	if activityDay.After(today) {
		return
	}

	// First activity ever.
	if s.LastActivityDate == nil {
		s.CurrentStreak = 1
		s.LongestStreak = 1
		s.LastActivityDate = &activityDay
		return
	}

	lastDay := storedDay(*s.LastActivityDate)

	switch {
	case activityDay.Equal(lastDay):
		// Same calendar day — already counted, no change.
		return
	case activityDay.Equal(lastDay.AddDate(0, 0, 1)):
		// Next calendar day — extend the streak.
		s.CurrentStreak++
		if s.CurrentStreak > s.LongestStreak {
			s.LongestStreak = s.CurrentStreak
		}
		s.LastActivityDate = &activityDay
	case activityDay.After(lastDay):
		// Missed at least one full day — streak restarts at 1.
		s.CurrentStreak = 1
		s.LastActivityDate = &activityDay
	}
	// activityDay before lastDay (backdated) — leave the streak untouched.
}

// CheckAndResetIfNeeded resets the streak to 0 when the last activity was before
// yesterday (a full calendar day was missed). Returns true if it changed.
func (s *Streak) CheckAndResetIfNeeded(timezone *time.Location) bool {
	return s.checkAndResetIfNeeded(time.Now(), timezone)
}

func (s *Streak) checkAndResetIfNeeded(now time.Time, tz *time.Location) bool {
	if s.LastActivityDate == nil {
		return false
	}

	today := calendarDay(now, tz)
	lastDay := storedDay(*s.LastActivityDate)
	yesterday := today.AddDate(0, 0, -1)

	// Last activity older than yesterday → the streak is broken.
	if lastDay.Before(yesterday) {
		s.CurrentStreak = 0
		return true
	}

	return false
}
