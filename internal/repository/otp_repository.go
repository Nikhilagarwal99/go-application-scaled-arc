package repository

import (
	"context"
	"time"

	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/cache"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/utils"
)

const otpTTL = 15 * time.Minute

type OTPRepository interface {
	Save(ctx context.Context, email, otp string) error
	Get(ctx context.Context, email string) (string, error)
	Delete(ctx context.Context, email string) error
}

type otpRepository struct {
	redis *cache.Client
}

func NewOTPRepository(redis *cache.Client) OTPRepository {
	return &otpRepository{redis: redis}
}

func (r *otpRepository) Save(ctx context.Context, email, otp string) error {
	return r.redis.Set(ctx, utils.OTPRedisKey(email), otp, otpTTL)
}
func (r *otpRepository) Get(ctx context.Context, email string) (string, error) {
	return r.redis.Get(ctx, utils.OTPRedisKey(email))
}

func (r *otpRepository) Delete(ctx context.Context, email string) error {
	return r.redis.Delete(ctx, utils.OTPRedisKey(email))
}
