package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/utils"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/errorType"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/response"
)

// Auth validates the Bearer JWT on protected routes and injects
// the authenticated user's ID and email into the Gin context.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errorType.ErrInvalidToken)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, errorType.ErrTokenFormat)
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			response.Error(c, errorType.ErrInvalidToken)
			c.Abort()
			return
		}

		// Make user info available to downstream handlers
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}
