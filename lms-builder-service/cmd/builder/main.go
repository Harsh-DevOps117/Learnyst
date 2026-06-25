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

	// Initialize S3 Uploader
	azureblob.InitAzureBlob()

	// Start Kafka Consumer
	ctx, cancel := context.WithCancel(context.Background())
	go kafka.StartConsumer(ctx)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Builder Service...")
	cancel()
	log.Println("Builder Service stopped safely.")
}
