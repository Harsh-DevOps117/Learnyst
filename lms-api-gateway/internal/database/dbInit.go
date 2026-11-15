package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"lms-api-gateway/internal/models"
	"os"
)

func Initdatabase() (*gorm.DB, error) {
	dsn := os.Getenv("DB_URI")
	if dsn == "" {
		return nil, fmt.Errorf("DB_URI is required; start with npm run dev")
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)})
}
func Automigrate(db *gorm.DB) error {
	// Core owns users, roles, courses and tenants, and starts before auth.
	return db.AutoMigrate(&models.AuthSession{}, &models.PasswordReset{})
}
