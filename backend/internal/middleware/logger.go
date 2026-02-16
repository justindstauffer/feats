package middleware

import (
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := redactSensitiveQuery(c.Request.URL.Query())

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		// Log format: timestamp | status | latency | ip | method | path
		log.Printf("%d | %13v | %15s | %-7s %s",
			status,
			latency,
			clientIP,
			method,
			path,
		)

		// Log errors separately for easier monitoring
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("ERROR: %s", err.Error())
			}
		}
	}
}

func redactSensitiveQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}

	redacted := make(url.Values, len(values))
	for key, val := range values {
		lowerKey := strings.ToLower(key)
		if lowerKey == "token" || lowerKey == "access_token" || lowerKey == "refresh_token" || lowerKey == "authorization" {
			redacted[key] = []string{"[REDACTED]"}
			continue
		}
		redacted[key] = val
	}

	return redacted.Encode()
}
