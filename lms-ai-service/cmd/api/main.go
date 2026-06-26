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
	db.InitDB()
	ai.InitGroq()
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/doubts", handlers.AskDoubtHandler)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002" 
	}
	log.Printf("LMS AI Service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
