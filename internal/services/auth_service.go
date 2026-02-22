package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/models"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/repository"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/utils"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/errorType"
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

type UserProfile struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
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
	GetProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	SendVerifyEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, email, otp string) error
}

// ---- Struct + Constructor ---------------------------------------------------

type authService struct {
	repo           repository.UserRepository
	otpService     repository.OTPRepository
	mailService    *utils.MailService
	jwtSecret      string
	jwtExpiryHours int
}

func NewAuthService(
	repo repository.UserRepository,
	otpService repository.OTPRepository,
	mailService *utils.MailService,
	jwtSecret string,
	jwtExpiryHours int,
) AuthService {
	return &authService{
		repo:           repo,
		otpService:     otpService,
		mailService:    mailService,
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

func (s *authService) GetProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errorType.ErrUserNotFound
	}
	return toProfile(user), nil
}

func (s *authService) UpdateProfile(ctx context.Context, id uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errorType.ErrUserNotFound
	}

	user.Name = req.Name
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errorType.ErrFailedToUpdateUser
	}
	return toProfile(user), nil
}

func (s *authService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return errorType.ErrFailedToDeleteUser
	}
	return nil
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

	if err := s.mailService.SendVerifyEmail(email, otp); err != nil {
		return errorType.ErrEmailServiceDown
	}

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
	}
}
