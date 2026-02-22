package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/logger"
	"go.uber.org/zap"
)

// RequestLogger replaces gin.Logger() with a structured zap logger.
// Logs every request with method, path, status, latency, and request ID.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Everything below runs AFTER the handler returns
		duration := time.Since(start)
		status := c.Writer.Status()

		// Pick log level based on status code
		//  5xx → error
		//  4xx → warn
		//  rest → info
		fields := []zap.Field{
			zap.String("requestId", getRequestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", duration),
			zap.String("ip", c.ClientIP()),
		}

		// Add query string if present — useful for debugging GET requests
		if c.Request.URL.RawQuery != "" {
			fields = append(fields, zap.String("query", c.Request.URL.RawQuery))
		}

		// Add error message if gin recorded one
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("ginError", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("request completed", fields...)
		case status >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}

// getRequestID reads the request ID from gin context.
// Set by the RequestID middleware which runs before this.
// Falls back to "-" if not present.
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("requestId"); exists {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return "-"
}
