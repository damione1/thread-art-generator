package queue

import "context"

// Queue is the cloud-agnostic job bus.
// Default impl: Postgres SKIP LOCKED. Pub/Sub is a legacy adapter.
type Queue interface {
	Publish(ctx context.Context, topic string, body []byte) error
	Subscribe(ctx context.Context, topic, consumer string, h Handler) error
	Close() error
}

// Handler processes one message. Return error to retry/NACK.
type Handler func(ctx context.Context, body []byte) error

const (
	TopicCompositionProcessing = "composition-processing"
)
