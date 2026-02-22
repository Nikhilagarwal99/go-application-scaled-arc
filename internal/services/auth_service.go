package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikhilAgarwal99/goapp/internal/models"
	"github.com/nikhilAgarwal99/goapp/internal/repository"
	"github.com/nikhilAgarwal99/goapp/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ---- Request / Response DTOs Request Validators------------------------------------------------

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
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type SendVerifyEmailOtpRequest struct {
	Email string `json:"email" binding:"required"`
}

type VerifyEmailOtpRequest struct {
	Email string `json:"email" binding:"required"`
	OTP   string `json:"otp" binding:"required"`
}

// ---- Service interface -------------------------------------------------------

type AuthService interface {
	Signup(req *SignupRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	GetProfile(id uuid.UUID) (*UserProfile, error)
	UpdateProfile(id uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error)
	DeleteAccount(id uuid.UUID) error
	SendVerifyEmail(ctx context.Context, email string) (map[string]any, error)
	VerifyEmail(ctx context.Context, email, otp string) (map[string]any, error)
}

type authService struct {
	repo           repository.UserRepository
	jwtSecret      string
	jwtExpiryHours int
	mailService    *utils.MailService
	otpService     repository.OTPRepository
}

// NewAuthService wires up the auth service with its dependencies.
func NewAuthService(repo repository.UserRepository, jwtSecret string, jwtExpiryHours int, otpService repository.OTPRepository, mailService *utils.MailService) AuthService {
	return &authService{
		repo:           repo,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
		mailService:    mailService,
		otpService:     otpService,
	}
}

// Signup creates a new user account and returns a JWT.
func (s *authService) Signup(req *SignupRequest) (*AuthResponse, error) {
	// Check for existing email
	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return s.buildAuthResponse(user)
}

// Login validates credentials and returns a JWT.
func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, errors.New("something went wrong")
	}

	if user.EmailVerified != true {
		return nil, errors.New("Email Not verified Error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return s.buildAuthResponse(user)
}

// GetProfile fetches a user's public profile by ID.
func (s *authService) GetProfile(id uuid.UUID) (*UserProfile, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return toProfile(user), nil
}

// UpdateProfile changes the user's name.
func (s *authService) UpdateProfile(id uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Name = req.Name
	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("failed to update profile")
	}
	return toProfile(user), nil
}

// DeleteAccount soft-deletes the user record.
func (s *authService) DeleteAccount(id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		return errors.New("failed to delete account")
	}
	return nil
}

func (s *authService) SendVerifyEmail(ctx context.Context, email string) (map[string]any, error) {
	//check if user exist else return
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	var emailDb string = user.Email

	if emailDb != email {
		return nil, errors.New("Email mismatch error type")
	}

	// check if existed user already verified his email if yes return

	if user.EmailVerified {
		return nil, errors.New("Email is already Verified")
	}

	otp, err := utils.GenerateOTP()

	if err != nil {
		return nil, errors.New("Failed to generate Otp")
	}

	// save Otp in redis

	_err := s.otpService.Save(ctx, user.Email, otp)
	if _err != nil {
		return nil, errors.New("Failed to save Otp")
	}
	// call utils function to send verify email on success send success response  else return error

	mailJetError := s.mailService.SendVerifyEmail(email, otp)

	if mailJetError != nil {
		return nil, errors.New("Email Service is temporary down")
	}

	return map[string]any{
		"message": "Email Send Successfully",
	}, nil

}

func (s *authService) VerifyEmail(ctx context.Context, email, otp string) (map[string]any, error) {

	// call utils function to verify email on success send success response  else return error

	fmt.Println("Nikhil00huihi", email)
	cacheOTP, err := s.otpService.Get(ctx, email)
	if err != nil {
		return nil, errors.New("Otp Expired, Please initiate again")
	}
	fmt.Println("Nikhil-------------", cacheOTP, otp)
	if cacheOTP != otp {
		return nil, errors.New("Invalid Otp")
	}

	user, err := s.repo.FindByEmail(email)

	if err != nil {
		return nil, errors.New("User not found")
	}

	if user.EmailVerified {
		return nil, errors.New("Email is already verified error")
	} else {
		user.EmailVerified = true
	}

	// call the model to update email verified

	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("Failed to update, Please try again")
	}

	return map[string]any{"user": user}, nil
}

// ---- Helpers ----------------------------------------------------------------

func (s *authService) buildAuthResponse(user *models.User) (*AuthResponse, error) {
	token, err := utils.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiryHours)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	return &AuthResponse{Token: token, User: *toProfile(user)}, nil
}

func toProfile(u *models.User) *UserProfile {
	return &UserProfile{ID: u.ID, Name: u.Name, Email: u.Email}
}
