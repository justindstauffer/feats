package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/config"
	"github.com/jstauff/feats-api/internal/models"
)

type RateLimiter struct {
	cfg     *config.Config
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
}

type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func NewRateLimiter(cfg *config.Config) *RateLimiter {
	rl := &RateLimiter{
		cfg:     cfg,
		buckets: make(map[string]*tokenBucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) getBucket(key string, maxTokens float64, refillRate float64) *tokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists = rl.buckets[key]; exists {
		return bucket
	}

	bucket = &tokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
	rl.buckets[key] = bucket
	return bucket
}

func (b *tokenBucket) take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	// Try to take a token
	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			bucket.mu.Lock()
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) rateLimitMiddleware(keyFunc func(*gin.Context) string, maxTokens float64, refillRate float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		bucket := rl.getBucket(key, maxTokens, refillRate)

		if !bucket.take() {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse(
				models.ErrCodeRateLimited,
				"Too many requests. Please try again later.",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit applies stricter rate limiting to auth endpoints.
//
// Keyed by the attempted email (falling back to client IP when the request has
// no email, e.g. token refresh) so one person's failed logins only throttle
// that account — not everyone sharing a home/NAT IP. Per-account lockout in the
// auth service remains the primary brute-force defense.
func (rl *RateLimiter) LoginRateLimit() gin.HandlerFunc {
	// 5 attempts per 15 minutes = 5/900 = 0.0056 per second
	maxTokens := float64(rl.cfg.RateLimitLogin)
	refillRate := maxTokens / (15 * 60) // refill over 15 minutes

	return rl.rateLimitMiddleware(loginRateKey, maxTokens, refillRate)
}

// loginRateKey keys auth rate limiting by the submitted email, falling back to
// client IP when no email is present.
func loginRateKey(c *gin.Context) string {
	if email := extractLoginEmail(c); email != "" {
		return "login:email:" + strings.ToLower(strings.TrimSpace(email))
	}
	return "login:ip:" + c.ClientIP()
}

// extractLoginEmail reads the "email" field from a JSON body and restores the
// body so the downstream handler can still bind it. Returns "" on any problem.
func extractLoginEmail(c *gin.Context) string {
	if c.Request == nil || c.Request.Body == nil {
		return ""
	}
	// Auth bodies are tiny; cap the read as a safety valve.
	const maxBody = 1 << 20
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBody))
	if err != nil {
		return ""
	}
	// Restore the consumed body for the handler.
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Email string `json:"email"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return ""
	}
	return payload.Email
}

// APIRateLimit applies general rate limiting to authenticated endpoints
func (rl *RateLimiter) APIRateLimit() gin.HandlerFunc {
	// 100 requests per minute
	maxTokens := float64(rl.cfg.RateLimitAPI)
	refillRate := maxTokens / 60 // refill per minute

	return rl.rateLimitMiddleware(func(c *gin.Context) string {
		userID, _ := GetCurrentUserID(c)
		if userID == "" {
			return "api:" + c.ClientIP()
		}
		return "api:" + userID
	}, maxTokens, refillRate)
}

// UploadRateLimit applies rate limiting to file uploads
func (rl *RateLimiter) UploadRateLimit() gin.HandlerFunc {
	// 20 uploads per hour
	maxTokens := float64(rl.cfg.RateLimitUpload)
	refillRate := maxTokens / 3600 // refill per hour

	return rl.rateLimitMiddleware(func(c *gin.Context) string {
		userID, _ := GetCurrentUserID(c)
		return "upload:" + userID
	}, maxTokens, refillRate)
}

// PostRateLimit applies rate limiting to post creation
func (rl *RateLimiter) PostRateLimit() gin.HandlerFunc {
	// 30 posts per hour
	maxTokens := 30.0
	refillRate := maxTokens / 3600

	return rl.rateLimitMiddleware(func(c *gin.Context) string {
		userID, _ := GetCurrentUserID(c)
		return "post:" + userID
	}, maxTokens, refillRate)
}

// InviteRedeemRateLimit applies rate limiting to invite code redemption
// This helps prevent brute-force attempts on invite codes
func (rl *RateLimiter) InviteRedeemRateLimit() gin.HandlerFunc {
	// 5 attempts per minute per user
	maxTokens := 5.0
	refillRate := maxTokens / 60 // refill per minute

	return rl.rateLimitMiddleware(func(c *gin.Context) string {
		userID, _ := GetCurrentUserID(c)
		if userID == "" {
			return "invite:" + c.ClientIP()
		}
		return "invite:" + userID
	}, maxTokens, refillRate)
}
