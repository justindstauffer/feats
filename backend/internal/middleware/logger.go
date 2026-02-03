package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

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
