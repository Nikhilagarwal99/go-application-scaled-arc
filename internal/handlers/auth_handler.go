package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/services"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/response"
)

/*
Simple rule to remember
Use When c.Request.Context()In the handler only,
to extract context from the HTTP request
context.ContextIn service and repository signatures — keeps them Gin-free*
gin.ContextIn the handler only — never pass it deeper */

// AuthHandler groups all auth-related HTTP handlers.
type AuthHandler struct {
	authSvc services.AuthService
}
type DefaultResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewAuthHandler(authSvc services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Signup godoc
//
//	POST /api/v1/auth/signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var req services.SignupRequest
	// request validation
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// Service Call
	res, err := h.authSvc.Signup(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "email already registered" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, "account created successfully", res)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, "login successful", res)
}

// GetProfile godoc
//
//	GET /api/v1/users/me
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := mustUserID(c)

	profile, err := h.authSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, "profile fetched", profile)
}

// UpdateProfile godoc
//
//	PUT /api/v1/users/me
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := mustUserID(c)

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	profile, err := h.authSvc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "profile updated", profile)
}

// DeleteAccount
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID := mustUserID(c)

	if err := h.authSvc.DeleteAccount(c.Request.Context(), userID); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "account deleted", nil)
}

// mustUserID reads the authenticated user ID set by the JWT middleware.
func mustUserID(c *gin.Context) uuid.UUID {
	return c.MustGet("userID").(uuid.UUID)
}

func (h *AuthHandler) SendVerifyEmail(c *gin.Context) {

	var req services.SendVerifyEmailOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// cal send Email Service
	res, err := h.authSvc.SendVerifyEmail(c.Request.Context(), req.Email)

	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Email Send Successfully", res)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req services.VerifyEmailOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Separate the context between http gin Context and c.Request.Context here for SOC

	res, err := h.authSvc.VerifyEmail(c.Request.Context(), req.Email, req.OTP)

	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.OK(c, "Email Verified Successfully", res)
}
