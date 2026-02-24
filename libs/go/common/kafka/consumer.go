package kafka

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader   *kafka.Reader
	producer Producer
	maxRetry int
}

func (c *Consumer) Consume(ctx context.Context, handler Handler) {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			break
		}

		retryCount := 0
		for _, h := range m.Headers {
			if h.Key == "x-retry-count" {
				retryCount, _ = strconv.Atoi(string(h.Value))
			}
		}

		msg := Message{
			Key:        m.Key,
			Value:      m.Value,
			Topic:      m.Topic,
			RetryCount: retryCount,
		}

		if err := handler(ctx, msg); err != nil {
			c.handleFailure(ctx, msg)
		}
	}
}

func (c *Consumer) handleFailure(ctx context.Context, msg Message) {
	if msg.RetryCount >= c.maxRetry {
		log.Printf("max retries reached. sending to DLQ: %s", msg.Topic+".dlq")
		msg.Topic = msg.Topic + ".dlq"

		c.producer.Publish(ctx, msg)

		return
	}

	msg.RetryCount++

	backoffDuration := c.getBackoffDuration(msg)
	log.Printf("retrying message in %v. attempt #%d", backoffDuration, msg.RetryCount)

	go func(d time.Duration, m Message) {
		time.Sleep(d)

		if err := c.producer.Publish(context.Background(), m); err != nil {
			log.Printf("failed to re-publish: %v", err)
		}
	}(backoffDuration, msg)

	c.producer.Publish(ctx, msg)
}

// exponential backoff -> 2 ^ retryCount
func (c *Consumer) getBackoffDuration(msg Message) time.Duration {
	return time.Duration(1<<uint(msg.RetryCount)) * time.Second
}
