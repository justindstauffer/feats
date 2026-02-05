package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type PostHandler struct {
	postService      *services.PostService
	streakService    *services.StreakService
	challengeService *services.ChallengeService
	goalService      *services.GoalService
	auditService     *services.AuditService
	cfg              *config.Config
}

func NewPostHandler(
	postService *services.PostService,
	streakService *services.StreakService,
	challengeService *services.ChallengeService,
	goalService *services.GoalService,
	auditService *services.AuditService,
	cfg *config.Config,
) *PostHandler {
	return &PostHandler{
		postService:      postService,
		streakService:    streakService,
		challengeService: challengeService,
		goalService:      goalService,
		auditService:     auditService,
		cfg:              cfg,
	}
}

func (h *PostHandler) ListPosts(c *gin.Context) {
	page := getPageParam(c)
	perPage := getPerPageParam(c)

	posts, total, err := h.postService.ListPosts(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedSuccessResponse(posts, page, perPage, int(total)))
}

func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("id")

	post, err := h.postService.GetPostByID(postID)
	if err != nil {
		if err == services.ErrPostNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(post))
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var input services.CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	post, err := h.postService.CreatePost(input, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			err.Error(),
		))
		return
	}

	// Update streak
	h.streakService.UpdateStreakForActivity(userID, time.Now())

	// Update challenge progress and get completed challenges
	completedChallengeIDs, _ := h.challengeService.UpdateProgressForActivity(userID, input.ActivityTypeID)

	// Create completion posts for any just-completed challenges
	for _, challengeID := range completedChallengeIDs {
		challenge, err := h.challengeService.GetChallengeByID(challengeID)
		if err == nil {
			h.postService.CreateChallengeCompletionPost(userID, challenge.Title)
		}
	}

	// Update goal progress
	h.goalService.UpdateProgressForActivity(userID, input.ActivityTypeID)

	c.JSON(http.StatusCreated, models.SuccessResponse(post))
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID := c.Param("id")

	var input services.UpdatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	post, err := h.postService.UpdatePost(postID, input, userID, isAdmin)
	if err != nil {
		switch err {
		case services.ErrPostNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
		case services.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not authorized to update this post",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(post))
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	postID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	err := h.postService.DeletePost(postID, userID, isAdmin)
	if err != nil {
		switch err {
		case services.ErrPostNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
		case services.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not authorized to delete this post",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Post deleted",
	}))
}

func (h *PostHandler) UploadImage(c *gin.Context) {
	postID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"No image file provided",
		))
		return
	}
	defer file.Close()

	image, err := h.postService.UploadImage(postID, userID, isAdmin, file, header.Filename)
	if err != nil {
		switch err {
		case services.ErrPostNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
		case services.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not authorized to add images to this post",
			))
		case services.ErrMaxImagesReached:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Maximum images per post reached",
			))
		case services.ErrInvalidImageType:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invalid image type. Allowed: JPEG, PNG",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(image))
}

func (h *PostHandler) DeleteImage(c *gin.Context) {
	postID := c.Param("id")
	imageID := c.Param("image_id")
	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	err := h.postService.DeleteImage(postID, imageID, userID, isAdmin)
	if err != nil {
		switch err {
		case services.ErrPostNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Post not found",
			))
		case services.ErrImageNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Image not found",
			))
		case services.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.ErrCodeForbidden,
				"Not authorized to delete this image",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Image deleted",
	}))
}

func (h *PostHandler) ServeImage(c *gin.Context) {
	imageID := c.Param("id")

	imagePath, err := h.postService.GetImagePath(imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"Image not found",
		))
		return
	}

	// Path traversal protection: ensure the image path is within the storage directory
	absImagePath, err := filepath.Abs(imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	absStoragePath, err := filepath.Abs(h.cfg.StoragePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	// Ensure the image path starts with the storage path (prevent path traversal)
	if !strings.HasPrefix(absImagePath, absStoragePath+string(filepath.Separator)) {
		// Log the traversal attempt
		userID, _ := middleware.GetCurrentUserID(c)
		log.Printf("SECURITY: Path traversal attempt by user %s, attempted path: %s", userID, imagePath)
		h.auditService.Log(services.AuditLogInput{
			UserID:  &userID,
			Action:  models.AuditActionAuthorizationFail,
			Details: map[string]interface{}{
				"reason":         "path_traversal_attempt",
				"attempted_path": imagePath,
			},
			Success: false,
		})
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.ErrCodeForbidden,
			"Access denied",
		))
		return
	}

	// Check file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.ErrCodeNotFound,
			"Image not found",
		))
		return
	}

	// Set security headers
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", "inline")

	// Serve the file
	c.File(imagePath)
}

// Helper functions

func getPageParam(c *gin.Context) int {
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed := parseInt(p); parsed > 0 {
			page = parsed
		}
	}
	return page
}

func getPerPageParam(c *gin.Context) int {
	perPage := 20
	if p := c.Query("per_page"); p != "" {
		if parsed := parseInt(p); parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}
	return perPage
}

func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		result = result*10 + int(c-'0')
	}
	return result
}

func getFileExtension(filename string) string {
	return filepath.Ext(filename)
}
