package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrActivityNotFound    = errors.New("activity type not found")
	ErrActivityExists      = errors.New("activity type already exists")
	ErrCannotDeleteSystem  = errors.New("cannot delete system activity type")
	ErrActivityInUse       = errors.New("activity type is in use")
)

type ActivityService struct {
	db *gorm.DB
}

func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{db: db}
}

type CreateActivityInput struct {
	Name string  `json:"name" binding:"required"`
	Icon *string `json:"icon"`
}

// ListActivities returns all activity types for a group (system-wide + group-specific, excludes Achievement which is system-only)
func (s *ActivityService) ListActivities(groupID string) ([]models.ActivityType, error) {
	var activities []models.ActivityType
	if err := s.db.
		Where("name != ? AND (group_id IS NULL OR group_id = ?)", "Achievement", groupID).
		Order("is_system DESC, name ASC").
		Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

// GetActivityByID retrieves an activity type by ID (must be system-wide or belong to the group)
func (s *ActivityService) GetActivityByID(groupID, id string) (*models.ActivityType, error) {
	var activity models.ActivityType
	if err := s.db.First(&activity, "id = ? AND (group_id IS NULL OR group_id = ?)", id, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// CreateActivity creates a custom activity type for a group
func (s *ActivityService) CreateActivity(groupID string, input CreateActivityInput, userID string) (*models.ActivityType, error) {
	name := strings.TrimSpace(input.Name)

	// Check for duplicate name (within system-wide or this group)
	var existing models.ActivityType
	if err := s.db.Where("LOWER(name) = LOWER(?) AND (group_id IS NULL OR group_id = ?)", name, groupID).First(&existing).Error; err == nil {
		return nil, ErrActivityExists
	}

	activity := models.ActivityType{
		ID:        uuid.New().String(),
		GroupID:   &groupID,
		Name:      name,
		Icon:      input.Icon,
		IsSystem:  false,
		CreatedBy: &userID,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&activity).Error; err != nil {
		return nil, err
	}

	return &activity, nil
}

// DeleteActivity deletes a custom activity type within a group
func (s *ActivityService) DeleteActivity(groupID, id, userID string, isAdmin bool) error {
	var activity models.ActivityType
	if err := s.db.First(&activity, "id = ? AND group_id = ?", id, groupID).Error; err != nil {
		return ErrActivityNotFound
	}

	// Cannot delete system activities
	if activity.IsSystem {
		return ErrCannotDeleteSystem
	}

	// Only creator or group admin can delete
	if !isAdmin && (activity.CreatedBy == nil || *activity.CreatedBy != userID) {
		return errors.New("not authorized to delete this activity")
	}

	// Check if activity is in use within the group
	var postCount int64
	s.db.Model(&models.Post{}).Where("activity_type_id = ? AND group_id = ?", id, groupID).Count(&postCount)
	if postCount > 0 {
		return ErrActivityInUse
	}

	return s.db.Delete(&activity).Error
}
