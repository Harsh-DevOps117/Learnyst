package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/IBM/sarama"

	"lms-video-worker/internal/transcoder"
)

type VideoUploadedEvent struct {
	TenantID uint   `json:"tenant_id"`
	CourseID uint   `json:"course_id"`
	VideoID  uint   `json:"video_id"`
	RawUrl   string `json:"raw_url"`
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

		// Trigger the adaptive streaming FFMPEG pipeline
		masterUrl, err := transcoder.ProcessAdaptiveStreaming(event.TenantID, event.VideoID, event.RawUrl)
		if err != nil {
			log.Printf("FFMPEG pipeline failed for video %d: %v", event.VideoID, err)
			// Depending on strategy, we could write to a dead-letter queue here
		} else {
			log.Printf("Successfully transcoded video %d to HLS", event.VideoID)
			
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

	client, err := sarama.NewConsumerGroup([]string{broker}, "video-workers-group", config)
	if err != nil {
		log.Fatalf("Error creating consumer group client: %v", err)
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
