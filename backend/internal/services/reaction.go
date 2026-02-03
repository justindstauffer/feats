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

// GetReactions returns all reactions for a post with summary
func (s *ReactionService) GetReactions(postID string) ([]models.ReactionSummary, []models.Reaction, error) {
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

// AddReaction adds or updates a reaction
func (s *ReactionService) AddReaction(postID, userID string, reactionType models.ReactionType) (*models.Reaction, bool, error) {
	if !models.IsValidReactionType(reactionType) {
		return nil, false, ErrInvalidReaction
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
		existing.ReactionType = reactionType
		if err := s.db.Save(&existing).Error; err != nil {
			return nil, false, err
		}
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

// RemoveReaction removes a user's reaction from a post
func (s *ReactionService) RemoveReaction(postID, userID string) error {
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
