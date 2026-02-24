package kafka

import "context"

type Message struct {
	Key        []byte
	Value      []byte
	Topic      string
	RetryCount int
}

type Producer interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

type Handler func(ctx context.Context, msg Message) error
