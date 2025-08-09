package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/Damione1/thread-art-generator/core/util"
)


// StorageServiceServer implements the StorageService RPC methods
type StorageServiceServer struct {
	config  util.Config
	storage storage.StorageProvider
}

// NewStorageService creates a new storage service with Firebase integration
func NewStorageService(config util.Config) (*StorageServiceServer, error) {
	ctx := context.Background()

	// Initialize single storage provider using the new interface approach
	storageProvider, err := storage.NewStorageProviderFromUtil(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage provider: %w", err)
	}

	log.Info().
		Str("bucket", config.Storage.Bucket).
		Bool("emulator_mode", config.IsEmulatorMode()).
		Msg("Storage service initialized with single StorageProvider interface")

	return &StorageServiceServer{
		config:  config,
		storage: storageProvider,
	}, nil
}

// GenerateUploadURL generates a signed URL for secure file uploads
func (s *StorageServiceServer) GenerateUploadURL(
	ctx context.Context,
	req *connect.Request[pb.GenerateUploadURLRequest],
) (*connect.Response[pb.GenerateUploadURLResponse], error) {
	
	log.Info().
		Str("user_id", req.Msg.UserId).
		Str("art_id", req.Msg.ArtId).
		Str("content_type", req.Msg.ContentType).
		Int64("file_size", req.Msg.FileSize).
		Msg("Generating upload URL")

	// Validate content type is provided and valid
	if req.Msg.ContentType == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("content type is required for uploads"),
		)
	}

	// Validate content type is an image
	if !strings.HasPrefix(req.Msg.ContentType, "image/") {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("only image files are allowed, got: %s", req.Msg.ContentType),
		)
	}

	// Validate file size against configured maximum
	maxFileSizeBytes := int64(s.config.StorageService.MaxFileSizeMB * 1024 * 1024)
	if req.Msg.FileSize > maxFileSizeBytes {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d MB", 
				req.Msg.FileSize, s.config.StorageService.MaxFileSizeMB),
		)
	}

	// Build storage path: users/{user_id}/arts/{art_id}
	storagePath := fmt.Sprintf("users/%s/arts/%s", req.Msg.UserId, req.Msg.ArtId)

	// Calculate expiration time from config
	ttl := time.Duration(s.config.StorageService.SignedURLTTLMinutes) * time.Minute
	expiresAt := time.Now().Add(ttl)

	// Generate signed URL for upload (PUT method)
	signedURLOpts := &storage.SignedURLOptions{
		Method:      "PUT",
		Expiry:      ttl,
		ContentType: req.Msg.ContentType,
	}

	// Use storage provider for art uploads (security handled by Firebase storage rules)
	uploadURL, err := s.storage.SignedURL(ctx, storagePath, signedURLOpts)
	if err != nil {
		log.Error().
			Err(err).
			Str("storage_path", storagePath).
			Str("content_type", req.Msg.ContentType).
			Msg("Failed to generate upload signed URL")
		
		return nil, connect.NewError(
			connect.CodeInternal,
			fmt.Errorf("failed to generate upload URL: %w", err),
		)
	}

	log.Info().
		Str("storage_path", storagePath).
		Str("content_type", req.Msg.ContentType).
		Str("expires_at", expiresAt.Format(time.RFC3339)).
		Msg("Upload URL generated successfully with content type")

	response := &pb.GenerateUploadURLResponse{
		UploadUrl:     uploadURL,
		StoragePath:   storagePath,
		ExpiresAt:     timestamppb.New(expiresAt),
		MaxFileSize:   maxFileSizeBytes,
	}

	return connect.NewResponse(response), nil
}

// GenerateDownloadURL generates a signed URL for secure file downloads
func (s *StorageServiceServer) GenerateDownloadURL(
	ctx context.Context,
	req *connect.Request[pb.GenerateDownloadURLRequest],
) (*connect.Response[pb.GenerateDownloadURLResponse], error) {
	
	log.Info().
		Str("user_id", req.Msg.UserId).
		Str("art_id", req.Msg.ArtId).
		Msg("Generating download URL")

	// Determine storage path
	var storagePath string
	if req.Msg.FilePath != nil && *req.Msg.FilePath != "" {
		// Use provided file path
		storagePath = *req.Msg.FilePath
	} else {
		// Use default art image path
		storagePath = fmt.Sprintf("users/%s/arts/%s", req.Msg.UserId, req.Msg.ArtId)
	}

	// Calculate expiration time from config
	ttl := time.Duration(s.config.StorageService.SignedURLTTLMinutes) * time.Minute
	expiresAt := time.Now().Add(ttl)

	// Check if file exists first
	exists, err := s.storage.Exists(ctx, storagePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("storage_path", storagePath).
			Msg("Failed to check if file exists")
		
		return nil, connect.NewError(
			connect.CodeInternal,
			fmt.Errorf("failed to check file existence: %w", err),
		)
	}

	if !exists {
		return nil, connect.NewError(
			connect.CodeNotFound,
			fmt.Errorf("file not found at path: %s", storagePath),
		)
	}

	// Generate signed URL for download (GET method)
	signedURLOpts := &storage.SignedURLOptions{
		Method: "GET",
		Expiry: ttl,
	}

	// Use storage provider for art downloads (security handled by Firebase storage rules)
	downloadURL, err := s.storage.SignedURL(ctx, storagePath, signedURLOpts)
	if err != nil {
		log.Error().
			Err(err).
			Str("storage_path", storagePath).
			Msg("Failed to generate download signed URL")
		
		return nil, connect.NewError(
			connect.CodeInternal,
			fmt.Errorf("failed to generate download URL: %w", err),
		)
	}

	log.Info().
		Str("storage_path", storagePath).
		Str("expires_at", expiresAt.Format(time.RFC3339)).
		Msg("Download URL generated successfully")

	response := &pb.GenerateDownloadURLResponse{
		DownloadUrl: downloadURL,
		StoragePath: storagePath,
		ExpiresAt:   timestamppb.New(expiresAt),
	}

	return connect.NewResponse(response), nil
}

// Close closes the storage service and cleans up resources
func (s *StorageServiceServer) Close() error {
	var err error
	
	if s.storage != nil {
		if closeErr := s.storage.Close(); closeErr != nil {
			err = closeErr
		}
	}
	
	if err != nil {
		log.Error().Err(err).Msg("Error closing storage service")
	} else {
		log.Info().Msg("Storage service closed successfully")
	}
	
	return err
}