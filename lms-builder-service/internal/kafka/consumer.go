package kafka
import (
	"context"
	"encoding/json"
	"log"
	"os"
	"github.com/IBM/sarama"
	"lms-builder-service/internal/engine"
)
type BuildRequestEvent struct {
	TenantSlug  string `json:"tenant_slug"`
	BuildType   string `json:"build_type"` 
	AcademyName string `json:"academy_name,omitempty"`
	Color       string `json:"color,omitempty"`
	ZipFilePath string `json:"zip_file_path,omitempty"` 
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
		log.Printf("Received build event: %s", string(message.Value))
		var event BuildRequestEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Failed to unmarshal build event: %v", err)
			continue
		}
		if event.BuildType == "template" {
			err := engine.BuildTemplateSite(event.TenantSlug, event.AcademyName, event.Color)
			if err != nil {
				log.Printf("Template build failed: %v", err)
			}
		} else if event.BuildType == "react" {
			err := engine.BuildCustomReact(event.TenantSlug, event.ZipFilePath)
			if err != nil {
				log.Printf("React build failed: %v", err)
			}
		}
		session.MarkMessage(message, "")
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
	client, err := sarama.NewConsumerGroup([]string{broker}, "builder-group", config)
	if err != nil {
		log.Fatalf("Error creating consumer group client: %v", err)
	}
	consumer := &Consumer{
		ready: make(chan bool),
	}
	go func() {
		for {
			if err := client.Consume(ctx, []string{"website.build_requested"}, consumer); err != nil {
				log.Fatalf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	log.Println("Builder Kafka Consumer up and running!")
}
