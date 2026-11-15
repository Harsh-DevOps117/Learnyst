package main

import (
	"github.com/joho/godotenv"
	"lms-core-service/internal/database"
	"lms-core-service/internal/handlers"
	"lms-core-service/internal/storage"
	"log"
	"os"
)

func main() {
	_ = godotenv.Load()
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatal("Core database connection failed")
	}
	if err = database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	storage.InitStorage()
	if err = database.SeedLocal(db); err != nil {
		log.Fatal("Local seed failed: ", err)
	}
	router := handlers.LocalRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3011"
	}
	log.Print("Learnyst core listening on http://127.0.0.1:", port)
	if err = router.Run("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}
