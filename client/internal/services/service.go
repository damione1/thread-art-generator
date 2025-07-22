package services

import (
	"context"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/client"
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
func (s *GeneratorService) GetCurrentUser(ctx context.Context, r *http.Request) (*client.User, error) {
	return s.UserService.GetCurrentUser(ctx, r)
}

// Art domain methods - delegate to ArtService
func (s *GeneratorService) CreateArt(ctx context.Context, createArtRequest *pb.CreateArtRequest) (*pb.Art, map[string][]string, error) {
	return s.ArtService.CreateArt(ctx, createArtRequest)
}

func (s *GeneratorService) GetArt(ctx context.Context, userID, artID string) (*pb.Art, error) {
	return s.ArtService.GetArt(ctx, userID, artID)
}

func (s *GeneratorService) GetArtUploadUrl(ctx context.Context, userID, artID, contentType string, fileSize int64) (*pb.GetArtUploadUrlResponse, error) {
	return s.ArtService.GetArtUploadUrl(ctx, userID, artID, contentType, fileSize)
}

func (s *GeneratorService) ConfirmArtImageUpload(ctx context.Context, artName string) (*pb.Art, error) {
	return s.ArtService.ConfirmArtImageUpload(ctx, artName)
}

func (s *GeneratorService) ListArts(ctx context.Context, userID string, pageSize int, pageToken string, orderBy, orderDirection string) (*pb.ListArtsResponse, error) {
	return s.ArtService.ListArts(ctx, userID, pageSize, pageToken, orderBy, orderDirection)
}

// Composition domain methods - delegate to CompositionService
func (s *GeneratorService) ListCompositions(ctx context.Context, pageSize int, pageToken string) (*pb.ListCompositionsResponse, error) {
	return s.CompositionService.ListCompositions(ctx, pageSize, pageToken)
}

func (s *GeneratorService) CreateComposition(ctx context.Context, createRequest *pb.CreateCompositionRequest) (*pb.Composition, map[string][]string, error) {
	return s.CompositionService.CreateComposition(ctx, createRequest)
}

func (s *GeneratorService) GetComposition(ctx context.Context, userID, artID, compositionID string) (*pb.Composition, error) {
	return s.CompositionService.GetComposition(ctx, userID, artID, compositionID)
}
