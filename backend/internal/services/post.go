package services

import (
	"bytes"
	"errors"
	"fmt"
	"html"
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
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/draw"
	"gorm.io/gorm"
)

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrImageNotFound    = errors.New("image not found")
	ErrMaxImagesReached = errors.New("maximum images per post reached")
	ErrInvalidImageType = errors.New("invalid image type")
	ErrNotAuthorized    = errors.New("not authorized")
)

const (
	MaxImagesPerPost   = 4
	MaxImageDimension  = 2048 // Max width or height in pixels
)

// Magic bytes for image format validation
var imageMagicBytes = map[string][]byte{
	"jpeg": {0xFF, 0xD8, 0xFF},
	"png":  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	"gif":  {0x47, 0x49, 0x46, 0x38},
}

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

// ListPosts returns paginated posts for a group
func (s *PostService) ListPosts(groupID string, page, perPage int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	if err := s.db.Model(&models.Post{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := s.db.
		Where("group_id = ?", groupID).
		Preload("User").
		Preload("ActivityType").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Reactions").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	// Get comment counts for all posts in a single query
	if len(posts) > 0 {
		postIDs := make([]string, len(posts))
		for i, post := range posts {
			postIDs[i] = post.ID
		}

		var counts []struct {
			PostID string
			Count  int
		}
		s.db.Model(&models.Comment{}).
			Select("post_id, COUNT(*) as count").
			Where("post_id IN ?", postIDs).
			Group("post_id").
			Scan(&counts)

		countMap := make(map[string]int)
		for _, c := range counts {
			countMap[c.PostID] = c.Count
		}

		for i := range posts {
			posts[i].CommentCount = countMap[posts[i].ID]
		}
	}

	return posts, total, nil
}

// GetPostByID retrieves a post by ID with all relationships
func (s *PostService) GetPostByID(groupID, id string) (*models.Post, error) {
	var post models.Post
	if err := s.db.
		Preload("User").
		Preload("ActivityType").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Reactions").
		Preload("Reactions.User").
		First(&post, "id = ? AND group_id = ?", id, groupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

// CreatePost creates a new post
func (s *PostService) CreatePost(groupID string, input CreatePostInput, userID string) (*models.Post, error) {
	// Verify activity type exists (system-wide or group-specific)
	var activityType models.ActivityType
	if err := s.db.First(&activityType, "id = ? AND (group_id IS NULL OR group_id = ?)", input.ActivityTypeID, groupID).Error; err != nil {
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
		GroupID:        groupID,
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
	return s.GetPostByID(groupID, post.ID)
}

// UpdatePost updates a post
func (s *PostService) UpdatePost(groupID, id string, input UpdatePostInput, userID string, isAdmin bool) (*models.Post, error) {
	var post models.Post
	if err := s.db.First(&post, "id = ? AND group_id = ?", id, groupID).Error; err != nil {
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

	return s.GetPostByID(groupID, post.ID)
}

// DeletePost soft-deletes a post
func (s *PostService) DeletePost(groupID, id, userID string, isAdmin bool) error {
	var post models.Post
	if err := s.db.First(&post, "id = ? AND group_id = ?", id, groupID).Error; err != nil {
		return ErrPostNotFound
	}

	// Check authorization
	if !isAdmin && post.UserID != userID {
		return ErrNotAuthorized
	}

	return s.db.Delete(&post).Error
}

// readExifOrientation reads the EXIF orientation tag from image data
// Returns orientation value 1-8, or 1 (normal) if not found
func readExifOrientation(data []byte) int {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return 1 // Default to normal orientation
	}

	orientTag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	orient, err := orientTag.Int(0)
	if err != nil {
		return 1
	}

	return orient
}

// fixOrientation applies EXIF orientation transformation to an image
// EXIF orientation values:
// 1 = Normal, 2 = Flip H, 3 = Rotate 180, 4 = Flip V
// 5 = Transpose, 6 = Rotate 90 CW, 7 = Transverse, 8 = Rotate 90 CCW
func fixOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 1:
		return img // Normal - no change
	case 2:
		return flipHorizontal(img)
	case 3:
		return rotate180(img)
	case 4:
		return flipVertical(img)
	case 5:
		return transpose(img)
	case 6:
		return rotate90CW(img)
	case 7:
		return transverse(img)
	case 8:
		return rotate90CCW(img)
	default:
		return img
	}
}

func flipHorizontal(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(bounds.Max.X-1-x, y, img.At(x, y))
		}
	}
	return result
}

func flipVertical(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, bounds.Max.Y-1-y, img.At(x, y))
		}
	}
	return result
}

func rotate90CW(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	result := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(h-1-y, x, img.At(x, y))
		}
	}
	return result
}

func rotate90CCW(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	result := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(y, w-1-x, img.At(x, y))
		}
	}
	return result
}

func rotate180(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(bounds.Max.X-1-x, bounds.Max.Y-1-y, img.At(x, y))
		}
	}
	return result
}

