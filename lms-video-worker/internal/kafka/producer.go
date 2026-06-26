package kafka

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

var Producer sarama.SyncProducer

func InitProducer() {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	var err error
	for {
		Producer, err = sarama.NewSyncProducer([]string{broker}, config)
		if err == nil {
			break
		}
		log.Printf("Error creating Kafka Producer client: %v, retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}
	log.Println("Kafka Producer initialized successfully")
}

type VideoTranscodedEvent struct {
	VideoID uint   `json:"video_id"`
	HlsUrl  string `json:"hls_url"`
}

func PublishVideoTranscoded(event VideoTranscodedEvent) error {
	if Producer == nil {
		log.Println("Kafka Producer not initialized, skipping event")
		return nil
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "video.transcoded",
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = Producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to publish video.transcoded event: %v", err)
		return err
	}
	log.Printf("Published video.transcoded event for Video ID: %d", event.VideoID)
	return nil
}
