package config

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenRepository interface {
	StoreOTP(ctx context.Context, target string, code string) error
	GetOTP(ctx context.Context, target string) (string, error)
	DeleteOTP(ctx context.Context, target string) error
}

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) StoreOTP(ctx context.Context, target string, code string) error {
	return r.client.Set(ctx, "otp:"+target, code, 2*time.Minute).Err()
}

func (r *RedisRepo) GetOTP(ctx context.Context, target string) (string, error) {
	return r.client.Get(ctx, "otp:"+target).Result()
}

func (r *RedisRepo) DeleteOTP(ctx context.Context, target string) error {
	return r.client.Del(ctx, "otp:"+target).Err()
}
