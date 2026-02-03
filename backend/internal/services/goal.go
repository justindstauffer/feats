package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrGoalNotFound  = errors.New("goal not found")
	ErrInvalidPeriod = errors.New("invalid period")
)

type GoalService struct {
	db       *gorm.DB
	cfg      *config.Config
	timezone *time.Location
}

func NewGoalService(db *gorm.DB, cfg *config.Config) *GoalService {
	tz, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		tz = time.UTC
	}

	return &GoalService{
		db:       db,
		cfg:      cfg,
		timezone: tz,
	}
}

type CreateGoalInput struct {
	ActivityTypeID *string `json:"activity_type_id"`
	TargetCount    int     `json:"target_count" binding:"required,min=1"`
	Period         string  `json:"period" binding:"required"`
}

type UpdateGoalInput struct {
	TargetCount *int    `json:"target_count"`
	Period      *string `json:"period"`
}

// GetUserGoals returns all goals for a user
func (s *GoalService) GetUserGoals(userID string) ([]models.Goal, error) {
	var goals []models.Goal

	if err := s.db.
		Preload("ActivityType").
		Where("user_id = ?", userID).
		Find(&goals).Error; err != nil {
		return nil, err
	}

	// Reset periods if needed
	for i := range goals {
		if goals[i].ResetIfNewPeriod(s.timezone) {
			s.db.Save(&goals[i])
		}
	}

	return goals, nil
}

// GetGoalByID retrieves a goal by ID
func (s *GoalService) GetGoalByID(goalID string) (*models.Goal, error) {
	var goal models.Goal
	if err := s.db.Preload("ActivityType").First(&goal, "id = ?", goalID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGoalNotFound
		}
		return nil, err
	}

	// Reset period if needed
	if goal.ResetIfNewPeriod(s.timezone) {
		s.db.Save(&goal)
	}

	return &goal, nil
}

// CreateGoal creates a new goal
func (s *GoalService) CreateGoal(input CreateGoalInput, userID string) (*models.Goal, error) {
	period := models.GoalPeriod(input.Period)
	if !models.IsValidPeriod(period) {
		return nil, ErrInvalidPeriod
	}

	// Validate activity type if provided
	if input.ActivityTypeID != nil {
		var activity models.ActivityType
		if err := s.db.First(&activity, "id = ?", *input.ActivityTypeID).Error; err != nil {
			return nil, errors.New("invalid activity type")
		}
	}

	now := time.Now().In(s.timezone)
	periodStart := s.getPeriodStart(now, period)

	goal := models.Goal{
		ID:              uuid.New().String(),
		UserID:          userID,
		ActivityTypeID:  input.ActivityTypeID,
		TargetCount:     input.TargetCount,
		Period:          period,
		CurrentProgress: 0,
		PeriodStart:     periodStart,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.db.Create(&goal).Error; err != nil {
		return nil, err
	}

	return s.GetGoalByID(goal.ID)
}

// UpdateGoal updates a goal
func (s *GoalService) UpdateGoal(goalID, userID string, input UpdateGoalInput) (*models.Goal, error) {
	var goal models.Goal
	if err := s.db.First(&goal, "id = ? AND user_id = ?", goalID, userID).Error; err != nil {
		return nil, ErrGoalNotFound
	}

	if input.TargetCount != nil && *input.TargetCount > 0 {
		goal.TargetCount = *input.TargetCount
	}

	if input.Period != nil {
		period := models.GoalPeriod(*input.Period)
		if !models.IsValidPeriod(period) {
			return nil, ErrInvalidPeriod
		}
		goal.Period = period
		// Reset period start and progress when changing period type
		now := time.Now().In(s.timezone)
		goal.PeriodStart = s.getPeriodStart(now, period)
		goal.CurrentProgress = 0
	}

	goal.UpdatedAt = time.Now()

	if err := s.db.Save(&goal).Error; err != nil {
		return nil, err
	}

	return s.GetGoalByID(goal.ID)
}

// DeleteGoal deletes a goal
func (s *GoalService) DeleteGoal(goalID, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", goalID, userID).Delete(&models.Goal{})
	if result.RowsAffected == 0 {
		return ErrGoalNotFound
	}
	return result.Error
}

// UpdateProgressForActivity updates goal progress when a user posts an activity
func (s *GoalService) UpdateProgressForActivity(userID string, activityTypeID string) ([]string, error) {
	var achievedGoals []string

	// Get all user's goals
	var goals []models.Goal
	if err := s.db.Where("user_id = ?", userID).Find(&goals).Error; err != nil {
		return nil, err
	}

	for i := range goals {
		goal := &goals[i]

		// Reset period if needed
		goal.ResetIfNewPeriod(s.timezone)

		// Check if activity type matches (if goal has specific type)
		if goal.ActivityTypeID != nil && *goal.ActivityTypeID != activityTypeID {
			continue
		}

		// Already achieved this period
		if goal.IsAchieved() {
			continue
		}

		// Increment progress
		justAchieved := goal.IncrementProgress()
		goal.UpdatedAt = time.Now()
		s.db.Save(goal)

		if justAchieved {
			achievedGoals = append(achievedGoals, goal.ID)
		}
	}

	return achievedGoals, nil
}

// getPeriodStart returns the start of the current period
func (s *GoalService) getPeriodStart(t time.Time, period models.GoalPeriod) time.Time {
	switch period {
	case models.PeriodDaily:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, s.timezone)
	case models.PeriodWeekly:
		return startOfWeek(t, s.timezone)
	case models.PeriodMonthly:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, s.timezone)
	default:
		return t
	}
}

func startOfWeek(t time.Time, tz *time.Location) time.Time {
	t = t.In(tz)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-(weekday-1), 0, 0, 0, 0, tz)
}
