package services

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

type AuditLogInput struct {
	UserID       *string
	Action       models.AuditAction
	ResourceType *string
	ResourceID   *string
	IPHash       *string
	UserAgent    *string
	Details      map[string]interface{}
	Success      bool
}

// Log creates an audit log entry
func (s *AuditService) Log(input AuditLogInput) error {
	var detailsJSON *string
	if input.Details != nil {
		bytes, err := json.Marshal(input.Details)
		if err == nil {
			str := string(bytes)
			detailsJSON = &str
		}
	}

	now := time.Now()
	log := models.AuditLog{
		ID:           uuid.New().String(),
		Timestamp:    now,
		UserID:       input.UserID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		IPHash:       input.IPHash,
		UserAgent:    input.UserAgent,
		Details:      detailsJSON,
		Success:      input.Success,
		CreatedAt:    now,
	}

	return s.db.Create(&log).Error
}

// LogLogin logs a login attempt
func (s *AuditService) LogLogin(userID *string, ipHash, userAgent string, success bool, reason string) {
	details := map[string]interface{}{
		"reason": reason,
	}
	s.Log(AuditLogInput{
		UserID:    userID,
		Action:    models.AuditActionLogin,
		IPHash:    &ipHash,
		UserAgent: &userAgent,
		Details:   details,
		Success:   success,
	})
}

// LogLoginFailed logs a failed login attempt
func (s *AuditService) LogLoginFailed(email, ipHash, userAgent, reason string) {
	details := map[string]interface{}{
		"email":  maskEmail(email),
		"reason": reason,
	}
	s.Log(AuditLogInput{
		Action:    models.AuditActionLoginFailed,
		IPHash:    &ipHash,
		UserAgent: &userAgent,
		Details:   details,
		Success:   false,
	})
}

// LogLogout logs a logout
func (s *AuditService) LogLogout(userID string) {
	s.Log(AuditLogInput{
		UserID:  &userID,
		Action:  models.AuditActionLogout,
		Success: true,
	})
}

// LogPasswordChange logs a password change
func (s *AuditService) LogPasswordChange(userID string, success bool) {
	s.Log(AuditLogInput{
		UserID:  &userID,
		Action:  models.AuditActionPasswordChange,
		Success: success,
	})
}

// LogPasswordReset logs a password reset
func (s *AuditService) LogPasswordReset(userID string, success bool) {
	s.Log(AuditLogInput{
		UserID:  &userID,
		Action:  models.AuditActionPasswordReset,
		Success: success,
	})
}

// LogUserCreated logs user creation
func (s *AuditService) LogUserCreated(adminID, newUserID, email string) {
	resourceType := "user"
	details := map[string]interface{}{
		"email": maskEmail(email),
	}
	s.Log(AuditLogInput{
		UserID:       &adminID,
		Action:       models.AuditActionUserCreated,
		ResourceType: &resourceType,
		ResourceID:   &newUserID,
		Details:      details,
		Success:      true,
	})
}

// LogUserDeleted logs user deletion
func (s *AuditService) LogUserDeleted(adminID, deletedUserID string) {
	resourceType := "user"
	s.Log(AuditLogInput{
		UserID:       &adminID,
		Action:       models.AuditActionUserDeleted,
		ResourceType: &resourceType,
		ResourceID:   &deletedUserID,
		Success:      true,
	})
}

// LogAccountLocked logs account lockout
func (s *AuditService) LogAccountLocked(userID, ipHash string, attempts int) {
	details := map[string]interface{}{
		"failed_attempts": attempts,
	}
	s.Log(AuditLogInput{
		UserID:  &userID,
		Action:  models.AuditActionAccountLocked,
		IPHash:  &ipHash,
		Details: details,
		Success: true,
	})
}

// LogAuthorizationFailed logs authorization failure
func (s *AuditService) LogAuthorizationFailed(userID, resource, action string) {
	details := map[string]interface{}{
		"resource": resource,
		"action":   action,
	}
	s.Log(AuditLogInput{
		UserID:  &userID,
		Action:  models.AuditActionAuthorizationFail,
		Details: details,
		Success: false,
	})
}

// LogRateLimitExceeded logs rate limit hit
func (s *AuditService) LogRateLimitExceeded(userID *string, ipHash, endpoint string) {
	details := map[string]interface{}{
		"endpoint": endpoint,
	}
	s.Log(AuditLogInput{
		UserID:  userID,
		Action:  models.AuditActionRateLimitExceeded,
		IPHash:  &ipHash,
		Details: details,
		Success: false,
	})
}

// GetLogs retrieves audit logs with pagination
func (s *AuditService) GetLogs(page, perPage int, userID *string, action *models.AuditAction) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.db.Model(&models.AuditLog{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != nil {
		query = query.Where("action = ?", *action)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("timestamp DESC").Offset(offset).Limit(perPage).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// CleanupOldLogs removes logs older than retention period
func (s *AuditService) CleanupOldLogs(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.db.Where("timestamp < ?", cutoff).Delete(&models.AuditLog{}).Error
}

// maskEmail partially masks an email for logging
func maskEmail(email string) string {
	if len(email) < 3 {
		return "***"
	}
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex <= 1 {
		return string(email[0]) + "***" + email[atIndex:]
	}
	return string(email[0]) + "***" + email[atIndex:]
}
