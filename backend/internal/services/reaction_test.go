package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReactionServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Group{}, &models.ActivityType{}, &models.Post{}, &models.Reaction{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestAddReaction_ChangesExistingReactionType(t *testing.T) {
	db := setupReactionServiceTestDB(t)
	service := NewReactionService(db)

	now := time.Now()
	groupID := uuid.New().String()
	userID := uuid.New().String()
	activityID := uuid.New().String()
	postID := uuid.New().String()

	user := models.User{
		ID:                userID,
		Email:             "reaction-change-" + uuid.New().String() + "@example.com",
		PasswordHash:      "hash",
		Name:              "Reactor",
		Role:              models.RoleUser,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	group := models.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	activity := models.ActivityType{
		ID:        activityID,
		Name:      "Run",
		IsSystem:  true,
		CreatedAt: now,
	}
	post := models.Post{
		ID:             postID,
		GroupID:        groupID,
		UserID:         userID,
		ActivityTypeID: activityID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	for _, record := range []any{&user, &group, &activity, &post} {
		if err := db.Create(record).Error; err != nil {
			t.Fatalf("failed to seed test record: %v", err)
		}
	}

	if _, _, err := service.AddReaction(groupID, postID, userID, models.ReactionLike); err != nil {
		t.Fatalf("failed to add initial reaction: %v", err)
	}
	if _, _, err := service.AddReaction(groupID, postID, userID, models.ReactionLove); err != nil {
		t.Fatalf("failed to change reaction: %v", err)
	}

	var reactions []models.Reaction
	if err := db.Where("post_id = ? AND user_id = ?", postID, userID).Find(&reactions).Error; err != nil {
		t.Fatalf("failed to query reactions: %v", err)
	}
	if len(reactions) != 1 {
		t.Fatalf("expected exactly one reaction row, got %d", len(reactions))
	}
	if reactions[0].ReactionType != models.ReactionLove {
		t.Fatalf("expected updated reaction type %d, got %d", models.ReactionLove, reactions[0].ReactionType)
	}
}
