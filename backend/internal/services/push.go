package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"google.golang.org/api/option"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushService struct {
	db          *gorm.DB
	client      *apns2.Client
	fcmClient   *messaging.Client
	bundleID    string
	enabled     bool
	apnsEnabled bool
	fcmEnabled  bool
}

type tokenColumnInfo struct {
	conflictColumn string
	selectExpr     string
	hasToken       bool
	hasLegacyToken bool
}

func NewPushService(db *gorm.DB, cfg *config.Config) *PushService {
	service := &PushService{
		db:          db,
		bundleID:    cfg.APNsBundleID,
		enabled:     false,
		apnsEnabled: false,
		fcmEnabled:  false,
	}

	// Initialize APNs when fully configured
	if strings.TrimSpace(cfg.APNsKeyPath) != "" &&
		strings.TrimSpace(cfg.APNsKeyID) != "" &&
		strings.TrimSpace(cfg.APNsTeamID) != "" &&
		strings.TrimSpace(cfg.APNsBundleID) != "" {
		authKey, err := token.AuthKeyFromFile(cfg.APNsKeyPath)
		if err != nil {
			log.Printf("APNs disabled: failed to load APNs key: %v", err)
		} else {
			apnsToken := &token.Token{
				AuthKey: authKey,
				KeyID:   cfg.APNsKeyID,
				TeamID:  cfg.APNsTeamID,
			}

			if cfg.APNsProduction {
				service.client = apns2.NewTokenClient(apnsToken).Production()
			} else {
				service.client = apns2.NewTokenClient(apnsToken).Development()
			}
			service.apnsEnabled = true
			log.Println("APNs notifications enabled")
		}
	} else {
		log.Println("APNs notifications disabled: APNs not fully configured")
	}

	// Initialize FCM (HTTP v1) from a service-account credentials file. The
	// project ID is read from the file, so no separate config is required.
	if strings.TrimSpace(cfg.FCMCredentialsPath) != "" {
		app, err := firebase.NewApp(
			context.Background(),
			nil,
			option.WithCredentialsFile(cfg.FCMCredentialsPath),
		)
		if err != nil {
			log.Printf("FCM disabled: failed to init Firebase app: %v", err)
		} else if fcmClient, err := app.Messaging(context.Background()); err != nil {
			log.Printf("FCM disabled: failed to init messaging client: %v", err)
		} else {
			service.fcmClient = fcmClient
			service.fcmEnabled = true
			log.Println("FCM notifications enabled")
		}
	} else {
		log.Println("FCM notifications disabled: FCM_CREDENTIALS_PATH not configured")
	}

	service.enabled = service.apnsEnabled || service.fcmEnabled
	if !service.enabled {
		log.Println("Push notifications disabled: no providers configured")
	}
	return service
}

// RegisterToken saves or updates a device token for a user
func (s *PushService) RegisterToken(userID, tokenStr, platform string) error {
	colInfo, err := s.resolveTokenColumnInfo()
	if err != nil {
		return err
	}

	now := time.Now()
	values := map[string]interface{}{
		"id":         uuid.New().String(),
		"user_id":    userID,
		"platform":   platform,
		"created_at": now,
		"updated_at": now,
	}
	if colInfo.hasToken {
		values["token"] = tokenStr
	}
	if colInfo.hasLegacyToken {
		values["device_token"] = tokenStr
	}

	updateValues := map[string]interface{}{
		"user_id":    userID,
		"platform":   platform,
		"updated_at": now,
	}
	if colInfo.hasToken {
		updateValues["token"] = tokenStr
	}
	if colInfo.hasLegacyToken {
		updateValues["device_token"] = tokenStr
	}

	// Atomic upsert avoids race conditions under repeated registration.
	return s.db.Table("device_tokens").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: colInfo.conflictColumn}},
		DoUpdates: clause.Assignments(updateValues),
	}).Create(values).Error
}

