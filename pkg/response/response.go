package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the standard JSON shape for every API response.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
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

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Error:   err,
	})
}

func Unauthorized(c *gin.Context, err string) {
	c.JSON(http.StatusUnauthorized, Envelope{
		Success: false,
		Error:   err,
	})
}

func NotFound(c *gin.Context, err string) {
	c.JSON(http.StatusNotFound, Envelope{
		Success: false,
		Error:   err,
	})
}

func Conflict(c *gin.Context, err string) {
	c.JSON(http.StatusConflict, Envelope{
		Success: false,
		Error:   err,
	})
}

func InternalServerError(c *gin.Context, err string) {
	c.JSON(http.StatusInternalServerError, Envelope{
		Success: false,
		Error:   err,
	})
}
