package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/database"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type apiTestApp struct {
	router   http.Handler
	db       *gorm.DB
	cfg      *config.Config
	services *appServices
}

func newAPITestApp(t *testing.T) *apiTestApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		GinMode:            "test",
		JWTSecret:          strings.Repeat("x", 32),
		JWTAccessTTL:       15 * time.Minute,
		JWTRefreshTTL:      24 * time.Hour,
		BcryptCost:         4,
		LoginMaxAttempts:   5,
		LockoutDuration:    15 * time.Minute,
		RateLimitAPI:       1000,
		RateLimitLogin:     1000,
		RateLimitUpload:    1000,
		PasswordResetTTL:   time.Hour,
		SessionInactiveTTL: 30 * 24 * time.Hour,
		Timezone:           "UTC",
		StoragePath:        t.TempDir(),
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	if err := database.CreateIndexes(db); err != nil {
		t.Fatalf("failed to create indexes: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("failed to seed test db: %v", err)
	}

	svc := initServices(db, cfg)
	h := initHandlers(svc, cfg, nil)
	m := initMiddleware(svc.auth, svc.group, cfg)
	router := setupRouter(cfg, h, m)

	return &apiTestApp{
		router:   router,
		db:       db,
		cfg:      cfg,
		services: svc,
	}
}

func (a *apiTestApp) request(t *testing.T, method, path string, body any, bearerToken string) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader *bytes.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(payload)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	rr := httptest.NewRecorder()
	a.router.ServeHTTP(rr, req)
	return rr
}

func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder) models.Response {
	t.Helper()

	var resp models.Response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func (a *apiTestApp) createUser(t *testing.T, email, password, name string, role models.UserRole) models.User {
	t.Helper()

	hash, err := a.services.auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	now := time.Now()
	user := models.User{
		ID:                uuid.New().String(),
		Email:             strings.ToLower(email),
		PasswordHash:      hash,
		Name:              name,
		Role:              role,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := a.db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func (a *apiTestApp) loginToken(t *testing.T, email, password string) string {
	t.Helper()

	tokens, _, err := a.services.auth.Login(email, password, "test-ip", "test-agent")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	return tokens.AccessToken
}

func TestRegisterConsumesInviteOnSuccess(t *testing.T) {
	app := newAPITestApp(t)
	creator := app.createUser(t, "admin@example.com", "ValidPass123!", "Admin", models.RoleAdmin)

	invite := models.BetaInvite{
		ID:        uuid.New().String(),
		Code:      "ABCD-EFGH-JKLM",
		CreatedBy: creator.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		MaxUses:   1,
		UseCount:  0,
		CreatedAt: time.Now(),
	}
	if err := app.db.Create(&invite).Error; err != nil {
		t.Fatalf("failed to create beta invite: %v", err)
	}

	rr := app.request(t, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":       "newuser@example.com",
		"password":    "ValidPass123!",
		"name":        "New User",
		"invite_code": invite.Code,
	}, "")

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeResponse(t, rr)
	if !resp.Success {
		t.Fatalf("expected success response, got %+v", resp)
	}

	var updated models.BetaInvite
	if err := app.db.First(&updated, "id = ?", invite.ID).Error; err != nil {
		t.Fatalf("failed to reload invite: %v", err)
	}
	if updated.UseCount != 1 {
		t.Fatalf("expected invite use_count=1, got %d", updated.UseCount)
	}
}

func TestRegisterRollsBackInviteOnUserCreationFailure(t *testing.T) {
	app := newAPITestApp(t)
	creator := app.createUser(t, "admin2@example.com", "ValidPass123!", "Admin2", models.RoleAdmin)
	app.createUser(t, "existing@example.com", "ValidPass123!", "Existing", models.RoleUser)

	invite := models.BetaInvite{
		ID:        uuid.New().String(),
		Code:      "WXYZ-2345-6789",
		CreatedBy: creator.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		MaxUses:   1,
		UseCount:  0,
		CreatedAt: time.Now(),
	}
	if err := app.db.Create(&invite).Error; err != nil {
		t.Fatalf("failed to create beta invite: %v", err)
	}

	rr := app.request(t, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":       "existing@example.com",
		"password":    "ValidPass123!",
		"name":        "Duplicate User",
		"invite_code": invite.Code,
	}, "")

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rr.Code, rr.Body.String())
	}

	var updated models.BetaInvite
	if err := app.db.First(&updated, "id = ?", invite.ID).Error; err != nil {
		t.Fatalf("failed to reload invite: %v", err)
	}
	if updated.UseCount != 0 {
		t.Fatalf("expected invite use_count rollback to 0, got %d", updated.UseCount)
	}
}