// UnregisterToken removes a device token for a specific user.
// This prevents one user from unregistering another user's device token.
func (s *PushService) UnregisterToken(userID, tokenStr string) error {
	colInfo, err := s.resolveTokenColumnInfo()
	if err != nil {
		return err
	}
	return s.db.Table("device_tokens").
		Where("user_id = ? AND "+colInfo.conflictColumn+" = ?", userID, tokenStr).
		Delete(nil).Error
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

	colInfo, err := s.resolveTokenColumnInfo()
	if err != nil {
		return err
	}

	var tokens []models.DeviceToken
	if err := s.db.Table("device_tokens").
		Select("id, user_id, "+colInfo.selectExpr+", platform, created_at, updated_at").
		Where("user_id = ?", userID).
		Find(&tokens).Error; err != nil {
		return err
	}

	for _, t := range tokens {
		if t.Platform == "ios" {
			go s.sendIOS(t.Token, title, body, data)
		}
		if t.Platform == "android" {
			go s.sendAndroid(t.Token, title, body, data)
		}
	}

	return nil
}

// SendToUsers sends a push notification to multiple users
func (s *PushService) SendToUsers(userIDs []string, excludeUserID string, title, body string, data map[string]interface{}) error {
	if !s.enabled {
		return nil
	}

	colInfo, err := s.resolveTokenColumnInfo()
	if err != nil {
		return err
	}

	var tokens []models.DeviceToken
	query := s.db.Table("device_tokens").
		Select("id, user_id, "+colInfo.selectExpr+", platform, created_at, updated_at").
		Where("user_id IN ?", userIDs)
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
		if t.Platform == "android" {
			go s.sendAndroid(t.Token, title, body, data)
		}
	}

	return nil
}

func (s *PushService) sendIOS(deviceToken, title, body string, data map[string]interface{}) {
	if !s.apnsEnabled || s.client == nil {
		return
	}

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
			if colInfo, err := s.resolveTokenColumnInfo(); err == nil {
				_ = s.db.Table("device_tokens").Where(colInfo.conflictColumn+" = ?", deviceToken).Delete(nil).Error
			}
		}
	}
}

func (s *PushService) sendAndroid(deviceToken, title, body string, data map[string]interface{}) {
	if !s.fcmEnabled || s.fcmClient == nil {
		return
	}

	// FCM HTTP v1 requires the data payload to be map[string]string. All current
	// callers pass strings, but convert defensively so a non-string value can't
	// silently drop the notification.
	stringData := make(map[string]string, len(data))
	for k, v := range data {
		if sv, ok := v.(string); ok {
			stringData[k] = sv
		} else {
			stringData[k] = fmt.Sprintf("%v", v)
		}
	}

	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: stringData,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
	}

	if _, err := s.fcmClient.Send(context.Background(), message); err != nil {
		log.Printf("FCM push failed: %v", err)
		// Drop tokens the server reports as dead, mirroring the APNs path.
		if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
			if colInfo, err := s.resolveTokenColumnInfo(); err == nil {
				_ = s.db.Table("device_tokens").Where(colInfo.conflictColumn+" = ?", deviceToken).Delete(nil).Error
			}
		}
	}
}

func (s *PushService) resolveTokenColumnInfo() (tokenColumnInfo, error) {
	type pragmaColumn struct {
		Name string `gorm:"column:name"`
	}

	var cols []pragmaColumn
	if err := s.db.Raw("PRAGMA table_info(device_tokens)").Scan(&cols).Error; err != nil {
		return tokenColumnInfo{}, err
	}

	hasToken := false
	hasLegacyToken := false
	for _, c := range cols {
		if c.Name == "token" {
			hasToken = true
		}
		if c.Name == "device_token" {
			hasLegacyToken = true
		}
	}

	if hasLegacyToken && hasToken {
		return tokenColumnInfo{
			conflictColumn: "device_token",
			selectExpr:     "COALESCE(device_token, token) AS token",
			hasToken:       true,
			hasLegacyToken: true,
		}, nil
	}
	if hasLegacyToken {
		return tokenColumnInfo{
			conflictColumn: "device_token",
			selectExpr:     "device_token AS token",
			hasToken:       false,
			hasLegacyToken: true,
		}, nil
	}
	if hasToken {
		return tokenColumnInfo{
			conflictColumn: "token",
			selectExpr:     "token",
			hasToken:       true,
			hasLegacyToken: false,
		}, nil
	}

	return tokenColumnInfo{}, fmt.Errorf("device_tokens table missing token columns (expected token or device_token)")
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
