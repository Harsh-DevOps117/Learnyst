package main
import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"lms-builder-service/internal/kafka"
	"lms-builder-service/internal/azureblob"
	"github.com/joho/godotenv"
)
func main() {
	_ = godotenv.Load()
	log.Println("Starting LMS Builder Service...")
	azureblob.InitAzureBlob()
	ctx, cancel := context.WithCancel(context.Background())
	go kafka.StartConsumer(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Builder Service...")
	cancel()
	log.Println("Builder Service stopped safely.")
}
