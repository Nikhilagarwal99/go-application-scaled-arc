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

func NewAuthHandler(authSvc services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req services.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := h.authSvc.Signup(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "account created successfully", res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	res, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "login successful", res)
}

func (h *AuthHandler) SendVerifyEmail(c *gin.Context) {

	var req services.SendVerifyEmailOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// cal send Email Service
	err := h.authSvc.SendVerifyEmail(c.Request.Context(), req.Email)

	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Email Send Successfully", nil)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req services.VerifyEmailOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Separate the context between http gin Context and c.Request.Context here for SOC

	err := h.authSvc.VerifyEmail(c.Request.Context(), req.Email, req.OTP)

	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Email Verified Successfully", nil)
}

// mustUserID reads the authenticated user ID set by the JWT middleware.
func mustUserID(c *gin.Context) uuid.UUID {
	return c.MustGet("userID").(uuid.UUID)
}
