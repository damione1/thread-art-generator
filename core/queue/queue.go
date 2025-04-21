package queue

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// QueueClient defines methods for interacting with the message queue
type QueueClient interface {
	// PublishMessage publishes a message to a queue
	PublishMessage(ctx context.Context, queueName string, message []byte) error

	// Close closes the queue connection
	Close() error
}

// RabbitMQClient implements QueueClient for RabbitMQ
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	client := &RabbitMQClient{
		conn:    conn,
		channel: ch,
	}

	return client, nil
}

// declareQueue declares a queue if it doesn't exist
func (c *RabbitMQClient) declareQueue(queueName string) error {
	_, err := c.channel.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	return err
}

// PublishMessage publishes a message to a queue
func (c *RabbitMQClient) PublishMessage(ctx context.Context, queueName string, message []byte) error {
	// Ensure the queue exists
	err := c.declareQueue(queueName)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Publish the message
	err = c.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Persistent to survive broker restarts
			Body:         message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Info().Str("queue", queueName).Int("messageSize", len(message)).Msg("Message published to queue")
	return nil
}

// Close closes the connection to RabbitMQ
func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close RabbitMQ channel")
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
		}
	}

	return nil
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
