package db

import (
	"lms-api-gateway/internal/models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func Initdatabase()(*gorm.DB,	error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	DB_URL:=os.Getenv("DB_URL")
	db,err:=gorm.Open(postgres.Open(DB_URL),&gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}


func Automigrate(db *gorm.DB) {
	err:=db.AutoMigrate(
		models.User{}, models.Role{}, models.Permission{}, models.Tenant{}, models.Refresh_Token{}, models.Subscription{}
	)
	if err != nil {
		config.Logger.Fatal(err)
	}
}
