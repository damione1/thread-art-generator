package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/sqlboiler/v4/boil"

	database "github.com/Damione1/thread-art-generator/core/db"
	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/util"
)

func main() {
	// Configure logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("🧵 Starting worker service")

	// Load configuration
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to load config")
	}

	// Connect to database (using same pattern as in API main.go)
	_, err = database.ConnectDb(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to connect to database")
	}

	// Create context with cancel for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	bucket, err := initializeStorage(ctx, config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	defer bucket.Close()

	// Connect to RabbitMQ and start processing
	if err := startQueueProcessing(ctx, config, bucket); err != nil {
		log.Fatal().Err(err).Msg("Failed to start queue processing")
	}
}

func initializeStorage(ctx context.Context, config util.Config) (*storage.BlobStorage, error) {
	// Convert provider string to StorageProvider type
	var provider storage.StorageProvider
	switch config.Storage.Provider {
	case "s3":
		provider = storage.ProviderS3
	case "minio":
		provider = storage.ProviderMinIO
	case "gcs":
		provider = storage.ProviderGCS
	default:
		if config.Environment == "development" {
			provider = storage.ProviderMinIO
		} else {
			provider = storage.ProviderS3
		}
	}

	storageConfig := storage.BlobStorageConfig{
		Provider:         provider,
		Bucket:           config.Storage.Bucket,
		Region:           config.Storage.Region,
		InternalEndpoint: config.Storage.InternalEndpoint,
		ExternalEndpoint: config.Storage.ExternalEndpoint,
		UseSSL:           config.Storage.UseSSL,
		ForceExternalSSL: config.Storage.ForceExternalSSL,
		AccessKey:        config.Storage.AccessKey,
		SecretKey:        config.Storage.SecretKey,
		GCPProjectID:     config.Storage.GCPProjectID,
	}

	// If config values are missing, provide reasonable defaults
	if storageConfig.Bucket == "" {
		storageConfig.Bucket = "local-bucket"
	}

	if storageConfig.Region == "" {
		storageConfig.Region = "us-east-1"
	}

	// Set up endpoints based on environment if not provided
	if config.Environment == "development" && provider == storage.ProviderMinIO {
		if storageConfig.InternalEndpoint == "" {
			storageConfig.InternalEndpoint = "http://minio:9000"
		}
		if storageConfig.ExternalEndpoint == "" {
			storageConfig.ExternalEndpoint = "http://localhost:9000"
		}
	}

	return storage.NewBlobStorage(ctx, storageConfig)
}

func startQueueProcessing(ctx context.Context, config util.Config, bucket *storage.BlobStorage) error {
	// Connect to RabbitMQ
	queueURL := config.Queue.URL
	if queueURL == "" {
		queueURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	conn, err := amqp.Dial(queueURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// Get queue name from config
	queueName := config.Queue.CompositionProcessing
	if queueName == "" {
		queueName = "composition-processing"
	}

	// Declare queue
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Set QoS
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Consume messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	go func() {
		for d := range msgs {
			log.Info().Int("size", len(d.Body)).Msg("Received a message")

			// Process message
			err := processMessage(ctx, d.Body, config.DB, bucket)
			if err != nil {
				log.Error().Err(err).Msg("Failed to process message")

				// Nack the message with requeue
				err = d.Nack(false, true)
				if err != nil {
					log.Error().Err(err).Msg("Failed to nack message")
				}

				// Wait a bit before continuing to avoid rapid requeuing
				time.Sleep(5 * time.Second)
				continue
			}

			// Ack the message
			err = d.Ack(false)
			if err != nil {
				log.Error().Err(err).Msg("Failed to ack message")
			}
		}
	}()

	log.Info().Str("queue", queueName).Msg("🧵 Worker is waiting for messages")

	// Wait for termination signal
	<-sigChan
	log.Info().Msg("Received termination signal, shutting down")
	return nil
}

// processMessage processes a single message from the queue
func processMessage(ctx context.Context, body []byte, db *sql.DB, bucket *storage.BlobStorage) error {
	// Parse the message
	var message queue.CompositionProcessingMessage
	err := message.FromJSON(body)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Verify message type
	if message.Type != queue.MessageTypeCompositionProcessing {
		return fmt.Errorf("unexpected message type: %s", message.Type)
	}

	log.Info().
		Str("type", message.Type).
		Str("artID", message.ArtID).
		Str("compositionID", message.CompositionID).
		Msg("Processing composition")

	// Get the composition
	composition, err := models.Compositions(
		models.CompositionWhere.ID.EQ(message.CompositionID),
		models.CompositionWhere.ArtID.EQ(message.ArtID),
	).One(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get composition: %w", err)
	}

	// Get the art (needed for accessing the image)
	_, err = models.Arts(
		models.ArtWhere.ID.EQ(message.ArtID),
	).One(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get art: %w", err)
	}

	// Update status to processing
	composition.Status = models.CompositionStatusEnumPROCESSING
	_, err = composition.Update(ctx, db, boil.Whitelist(models.CompositionColumns.Status))
	if err != nil {
		return fmt.Errorf("failed to update composition status: %w", err)
	}

	// TODO: Actual thread generation processing here

	// Simulate processing time for now
	time.Sleep(5 * time.Second)

	// For now, just update the status to COMPLETE
	// In the real implementation, we will:
	// 1. Generate the thread art using the composition settings
	// 2. Save the preview image to storage
	// 3. Save the GCode and path list files to storage
	// 4. Update the composition record with file URLs

	// Update status to complete
	composition.Status = models.CompositionStatusEnumCOMPLETE
	_, err = composition.Update(ctx, db, boil.Whitelist(models.CompositionColumns.Status))
	if err != nil {
		return fmt.Errorf("failed to update composition status: %w", err)
	}

	log.Info().
		Str("compositionID", composition.ID).
		Msg("Composition processing completed")

	return nil
}
