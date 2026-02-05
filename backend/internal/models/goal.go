package models

import (
	"time"
)

type GoalPeriod string

const (
	PeriodDaily   GoalPeriod = "daily"
	PeriodWeekly  GoalPeriod = "weekly"
	PeriodMonthly GoalPeriod = "monthly"
)

type Goal struct {
	ID              string     `gorm:"type:text;primaryKey" json:"id"`
	GroupID         string     `gorm:"type:text;not null;index" json:"group_id"`
	UserID          string     `gorm:"type:text;not null;index" json:"user_id"`
	ActivityTypeID  *string    `gorm:"type:text" json:"activity_type_id,omitempty"`
	TargetCount     int        `gorm:"not null" json:"target_count"`
	Period          GoalPeriod `gorm:"type:text;not null" json:"period"`
	CurrentProgress int        `gorm:"not null;default:0" json:"current_progress"`
	PeriodStart     time.Time  `gorm:"type:date;not null" json:"period_start"`
	CreatedAt       time.Time  `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"type:datetime;not null" json:"updated_at"`

	// Relationships
	Group        Group         `gorm:"foreignKey:GroupID" json:"-"`
	User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ActivityType *ActivityType `gorm:"foreignKey:ActivityTypeID" json:"activity_type,omitempty"`
}

func (Goal) TableName() string {
	return "goals"
}

// IsAchieved checks if the goal has been achieved
func (g *Goal) IsAchieved() bool {
	return g.CurrentProgress >= g.TargetCount
}

// ResetIfNewPeriod checks if we're in a new period and resets progress if needed
func (g *Goal) ResetIfNewPeriod(timezone *time.Location) bool {
	now := time.Now().In(timezone)
	periodStart := g.PeriodStart.In(timezone)

	var needsReset bool
	var newPeriodStart time.Time

	switch g.Period {
	case PeriodDaily:
		currentDay := now.Truncate(24 * time.Hour)
		periodDay := periodStart.Truncate(24 * time.Hour)
		needsReset = currentDay.After(periodDay)
		newPeriodStart = currentDay

	case PeriodWeekly:
		// Get start of current week (Monday)
		currentWeekStart := startOfWeek(now)
		periodWeekStart := startOfWeek(periodStart)
		needsReset = currentWeekStart.After(periodWeekStart)
		newPeriodStart = currentWeekStart

	case PeriodMonthly:
		// Get start of current month
		currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, timezone)
		periodMonthStart := time.Date(periodStart.Year(), periodStart.Month(), 1, 0, 0, 0, 0, timezone)
		needsReset = currentMonthStart.After(periodMonthStart)
		newPeriodStart = currentMonthStart
	}

	if needsReset {
		g.CurrentProgress = 0
		g.PeriodStart = newPeriodStart
		return true
	}

	return false
}

// IncrementProgress increases the progress count
func (g *Goal) IncrementProgress() bool {
	wasAchieved := g.IsAchieved()
	g.CurrentProgress++
	isNowAchieved := g.IsAchieved()

	// Return true if this increment achieved the goal
	return !wasAchieved && isNowAchieved
}

// startOfWeek returns the Monday of the week containing t
func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
}

// IsValidPeriod checks if the period is valid
func IsValidPeriod(p GoalPeriod) bool {
	return p == PeriodDaily || p == PeriodWeekly || p == PeriodMonthly
}
