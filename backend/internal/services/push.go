package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushService struct {
	db          *gorm.DB
	client      *apns2.Client
	fcmClient   *http.Client
	fcmServerKey string
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
		db:           db,
		bundleID:     cfg.APNsBundleID,
		fcmServerKey: strings.TrimSpace(cfg.FCMServerKey),
		enabled:      false,
		apnsEnabled:  false,
		fcmEnabled:   false,
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

	// Initialize FCM when server key is provided
	if service.fcmServerKey != "" {
		service.fcmClient = &http.Client{Timeout: 10 * time.Second}
		service.fcmEnabled = true
		log.Println("FCM notifications enabled")
	} else {
		log.Println("FCM notifications disabled: FCM_SERVER_KEY not configured")
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
	if !s.fcmEnabled || s.fcmClient == nil || s.fcmServerKey == "" {
		return
	}

	payload := map[string]interface{}{
		"to": deviceToken,
		"notification": map[string]interface{}{
			"title": title,
			"body":  body,
			"sound": "default",
		},
		"data": data,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("FCM payload marshal error: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("FCM request build error: %v", err)
		return
	}
	req.Header.Set("Authorization", "key="+s.fcmServerKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.fcmClient.Do(req)
	if err != nil {
		log.Printf("FCM push error: %v", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Success int `json:"success"`
		Failure int `json:"failure"`
		Results []struct {
			Error string `json:"error"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("FCM decode error: %v", err)
		return
	}

	if resp.StatusCode >= 400 || result.Failure > 0 {
		reason := ""
		if len(result.Results) > 0 {
			reason = result.Results[0].Error
		}
		log.Printf("FCM push failed: status=%d reason=%s", resp.StatusCode, reason)
		if reason == "NotRegistered" || reason == "InvalidRegistration" {
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
