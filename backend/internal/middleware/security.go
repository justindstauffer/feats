package middleware

import (
	"github.com/gin-gonic/gin"
)

type SecurityMiddleware struct{}

func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{}
}

func (m *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// XSS protection (legacy but still useful)
		c.Header("X-XSS-Protection", "1; mode=block")

		// HSTS - enforce HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy - very restrictive for API
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Prevent caching of authenticated responses
		if c.GetHeader("Authorization") != "" {
			c.Header("Cache-Control", "no-store")
			c.Header("Pragma", "no-cache")
		}

		c.Next()
	}
}
