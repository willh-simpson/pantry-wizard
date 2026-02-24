package main

import (
	"context"
	"log"
	"time"

	"github.com/willh-simpson/pantry-wizard/libs/go/common/kafka"
)

func main() {
	brokers := []string{"localhost:9092"}
	producer := kafka.NewProducer(brokers)

	msg := kafka.Message{
		Topic: "user-signals",
		Key:   []byte("test-user-123"),
		Value: []byte(`{"event": "test_ping", "status": "success"}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("attempting to send message to Kafka...")
	if err := producer.Publish(ctx, msg); err != nil {
		log.Fatalf("failed to publish: %v", err)
	}

	log.Println("message successfully sent to 'user-signals' topic")
}
