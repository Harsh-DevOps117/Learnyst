package service

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
	redis *redis.Client
}

func NewAuthService(db *gorm.DB, redis *redis.Client) *AuthService {
	return &AuthService{
		db: db,
		redis: redis,
	}
}

func (a *AuthService) Singup(

) {}

func (a *AuthService) Login() {}
