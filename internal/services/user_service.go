package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/repository"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/tasks"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/pkg/errorType"
)

// ---- DTOs -------------------------------------------------------------------

type UserProfile struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	ImageUrl      string    `json:"image_url"`
	DateOfBirth   time.Time `json:"date_of_birth"`
	Address       string    `json:"address"`
	PhoneNumber   string    `json:"phone_number"`
}

type UpdateProfileRequest struct {
	Name        string    `json:"name" binding:"required,min=2,max=100"`
	ImageUrl    string    `json:"image_url"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Address     string    `json:"address"`
	PhoneNumber string    `json:"phone_number"`
}

type UserService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error)
	DeleteProfile(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	userRepo   repository.UserRepository
	taskClient *tasks.Client
}

func NewUserService(userRepo repository.UserRepository, taskClient *tasks.Client) *userService {
	return &userService{
		userRepo:   userRepo,
		taskClient: taskClient,
	}
}

// ---- Methods ----------------------------------------------------------------

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errorType.ErrUserNotFound
	}
	return toProfile(user), nil
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errorType.ErrUserNotFound
	}

	//if image binary file is there then upload to cloud storage using go routine and update image url

	user.Name = req.Name
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errorType.ErrFailedToUpdateUser
	}
	return toProfile(user), nil
}

func (s *userService) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return errorType.ErrFailedToDeleteUser
	}
	return nil
}
