package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailExists    = errors.New("email already exists")
	ErrInvalidRole    = errors.New("invalid role")
)

type UserService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewUserService(db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{
		db:  db,
		cfg: cfg,
	}
}

type CreateUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"`
}

type UpdateUserInput struct {
	Name           *string `json:"name"`
	Bio            *string `json:"bio"`
	ProfilePicture *string `json:"profile_picture"`
}

// CreateUser creates a new user (admin only)
func (s *UserService) CreateUser(input CreateUserInput, authService *AuthService) (*models.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// Check if email exists
	var existing models.User
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailExists
	}

	// Validate role
	role := models.RoleUser
	if input.Role != "" {
		if input.Role != string(models.RoleAdmin) && input.Role != string(models.RoleUser) {
			return nil, ErrInvalidRole
		}
		role = models.UserRole(input.Role)
	}

	// Hash password
	passwordHash, err := authService.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := models.User{
		ID:                uuid.New().String(),
		Email:             email,
		PasswordHash:      passwordHash,
		Name:              strings.TrimSpace(input.Name),
		Role:              role,
		PasswordChangedAt: now,
		ForcePasswordChange: true, // Require password change on first login
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Create initial streak record
	streak := models.Streak{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		CurrentStreak: 0,
		LongestStreak: 0,
		UpdatedAt:     now,
	}
	s.db.Create(&streak)

	return &user, nil
}

// RegisterUser creates a new user via self-registration (requires beta invite)
func (s *UserService) RegisterUser(email, password, name string, authService *AuthService) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if email exists
	var existing models.User
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailExists
	}

	// Validate password strength
	if err := authService.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := authService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := models.User{
		ID:                  uuid.New().String(),
		Email:               email,
		PasswordHash:        passwordHash,
		Name:                strings.TrimSpace(name),
		Role:                models.RoleUser, // Self-registered users are always regular users
		PasswordChangedAt:   now,
		ForcePasswordChange: false, // User set their own password
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Create initial streak record
	streak := models.Streak{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		CurrentStreak: 0,
		LongestStreak: 0,
		UpdatedAt:     now,
	}
	s.db.Create(&streak)

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user's profile
func (s *UserService) UpdateUser(userID string, input UpdateUserInput) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if input.Name != nil {
		user.Name = strings.TrimSpace(*input.Name)
	}
	if input.Bio != nil {
		bio := strings.TrimSpace(*input.Bio)
		if len(bio) > 500 {
			bio = bio[:500]
		}
		user.Bio = &bio
	}
	if input.ProfilePicture != nil {
		user.ProfilePicture = input.ProfilePicture
	}

	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser soft-deletes a user (admin only)
func (s *UserService) DeleteUser(userID string) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return ErrUserNotFound
	}

	// Don't allow deleting the last admin
	if user.Role == models.RoleAdmin {
		var adminCount int64
		s.db.Model(&models.User{}).Where("role = ? AND deleted_at IS NULL", models.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("cannot delete the last admin")
		}
	}

	return s.db.Delete(&user).Error
}

// ListUsers returns all users (admin only)
func (s *UserService) ListUsers(includeDeleted bool) ([]models.User, error) {
	var users []models.User
	query := s.db.Model(&models.User{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetAllUserIDs returns all active user IDs (for notifications)
func (s *UserService) GetAllUserIDs() ([]string, error) {
	var ids []string
	if err := s.db.Model(&models.User{}).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// GetAllUserIDsExcept returns all active user IDs except the specified one
func (s *UserService) GetAllUserIDsExcept(excludeID string) ([]string, error) {
	var ids []string
	if err := s.db.Model(&models.User{}).Where("id != ?", excludeID).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
