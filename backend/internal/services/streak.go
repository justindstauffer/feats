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

// GetUserStreak returns a user's streak information
func (s *StreakService) GetUserStreak(userID string) (*models.Streak, error) {
	var streak models.Streak
	if err := s.db.Where("user_id = ?", userID).First(&streak).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create initial streak
			streak = models.Streak{
				ID:            uuid.New().String(),
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

// UpdateStreakForActivity updates a user's streak after posting an activity
func (s *StreakService) UpdateStreakForActivity(userID string, activityTime time.Time) (*models.Streak, bool, error) {
	streak, err := s.GetUserStreak(userID)
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

// GetLeaderboard returns all users' streaks sorted by current streak
func (s *StreakService) GetLeaderboard() ([]models.Streak, error) {
	var streaks []models.Streak

	// First, update any stale streaks
	s.updateStaleStreaks()

	if err := s.db.
		Preload("User").
		Order("current_streak DESC, longest_streak DESC").
		Find(&streaks).Error; err != nil {
		return nil, err
	}

	return streaks, nil
}

// updateStaleStreaks resets streaks that haven't been updated
func (s *StreakService) updateStaleStreaks() {
	var streaks []models.Streak
	s.db.Find(&streaks)

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
