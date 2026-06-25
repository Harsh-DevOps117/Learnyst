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

	StorePendingUser(ctx context.Context, phone string, userDataJSON []byte) error
	GetPendingUser(ctx context.Context, phone string) ([]byte, error)
	DeletePendingUser(ctx context.Context, phone string) error
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

func (r *RedisRepo) StorePendingUser(ctx context.Context, phone string, userDataJSON []byte) error {
	return r.client.Set(ctx, "pending_user:"+phone, userDataJSON, 10*time.Minute).Err()
}

func (r *RedisRepo) GetPendingUser(ctx context.Context, phone string) ([]byte, error) {
	return r.client.Get(ctx, "pending_user:"+phone).Bytes()
}

func (r *RedisRepo) DeletePendingUser(ctx context.Context, phone string) error {
	return r.client.Del(ctx, "pending_user:"+phone).Err()
}

func (r *RedisRepo) StoreVerificationToken(ctx context.Context, token string, email string) error {
	return r.client.Set(ctx, "verify_email:"+token, email, 15*time.Minute).Err()
}

func (r *RedisRepo) GetEmailFromToken(ctx context.Context, token string) (string, error) {
	return r.client.Get(ctx, "verify_email:"+token).Result()
}

func (r *RedisRepo) DeleteVerificationToken(ctx context.Context, token string) error {
	return r.client.Del(ctx, "verify_email:"+token).Err()
}

func (r *RedisRepo) SetOTPCooldown(ctx context.Context, phone string) error {
	return r.client.Set(ctx, "otp_cooldown:"+phone, "1", 60*time.Second).Err()
}

func (r *RedisRepo) IsOTPOnCooldown(ctx context.Context, phone string) bool {
	res, err := r.client.Get(ctx, "otp_cooldown:"+phone).Result()
	if err != nil {
		return false
	}
	return res == "1"
}
