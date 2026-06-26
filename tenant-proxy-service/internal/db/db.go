package db
import (
	"log"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
type Tenant struct {
	ID           uint
	AcademySlug  string
	CustomDomain string
	IsActive     bool
	IsSuspended  bool
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
		log.Fatalf("Failed to connect to proxy database: %v", err)
	}
	log.Println("Proxy Database connected successfully")
}
func GetTenantByHost(host string) (*Tenant, error) {
	var tenant Tenant
	err := DB.Where("custom_domain = ? AND is_active = ? AND is_suspended = ?", host, true, false).First(&tenant).Error
	if err == nil {
		return &tenant, nil
	}
	err = DB.Where("academy_slug = ? AND is_active = ? AND is_suspended = ?", host, true, false).First(&tenant).Error
	if err == nil {
		return &tenant, nil
	}
	return nil, err
}
