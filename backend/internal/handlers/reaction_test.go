package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testReactionPushNotifier struct {
	called      bool
	postOwnerID string
	reactorName string
	emoji       string
	postID      string
}

func (n *testReactionPushNotifier) NotifyNewReaction(postOwnerID, reactorName, emoji, postID string) {
	n.called = true
	n.postOwnerID = postOwnerID
	n.reactorName = reactorName
	n.emoji = emoji
	n.postID = postID
}

func setupReactionHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Group{}, &models.GroupMember{}, &models.ActivityType{}, &models.Post{}, &models.Reaction{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
	return db
}

func TestAddReactionSendsPushToPostOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReactionHandlerTestDB(t)
	reactionService := services.NewReactionService(db)
	push := &testReactionPushNotifier{}
	handler := NewReactionHandler(reactionService, push, nil)

	now := time.Now()
	groupID := uuid.New().String()
	ownerID := uuid.New().String()
	reactorID := uuid.New().String()
	activityID := uuid.New().String()
	postID := uuid.New().String()

	group := models.Group{ID: groupID, Name: "G", CreatedBy: ownerID, CreatedAt: now, UpdatedAt: now}
	owner := models.User{ID: ownerID, Email: "owner@test.com", Name: "Owner", PasswordHash: "x", Role: models.RoleUser, PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now}
	reactor := models.User{ID: reactorID, Email: "reactor@test.com", Name: "Reactor", PasswordHash: "x", Role: models.RoleUser, PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now}
	activity := models.ActivityType{ID: activityID, Name: "Running", IsSystem: true, CreatedAt: now}
	post := models.Post{ID: postID, GroupID: groupID, UserID: ownerID, ActivityTypeID: activityID, CreatedAt: now, UpdatedAt: now}

	if err := db.Create(&group).Error; err != nil { t.Fatalf("create group: %v", err) }
	if err := db.Create(&owner).Error; err != nil { t.Fatalf("create owner: %v", err) }
	if err := db.Create(&reactor).Error; err != nil { t.Fatalf("create reactor: %v", err) }
	if err := db.Create(&activity).Error; err != nil { t.Fatalf("create activity: %v", err) }
	if err := db.Create(&post).Error; err != nil { t.Fatalf("create post: %v", err) }

	body, _ := json.Marshal(map[string]any{"reaction_type": int(models.ReactionLike)})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: postID}}
	c.Set("group_id", groupID)
	c.Set("user_id", reactorID)
	c.Set("user", &reactor)

	handler.AddReaction(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	if !push.called {
		t.Fatal("expected push notifier to be called")
	}
	if push.postOwnerID != ownerID {
		t.Fatalf("expected owner %s, got %s", ownerID, push.postOwnerID)
	}
	if push.postID != postID {
		t.Fatalf("expected post id %s, got %s", postID, push.postID)
	}
	if push.reactorName != reactor.Name {
		t.Fatalf("expected reactor name %s, got %s", reactor.Name, push.reactorName)
	}
	if push.emoji == "" {
		t.Fatal("expected reaction emoji to be populated")
	}
}

func TestAddReactionDoesNotSendPushForSelfReaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupReactionHandlerTestDB(t)
	reactionService := services.NewReactionService(db)
	push := &testReactionPushNotifier{}
	handler := NewReactionHandler(reactionService, push, nil)

	now := time.Now()
	groupID := uuid.New().String()
	userID := uuid.New().String()
	activityID := uuid.New().String()
	postID := uuid.New().String()

	group := models.Group{ID: groupID, Name: "G", CreatedBy: userID, CreatedAt: now, UpdatedAt: now}
	user := models.User{ID: userID, Email: "self@test.com", Name: "Self", PasswordHash: "x", Role: models.RoleUser, PasswordChangedAt: now, CreatedAt: now, UpdatedAt: now}
	activity := models.ActivityType{ID: activityID, Name: "Running", IsSystem: true, CreatedAt: now}
	post := models.Post{ID: postID, GroupID: groupID, UserID: userID, ActivityTypeID: activityID, CreatedAt: now, UpdatedAt: now}

	if err := db.Create(&group).Error; err != nil { t.Fatalf("create group: %v", err) }
	if err := db.Create(&user).Error; err != nil { t.Fatalf("create user: %v", err) }
	if err := db.Create(&activity).Error; err != nil { t.Fatalf("create activity: %v", err) }
	if err := db.Create(&post).Error; err != nil { t.Fatalf("create post: %v", err) }

	body, _ := json.Marshal(map[string]any{"reaction_type": int(models.ReactionLike)})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: postID}}
	c.Set("group_id", groupID)
	c.Set("user_id", userID)
	c.Set("user", &user)

	handler.AddReaction(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	if push.called {
		t.Fatal("expected push notifier to not be called for self-reaction")
	}
}
