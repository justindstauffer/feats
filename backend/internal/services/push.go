package services

import (
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"gorm.io/gorm"
)

type PushService struct {
	db       *gorm.DB
	client   *apns2.Client
	bundleID string
	enabled  bool
}

func NewPushService(db *gorm.DB, cfg *config.Config) *PushService {
	service := &PushService{
		db:       db,
		bundleID: cfg.APNsBundleID,
		enabled:  false,
	}

	// Only initialize if APNs is configured
	if strings.TrimSpace(cfg.APNsKeyPath) == "" ||
		strings.TrimSpace(cfg.APNsKeyID) == "" ||
		strings.TrimSpace(cfg.APNsTeamID) == "" ||
		strings.TrimSpace(cfg.APNsBundleID) == "" {
		log.Println("Push notifications disabled: APNs not fully configured (key path/id, team id, bundle id required)")
		return service
	}

	authKey, err := token.AuthKeyFromFile(cfg.APNsKeyPath)
	if err != nil {
		log.Printf("Push notifications disabled: failed to load APNs key: %v", err)
		return service
	}

	apnsToken := &token.Token{
		AuthKey: authKey,
		KeyID:   cfg.APNsKeyID,
		TeamID:  cfg.APNsTeamID,
	}

	// Use production or development based on config
	if cfg.APNsProduction {
		service.client = apns2.NewTokenClient(apnsToken).Production()
	} else {
		service.client = apns2.NewTokenClient(apnsToken).Development()
	}

	service.enabled = true
	log.Println("Push notifications enabled")
	return service
}

// RegisterToken saves or updates a device token for a user
func (s *PushService) RegisterToken(userID, tokenStr, platform string) error {
	// Check if token already exists
	var existing models.DeviceToken
	err := s.db.Where("token = ?", tokenStr).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new token
		deviceToken := models.DeviceToken{
			ID:        uuid.New().String(),
			UserID:    userID,
			Token:     tokenStr,
			Platform:  platform,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return s.db.Create(&deviceToken).Error
	} else if err != nil {
		return err
	}

	// Update existing token if user changed
	if existing.UserID != userID {
		existing.UserID = userID
		existing.UpdatedAt = time.Now()
		return s.db.Save(&existing).Error
	}

	return nil
}

// UnregisterToken removes a device token for a specific user.
// This prevents one user from unregistering another user's device token.
func (s *PushService) UnregisterToken(userID, tokenStr string) error {
	return s.db.Where("user_id = ? AND token = ?", userID, tokenStr).Delete(&models.DeviceToken{}).Error
}

// UnregisterUserTokens removes all device tokens for a user
func (s *PushService) UnregisterUserTokens(userID string) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.DeviceToken{}).Error
}

// SendToUser sends a push notification to all devices for a user
func (s *PushService) SendToUser(userID, title, body string, data map[string]interface{}) error {
	if !s.enabled {
		return nil
	}

	var tokens []models.DeviceToken
	if err := s.db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return err
	}

	for _, t := range tokens {
		if t.Platform == "ios" {
			go s.sendIOS(t.Token, title, body, data)
		}
		// Add Android support here later if needed
	}

	return nil
}

// SendToUsers sends a push notification to multiple users
func (s *PushService) SendToUsers(userIDs []string, excludeUserID string, title, body string, data map[string]interface{}) error {
	if !s.enabled {
		return nil
	}

	var tokens []models.DeviceToken
	query := s.db.Where("user_id IN ?", userIDs)
	if excludeUserID != "" {
		query = query.Where("user_id != ?", excludeUserID)
	}
	if err := query.Find(&tokens).Error; err != nil {
		return err
	}

	for _, t := range tokens {
		if t.Platform == "ios" {
			go s.sendIOS(t.Token, title, body, data)
		}
	}

	return nil
}

func (s *PushService) sendIOS(deviceToken, title, body string, data map[string]interface{}) {
	p := payload.NewPayload().
		AlertTitle(title).
		AlertBody(body).
		Sound("default").
		Badge(1)

	// Add custom data
	for key, value := range data {
		p.Custom(key, value)
	}

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       s.bundleID,
		Payload:     p,
	}

	res, err := s.client.Push(notification)
	if err != nil {
		log.Printf("Push notification error: %v", err)
		return
	}

		if !res.Sent() {
			log.Printf("Push notification failed: %d - %s", res.StatusCode, res.Reason)
			// Remove invalid tokens
			if res.Reason == apns2.ReasonBadDeviceToken ||
				res.Reason == apns2.ReasonUnregistered ||
				res.Reason == apns2.ReasonExpiredToken {
				_ = s.db.Where("token = ?", deviceToken).Delete(&models.DeviceToken{}).Error
			}
		}
	}

// Notification helpers for common events

func (s *PushService) NotifyNewComment(postOwnerID, commenterName, postPreview string, postID string) {
	title := "New Comment"
	body := commenterName + " commented on your post"
	if postPreview != "" {
		body = commenterName + " commented: \"" + truncate(postPreview, 50) + "\""
	}
	data := map[string]interface{}{
		"type":    "comment",
		"post_id": postID,
	}
	s.SendToUser(postOwnerID, title, body, data)
}

func (s *PushService) NotifyNewReaction(postOwnerID, reactorName, emoji, postID string) {
	title := "New Reaction"
	body := reactorName + " reacted " + emoji + " to your post"
	data := map[string]interface{}{
		"type":    "reaction",
		"post_id": postID,
	}
	s.SendToUser(postOwnerID, title, body, data)
}

func (s *PushService) NotifyNewPost(groupMemberIDs []string, excludeUserID, posterName, activityName string, postID string) {
	title := "New Post"
	body := posterName + " logged a " + activityName
	data := map[string]interface{}{
		"type":    "post",
		"post_id": postID,
	}
	s.SendToUsers(groupMemberIDs, excludeUserID, title, body, data)
}

func (s *PushService) NotifyChallengeJoined(challengeCreatorID, joinerName, challengeTitle, challengeID string) {
	title := "Challenge Joined"
	body := joinerName + " joined your challenge \"" + challengeTitle + "\""
	data := map[string]interface{}{
		"type":         "challenge",
		"challenge_id": challengeID,
	}
	s.SendToUser(challengeCreatorID, title, body, data)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
