package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrBetaInviteNotFound    = errors.New("beta invite not found")
	ErrBetaInviteExpired     = errors.New("beta invite has expired")
	ErrBetaInviteMaxUses     = errors.New("beta invite has reached maximum uses")
	ErrBetaInviteInvalid     = errors.New("invalid beta invite code")
	ErrBetaInviteCodeExists  = errors.New("invite code already exists")
)

type BetaInviteService struct {
	db *gorm.DB
}

func NewBetaInviteService(db *gorm.DB) *BetaInviteService {
	return &BetaInviteService{db: db}
}

// GenerateCode creates a unique invite code in format XXXX-XXXX-XXXX
func (s *BetaInviteService) GenerateCode() (string, error) {
	// Use alphanumeric characters, excluding ambiguous ones (0/O, 1/I/L)
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	const codeLength = 12

	code := make([]byte, codeLength)
	randomBytes := make([]byte, codeLength)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < codeLength; i++ {
		code[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	// Format as XXXX-XXXX-XXXX
	return fmt.Sprintf("%s-%s-%s", string(code[0:4]), string(code[4:8]), string(code[8:12])), nil
}

// Create creates a new beta invite
func (s *BetaInviteService) Create(creatorID string, req models.CreateBetaInviteRequest) (*models.BetaInvite, error) {
	code, err := s.GenerateCode()
	if err != nil {
		return nil, err
	}

	// Default expiration to 7 days
	expiresIn := req.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 168 // 7 days in hours
	}

	// Default max uses to 1
	maxUses := req.MaxUses
	if maxUses < 0 {
		maxUses = 1
	}

	invite := &models.BetaInvite{
		ID:        uuid.New().String(),
		Code:      code,
		CreatedBy: creatorID,
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Hour),
		MaxUses:   maxUses,
		UseCount:  0,
		Note:      strings.TrimSpace(req.Note),
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(invite).Error; err != nil {
		return nil, err
	}

	// Load creator
	if err := s.db.Preload("Creator").First(invite, "id = ?", invite.ID).Error; err != nil {
		return nil, err
	}

	return invite, nil
}

// ValidateAndConsume validates a beta invite code and increments its use count
// Returns the invite if valid, or an error if not
func (s *BetaInviteService) ValidateAndConsume(code string) (*models.BetaInvite, error) {
	normalizedCode := models.NormalizeBetaCode(code)

	var invite models.BetaInvite
	// Find by normalized code (stored codes have dashes, so we need to compare without)
	if err := s.db.Where("REPLACE(UPPER(code), '-', '') = ?", normalizedCode).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBetaInviteInvalid
		}
		return nil, err
	}

	// Check if expired
	if invite.IsExpired() {
		return nil, ErrBetaInviteExpired
	}

	// Check if max uses reached
	if !invite.HasUsesRemaining() {
		return nil, ErrBetaInviteMaxUses
	}

	// Increment use count
	if err := s.db.Model(&invite).Update("use_count", gorm.Expr("use_count + 1")).Error; err != nil {
		return nil, err
	}

	return &invite, nil
}

// GetByID retrieves a beta invite by ID
func (s *BetaInviteService) GetByID(id string) (*models.BetaInvite, error) {
	var invite models.BetaInvite
	if err := s.db.Preload("Creator").First(&invite, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBetaInviteNotFound
		}
		return nil, err
	}
	return &invite, nil
}

// List retrieves all beta invites
func (s *BetaInviteService) List() ([]models.BetaInvite, error) {
	var invites []models.BetaInvite
	if err := s.db.Preload("Creator").Order("created_at DESC").Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

// Delete removes a beta invite
func (s *BetaInviteService) Delete(id string) error {
	result := s.db.Delete(&models.BetaInvite{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBetaInviteNotFound
	}
	return nil
}
