package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrPasswordTooWeak    = errors.New("password does not meet requirements")
	ErrPasswordReused     = errors.New("password was recently used")
)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

type TokenClaims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	JTI    string `json:"jti"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(email, password, ipHash, userAgent string) (*TokenPair, *models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment failed attempts
		user.FailedLoginAttempts++

		// Lock account if threshold reached
		if user.FailedLoginAttempts >= s.cfg.LoginMaxAttempts {
			lockUntil := time.Now().Add(s.cfg.LockoutDuration)
			user.LockedUntil = &lockUntil
		}

		s.db.Save(&user)
		return nil, nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	now := time.Now()
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	user.LastLoginAt = &now
	user.LastLoginIPHash = &ipHash
	s.db.Save(&user)

	// Generate tokens
	tokens, err := s.generateTokenPair(&user)
	if err != nil {
		return nil, nil, err
	}

	return tokens, &user, nil
}

// RefreshToken validates a refresh token and issues new tokens
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	var storedToken models.RefreshToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&storedToken).Error; err != nil {
		return nil, ErrTokenInvalid
	}

	if storedToken.IsExpired() {
		s.db.Delete(&storedToken)
		return nil, ErrTokenExpired
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, "id = ?", storedToken.UserID).Error; err != nil {
		return nil, ErrTokenInvalid
	}

	// Delete old refresh token (rotation)
	s.db.Delete(&storedToken)

	// Generate new token pair
	return s.generateTokenPair(&user)
}

// Logout invalidates the refresh token
func (s *AuthService) Logout(userID string) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

// ValidateAccessToken validates an access token and returns claims
func (s *AuthService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := s.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Check password history
	if err := s.checkPasswordHistory(userID, newPassword); err != nil {
		return err
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
	if err != nil {
		return err
	}

	// Save to password history
	s.saveToPasswordHistory(userID, user.PasswordHash)

	// Update user
	now := time.Now()
	user.PasswordHash = string(hash)
	user.PasswordChangedAt = now
	user.ForcePasswordChange = false

	// Invalidate all refresh tokens
	s.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{})

	return s.db.Save(&user).Error
}

// CreatePasswordResetToken generates a password reset token
func (s *AuthService) CreatePasswordResetToken(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		// Don't reveal if user exists
		return "", nil
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store hashed token
	resetToken := models.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(s.cfg.PasswordResetTTL),
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&resetToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	tokenHash := hashToken(token)

	var resetToken models.PasswordResetToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&resetToken).Error; err != nil {
		return ErrTokenInvalid
	}

	if resetToken.IsExpired() || resetToken.IsUsed() {
		return ErrTokenExpired
	}

	// Validate new password
	if err := s.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, "id = ?", resetToken.UserID).Error; err != nil {
		return err
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
	if err != nil {
		return err
	}

	// Save to password history
	s.saveToPasswordHistory(user.ID, user.PasswordHash)

	// Update user
	now := time.Now()
	user.PasswordHash = string(hash)
	user.PasswordChangedAt = now
	user.ForcePasswordChange = false
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	// Mark token as used
	resetToken.UsedAt = &now
	s.db.Save(&resetToken)

	// Invalidate all refresh tokens
	s.db.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{})

	return nil
}

// ValidatePassword checks if password meets requirements
func (s *AuthService) ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("%w: must be at least 12 characters", ErrPasswordTooWeak)
	}

	if len(password) > 128 {
		return fmt.Errorf("%w: must be at most 128 characters", ErrPasswordTooWeak)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("%w: must contain at least one uppercase letter", ErrPasswordTooWeak)
	}
	if !hasLower {
		return fmt.Errorf("%w: must contain at least one lowercase letter", ErrPasswordTooWeak)
	}
	if !hasDigit {
		return fmt.Errorf("%w: must contain at least one digit", ErrPasswordTooWeak)
	}
	if !hasSpecial {
		return fmt.Errorf("%w: must contain at least one special character", ErrPasswordTooWeak)
	}

	return nil
}

// checkPasswordHistory checks if password was recently used
func (s *AuthService) checkPasswordHistory(userID, newPassword string) error {
	var history []models.PasswordHistory
	s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&history)

	for _, h := range history {
		if err := bcrypt.CompareHashAndPassword([]byte(h.PasswordHash), []byte(newPassword)); err == nil {
			return ErrPasswordReused
		}
	}

	return nil
}

// saveToPasswordHistory saves a password hash to history
func (s *AuthService) saveToPasswordHistory(userID, passwordHash string) {
	history := models.PasswordHistory{
		ID:           uuid.New().String(),
		UserID:       userID,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.db.Create(&history)

	// Keep only last 5 entries
	var oldEntries []models.PasswordHistory
	s.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(5).Find(&oldEntries)
	for _, entry := range oldEntries {
		s.db.Delete(&entry)
	}
}

// generateTokenPair creates access and refresh tokens
func (s *AuthService) generateTokenPair(user *models.User) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.cfg.JWTAccessTTL)

	// Create access token
	accessClaims := TokenClaims{
		UserID: user.ID,
		Role:   string(user.Role),
		JTI:    uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshToken := base64.URLEncoding.EncodeToString(refreshTokenBytes)

	// Store refresh token hash
	storedToken := models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: now.Add(s.cfg.JWTRefreshTTL),
		CreatedAt: now,
	}

	if err := s.db.Create(&storedToken).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

// hashToken creates a SHA-256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// HashIP creates a SHA-256 hash of an IP address
func HashIP(ip string) string {
	return hashToken(ip)
}

// HashPassword hashes a password with bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	if err := s.ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
