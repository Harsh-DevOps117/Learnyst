package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FirstName      string    `gorm:"type:varchar(100);not null" json:"first_name"`
	MiddleName     string    `gorm:"type:varchar(100)" json:"middle_name"`
	LastName       string    `gorm:"type:varchar(100);not null" json:"last_name"`
	Email          string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone          string    `gorm:"type:varchar(20)" json:"phone"`
	PasswordHash   string    `gorm:"type:varchar(255);not null" json:"-"`
	EmailVerified  bool      `gorm:"default:false" json:"email_verified"`
	PhoneVerified  bool      `gorm:"default:false" json:"phone_verified"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	Location       string    `gorm:"type:varchar(255)" json:"location"`
	RoleID         int       `gorm:"not null" json:"role_id"`
}

func main() {
	dsn := "postgresql://adminFutre123:adminFutreKanye123@localhost:5432/adminFutreMetro123?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	password := "student123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	user := User{
		FirstName:    "Test",
		LastName:     "Student",
		Email:        "student@learnyst.com",
		Phone:        "+910000000001",
		PasswordHash: string(hashedPassword),
		RoleID:       3, // Role 3 is Student
		IsActive:     true,
	}

	result := db.Create(&user)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	fmt.Printf("Created student successfully. Email: %s, Password: %s\n", user.Email, password)
}
