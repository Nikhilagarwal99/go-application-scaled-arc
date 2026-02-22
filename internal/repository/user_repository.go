package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/middleware"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines the contract for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

// getDB is the key helper.
// If a transaction exists in the context → use it.
// Otherwise fall back to the regular db connection.
// This is the ONLY place in the entire codebase that knows about this logic.
func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	if tx := middleware.TxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

// NewUserRepository returns a concrete UserRepository backed by *gorm.DB.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.getDB(ctx).Create(user).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.getDB(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.getDB(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.getDB(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).Delete(&models.User{}, "id = ?", id).Error
}
