package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/services"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/response"
)

type UserHandler struct {
	userSvc services.UserService
}

func NewUserHandler(userSvc services.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := mustUserID(c)

	profile, err := h.userSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "profile fetched", profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := mustUserID(c)

	// Parse multipart form — 10MB max file size
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		response.ValidationError(c, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Attach file if present — optional
	file, header, err := c.Request.FormFile("image")
	if err == nil {
		// File was uploaded
		defer file.Close()
		req.ImageUrl = header
	}

	profile, err := h.userSvc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "profile updated", profile)
}

func (h *UserHandler) DeleteProfile(c *gin.Context) {
	userID := mustUserID(c)

	if err := h.userSvc.DeleteProfile(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "account deleted", nil)
}
