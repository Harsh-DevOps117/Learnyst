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
	cassandra.InitCassandra()
	kafka.InitProducer()
	ctx, cancel := context.WithCancel(context.Background())
	go kafka.StartConsumer(ctx)
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
