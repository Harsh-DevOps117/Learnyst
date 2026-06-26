package kafka
import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
	"github.com/IBM/sarama"
	"lms-video-worker/internal/azureblob"
	"lms-video-worker/internal/transcoder"
)
type VideoUploadedEvent struct {
	TenantID uint   `json:"tenant_id"`
	CourseID uint   `json:"course_id"`
	VideoID  uint   `json:"video_id"`
	RawUrl   string `json:"raw_url"`
	Slug     string `json:"slug"`
	Phone    string `json:"phone"`
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
		log.Printf("Received video event: value=%s, timestamp=%v, topic=%s", string(message.Value), message.Timestamp, message.Topic)
		var event VideoUploadedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}
		masterUrl, err := transcoder.ProcessAdaptiveStreaming(event.TenantID, event.CourseID, event.VideoID, event.RawUrl, event.Slug, event.Phone)
		if err != nil {
			log.Printf("FFMPEG pipeline failed for video %d: %v", event.VideoID, err)
		} else {
			log.Printf("Successfully transcoded video %d to HLS", event.VideoID)
			if strings.HasPrefix(event.RawUrl, "local://") {
				localPath := strings.TrimPrefix(event.RawUrl, "local://")
				if err := os.Remove(localPath); err != nil {
					log.Printf("Warning: Failed to delete local video %d: %v", event.VideoID, err)
				} else {
					log.Printf("Successfully deleted local video %d", event.VideoID)
				}
			} else {
				if err := azureblob.DeleteBlobFromUrl(event.RawUrl); err != nil {
					log.Printf("Warning: Failed to delete raw video from Azure for video %d: %v", event.VideoID, err)
				} else {
					log.Printf("Successfully deleted raw video from Azure for video %d", event.VideoID)
				}
			}
			if masterUrl != "" {
				err = PublishVideoTranscoded(VideoTranscodedEvent{
					VideoID: event.VideoID,
					HlsUrl:  masterUrl,
				})
				if err != nil {
					log.Printf("Failed to emit transcoded event: %v", err)
				}
			}
			session.MarkMessage(message, "")
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
		client, err = sarama.NewConsumerGroup([]string{broker}, "video-workers-group", config)
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
			if err := client.Consume(ctx, []string{"video.uploaded"}, consumer); err != nil {
				log.Fatalf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	log.Println("Kafka Consumer up and running!")
}
