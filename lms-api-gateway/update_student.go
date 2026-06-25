package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID             uint           `gorm:"primaryKey;autoIncrement"`
	First_Name     string         `gorm:"type:varchar(50);not null"`
	Last_Name      string         `gorm:"type:varchar(50);not null"`
	Email          string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password_Hash  string         `gorm:"type:varchar(255);not null"`
}

func main() {
	dsn := "postgresql://adminFutre123:adminFutreKanye123@localhost:5432/adminFutreMetro123?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	pepper := os.Getenv("APP_SECRET_PEPPER")
	password := "student123"
	pepperedPassword := password + pepper

	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	result := db.Model(&User{}).Where("email = ?", "student@learnyst.com").Update("password_hash", string(hash))
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	fmt.Printf("Updated student successfully. Email: student@learnyst.com, Password: %s, Pepper: '%s'\n", password, pepper)
}
