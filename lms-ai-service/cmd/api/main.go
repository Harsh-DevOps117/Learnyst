package main

import (
	"log"
	"os"

	"lms-ai-service/internal/ai"
	"lms-ai-service/internal/db"
	"lms-ai-service/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	log.Println("Starting LMS AI Service...")

	// Init DB
	db.InitDB()

	// Init Groq AI
	ai.InitGroq()

	// Setup Gin
	r := gin.Default()

	// Routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/doubts", handlers.AskDoubtHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002" // Different port to avoid conflict with Core and Gateway
	}

	log.Printf("LMS AI Service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
