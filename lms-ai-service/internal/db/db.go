package db

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
type AIDoubt struct {
	ID        uint      `gorm:"primaryKey"`
	StudentID string    `gorm:"type:uuid;not null"`
	VideoID   uint      `gorm:"not null"`
	Question  string    `gorm:"type:text;not null"`
	Answer    string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
var DB *gorm.DB
func InitDB() {
	dbUri := os.Getenv("DB_URI")
	if dbUri == "" {
		log.Fatal("DB_URI environment variable is required")
	}
	var err error
	DB, err = gorm.Open(postgres.Open(dbUri), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	err = DB.AutoMigrate(&AIDoubt{})
	if err != nil {
		log.Fatalf("Failed to migrate AIDoubt table: %v", err)
	}
	log.Println("Database connected successfully")
}
