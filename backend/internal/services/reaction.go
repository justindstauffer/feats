package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrReactionNotFound = errors.New("reaction not found")
	ErrInvalidReaction  = errors.New("invalid reaction type")
)

type ReactionService struct {
	db *gorm.DB
}

func NewReactionService(db *gorm.DB) *ReactionService {
	return &ReactionService{db: db}
}

type AddReactionInput struct {
	ReactionType int `json:"reaction_type" binding:"required"`
}

// GetReactions returns all reactions for a post with summary (validates post belongs to group)
func (s *ReactionService) GetReactions(groupID, postID string) ([]models.ReactionSummary, []models.Reaction, error) {
	// Verify post belongs to group
	var post models.Post
	if err := s.db.First(&post, "id = ? AND group_id = ?", postID, groupID).Error; err != nil {
		return nil, nil, ErrPostNotFound
	}

	var reactions []models.Reaction
	if err := s.db.
		Preload("User").
		Where("post_id = ?", postID).
		Find(&reactions).Error; err != nil {
		return nil, nil, err
	}

	// Build summary
	counts := make(map[models.ReactionType]int)
	for _, r := range reactions {
		counts[r.ReactionType]++
	}

	var summaries []models.ReactionSummary
	for rt, count := range counts {
		summaries = append(summaries, models.ReactionSummary{
			Type:  rt,
			Emoji: models.ReactionEmojis[rt],
			Count: count,
		})
	}

	return summaries, reactions, nil
}

// AddReaction adds or updates a reaction (validates post belongs to group)
func (s *ReactionService) AddReaction(groupID, postID, userID string, reactionType models.ReactionType) (*models.Reaction, bool, error) {
	if !models.IsValidReactionType(reactionType) {
		return nil, false, ErrInvalidReaction
	}

	// Verify post belongs to group
	var post models.Post
	if err := s.db.First(&post, "id = ? AND group_id = ?", postID, groupID).Error; err != nil {
		return nil, false, ErrPostNotFound
	}

	// Check if user already reacted
	var existing models.Reaction
	err := s.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&existing).Error

	if err == nil {
		// Update existing reaction
		if existing.ReactionType == reactionType {
			// Same reaction, no change needed
			return &existing, false, nil
		}
		if err := s.db.Model(&models.Reaction{}).
			Where("id = ?", existing.ID).
			Update("reaction_type", reactionType).Error; err != nil {
			return nil, false, err
		}
		existing.ReactionType = reactionType
		_ = s.db.Preload("User").First(&existing, "id = ?", existing.ID).Error
		return &existing, false, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	// Create new reaction
	reaction := models.Reaction{
		ID:           uuid.New().String(),
		UserID:       userID,
		PostID:       postID,
		ReactionType: reactionType,
		CreatedAt:    time.Now(),
	}

	if err := s.db.Create(&reaction).Error; err != nil {
		return nil, false, err
	}

	// Load user for response
	s.db.Preload("User").First(&reaction, "id = ?", reaction.ID)

	return &reaction, true, nil
}

// RemoveReaction removes a user's reaction from a post (validates post belongs to group)
func (s *ReactionService) RemoveReaction(groupID, postID, userID string) error {
	// Verify post belongs to group
	var post models.Post
	if err := s.db.First(&post, "id = ? AND group_id = ?", postID, groupID).Error; err != nil {
		return ErrPostNotFound
	}

	result := s.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&models.Reaction{})
	if result.RowsAffected == 0 {
		return ErrReactionNotFound
	}
	return result.Error
}

// GetPostOwnerID returns the owner ID of a post (for notifications)
func (s *ReactionService) GetPostOwnerID(postID string) (string, error) {
	var post models.Post
	if err := s.db.Select("user_id").First(&post, "id = ?", postID).Error; err != nil {
		return "", err
	}
	return post.UserID, nil
}
