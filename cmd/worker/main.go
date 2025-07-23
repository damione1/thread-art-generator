package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	database "github.com/Damione1/thread-art-generator/core/db"
	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pbx"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/Damione1/thread-art-generator/threadGenerator"
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

	// Initialize dual bucket storage
	storage, err := initializeDualStorage(ctx, config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	defer storage.Close()

	// Connect to RabbitMQ and start processing
	if err := startQueueProcessing(ctx, config, storage); err != nil {
		log.Fatal().Err(err).Msg("Failed to start queue processing")
	}
}

func initializeDualStorage(ctx context.Context, config util.Config) (*storage.DualBucketStorage, error) {
	return storage.NewDualBucketStorage(ctx, config.Storage)
}

func startQueueProcessing(ctx context.Context, config util.Config, dualStorage *storage.DualBucketStorage) error {
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
			err := processMessage(ctx, d.Body, config.DB, dualStorage)
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
func processMessage(ctx context.Context, body []byte, db *sql.DB, dualStorage *storage.DualBucketStorage) error {
	processingStartTime := time.Now()

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
	art, err := models.Arts(
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

	// Create temporary directory for processing
	tempDir, err := os.MkdirTemp("", "composition-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download source image
	sourceImagePath := filepath.Join(tempDir, "source.jpg")
	sourceFile, err := os.Create(sourceImagePath)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to create source file: %v", err))
		return fmt.Errorf("failed to create source file: %w", err)
	}
	defer sourceFile.Close()

	imageKey := pbx.GetResourceName([]pbx.Resource{
		{Type: pbx.RessourceTypeUsers, ID: art.AuthorID},
		{Type: pbx.RessourceTypeArts, ID: art.ImageID.String},
	})

	log.Info().
		Str("imageKey", imageKey).
		Str("artID", art.ID).
		Str("imageID", art.ImageID.String).
		Msg("Attempting to download source image")

	reader, err := dualStorage.GetPublicStorage().Download(ctx, imageKey)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to download source image: %v", err))
		return fmt.Errorf("failed to download source image: %w", err)
	}
	defer reader.Close()

	written, err := io.Copy(sourceFile, reader)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to write source image: %v", err))
		return fmt.Errorf("failed to write source image: %w", err)
	}

	log.Info().
		Int64("bytesWritten", written).
		Str("path", sourceImagePath).
		Msg("Source image downloaded and saved")

	// Verify the file exists and has content
	if fi, err := os.Stat(sourceImagePath); err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to verify source image: %v", err))
		return fmt.Errorf("failed to verify source image: %w", err)
	} else {
		log.Info().
			Int64("size", fi.Size()).
			Msg("Source image file verified")
	}

	// Initialize thread generator with composition settings
	config := threadGenerator.DefaultConfig()
	config.NailsQuantity = composition.NailsQuantity
	config.ImgSize = composition.ImgSize
	config.MaxPaths = composition.MaxPaths
	config.StartingNail = composition.StartingNail
	config.MinimumDifference = composition.MinimumDifference
	config.BrightnessFactor = composition.BrightnessFactor
	config.ImageContrast = composition.ImageContrast
	config.PhysicalRadius = composition.PhysicalRadius

	// Log the configuration settings being used
	log.Info().
		Int("nailsQuantity", composition.NailsQuantity).
		Int("imgSize", composition.ImgSize).
		Int("maxPaths", composition.MaxPaths).
		Int("startingNail", composition.StartingNail).
		Int("minimumDifference", composition.MinimumDifference).
		Int("brightnessFactor", composition.BrightnessFactor).
		Float64("imageContrast", composition.ImageContrast).
		Float64("physicalRadius", composition.PhysicalRadius).
		Msg("Applying thread generator settings")

	generator := threadGenerator.NewThreadGenerator(config)
	generator.SetImage(sourceImagePath)

	// Generate thread art - now we can just pass the image name
	startTime := time.Now()
	stats, err := generator.Generate(threadGenerator.Args{
		ImageName: sourceImagePath,
	})
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to generate thread art: %v", err))
		return fmt.Errorf("failed to generate thread art: %w", err)
	}

	generationTime := time.Since(startTime)
	log.Info().
		Int("threadLength", stats.ThreadLength).
		Int("totalLines", stats.TotalLines).
		Msg("Thread art generation completed")

	// Generate preview image
	previewStartTime := time.Now()
	previewImage, err := generator.GeneratePathsImage()
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to generate preview image: %v", err))
		return fmt.Errorf("failed to generate preview image: %w", err)
	}

	previewGenerationTime := time.Since(previewStartTime)
	log.Info().Msg("Preview image generated")

	// Save preview image
	previewPath := filepath.Join(tempDir, "preview.png")
	previewFile, err := os.Create(previewPath)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to create preview file: %v", err))
		return fmt.Errorf("failed to create preview file: %w", err)
	}
	defer previewFile.Close()

	err = png.Encode(previewFile, previewImage)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to encode preview image: %v", err))
		return fmt.Errorf("failed to encode preview image: %w", err)
	}

	log.Info().Msg("Preview image saved to temp file")

	// Generate GCode
	gcode := generator.GetGcode()
	gcodePath := filepath.Join(tempDir, "gcode.txt")
	err = os.WriteFile(gcodePath, []byte(strings.Join(gcode, "\n")), 0644)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to write gcode file: %v", err))
		return fmt.Errorf("failed to write gcode file: %w", err)
	}

	log.Info().Msg("GCode file generated")

	// Get paths list
	paths := generator.GetPathsList()
	pathsJSON, err := json.Marshal(paths)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to marshal paths list: %v", err))
		return fmt.Errorf("failed to marshal paths list: %w", err)
	}

	pathsPath := filepath.Join(tempDir, "paths.json")
	err = os.WriteFile(pathsPath, pathsJSON, 0644)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to write paths file: %v", err))
		return fmt.Errorf("failed to write paths file: %w", err)
	}

	log.Info().Msg("Paths list file generated")

	// Upload files to storage
	uploadStartTime := time.Now()
	previewKey := fmt.Sprintf("users/%s/arts/%s/compositions/%s/preview.png", art.AuthorID, art.ID, composition.ID)
	gcodeKey := fmt.Sprintf("users/%s/arts/%s/compositions/%s/gcode.txt", art.AuthorID, art.ID, composition.ID)
	pathsKey := fmt.Sprintf("users/%s/arts/%s/compositions/%s/paths.json", art.AuthorID, art.ID, composition.ID)

	// Upload preview image
	previewFile, err = os.Open(previewPath)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to open preview file: %v", err))
		return fmt.Errorf("failed to open preview file: %w", err)
	}
	defer previewFile.Close()

	err = dualStorage.GetPublicStorage().Upload(ctx, previewKey, previewFile, "image/png")
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to upload preview image: %v", err))
		return fmt.Errorf("failed to upload preview image: %w", err)
	}

	log.Info().Str("key", previewKey).Msg("Preview image uploaded to bucket")

	// Upload GCode file
	gcodeFile, err := os.Open(gcodePath)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to open gcode file: %v", err))
		return fmt.Errorf("failed to open gcode file: %w", err)
	}
	defer gcodeFile.Close()

	err = dualStorage.GetPublicStorage().Upload(ctx, gcodeKey, gcodeFile, "text/plain")
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to upload gcode file: %v", err))
		return fmt.Errorf("failed to upload gcode file: %w", err)
	}

	log.Info().Str("key", gcodeKey).Msg("GCode file uploaded to bucket")

	// Upload paths file
	pathsFile, err := os.Open(pathsPath)
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to open paths file: %v", err))
		return fmt.Errorf("failed to open paths file: %w", err)
	}
	defer pathsFile.Close()

	err = dualStorage.GetPublicStorage().Upload(ctx, pathsKey, pathsFile, "application/json")
	if err != nil {
		setCompositionError(ctx, db, composition, fmt.Sprintf("failed to upload paths file: %v", err))
		return fmt.Errorf("failed to upload paths file: %w", err)
	}

	log.Info().Str("key", pathsKey).Msg("Paths file uploaded to bucket")

	uploadTime := time.Since(uploadStartTime)
	log.Info().Msg("All files uploaded")

	// Update composition with results
	composition.Status = models.CompositionStatusEnumCOMPLETE
	composition.PreviewURL = null.StringFrom(previewKey)
	composition.GcodeURL = null.StringFrom(gcodeKey)
	composition.PathlistURL = null.StringFrom(pathsKey)
	composition.ThreadLength = null.IntFrom(stats.ThreadLength)
	composition.TotalLines = null.IntFrom(stats.TotalLines)

	_, err = composition.Update(ctx, db, boil.Whitelist(
		models.CompositionColumns.Status,
		models.CompositionColumns.PreviewURL,
		models.CompositionColumns.GcodeURL,
		models.CompositionColumns.PathlistURL,
		models.CompositionColumns.ThreadLength,
		models.CompositionColumns.TotalLines,
	))
	if err != nil {
		return fmt.Errorf("failed to update composition with results: %w", err)
	}

	log.Info().
		Str("compositionID", composition.ID).
		Int("threadLength", stats.ThreadLength).
		Int("totalLines", stats.TotalLines).
		Msg("Composition processing completed successfully")

	totalProcessingTime := time.Since(processingStartTime)
	log.Info().
		Str("compositionID", composition.ID).
		Dur("totalTime", totalProcessingTime).
		Dur("generationTime", generationTime).
		Dur("previewGenerationTime", previewGenerationTime).
		Dur("uploadTime", uploadTime).
		Int("threadLength", stats.ThreadLength).
		Int("totalLines", stats.TotalLines).
		Msgf("🎉 Processing summary: Total: %s | Thread art: %s | Preview: %s | Upload: %s",
			totalProcessingTime,
			generationTime,
			previewGenerationTime,
			uploadTime)

	return nil
}

func setCompositionError(ctx context.Context, db *sql.DB, composition *models.Composition, errorMessage string) {
	composition.Status = models.CompositionStatusEnumFAILED
	composition.ErrorMessage = null.StringFrom(errorMessage)
	_, err := composition.Update(ctx, db, boil.Whitelist(
		models.CompositionColumns.Status,
		models.CompositionColumns.ErrorMessage,
	))
	if err != nil {
		log.Error().Err(err).Msg("Failed to update composition error status")
	}
}
