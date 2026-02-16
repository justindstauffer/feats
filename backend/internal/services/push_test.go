package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPushTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.DeviceToken{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	user := models.User{
		ID:                uuid.New().String(),
		Email:             "push-test-" + uuid.New().String() + "@example.com",
		PasswordHash:      "hash",
		Name:              "Push Test",
		Role:              models.RoleUser,
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return db
}

func TestRegisterTokenCreatesNewTokenWhenNotFound(t *testing.T) {
	db := setupPushTestDB(t)
	service := &PushService{db: db}

	token := "test-device-token-123"

	var user models.User
	if err := db.First(&user).Error; err != nil {
		t.Fatalf("failed to get test user: %v", err)
	}

	if err := service.RegisterToken(user.ID, token, "ios"); err != nil {
		t.Fatalf("expected RegisterToken to create new token, got error: %v", err)
	}

	var saved models.DeviceToken
	if err := db.Where("token = ?", token).First(&saved).Error; err != nil {
		t.Fatalf("expected token to be persisted, got error: %v", err)
	}

	if saved.UserID != user.ID {
		t.Fatalf("expected token user_id %s, got %s", user.ID, saved.UserID)
	}
}

func TestRegisterTokenReassignsExistingTokenToNewUser(t *testing.T) {
	db := setupPushTestDB(t)
	service := &PushService{db: db}

	user1 := models.User{
		ID:                uuid.New().String(),
		Email:             "push-test-1@example.com",
		PasswordHash:      "hash",
		Name:              "Push Test 1",
		Role:              models.RoleUser,
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	user2 := models.User{
		ID:                uuid.New().String(),
		Email:             "push-test-2@example.com",
		PasswordHash:      "hash",
		Name:              "Push Test 2",
		Role:              models.RoleUser,
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("failed to create second user: %v", err)
	}

	token := "test-device-token-reassign"

	if err := service.RegisterToken(user1.ID, token, "ios"); err != nil {
		t.Fatalf("failed to register token for first user: %v", err)
	}
	if err := service.RegisterToken(user2.ID, token, "ios"); err != nil {
		t.Fatalf("failed to re-register token for second user: %v", err)
	}

	var tokens []models.DeviceToken
	if err := db.Where("token = ?", token).Find(&tokens).Error; err != nil {
		t.Fatalf("failed to query device tokens: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected exactly one token row, got %d", len(tokens))
	}
	if tokens[0].UserID != user2.ID {
		t.Fatalf("expected token user_id %s, got %s", user2.ID, tokens[0].UserID)
	}
}
