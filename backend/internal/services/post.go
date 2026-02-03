package services

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrImageNotFound    = errors.New("image not found")
	ErrMaxImagesReached = errors.New("maximum images per post reached")
	ErrInvalidImageType = errors.New("invalid image type")
	ErrNotAuthorized    = errors.New("not authorized")
)

const MaxImagesPerPost = 4

type PostService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewPostService(db *gorm.DB, cfg *config.Config) *PostService {
	return &PostService{
		db:  db,
		cfg: cfg,
	}
}

type CreatePostInput struct {
	ActivityTypeID string  `json:"activity_type_id" binding:"required"`
	Description    *string `json:"description"`
}

type UpdatePostInput struct {
	Description *string `json:"description"`
}

// ListPosts returns paginated posts
func (s *PostService) ListPosts(page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	if err := s.db.Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := s.db.
		Preload("User").
		Preload("ActivityType").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostByID retrieves a post by ID with all relationships
func (s *PostService) GetPostByID(id string) (*models.Post, error) {
	var post models.Post
	if err := s.db.
		Preload("User").
		Preload("ActivityType").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Reactions").
		Preload("Reactions.User").
		First(&post, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

// CreatePost creates a new post
func (s *PostService) CreatePost(input CreatePostInput, userID string) (*models.Post, error) {
	// Verify activity type exists
	var activityType models.ActivityType
	if err := s.db.First(&activityType, "id = ?", input.ActivityTypeID).Error; err != nil {
		return nil, errors.New("invalid activity type")
	}

	// Sanitize description
	var description *string
	if input.Description != nil {
		desc := sanitizeText(*input.Description, 2000)
		description = &desc
	}

	now := time.Now()
	post := models.Post{
		ID:             uuid.New().String(),
		UserID:         userID,
		ActivityTypeID: input.ActivityTypeID,
		Description:    description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.Create(&post).Error; err != nil {
		return nil, err
	}

	// Reload with relationships
	return s.GetPostByID(post.ID)
}

// UpdatePost updates a post
func (s *PostService) UpdatePost(id string, input UpdatePostInput, userID string, isAdmin bool) (*models.Post, error) {
	var post models.Post
	if err := s.db.First(&post, "id = ?", id).Error; err != nil {
		return nil, ErrPostNotFound
	}

	// Check authorization
	if !isAdmin && post.UserID != userID {
		return nil, ErrNotAuthorized
	}

	if input.Description != nil {
		desc := sanitizeText(*input.Description, 2000)
		post.Description = &desc
	}
	post.UpdatedAt = time.Now()

	if err := s.db.Save(&post).Error; err != nil {
		return nil, err
	}

	return s.GetPostByID(post.ID)
}

// DeletePost soft-deletes a post
func (s *PostService) DeletePost(id, userID string, isAdmin bool) error {
	var post models.Post
	if err := s.db.First(&post, "id = ?", id).Error; err != nil {
		return ErrPostNotFound
	}

	// Check authorization
	if !isAdmin && post.UserID != userID {
		return ErrNotAuthorized
	}

	return s.db.Delete(&post).Error
}

// UploadImage uploads an image to a post
func (s *PostService) UploadImage(postID, userID string, isAdmin bool, file io.Reader, filename string) (*models.PostImage, error) {
	var post models.Post
	if err := s.db.Preload("Images").First(&post, "id = ?", postID).Error; err != nil {
		return nil, ErrPostNotFound
	}

	// Check authorization
	if !isAdmin && post.UserID != userID {
		return nil, ErrNotAuthorized
	}

	// Check max images
	if len(post.Images) >= MaxImagesPerPost {
		return nil, ErrMaxImagesReached
	}

	// Validate and process image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, ErrInvalidImageType
	}

	// Only allow specific formats
	if format != "jpeg" && format != "png" && format != "gif" {
		return nil, ErrInvalidImageType
	}

	// Generate file path
	imageID := uuid.New().String()
	dir := filepath.Join(s.cfg.StoragePath, "images", userID, postID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dir, imageID+".jpg")

	// Create output file
	outFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	// Encode as JPEG (re-encoding for security)
	if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: 85}); err != nil {
		os.Remove(filePath)
		return nil, err
	}

	// Create database record
	displayOrder := len(post.Images)
	postImage := models.PostImage{
		ID:           imageID,
		PostID:       postID,
		FilePath:     filePath,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
	}

	if err := s.db.Create(&postImage).Error; err != nil {
		os.Remove(filePath)
		return nil, err
	}

	return &postImage, nil
}

// DeleteImage removes an image from a post
func (s *PostService) DeleteImage(postID, imageID, userID string, isAdmin bool) error {
	var post models.Post
	if err := s.db.First(&post, "id = ?", postID).Error; err != nil {
		return ErrPostNotFound
	}

	// Check authorization
	if !isAdmin && post.UserID != userID {
		return ErrNotAuthorized
	}

	var postImage models.PostImage
	if err := s.db.First(&postImage, "id = ? AND post_id = ?", imageID, postID).Error; err != nil {
		return ErrImageNotFound
	}

	// Delete file
	os.Remove(postImage.FilePath)

	return s.db.Delete(&postImage).Error
}

// GetImage retrieves an image by ID
func (s *PostService) GetImage(imageID string) (*models.PostImage, error) {
	var image models.PostImage
	if err := s.db.First(&image, "id = ?", imageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	return &image, nil
}

// GetUserPostsForDate returns posts by a user on a specific date
func (s *PostService) GetUserPostsForDate(userID string, date time.Time, timezone *time.Location) ([]models.Post, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, timezone)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var posts []models.Post
	if err := s.db.
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfDay, endOfDay).
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetUserPostsInPeriod returns posts by a user in a time period
func (s *PostService) GetUserPostsInPeriod(userID string, start, end time.Time, activityTypeID *string) ([]models.Post, error) {
	query := s.db.Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end)

	if activityTypeID != nil {
		query = query.Where("activity_type_id = ?", *activityTypeID)
	}

	var posts []models.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// sanitizeText removes potentially dangerous content and truncates
func sanitizeText(s string, maxLen int) string {
	// Remove HTML tags (simple approach)
	s = stripHTMLTags(s)
	s = strings.TrimSpace(s)

	if len(s) > maxLen {
		s = s[:maxLen]
	}

	return s
}

func stripHTMLTags(s string) string {
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

// GetImagePath returns the full filesystem path for an image
func (s *PostService) GetImagePath(imageID string) (string, error) {
	var image models.PostImage
	if err := s.db.First(&image, "id = ?", imageID).Error; err != nil {
		return "", fmt.Errorf("image not found")
	}
	return image.FilePath, nil
}

// CreateChallengeCompletionPost creates a post announcing challenge completion
func (s *PostService) CreateChallengeCompletionPost(userID string, challengeTitle string) (*models.Post, error) {
	// Find the Achievement activity type
	var achievementActivity models.ActivityType
	if err := s.db.Where("name = ?", "Achievement").First(&achievementActivity).Error; err != nil {
		return nil, fmt.Errorf("achievement activity type not found")
	}

	description := fmt.Sprintf("🎉 Completed the \"%s\" challenge!", challengeTitle)

	now := time.Now()
	post := models.Post{
		ID:             uuid.New().String(),
		UserID:         userID,
		ActivityTypeID: achievementActivity.ID,
		Description:    &description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.Create(&post).Error; err != nil {
		return nil, err
	}

	return s.GetPostByID(post.ID)
}
