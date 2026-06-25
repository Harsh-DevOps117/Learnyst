package db

import (
	"lms-api-gateway/internal/config"
	"lms-api-gateway/internal/models"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func Initdatabase()(*gorm.DB,	error) {

	DB_URL:=os.Getenv("DB_URI")
	db,err:=gorm.Open(postgres.Open(DB_URL),&gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}


func Automigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		models.User{},
		models.Role{},
		models.Permission{},
		models.Tenant{},
		models.Refresh_Token{},
		models.Subscription{},
	)
	config.Log.Info("Database Migrated Successfully...")
	if err != nil {
		config.Log.Warn(err.Error())
	}

	// Automatically seed default roles
	seedDefaultData(db)
}

func seedDefaultData(db *gorm.DB) {
	roles := []models.Role{
		{ID: 1, Name: "Tenant"},
		{ID: 2, Name: "Teacher"},
		{ID: 3, Name: "Student"},
		{ID: 4, Name: "SuperAdmin"},
	}

	for _, role := range roles {
		// Use FirstOrCreate to only insert if the role doesn't exist
		if err := db.Where(models.Role{ID: role.ID}).FirstOrCreate(&role).Error; err != nil {
			config.Log.Error("Failed to seed role: " + err.Error())
		}
	}
	
	// Seed SuperAdmin User
	var superAdminCount int64
	db.Model(&models.User{}).Where("email = ?", "superadmin@learnyst.com").Count(&superAdminCount)
	if superAdminCount == 0 {
		pepper := os.Getenv("APP_SECRET_PEPPER")
		hash, _ := bcrypt.GenerateFromPassword([]byte("superadmin123"+pepper), bcrypt.DefaultCost)
		superAdmin := models.User{
			First_Name:     "Super",
			Last_Name:      "Admin",
			Email:          "superadmin@learnyst.com",
			Phone:          "0000000000",
			RoleID:         4,
			Password_Hash:  string(hash),
			Phone_Verified: true,
			Email_Verified: true,
			Is_Active:      true,
		}
		if err := db.Create(&superAdmin).Error; err != nil {
			config.Log.Error("Failed to seed SuperAdmin user: " + err.Error())
		} else {
			config.Log.Info("SuperAdmin seeded successfully (email: superadmin@learnyst.com, password: superadmin123)")
		}
	}

	config.Log.Info("Default data seeded successfully")
}
