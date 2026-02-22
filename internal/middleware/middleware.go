package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger writes a structured log line for every request.
// Swap out for zap/zerolog in production.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		log.Printf("[%s] %s %s | status=%d | duration=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Writer.Status(),
			time.Since(start).String(),
		)
	}
}

// CORS adds permissive CORS headers. Tighten AllowOrigins for production.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}