package kafka

import (
	"encoding/json"
	"log"
	"os"

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
	Producer, err = sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		log.Println("WARNING: Failed to initialize Kafka Producer:", err)
		return
	}
	log.Println("Kafka Producer initialized successfully")
}

type VideoUploadedEvent struct {
	TenantID uint   `json:"tenant_id"`
	CourseID uint   `json:"course_id"`
	VideoID  uint   `json:"video_id"`
	RawUrl   string `json:"raw_url"`
}

func PublishVideoUploaded(event VideoUploadedEvent) error {
	if Producer == nil {
		log.Println("Kafka Producer not initialized, skipping event")
		return nil
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "video.uploaded",
		Value: sarama.ByteEncoder(bytes),
	}

	partition, offset, err := Producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to publish video.uploaded event: %v", err)
		return err
	}
	log.Printf("Published video.uploaded event to partition %d at offset %d", partition, offset)
	return nil
}