func TestRedeemInviteRejectsMalformedCode(t *testing.T) {
	app := newAPITestApp(t)
	user := app.createUser(t, "member@example.com", "ValidPass123!", "Member", models.RoleUser)
	token := app.loginToken(t, user.Email, "ValidPass123!")

	rr := app.request(t, http.MethodPost, "/api/v1/invites/redeem", map[string]any{
		"code": "BAD",
	}, token)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreatePostUpdatesStreakForGroup(t *testing.T) {
	app := newAPITestApp(t)
	user := app.createUser(t, "poster@example.com", "ValidPass123!", "Poster", models.RoleUser)
	token := app.loginToken(t, user.Email, "ValidPass123!")

	groupID := uuid.New().String()
	now := time.Now()
	group := models.Group{
		ID:        groupID,
		Name:      "Test Group",
		CreatedBy: user.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := app.db.Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	member := models.GroupMember{
		ID:       uuid.New().String(),
		GroupID:  groupID,
		UserID:   user.ID,
		Role:     models.GroupRoleAdmin,
		JoinedAt: now,
	}
	if err := app.db.Create(&member).Error; err != nil {
		t.Fatalf("failed to create membership: %v", err)
	}

	var activity models.ActivityType
	if err := app.db.Where("name = ?", "Running").First(&activity).Error; err != nil {
		t.Fatalf("failed to load seeded activity: %v", err)
	}

	rr := app.request(t, http.MethodPost, "/api/v1/groups/"+groupID+"/posts", map[string]any{
		"activity_type_id": activity.ID,
		"description":      "Morning run",
	}, token)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	var streak models.Streak
	if err := app.db.Where("group_id = ? AND user_id = ?", groupID, user.ID).First(&streak).Error; err != nil {
		t.Fatalf("failed to load streak: %v", err)
	}
	if streak.CurrentStreak != 1 {
		t.Fatalf("expected current_streak=1, got %d", streak.CurrentStreak)
	}
}

func TestServeImageRejectsUserOutsideGroup(t *testing.T) {
	app := newAPITestApp(t)
	owner := app.createUser(t, "owner@example.com", "ValidPass123!", "Owner", models.RoleUser)
	outsider := app.createUser(t, "outsider@example.com", "ValidPass123!", "Outsider", models.RoleUser)
	outsiderToken := app.loginToken(t, outsider.Email, "ValidPass123!")

	groupID := uuid.New().String()
	now := time.Now()
	group := models.Group{
		ID:        groupID,
		Name:      "Private Group",
		CreatedBy: owner.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := app.db.Create(&group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	member := models.GroupMember{
		ID:       uuid.New().String(),
		GroupID:  groupID,
		UserID:   owner.ID,
		Role:     models.GroupRoleAdmin,
		JoinedAt: now,
	}
	if err := app.db.Create(&member).Error; err != nil {
		t.Fatalf("failed to create owner membership: %v", err)
	}

	post := models.Post{
		ID:             uuid.New().String(),
		GroupID:        groupID,
		UserID:         owner.ID,
		ActivityTypeID: mustSeededActivityID(t, app.db, "Running"),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := app.db.Create(&post).Error; err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	image := models.PostImage{
		ID:           uuid.New().String(),
		PostID:       post.ID,
		FilePath:     app.cfg.StoragePath + "/images/" + uuid.New().String() + ".jpg",
		DisplayOrder: 0,
		CreatedAt:    now,
	}
	if err := app.db.Create(&image).Error; err != nil {
		t.Fatalf("failed to create post image: %v", err)
	}

	rr := app.request(t, http.MethodGet, "/images/"+image.ID, nil, outsiderToken)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for outsider image access, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func mustSeededActivityID(t *testing.T, db *gorm.DB, name string) string {
	t.Helper()

	var activity models.ActivityType
	if err := db.Where("name = ?", name).First(&activity).Error; err != nil {
		t.Fatalf("failed to load seeded activity %s: %v", name, err)
	}
	return activity.ID
}
