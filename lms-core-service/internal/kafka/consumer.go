package kafka
import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
	"github.com/IBM/sarama"
	"lms-core-service/internal/database"
	"lms-core-service/internal/models"
)
type VideoTranscodedEvent struct {
	VideoID uint   `json:"video_id"`
	HlsUrl  string `json:"hls_url"`
}
type Consumer struct {
	ready chan bool
}
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Received video.transcoded event: value=%s", string(message.Value))
		var event VideoTranscodedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Failed to unmarshal transcoded event: %v", err)
			continue
		}
		var video models.CourseVideo
		if err := database.DB.First(&video, event.VideoID).Error; err == nil {
			video.VideoUrl = event.HlsUrl
			if err := database.DB.Save(&video).Error; err != nil {
				log.Printf("Failed to update video %d url: %v", event.VideoID, err)
			} else {
				log.Printf("Successfully updated video %d with HLS url %s", event.VideoID, event.HlsUrl)
				session.MarkMessage(message, "")
			}
		} else {
			log.Printf("Video %d not found in DB", event.VideoID)
		}
	}
	return nil
}
func StartConsumer(ctx context.Context) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	var client sarama.ConsumerGroup
	var err error
	for {
		client, err = sarama.NewConsumerGroup([]string{broker}, "core-video-consumer-group", config)
		if err == nil {
			break
		}
		log.Printf("Error creating consumer group client: %v, retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}
	consumer := &Consumer{
		ready: make(chan bool),
	}
	go func() {
		for {
			if err := client.Consume(ctx, []string{"video.transcoded"}, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	log.Println("Core Kafka Consumer for video.transcoded is running!")
}
