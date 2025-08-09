package queue

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"
)

// PubSubClient implements QueueClient for Google Cloud Pub/Sub
type PubSubClient struct {
	client      *pubsub.Client
	projectID   string
	environment string
}

// NewPubSubClient creates a new Pub/Sub client
func NewPubSubClient(ctx context.Context, projectID string, emulatorHost string, environment string) (*PubSubClient, error) {
	// Check if we're using the emulator
	if emulatorHost != "" {
		log.Info().Str("emulatorHost", emulatorHost).Msg("Using Pub/Sub emulator")
		// The PUBSUB_EMULATOR_HOST environment variable should be set at application startup
		// The client library automatically detects and uses the emulator
	} else {
		log.Info().Msg("Using production Pub/Sub service")
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		if emulatorHost != "" {
			return nil, fmt.Errorf("failed to create Pub/Sub client (emulator at %s): %w. Make sure the Pub/Sub emulator is running", emulatorHost, err)
		}
		return nil, fmt.Errorf("failed to create Pub/Sub client: %w", err)
	}

	return &PubSubClient{
		client:      client,
		projectID:   projectID,
		environment: environment,
	}, nil
}

// ensureTopic creates a topic if it doesn't exist
func (c *PubSubClient) ensureTopic(ctx context.Context, topicName string) (*pubsub.Topic, error) {
	topic := c.client.Topic(topicName)

	// Check if topic exists
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check topic existence: %w", err)
	}

	if !exists {
		// Create the topic
		topic, err = c.client.CreateTopic(ctx, topicName)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic: %w", err)
		}
		log.Info().Str("topic", topicName).Msg("Created Pub/Sub topic")
	}

	return topic, nil
}

// PublishMessage publishes a message to a topic
func (c *PubSubClient) PublishMessage(ctx context.Context, topicName string, message []byte) error {
	// Ensure topic exists
	topic, err := c.ensureTopic(ctx, topicName)
	if err != nil {
		return err
	}

	// Publish the message
	result := topic.Publish(ctx, &pubsub.Message{
		Data: message,
		Attributes: map[string]string{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})

	// Wait for the publish to complete
	messageID, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Info().
		Str("topic", topicName).
		Str("messageID", messageID).
		Int("messageSize", len(message)).
		Msg("Message published to Pub/Sub")

	return nil
}

// Close closes the Pub/Sub client
func (c *PubSubClient) Close() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("failed to close Pub/Sub client: %w", err)
		}
	}
	return nil
}

// CreateSubscription creates a subscription for a topic if it doesn't exist
func (c *PubSubClient) CreateSubscription(ctx context.Context, subscriptionName, topicName string) (*pubsub.Subscription, error) {
	sub := c.client.Subscription(subscriptionName)

	// Check if subscription exists
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription existence: %w", err)
	}

	if exists {
		return sub, nil
	}

	// Ensure topic exists first
	topic, err := c.ensureTopic(ctx, topicName)
	if err != nil {
		return nil, err
	}

	// Create subscription configuration
	config := pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 10 * time.Minute, // 10 minutes for processing
		RetryPolicy: &pubsub.RetryPolicy{
			MinimumBackoff: 10 * time.Second,
			MaximumBackoff: 300 * time.Second,
		},
		// Enable message ordering if needed
		EnableMessageOrdering: false,
	}
	
	// Only add dead letter policy for production (not emulator)
	// Check if we're using emulator by looking at the environment
	if c.environment != "development" {
		// Production configuration with dead letter policy
		config.DeadLetterPolicy = &pubsub.DeadLetterPolicy{
			DeadLetterTopic:     c.client.Topic(topicName + "-dead-letter").String(),
			MaxDeliveryAttempts: 5,
		}
	}
	
	// Create the subscription
	sub, err = c.client.CreateSubscription(ctx, subscriptionName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	log.Info().
		Str("subscription", subscriptionName).
		Str("topic", topicName).
		Msg("Created Pub/Sub subscription")

	return sub, nil
}

// ReceiveMessages receives messages from a subscription
func (c *PubSubClient) ReceiveMessages(ctx context.Context, subscriptionName string, handler func(context.Context, []byte) error) error {
	sub := c.client.Subscription(subscriptionName)

	// Configure receive settings
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute
	sub.ReceiveSettings.MaxOutstandingMessages = 1 // Process one at a time

	log.Info().Str("subscription", subscriptionName).Msg("Starting to receive messages from Pub/Sub")

	// Receive messages
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Info().
			Str("messageID", msg.ID).
			Int("size", len(msg.Data)).
			Msg("Received message from Pub/Sub")

		// Process the message
		if err := handler(ctx, msg.Data); err != nil {
			log.Error().Err(err).Str("messageID", msg.ID).Msg("Failed to process message")
			// Nack the message to retry
			msg.Nack()
			return
		}

		// Ack the message
		msg.Ack()
		log.Info().Str("messageID", msg.ID).Msg("Message processed successfully")
	})

	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	return nil
}
