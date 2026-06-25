package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// We only need the fields relevant for routing here
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
		dbUri = "postgresql://adminFutre123:adminFutreKanye123@localhost:5432/adminFutreMetro123?sslmode=disable"
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
	// Try looking up by CustomDomain first
	err := DB.Where("custom_domain = ? AND is_active = ? AND is_suspended = ?", host, true, false).First(&tenant).Error
	if err == nil {
		return &tenant, nil
	}

	// If not found by custom domain, it might be a subdomain (e.g., slug.learnyst.com)
	// The host will be "slug.learnyst.com". We just look up by slug if we can extract it.
	// For now, let's assume if it fails custom domain, we check if it matches a slug.
	err = DB.Where("academy_slug = ? AND is_active = ? AND is_suspended = ?", host, true, false).First(&tenant).Error
	if err == nil {
		return &tenant, nil
	}

	return nil, err
}
