package kafka

import (
	"context"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type Writer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Writer {
	return &Writer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (w *Writer) Publish(ctx context.Context, msg Message) error {
	return w.writer.WriteMessages(ctx, kafka.Message{
		Topic: msg.Topic,
		Key:   msg.Key,
		Value: msg.Value,
		Headers: []kafka.Header{
			{
				Key:   "x-retry-count",
				Value: []byte(strconv.Itoa(msg.RetryCount)),
			},
		},
	})
}
