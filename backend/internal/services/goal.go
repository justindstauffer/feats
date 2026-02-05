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

// GetUserGoals returns all goals for a user within a group
func (s *GoalService) GetUserGoals(groupID, userID string) ([]models.Goal, error) {
	var goals []models.Goal

	if err := s.db.
		Preload("ActivityType").
		Where("group_id = ? AND user_id = ?", groupID, userID).
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

// GetGoalByID retrieves a goal by ID within a group
func (s *GoalService) GetGoalByID(groupID, goalID string) (*models.Goal, error) {
	var goal models.Goal
	if err := s.db.Preload("ActivityType").First(&goal, "id = ? AND group_id = ?", goalID, groupID).Error; err != nil {
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

// CreateGoal creates a new goal within a group
func (s *GoalService) CreateGoal(groupID string, input CreateGoalInput, userID string) (*models.Goal, error) {
	period := models.GoalPeriod(input.Period)
	if !models.IsValidPeriod(period) {
		return nil, ErrInvalidPeriod
	}

	// Validate activity type if provided (system-wide or group-specific)
	if input.ActivityTypeID != nil {
		var activity models.ActivityType
		if err := s.db.First(&activity, "id = ? AND (group_id IS NULL OR group_id = ?)", *input.ActivityTypeID, groupID).Error; err != nil {
			return nil, errors.New("invalid activity type")
		}
	}

	now := time.Now().In(s.timezone)
	periodStart := s.getPeriodStart(now, period)

	goal := models.Goal{
		ID:              uuid.New().String(),
		GroupID:         groupID,
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

	return s.GetGoalByID(groupID, goal.ID)
}

// UpdateGoal updates a goal within a group
func (s *GoalService) UpdateGoal(groupID, goalID, userID string, input UpdateGoalInput) (*models.Goal, error) {
	var goal models.Goal
	if err := s.db.First(&goal, "id = ? AND group_id = ? AND user_id = ?", goalID, groupID, userID).Error; err != nil {
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

	return s.GetGoalByID(groupID, goal.ID)
}

// DeleteGoal deletes a goal within a group
func (s *GoalService) DeleteGoal(groupID, goalID, userID string) error {
	result := s.db.Where("id = ? AND group_id = ? AND user_id = ?", goalID, groupID, userID).Delete(&models.Goal{})
	if result.RowsAffected == 0 {
		return ErrGoalNotFound
	}
	return result.Error
}

// UpdateProgressForActivity updates goal progress when a user posts an activity within a group
func (s *GoalService) UpdateProgressForActivity(groupID, userID string, activityTypeID string) ([]string, error) {
	var achievedGoals []string

	// Get all user's goals in this group
	var goals []models.Goal
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).Find(&goals).Error; err != nil {
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
