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
	Is_Active      bool           `gorm:"default:true"`
}

func main() {
	dsn := "postgresql://adminFutre123:adminFutreKanye123@localhost:5432/adminFutreMetro123?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	var user User
	if err := db.Where("email = ?", "student@learnyst.com").First(&user).Error; err != nil {
		log.Fatal("User not found: ", err)
	}

	fmt.Printf("User found: %+v\n", user)

	pepper := os.Getenv("APP_SECRET_PEPPER")
	password := "student123"
	pepperedPassword := password + pepper

	err = bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(pepperedPassword))
	if err != nil {
		fmt.Printf("Password mismatch! Error: %v\n", err)
	} else {
		fmt.Println("Password matched successfully!")
	}
}
