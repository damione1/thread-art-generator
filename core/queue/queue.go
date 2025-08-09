package queue

import (
	"context"
	"encoding/json"
	"fmt"
)

// QueueClient defines methods for interacting with the message queue
type QueueClient interface {
	// PublishMessage publishes a message to a queue
	PublishMessage(ctx context.Context, queueName string, message []byte) error

	// Close closes the queue connection
	Close() error
}

// Message types for different queue operations
const (
	MessageTypeCompositionProcessing = "composition_processing"
)

// BaseMessage is the base structure for all queue messages
type BaseMessage struct {
	Type string `json:"type"`
}

// CompositionProcessingMessage is the message for composition processing
type CompositionProcessingMessage struct {
	BaseMessage
	ArtID         string `json:"art_id"`
	CompositionID string `json:"composition_id"`
}

// NewCompositionProcessingMessage creates a new composition processing message
func NewCompositionProcessingMessage(artID, compositionID string) *CompositionProcessingMessage {
	return &CompositionProcessingMessage{
		BaseMessage: BaseMessage{
			Type: MessageTypeCompositionProcessing,
		},
		ArtID:         artID,
		CompositionID: compositionID,
	}
}

// ToJSON serializes the message to JSON
func (m *CompositionProcessingMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON deserializes the message from JSON
func (m *CompositionProcessingMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// PubSubConfig holds configuration for Pub/Sub client
type PubSubConfig struct {
	ProjectID    string
	EmulatorHost string
	Environment  string
}

// NewQueueClient creates a new queue client based on the environment
func NewQueueClient(ctx context.Context, queueProvider string, config any) (QueueClient, error) {
	switch queueProvider {
	case "pubsub":
		// Config should be a PubSubConfig struct
		pubsubConfig, ok := config.(PubSubConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for pubsub provider, expected PubSubConfig")
		}
		return NewPubSubClient(ctx, pubsubConfig.ProjectID, pubsubConfig.EmulatorHost, pubsubConfig.Environment)
	default:
		return nil, fmt.Errorf("unsupported queue provider: %s", queueProvider)
	}
}
