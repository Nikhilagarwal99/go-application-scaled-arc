package services

import (
	"context"
	"errors"

	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/logger"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/models"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/repository"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/tasks"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/utils"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/errorType"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ---- DTOs -------------------------------------------------------------------

type SignupRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

type SendVerifyEmailOtpRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailOtpRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp"   binding:"required"`
}

// ---- Interface --------------------------------------------------------------

type AuthService interface {
	Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	SendVerifyEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, email, otp string) error
}

// ---- Struct + Constructor ---------------------------------------------------

type authService struct {
	repo           repository.UserRepository
	otpService     repository.OTPRepository
	mailService    *utils.MailService
	taskClient     *tasks.Client
	jwtSecret      string
	jwtExpiryHours int
}

func NewAuthService(
	repo repository.UserRepository,
	otpService repository.OTPRepository,
	mailService *utils.MailService,
	taskClient *tasks.Client,
	jwtSecret string,
	jwtExpiryHours int,
) AuthService {
	return &authService{
		repo:           repo,
		otpService:     otpService,
		mailService:    mailService,
		taskClient:     taskClient,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
	}
}

// ---- Methods ----------------------------------------------------------------

func (s *authService) Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error) {
	if _, err := s.repo.FindByEmail(ctx, req.Email); err == nil {
		return nil, errorType.ErrEmailAlreadyRegistered
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errorType.ErrInternalServer
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, errorType.ErrFailedToCreateUser
	}

	// Enqueue welcome email — fire and forget, don't fail signup if this fails
	if err := s.taskClient.EnqueueWelcomeEmail(user.Email, user.Name); err != nil {
		logger.Warn("failed to enqueue welcome email — signup still succeeded", zap.String("email", user.Email), zap.Error(err))
	}

	return s.buildAuthResponse(user)
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorType.ErrInvalidCredentials
		}
		return nil, errorType.ErrInternalServer
	}

	if !user.EmailVerified {
		return nil, errorType.ErrEmailNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errorType.ErrInvalidCredentials
	}

	return s.buildAuthResponse(user)
}

func (s *authService) SendVerifyEmail(ctx context.Context, email string) error {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return errorType.ErrUserNotFound
	}

	if user.EmailVerified {
		return errorType.ErrEmailAlreadyVerified
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return errorType.ErrFailedToGenerateOTP
	}

	if err := s.otpService.Save(ctx, user.Email, otp); err != nil {
		return errorType.ErrFailedToStoreOTP
	}

	// Enqueue email task — returns instantly, worker handles sending
	// If this fails the OTP is already saved — user can retry sending
	if err := s.taskClient.EnqueueVerifyEmail(user.Email, otp); err != nil {
		logger.Error("failed to enqueue verify email task", zap.String("email", user.Email), zap.Error(err))
		return errorType.ErrFailedToEnqueueTask
	}

	logger.Info("verify email task enqueued", zap.String("email", user.Email))
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, email, otp string) error {
	cachedOTP, err := s.otpService.Get(ctx, email)
	if err != nil {
		return errorType.ErrOTPExpired
	}

	if cachedOTP != otp {
		return errorType.ErrInvalidOTP
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return errorType.ErrUserNotFound
	}

	if user.EmailVerified {
		return errorType.ErrEmailAlreadyVerified
	}

	user.EmailVerified = true
	if err := s.repo.Update(ctx, user); err != nil {
		return errorType.ErrFailedToUpdateUser
	}

	return nil
}

// ---- Helpers ----------------------------------------------------------------

func (s *authService) buildAuthResponse(user *models.User) (*AuthResponse, error) {
	token, err := utils.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiryHours)
	if err != nil {
		return nil, errorType.ErrFailedToGenerateToken
	}
	return &AuthResponse{Token: token, User: *toProfile(user)}, nil
}

func toProfile(u *models.User) *UserProfile {
	return &UserProfile{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		ImageUrl:      u.ImageUrl,
		DateOfBirth:   u.DateOfBirth,
		Address:       u.Address,
		PhoneNumber:   u.PhoneNumber,
	}
}
