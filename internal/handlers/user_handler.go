package handlers

import (
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

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
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
