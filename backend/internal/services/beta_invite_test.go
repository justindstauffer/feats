package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBetaInviteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&models.BetaInvite{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestRollbackConsumeDecrementsUseCount(t *testing.T) {
	db := setupBetaInviteTestDB(t)
	service := NewBetaInviteService(db)

	invite := models.BetaInvite{
		ID:        uuid.New().String(),
		Code:      "ABCD-EFGH-JKLM",
		CreatedBy: "admin-1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		MaxUses:   5,
		UseCount:  1,
		CreatedAt: time.Now(),
	}
	if err := db.Create(&invite).Error; err != nil {
		t.Fatalf("failed to create invite: %v", err)
	}

	if err := service.RollbackConsume(invite.ID); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	var updated models.BetaInvite
	if err := db.First(&updated, "id = ?", invite.ID).Error; err != nil {
		t.Fatalf("failed to reload invite: %v", err)
	}

	if updated.UseCount != 0 {
		t.Fatalf("expected use_count 0, got %d", updated.UseCount)
	}
}

func TestRollbackConsumeDoesNotGoNegative(t *testing.T) {
	db := setupBetaInviteTestDB(t)
	service := NewBetaInviteService(db)

	invite := models.BetaInvite{
		ID:        uuid.New().String(),
		Code:      "WXYZ-2345-6789",
		CreatedBy: "admin-1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		MaxUses:   5,
		UseCount:  0,
		CreatedAt: time.Now(),
	}
	if err := db.Create(&invite).Error; err != nil {
		t.Fatalf("failed to create invite: %v", err)
	}

	if err := service.RollbackConsume(invite.ID); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	var updated models.BetaInvite
	if err := db.First(&updated, "id = ?", invite.ID).Error; err != nil {
		t.Fatalf("failed to reload invite: %v", err)
	}

	if updated.UseCount != 0 {
		t.Fatalf("expected use_count to remain 0, got %d", updated.UseCount)
	}
}

