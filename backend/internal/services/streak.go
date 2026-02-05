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
	ErrStreakNotFound = errors.New("streak not found")
)

type StreakService struct {
	db       *gorm.DB
	cfg      *config.Config
	timezone *time.Location
}

func NewStreakService(db *gorm.DB, cfg *config.Config) *StreakService {
	tz, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		tz = time.UTC
	}

	return &StreakService{
		db:       db,
		cfg:      cfg,
		timezone: tz,
	}
}

// GetUserStreak returns a user's streak information within a group
func (s *StreakService) GetUserStreak(groupID, userID string) (*models.Streak, error) {
	var streak models.Streak
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&streak).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create initial streak for this group
			streak = models.Streak{
				ID:            uuid.New().String(),
				GroupID:       groupID,
				UserID:        userID,
				CurrentStreak: 0,
				LongestStreak: 0,
				UpdatedAt:     time.Now(),
			}
			if err := s.db.Create(&streak).Error; err != nil {
				return nil, err
			}
			return &streak, nil
		}
		return nil, err
	}

	// Check if streak needs reset
	if streak.CheckAndResetIfNeeded(s.timezone) {
		s.db.Save(&streak)
	}

	return &streak, nil
}

// UpdateStreakForActivity updates a user's streak after posting an activity within a group
func (s *StreakService) UpdateStreakForActivity(groupID, userID string, activityTime time.Time) (*models.Streak, bool, error) {
	streak, err := s.GetUserStreak(groupID, userID)
	if err != nil {
		return nil, false, err
	}

	previousStreak := streak.CurrentStreak

	streak.UpdateForNewActivity(activityTime, s.timezone)
	streak.UpdatedAt = time.Now()

	if err := s.db.Save(streak).Error; err != nil {
		return nil, false, err
	}

	// Check if this is a milestone (every 7 days)
	isMilestone := streak.CurrentStreak > 0 &&
		streak.CurrentStreak%7 == 0 &&
		streak.CurrentStreak > previousStreak

	return streak, isMilestone, nil
}

// GetLeaderboard returns all users' streaks within a group sorted by current streak
func (s *StreakService) GetLeaderboard(groupID string) ([]models.Streak, error) {
	var streaks []models.Streak

	// First, update any stale streaks in this group
	s.updateStaleStreaks(groupID)

	if err := s.db.
		Where("group_id = ?", groupID).
		Preload("User").
		Order("current_streak DESC, longest_streak DESC").
		Find(&streaks).Error; err != nil {
		return nil, err
	}

	return streaks, nil
}

// updateStaleStreaks resets streaks that haven't been updated in a group
func (s *StreakService) updateStaleStreaks(groupID string) {
	var streaks []models.Streak
	s.db.Where("group_id = ?", groupID).Find(&streaks)

	for _, streak := range streaks {
		if streak.CheckAndResetIfNeeded(s.timezone) {
			s.db.Save(&streak)
		}
	}
}

// GetStreakMilestones returns milestone days (7, 14, 21, 30, 60, 90, etc.)
func GetStreakMilestones() []int {
	return []int{7, 14, 21, 30, 60, 90, 180, 365}
}

// IsStreakMilestone checks if a streak count is a milestone
func IsStreakMilestone(streak int) bool {
	milestones := GetStreakMilestones()
	for _, m := range milestones {
		if streak == m {
			return true
		}
	}
	// Also consider multiples of 100
	return streak > 0 && streak%100 == 0
}
