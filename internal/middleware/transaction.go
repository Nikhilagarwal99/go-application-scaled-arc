package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// txKey is an unexported type used as the context key for the transaction.
// Using a custom type prevents key collisions with other packages
// that might also store values in the context.
type txKey struct{}

// Transaction middleware begins a DB transaction at the start of the request.
// It stores the tx in both:
//   - Gin context  (so the handler layer can access it if needed)
//   - Request context (so service/repository layers can access it — Gin free)
//
// On the way out:
//   - If any handler/service wrote an error status (>= 400) → ROLLBACK
//   - If the response writer panicked → ROLLBACK  (gin.Recovery() runs first)
//   - Everything went fine → COMMIT
func Transaction(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Begin the transaction
		tx := db.Begin()
		if tx.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "could not start transaction",
			})
			return
		}

		// Store tx in request context so service/repository layers can read it
		// without importing Gin
		ctx := context.WithValue(c.Request.Context(), txKey{}, tx)
		c.Request = c.Request.WithContext(ctx)

		// Also store in Gin context for convenience if handler needs it
		c.Set("tx", tx)

		// Defer runs AFTER the handler returns —
		// this is where we commit or rollback
		defer func() {
			// If a panic happened, gin.Recovery() already caught it above us.
			// We still need to rollback to release the connection.
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r) // re-panic so gin.Recovery() can handle the response
			}

			// If response status is 4xx or 5xx → rollback
			if c.Writer.Status() >= http.StatusBadRequest {
				tx.Rollback()
				return
			}

			// All good → commit
			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "transaction commit failed",
				})
			}
		}()

		// Pass control to the next handler
		c.Next()
	}
}

// TxFromContext extracts the transaction from the context.
// Called inside the repository to get the active tx if one exists.
// Falls back to nil if no transaction is present —
// the repository will then use its regular db connection.
func TxFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(txKey{}).(*gorm.DB)
	return tx
}
