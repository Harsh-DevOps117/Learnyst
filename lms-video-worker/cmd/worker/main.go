package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"lms-video-worker/internal/cassandra"
	"lms-video-worker/internal/kafka"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	log.Println("Starting LMS Video Worker...")

	// Initialize Cassandra for Analytics
	cassandra.InitCassandra()

	// Initialize Kafka Consumer for Video Transcoding
	kafka.InitProducer()
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start consumer in a goroutine
	go kafka.StartConsumer(ctx)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down LMS Video Worker...")
	cancel()
	if cassandra.Session != nil {
		cassandra.Session.Close()
	}
	log.Println("Worker stopped safely.")
}
