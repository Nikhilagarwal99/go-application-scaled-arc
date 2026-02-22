package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/errorType"
)

// Envelope is the standard JSON shape for every API response.
type Envelope struct {
	Success bool        `json:"success"`
	Code    string      `json:"code,omitempty"` // app-specific error code
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error is the single function handlers call for ANY error.
// It unwraps the error, checks if it's an AppError,
// and maps it to the correct HTTP status + response body automatically.
//
// If it's not an AppError (unexpected error) it falls back to 500.
func Error(c *gin.Context, err error) {
	var appErr errorType.AppError

	// Try to unwrap as AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, Envelope{
			Success: false,
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	// Unknown error — never expose internal details to the client
	c.JSON(http.StatusInternalServerError, Envelope{
		Success: false,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "something went wrong",
	})
}

// ValidationError handles request binding and validation failures.
// Gin returns a detailed error message — we wrap it in our standard envelope.
func ValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Code:    "VALIDATION_ERROR",
		Message: err.Error(),
	})
}
