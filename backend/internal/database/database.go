package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jstauff/feats-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dbPath string, debug bool) (*gorm.DB, error) {
	logLevel := logger.Silent
	if debug {
		logLevel = logger.Info
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(1) // SQLite only supports one writer
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.PasswordHistory{},
		&models.PasswordResetToken{},
		&models.ActivityType{},
		&models.Post{},
		&models.PostImage{},
		&models.Reaction{},
		&models.Comment{},
		&models.Streak{},
		&models.Challenge{},
		&models.ChallengeParticipant{},
		&models.Goal{},
		&models.DeviceToken{},
		&models.AuditLog{},
		&models.RateLimit{},
	)
}

func Seed(db *gorm.DB) error {
	// Seed core activity types
	for _, activity := range models.CoreActivityTypes() {
		var existing models.ActivityType
		result := db.Where("name = ?", activity.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			activity.ID = uuid.New().String()
			activity.CreatedAt = time.Now()
			if err := db.Create(&activity).Error; err != nil {
				return fmt.Errorf("failed to seed activity type %s: %w", activity.Name, err)
			}
		}
	}

	return nil
}

// CreateIndexes creates additional indexes not handled by AutoMigrate
func CreateIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_posts_deleted_at ON posts(deleted_at)",
		"CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_user_post ON reactions(user_id, post_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_challenge_participants_unique ON challenge_participants(challenge_id, user_id)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
