package services

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrAlreadyParticipating = errors.New("already participating in challenge")
	ErrNotParticipating     = errors.New("not participating in challenge")
	ErrChallengeEnded       = errors.New("challenge has ended")
)

type ChallengeService struct {
	db *gorm.DB
}

func NewChallengeService(db *gorm.DB) *ChallengeService {
	return &ChallengeService{db: db}
}

type CreateChallengeInput struct {
	Title          string     `json:"title" binding:"required"`
	Description    *string    `json:"description"`
	ActivityTypeID *string    `json:"activity_type_id"`
	TargetCount    int        `json:"target_count" binding:"required,min=1"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
}

// ListChallenges returns all active challenges for a group
func (s *ChallengeService) ListChallenges(groupID string, includeExpired bool) ([]models.Challenge, error) {
	var challenges []models.Challenge

	query := s.db.
		Where("group_id = ?", groupID).
		Preload("Creator").
		Preload("ActivityType").
		Preload("Participants").
		Preload("Participants.User")

	if !includeExpired {
		now := time.Now().In(config.Location)
		// Convert to UTC for SQLite comparison (SQLite compares dates as strings)
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.Location).UTC()
		query = query.Where("end_date IS NULL OR end_date >= ?", startOfToday)
	}

	if err := query.Order("created_at DESC").Find(&challenges).Error; err != nil {
		return nil, err
	}

	return challenges, nil
}

// GetChallengeByID retrieves a challenge by ID within a group
func (s *ChallengeService) GetChallengeByID(groupID, id string) (*models.Challenge, error) {
	var challenge models.Challenge
	if err := s.db.
		Preload("Creator").
		Preload("ActivityType").
		Preload("Participants", func(db *gorm.DB) *gorm.DB {
			return db.Order("progress DESC, joined_at ASC")
		}).
		Preload("Participants.User").
		First(&challenge, "id = ? AND group_id = ?", id, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChallengeNotFound
		}
		return nil, err
	}
	return &challenge, nil
}

// CreateChallenge creates a new challenge within a group
func (s *ChallengeService) CreateChallenge(groupID string, input CreateChallengeInput, userID string) (*models.Challenge, error) {
	title := strings.TrimSpace(input.Title)
	if len(title) > 100 {
		title = title[:100]
	}

	var description *string
	if input.Description != nil {
		desc := strings.TrimSpace(*input.Description)
		if len(desc) > 500 {
			desc = desc[:500]
		}
		description = &desc
	}

	// Validate activity type if provided (system-wide or group-specific)
	if input.ActivityTypeID != nil {
		var activity models.ActivityType
		if err := s.db.First(&activity, "id = ? AND (group_id IS NULL OR group_id = ?)", *input.ActivityTypeID, groupID).Error; err != nil {
			return nil, errors.New("invalid activity type")
		}
	}

	// Validate dates
	if input.StartDate != nil && input.EndDate != nil {
		if input.EndDate.Before(*input.StartDate) {
			return nil, errors.New("end date must be after start date")
		}
	}

	challenge := models.Challenge{
		ID:             uuid.New().String(),
		GroupID:        groupID,
		CreatedBy:      userID,
		Title:          title,
		Description:    description,
		ActivityTypeID: input.ActivityTypeID,
		TargetCount:    input.TargetCount,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
		CreatedAt:      time.Now(),
	}

	if err := s.db.Create(&challenge).Error; err != nil {
		return nil, err
	}

	// Auto-join creator
	participant := models.ChallengeParticipant{
		ID:          uuid.New().String(),
		ChallengeID: challenge.ID,
		UserID:      userID,
		Progress:    0,
		JoinedAt:    time.Now(),
	}
	s.db.Create(&participant)

	return s.GetChallengeByID(groupID, challenge.ID)
}

// JoinChallenge adds a user to a challenge within a group
func (s *ChallengeService) JoinChallenge(groupID, challengeID, userID string) error {
	challenge, err := s.GetChallengeByID(groupID, challengeID)
	if err != nil {
		return err
	}

	// Check if challenge has ended (allow joining until end of the end date day)
	if challenge.EndDate != nil {
		now := time.Now().UTC()
		// Convert end date to UTC for comparison
		endDateUTC := challenge.EndDate.UTC()
		endOfEndDate := time.Date(
			endDateUTC.Year(), endDateUTC.Month(), endDateUTC.Day(),
			23, 59, 59, 0, time.UTC,
		)
		if now.After(endOfEndDate) {
			return ErrChallengeEnded
		}
	}

	// Check if already participating
	var existing models.ChallengeParticipant
	if err := s.db.Where("challenge_id = ? AND user_id = ?", challengeID, userID).First(&existing).Error; err == nil {
		return ErrAlreadyParticipating
	}

	participant := models.ChallengeParticipant{
		ID:          uuid.New().String(),
		ChallengeID: challengeID,
		UserID:      userID,
		Progress:    0,
		JoinedAt:    time.Now(),
	}

	return s.db.Create(&participant).Error
}

// LeaveChallenge removes a user from a challenge within a group
func (s *ChallengeService) LeaveChallenge(groupID, challengeID, userID string) error {
	// Verify challenge belongs to group
	var challenge models.Challenge
	if err := s.db.First(&challenge, "id = ? AND group_id = ?", challengeID, groupID).Error; err != nil {
		return ErrChallengeNotFound
	}

	result := s.db.Where("challenge_id = ? AND user_id = ?", challengeID, userID).Delete(&models.ChallengeParticipant{})
	if result.RowsAffected == 0 {
		return ErrNotParticipating
	}
	return result.Error
}

// DeleteChallenge deletes a challenge (creator or admin only)
func (s *ChallengeService) DeleteChallenge(groupID, challengeID, userID string, isAdmin bool) error {
	var challenge models.Challenge
	if err := s.db.First(&challenge, "id = ? AND group_id = ?", challengeID, groupID).Error; err != nil {
		return ErrChallengeNotFound
	}

	if !isAdmin && challenge.CreatedBy != userID {
		return ErrNotAuthorized
	}

	// Delete participants first
	s.db.Where("challenge_id = ?", challengeID).Delete(&models.ChallengeParticipant{})

	return s.db.Delete(&challenge).Error
}

// UpdateProgressForActivity updates challenge progress when a user posts an activity within a group
func (s *ChallengeService) UpdateProgressForActivity(groupID, userID string, activityTypeID string) ([]string, error) {
	var completedChallenges []string

	log.Printf("[Challenge] UpdateProgressForActivity called: groupID=%s, userID=%s, activityTypeID=%s", groupID, userID, activityTypeID)

	// Get current date boundaries for proper comparison using configured timezone
	// Convert to UTC for SQLite comparison (SQLite compares dates as strings)
	now := time.Now().In(config.Location)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.Location).UTC()
	endOfToday := startOfToday.Add(24 * time.Hour)
	log.Printf("[Challenge] Date boundaries (UTC): startOfToday=%v, endOfToday=%v", startOfToday, endOfToday)

	// Find all active challenges the user is participating in within this group
	var participants []models.ChallengeParticipant
	if err := s.db.
		Joins("JOIN challenges ON challenges.id = challenge_participants.challenge_id").
		Where("challenges.group_id = ?", groupID).
		Where("challenge_participants.user_id = ?", userID).
		Where("challenge_participants.completed_at IS NULL").
		Where("challenges.end_date IS NULL OR challenges.end_date >= ?", startOfToday).
		Where("challenges.start_date IS NULL OR challenges.start_date < ?", endOfToday).
		Find(&participants).Error; err != nil {
		log.Printf("[Challenge] Error finding participants: %v", err)
		return nil, err
	}

	log.Printf("[Challenge] Found %d active challenge participations for user", len(participants))

	for _, participant := range participants {
		// Get the challenge
		var challenge models.Challenge
		if err := s.db.First(&challenge, "id = ?", participant.ChallengeID).Error; err != nil {
			log.Printf("[Challenge] Error loading challenge %s: %v", participant.ChallengeID, err)
			continue
		}

		log.Printf("[Challenge] Checking challenge '%s' (ID: %s), activityTypeID: %v",
			challenge.Title, challenge.ID, challenge.ActivityTypeID)

		// Check if activity type matches (if challenge has specific type)
		if challenge.ActivityTypeID != nil && *challenge.ActivityTypeID != activityTypeID {
			log.Printf("[Challenge] Skipping - activity type mismatch: challenge wants %s, post has %s",
				*challenge.ActivityTypeID, activityTypeID)
			continue
		}

		// Increment progress
		oldProgress := participant.Progress
		justCompleted := participant.IncrementProgress(challenge.TargetCount)
		s.db.Save(&participant)

		log.Printf("[Challenge] Updated progress for challenge '%s': %d -> %d (target: %d, completed: %v)",
			challenge.Title, oldProgress, participant.Progress, challenge.TargetCount, justCompleted)

		if justCompleted {
			completedChallenges = append(completedChallenges, challenge.ID)
		}
	}

	return completedChallenges, nil
}

// GetParticipantIDs returns all participant user IDs for a challenge
func (s *ChallengeService) GetParticipantIDs(challengeID string) ([]string, error) {
	var ids []string
	if err := s.db.Model(&models.ChallengeParticipant{}).
		Where("challenge_id = ?", challengeID).
		Pluck("user_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