func transpose(img image.Image) image.Image {
	// Rotate 90 CW then flip horizontal
	return flipHorizontal(rotate90CW(img))
}

func transverse(img image.Image) image.Image {
	// Rotate 90 CCW then flip horizontal
	return flipHorizontal(rotate90CCW(img))
}

// resizeIfNeeded scales down an image if it exceeds MaxImageDimension
// Maintains aspect ratio, only shrinks (never enlarges)
func resizeIfNeeded(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Find the longest side
	maxSide := max(w, h)

	// If within limits, return original
	if maxSide <= MaxImageDimension {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newW, newH int
	if w > h {
		newW = MaxImageDimension
		newH = h * MaxImageDimension / w
	} else {
		newH = MaxImageDimension
		newW = w * MaxImageDimension / h
	}

	// Create resized image using high-quality CatmullRom interpolation
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

// UploadImage uploads an image to a post
func (s *PostService) UploadImage(groupID, postID, userID string, isAdmin bool, file io.Reader, filename string) (*models.PostImage, error) {
	var post models.Post
	if err := s.db.Preload("Images").First(&post, "id = ? AND group_id = ?", postID, groupID).Error; err != nil {
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

	// Read all image data into buffer (needed for EXIF reading and decoding)
	imageData, err := io.ReadAll(file)
	if err != nil {
		return nil, ErrInvalidImageType
	}

	// Validate magic bytes
	if len(imageData) < 8 {
		return nil, ErrInvalidImageType
	}
	validFormat := false
	for _, magic := range imageMagicBytes {
		if len(imageData) >= len(magic) && bytes.Equal(imageData[:len(magic)], magic) {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return nil, ErrInvalidImageType
	}

	// Read EXIF orientation before decoding (only works for JPEG)
	orientation := readExifOrientation(imageData)

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, ErrInvalidImageType
	}

	// Only allow specific formats
	if format != "jpeg" && format != "png" && format != "gif" {
		return nil, ErrInvalidImageType
	}

	// Apply EXIF orientation fix
	img = fixOrientation(img, orientation)

	// Resize if larger than max dimension
	img = resizeIfNeeded(img)

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

	// Encode as JPEG (re-encoding for security, orientation already fixed)
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
func (s *PostService) DeleteImage(groupID, postID, imageID, userID string, isAdmin bool) error {
	var post models.Post
	if err := s.db.First(&post, "id = ? AND group_id = ?", postID, groupID).Error; err != nil {
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

// GetUserPostsForDate returns posts by a user on a specific date within a group
func (s *PostService) GetUserPostsForDate(groupID, userID string, date time.Time, timezone *time.Location) ([]models.Post, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, timezone)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var posts []models.Post
	if err := s.db.
		Where("group_id = ? AND user_id = ? AND created_at >= ? AND created_at < ?", groupID, userID, startOfDay, endOfDay).
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetUserPostsInPeriod returns posts by a user in a time period within a group
func (s *PostService) GetUserPostsInPeriod(groupID, userID string, start, end time.Time, activityTypeID *string) ([]models.Post, error) {
	query := s.db.Where("group_id = ? AND user_id = ? AND created_at >= ? AND created_at < ?", groupID, userID, start, end)

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

	// Escape any remaining HTML special characters
	return html.EscapeString(s)
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

// GetAuthorizedImagePath returns an image path only if the requester is authorized.
func (s *PostService) GetAuthorizedImagePath(imageID, userID string, isAdmin bool) (string, error) {
	var row struct {
		FilePath string
		GroupID  string
	}

	if err := s.db.
		Table("post_images").
		Select("post_images.file_path, posts.group_id").
		Joins("JOIN posts ON posts.id = post_images.post_id").
		Where("post_images.id = ?", imageID).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrImageNotFound
		}
		return "", err
	}

	if isAdmin {
		return row.FilePath, nil
	}

	var membershipCount int64
	if err := s.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND left_at IS NULL", row.GroupID, userID).
		Count(&membershipCount).Error; err != nil {
		return "", err
	}
	if membershipCount == 0 {
		return "", ErrNotAuthorized
	}

	return row.FilePath, nil
}

// CreateChallengeCompletionPost creates a post announcing challenge completion
func (s *PostService) CreateChallengeCompletionPost(groupID, userID string, challengeTitle string) (*models.Post, error) {
	// Find the Achievement activity type (system-wide)
	var achievementActivity models.ActivityType
	if err := s.db.Where("name = ? AND group_id IS NULL", "Achievement").First(&achievementActivity).Error; err != nil {
		return nil, fmt.Errorf("achievement activity type not found")
	}

	description := fmt.Sprintf("🎉 Completed the \"%s\" challenge!", challengeTitle)

	now := time.Now()
	post := models.Post{
		ID:             uuid.New().String(),
		GroupID:        groupID,
		UserID:         userID,
		ActivityTypeID: achievementActivity.ID,
		Description:    &description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.db.Create(&post).Error; err != nil {
		return nil, err
	}

	return s.GetPostByID(groupID, post.ID)
}
