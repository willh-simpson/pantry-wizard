package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type Writer struct {
	writer  *kafka.Writer
	brokers []string
}

func NewProducer(brokers []string) *Writer {
	return &Writer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
		brokers: brokers,
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

func (w *Writer) Ping(ctx context.Context) error {
	dialer := &kafka.Dialer{
		Timeout:   5 * time.Second,
		DualStack: true,
	}

	if len(w.brokers) == 0 {
		return fmt.Errorf("no brokers configured")
	}

	conn, err := dialer.DialContext(ctx, "tcp", w.brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func (w *Writer) Close() error {
	return w.writer.Close()
}
