package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "requestId"
const RequestIDHeader = "X-Request-ID"

/*
RequestID assigns a unique ID to every incoming request.

If the client sends an "X-Request-ID" header (e.g. from a load balancer
or upstream service) — we reuse it. This preserves the ID across
microservices so you can trace a request end to end across systems.

If no header is present — we generate a fresh UUID.

The ID is:
  - stored in Gin context  (for handlers and middleware to read)
  - sent back in the response header (so clients can correlate requests)
*/
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)

		// If upstream didn't provide one — generate a new UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in Gin context — RequestLogger middleware reads it from here
		c.Set(RequestIDKey, requestID)

		// Send it back in the response so the client/load balancer
		// can correlate their request with your logs
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}
