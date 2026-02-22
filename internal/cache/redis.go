package cache

import (
	"context"
	"log"
	"time"

	"github.com/nikhilAgarwal99/goapp/internal/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(cfg *config.Config) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,

		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established")
	return &Client{rdb: rdb}
}

func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}
