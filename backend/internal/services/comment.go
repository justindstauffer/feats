package services

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrCommentTooLong  = errors.New("comment exceeds maximum length")
)

const MaxCommentLength = 1000

type CommentService struct {
	db *gorm.DB
}

func NewCommentService(db *gorm.DB) *CommentService {
	return &CommentService{db: db}
}

type CreateCommentInput struct {
	Content  string  `json:"content" binding:"required"`
	ParentID *string `json:"parent_id"`
}

type UpdateCommentInput struct {
	Content string `json:"content" binding:"required"`
}

// GetComments returns threaded comments for a post
func (s *CommentService) GetComments(postID string) ([]models.Comment, error) {
	var comments []models.Comment

	// Get top-level comments first
	if err := s.db.
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Replies.User").
		Where("post_id = ? AND parent_id IS NULL", postID).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, err
	}

	return comments, nil
}

// CreateComment creates a new comment or reply
func (s *CommentService) CreateComment(postID, userID string, input CreateCommentInput) (*models.Comment, error) {
	content := sanitizeComment(input.Content)
	if len(content) > MaxCommentLength {
		return nil, ErrCommentTooLong
	}

	// Verify post exists
	var post models.Post
	if err := s.db.First(&post, "id = ?", postID).Error; err != nil {
		return nil, ErrPostNotFound
	}

	// If reply, verify parent exists and belongs to same post
	if input.ParentID != nil {
		var parent models.Comment
		if err := s.db.First(&parent, "id = ? AND post_id = ?", *input.ParentID, postID).Error; err != nil {
			return nil, errors.New("parent comment not found")
		}
	}

	now := time.Now()
	comment := models.Comment{
		ID:        uuid.New().String(),
		PostID:    postID,
		UserID:    userID,
		ParentID:  input.ParentID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.db.Create(&comment).Error; err != nil {
		return nil, err
	}

	// Reload with user
	s.db.Preload("User").First(&comment, "id = ?", comment.ID)

	return &comment, nil
}

// UpdateComment updates a comment's content
func (s *CommentService) UpdateComment(commentID, userID string, isAdmin bool, input UpdateCommentInput) (*models.Comment, error) {
	var comment models.Comment
	if err := s.db.First(&comment, "id = ?", commentID).Error; err != nil {
		return nil, ErrCommentNotFound
	}

	// Check authorization
	if !isAdmin && comment.UserID != userID {
		return nil, ErrNotAuthorized
	}

	content := sanitizeComment(input.Content)
	if len(content) > MaxCommentLength {
		return nil, ErrCommentTooLong
	}

	comment.Content = content
	comment.UpdatedAt = time.Now()

	if err := s.db.Save(&comment).Error; err != nil {
		return nil, err
	}

	// Reload with user
	s.db.Preload("User").First(&comment, "id = ?", comment.ID)

	return &comment, nil
}

// DeleteComment soft-deletes a comment
func (s *CommentService) DeleteComment(commentID, userID string, isAdmin bool) error {
	var comment models.Comment
	if err := s.db.First(&comment, "id = ?", commentID).Error; err != nil {
		return ErrCommentNotFound
	}

	// Check authorization
	if !isAdmin && comment.UserID != userID {
		return ErrNotAuthorized
	}

	return s.db.Delete(&comment).Error
}

// GetCommentByID retrieves a comment by ID
func (s *CommentService) GetCommentByID(commentID string) (*models.Comment, error) {
	var comment models.Comment
	if err := s.db.Preload("User").First(&comment, "id = ?", commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

// GetPostOwnerID returns the post owner for a comment's post
func (s *CommentService) GetPostOwnerID(commentID string) (string, error) {
	var comment models.Comment
	if err := s.db.First(&comment, "id = ?", commentID).Error; err != nil {
		return "", err
	}

	var post models.Post
	if err := s.db.First(&post, "id = ?", comment.PostID).Error; err != nil {
		return "", err
	}

	return post.UserID, nil
}

// GetParentCommentOwnerID returns the owner of a comment's parent (for reply notifications)
func (s *CommentService) GetParentCommentOwnerID(commentID string) (*string, error) {
	var comment models.Comment
	if err := s.db.First(&comment, "id = ?", commentID).Error; err != nil {
		return nil, err
	}

	if comment.ParentID == nil {
		return nil, nil
	}

	var parent models.Comment
	if err := s.db.First(&parent, "id = ?", *comment.ParentID).Error; err != nil {
		return nil, err
	}

	return &parent.UserID, nil
}

func sanitizeComment(s string) string {
	// Strip HTML
	s = stripHTML(s)
	s = strings.TrimSpace(s)
	// Escape any remaining HTML special characters
	return html.EscapeString(s)
}

func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}
