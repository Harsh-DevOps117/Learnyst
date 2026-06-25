package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Tenant struct {
	ID        uint
	LogoUrl   string
	BannerUrl string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found or error loading")
	}

	dsn := "postgresql://adminFutre123:adminFutreKanye123@localhost:5432/adminFutreMetro123?sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	var tenants []Tenant
	db.Find(&tenants)

	fmt.Println("Current Tenants in DB:")
	for _, t := range tenants {
		fmt.Printf("Tenant %d - Logo: '%s', Banner: '%s'\n", t.ID, t.LogoUrl, t.BannerUrl)
		
		// If they are empty, let's just insert some cool dummy images to prove it works
		if t.LogoUrl == "" {
			t.LogoUrl = "https://cdn-icons-png.flaticon.com/512/2950/2950137.png" // Dummy logo
			db.Save(&t)
			fmt.Printf("  -> Injected dummy logo for Tenant %d\n", t.ID)
		}
		if t.BannerUrl == "" {
			t.BannerUrl = "https://images.unsplash.com/photo-1524178232363-1fb2b075b655?q=80&w=2070&auto=format&fit=crop" // Beautiful classroom banner
			db.Save(&t)
			fmt.Printf("  -> Injected dummy banner for Tenant %d\n", t.ID)
		}
	}
}
