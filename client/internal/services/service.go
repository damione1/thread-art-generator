package services

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/transport"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"
)

// GeneratorService is the main service container that provides access to all domain services
// while maintaining the same interface as the original monolithic service
type GeneratorService struct {
	// Domain services
	UserService        *UserService
	ArtService         *ArtService
	CompositionService *CompositionService
}

// ArtFormData represents the form data for creating art
type ArtFormData struct {
	Title   string
	Errors  map[string][]string
	Success bool
}

// UserProfile represents user information within the application
type UserProfile struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Avatar    string
}

// NewGeneratorService creates a new generator service with all domain services
func NewGeneratorService(client pbconnect.ArtGeneratorServiceClient, sessionManager *auth.SCSSessionManager) *GeneratorService {
	// Create shared base service
	baseService := NewBaseService(client, sessionManager)

	return &GeneratorService{
		UserService:        NewUserService(baseService),
		ArtService:         NewArtService(baseService),
		CompositionService: NewCompositionService(baseService),
	}
}

// User domain methods - delegate to UserService
func (s *GeneratorService) GetCurrentUser(ctx context.Context, r *http.Request) (*User, error) {
	return s.UserService.GetCurrentUser(ctx, r)
}

// Art domain methods - delegate to ArtService
func (s *GeneratorService) CreateArt(ctx context.Context, r *http.Request, userID string, title string) (*pb.Art, map[string][]string, error) {
	return s.ArtService.CreateArt(ctx, r, userID, title)
}

func (s *GeneratorService) GetArt(ctx context.Context, r *http.Request, userID, artID string) (*pb.Art, error) {
	return s.ArtService.GetArt(ctx, r, userID, artID)
}

func (s *GeneratorService) ListArts(ctx context.Context, r *http.Request, userID string, pageSize int, pageToken string, orderBy, orderDirection string) (*pb.ListArtsResponse, error) {
	return s.ArtService.ListArts(ctx, r, userID, pageSize, pageToken, orderBy, orderDirection)
}

// Composition domain methods - delegate to CompositionService
func (s *GeneratorService) ListCompositions(ctx context.Context, r *http.Request, pageSize int, pageToken string) (*pb.ListCompositionsResponse, error) {
	return s.CompositionService.ListCompositions(ctx, r, pageSize, pageToken)
}

func (s *GeneratorService) GetComposition(ctx context.Context, r *http.Request, compositionID string) (*pb.Composition, error) {
	return s.CompositionService.GetComposition(ctx, r, compositionID)
}

func (s *GeneratorService) CreateComposition(ctx context.Context, r *http.Request, userID, artID string, composition *pb.Composition) (*pb.Composition, map[string][]string, error) {
	return s.CompositionService.CreateComposition(ctx, r, userID, artID, composition)
}

func (s *GeneratorService) DeleteComposition(ctx context.Context, r *http.Request, compositionName string) error {
	return s.CompositionService.DeleteComposition(ctx, r, compositionName)
}

// Storage domain methods - use the same client and pattern as other services  
func (s *GeneratorService) GenerateUploadURL(ctx context.Context, r *http.Request, userID string, req StorageUploadURLRequest) (*StorageUploadURLResponse, error) {
	// Add session to context for authentication (same pattern as other methods)
	ctxWithSession := ctx
	if r != nil {
		ctxWithSession = transport.WithSessionRequest(ctx, r)
	}

	// Create the Connect request using the storage proto messages
	rpcRequest := connect.NewRequest(&pb.GenerateUploadURLRequest{
		UserId:      userID,
		ArtId:       req.ArtID,
		ContentType: req.ContentType,
		FileSize:    req.FileSize,
	})

	// Set filename if provided
	if req.Filename != "" {
		rpcRequest.Msg.Filename = &req.Filename
	}

	// Call the API - the transport layer will automatically add authentication
	baseService := s.UserService.BaseService // Access the shared BaseService
	response, err := baseService.client.GenerateUploadURL(ctxWithSession, rpcRequest)
	if err != nil {
		return nil, err
	}

	// Convert the response
	return &StorageUploadURLResponse{
		UploadURL:     response.Msg.UploadUrl,
		StoragePath:   response.Msg.StoragePath,
		ExpiresAt:     response.Msg.ExpiresAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		MaxFileSize:   response.Msg.MaxFileSize,
	}, nil
}

func (s *GeneratorService) GenerateDownloadURL(ctx context.Context, r *http.Request, userID string, req StorageDownloadURLRequest) (*StorageDownloadURLResponse, error) {
	// Add session to context for authentication (same pattern as other methods)
	ctxWithSession := ctx
	if r != nil {
		ctxWithSession = transport.WithSessionRequest(ctx, r)
	}

	// Create the Connect request
	rpcRequest := connect.NewRequest(&pb.GenerateDownloadURLRequest{
		UserId: userID,
		ArtId:  req.ArtID,
	})

	// Set file path if provided
	if req.FilePath != nil {
		rpcRequest.Msg.FilePath = req.FilePath
	}

	// Call the API - the transport layer will automatically add authentication
	baseService := s.UserService.BaseService // Access the shared BaseService
	response, err := baseService.client.GenerateDownloadURL(ctxWithSession, rpcRequest)
	if err != nil {
		return nil, err
	}

	// Convert the response
	return &StorageDownloadURLResponse{
		DownloadURL: response.Msg.DownloadUrl,
		StoragePath: response.Msg.StoragePath,
		ExpiresAt:   response.Msg.ExpiresAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// Storage request/response types - use unique names to avoid conflicts
type StorageUploadURLRequest struct {
	ArtID       string
	ContentType string
	FileSize    int64
	Filename    string
}

type StorageUploadURLResponse struct {
	UploadURL     string
	StoragePath   string
	ExpiresAt     string
	MaxFileSize   int64
}

type StorageDownloadURLRequest struct {
	ArtID    string
	FilePath *string
}

type StorageDownloadURLResponse struct {
	DownloadURL string
	StoragePath string
	ExpiresAt   string
}