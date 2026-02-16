package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var Location *time.Location

func init() {
	// Load .env file if it exists (ignore error if not found)
	godotenv.Load()

	// Set the application timezone
	tz := os.Getenv("TIMEZONE")
	if tz == "" {
		tz = "UTC"
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("Warning: could not load timezone %s, using UTC: %v", tz, err)
		loc = time.UTC
	}
	Location = loc
	log.Printf("Application timezone set to: %s", Location.String())
}

type Config struct {
	// Server
	Port           string
	GinMode        string
	TrustedProxies []string
	AllowedOrigins []string

	// Database
	DatabasePath string

	// JWT
	JWTSecret        string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	// Storage
	StoragePath    string
	StorageBackend string
	MaxImageSizeMB int

	// Security
	BcryptCost           int
	LoginMaxAttempts     int
	LockoutDuration      time.Duration
	RateLimitAPI         int
	RateLimitLogin       int
	RateLimitUpload      int
	PasswordResetTTL     time.Duration
	SessionInactiveTTL   time.Duration
	AuditLogRetention    int
	SoftDeleteRetention  int

	// APNs (Push Notifications)
	APNsKeyPath    string
	APNsKeyID      string
	APNsTeamID     string
	APNsBundleID   string
	APNsProduction bool

	// General
	Timezone string
}

func Load() (*Config, error) {
	cfg := &Config{
		// Server defaults
		Port:           getEnv("PORT", "8080"),
		GinMode:        getEnv("GIN_MODE", "release"),
		AllowedOrigins: getStringSliceEnv("ALLOWED_ORIGINS"),

		// Database defaults
		DatabasePath: getEnv("DATABASE_PATH", "./feats.db"),

		// JWT defaults
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTAccessTTL:  getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getDurationEnv("JWT_REFRESH_TTL", 168*time.Hour),

		// Storage defaults
		StoragePath:    getEnv("STORAGE_PATH", "./storage"),
		StorageBackend: getEnv("STORAGE_BACKEND", "local"),
		MaxImageSizeMB: getIntEnv("MAX_IMAGE_SIZE_MB", 10),

		// Security defaults
		BcryptCost:          getIntEnv("BCRYPT_COST", 12),
		LoginMaxAttempts:    getIntEnv("LOGIN_MAX_ATTEMPTS", 5),
		LockoutDuration:     getDurationEnv("LOCKOUT_DURATION", 15*time.Minute),
		RateLimitAPI:        getIntEnv("RATE_LIMIT_API", 100),
		RateLimitLogin:      getIntEnv("RATE_LIMIT_LOGIN", 5),
		RateLimitUpload:     getIntEnv("RATE_LIMIT_UPLOAD", 20),
		PasswordResetTTL:    getDurationEnv("PASSWORD_RESET_TTL", 1*time.Hour),
		SessionInactiveTTL:  getDurationEnv("SESSION_INACTIVE_TTL", 720*time.Hour),
		AuditLogRetention:   getIntEnv("AUDIT_LOG_RETENTION", 90),
		SoftDeleteRetention: getIntEnv("SOFT_DELETE_RETENTION", 30),

		// APNs
		APNsKeyPath:    getEnv("APNS_KEY_PATH", ""),
		APNsKeyID:      getEnv("APNS_KEY_ID", ""),
		APNsTeamID:     getEnv("APNS_TEAM_ID", ""),
		APNsBundleID:   getEnv("APNS_BUNDLE_ID", ""),
		APNsProduction: getEnv("APNS_PRODUCTION", "false") == "true",

		// General
		Timezone: getEnv("TIMEZONE", "UTC"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	var result []string
	for _, s := range strings.Split(value, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
