package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
	"github.com/Damione1/thread-art-generator/core/queue"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/bufbuild/protovalidate-go"
	"github.com/friendsofgo/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateComposition creates a new composition for an art
func (server *Server) CreateComposition(ctx context.Context, req *pb.CreateCompositionRequest) (*pb.Composition, error) {
	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Validate the request
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Parse the art resource name to get the art ID
	artResource, err := resource.ParseResourceName(req.GetParent())
	if err != nil {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("parent", errors.New("invalid resource name")),
		})
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("parent", errors.New("invalid art resource name")),
		})
	}

	// Verify the user is authorized to create a composition for this art
	if art.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can create compositions for this art")
	}

	// Get the art
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(art.UserID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("art not found")
		}
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	// Check if the art has an image
	if !artDb.ImageID.Valid {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("parent", errors.New("art must have an image to create compositions")),
		})
	}

	// Convert proto to database model
	compositionDb := &models.Composition{
		ID:                uuid.New().String(),
		ArtID:             art.ArtID,
		Status:            models.CompositionStatusEnumPENDING,
		NailsQuantity:     int(req.GetComposition().GetNailsQuantity()),
		ImgSize:           int(req.GetComposition().GetImgSize()),
		MaxPaths:          int(req.GetComposition().GetMaxPaths()),
		StartingNail:      int(req.GetComposition().GetStartingNail()),
		MinimumDifference: int(req.GetComposition().GetMinimumDifference()),
		BrightnessFactor:  int(req.GetComposition().GetBrightnessFactor()),
		ImageContrast:     float64(req.GetComposition().GetImageContrast()),
		PhysicalRadius:    float64(req.GetComposition().GetPhysicalRadius()),
	}

	// Insert the composition
	err = compositionDb.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, pbErrors.InternalError("failed to insert composition", err)
	}

	// Enqueue the composition for processing
	err = server.enqueueCompositionForProcessing(ctx, compositionDb, artDb)
	if err != nil {
		log.Error().Err(err).Str("compositionID", compositionDb.ID).Msg("Failed to enqueue composition for processing")
		// We don't return an error here, as the composition is created and can be requeued later
	}

	// Return the created composition
	return pbx.CompositionDbToProto(ctx, server.bucket, artDb, compositionDb), nil
}

// GetComposition retrieves a composition by ID
func (server *Server) GetComposition(ctx context.Context, req *pb.GetCompositionRequest) (*pb.Composition, error) {
	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Validate the request
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Parse the composition resource name
	compositionResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		})
	}

	composition, ok := compositionResource.(*resource.Composition)
	if !ok {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid composition resource name")),
		})
	}

	// Verify the user is authorized to get this composition
	if composition.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can get this composition")
	}

	// Get the art
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(composition.ArtID),
		models.ArtWhere.AuthorID.EQ(composition.UserID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("art not found")
		}
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	// Get the composition
	compositionDb, err := models.Compositions(
		models.CompositionWhere.ID.EQ(composition.CompositionID),
		models.CompositionWhere.ArtID.EQ(composition.ArtID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("composition not found")
		}
		return nil, pbErrors.InternalError("failed to get composition", err)
	}

	// Return the composition
	return pbx.CompositionDbToProto(ctx, server.bucket, artDb, compositionDb), nil
}

// UpdateComposition updates an existing composition
func (server *Server) UpdateComposition(ctx context.Context, req *pb.UpdateCompositionRequest) (*pb.Composition, error) {
	// Since compositions are processed asynchronously and their config shouldn't change
	// after they're created, we don't allow updates to compositions for now.
	return nil, status.Error(codes.Unimplemented, "updating compositions is not supported")
}

// ListCompositions lists all compositions for an art
func (server *Server) ListCompositions(ctx context.Context, req *pb.ListCompositionsRequest) (*pb.ListCompositionsResponse, error) {
	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Validate the request
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Parse the art resource name
	artResource, err := resource.ParseResourceName(req.GetParent())
	if err != nil {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("parent", errors.New("invalid resource name")),
		})
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("parent", errors.New("invalid art resource name")),
		})
	}

	// Verify the user is authorized to list compositions for this art
	if art.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can list compositions for this art")
	}

	// Get the art
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(art.UserID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("art not found")
		}
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	// Set default page size if not specified
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse page token if provided
	offset := 0
	if req.GetPageToken() != "" {
		var err error
		offset, err = parseInt32PageToken(req.GetPageToken())
		if err != nil {
			return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("page_token", err),
			})
		}
	}

	// Query compositions
	queryMods := []qm.QueryMod{
		models.CompositionWhere.ArtID.EQ(art.ArtID),
		qm.OrderBy(fmt.Sprintf("%s DESC", models.CompositionColumns.CreatedAt)), // Latest first
		qm.Limit(pageSize + 1), // +1 to check if there are more
		qm.Offset(offset),
	}

	compositions, err := models.Compositions(queryMods...).All(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to list compositions", err)
	}

	// Determine if there are more results
	hasNextPage := false
	if len(compositions) > pageSize {
		hasNextPage = true
		compositions = compositions[:pageSize]
	}

	// Convert to proto
	var protoCompositions []*pb.Composition
	for _, comp := range compositions {
		protoCompositions = append(protoCompositions, pbx.CompositionDbToProto(ctx, server.bucket, artDb, comp))
	}

	// Create response
	response := &pb.ListCompositionsResponse{
		Compositions: protoCompositions,
	}

	// Set next page token if there are more results
	if hasNextPage {
		response.NextPageToken = createPageToken(offset + pageSize)
	}

	return response, nil
}

// DeleteComposition deletes a composition
func (server *Server) DeleteComposition(ctx context.Context, req *pb.DeleteCompositionRequest) (*emptypb.Empty, error) {
	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Validate the request
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Parse the composition resource name
	compositionResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		})
	}

	composition, ok := compositionResource.(*resource.Composition)
	if !ok {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid composition resource name")),
		})
	}

	// Verify the user is authorized to delete this composition
	if composition.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can delete this composition")
	}

	// Get the composition
	compositionDb, err := models.Compositions(
		models.CompositionWhere.ID.EQ(composition.CompositionID),
		models.CompositionWhere.ArtID.EQ(composition.ArtID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("composition not found")
		}
		return nil, pbErrors.InternalError("failed to get composition", err)
	}

	// Delete the composition
	_, err = compositionDb.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to delete composition", err)
	}

	// Delete associated files from storage if they exist
	if compositionDb.PreviewURL.Valid {
		err = server.bucket.Delete(ctx, compositionDb.PreviewURL.String)
		if err != nil {
			log.Error().Err(err).Str("key", compositionDb.PreviewURL.String).Msg("Failed to delete preview file")
		}
	}

	if compositionDb.GcodeURL.Valid {
		err = server.bucket.Delete(ctx, compositionDb.GcodeURL.String)
		if err != nil {
			log.Error().Err(err).Str("key", compositionDb.GcodeURL.String).Msg("Failed to delete gcode file")
		}
	}

	if compositionDb.PathlistURL.Valid {
		err = server.bucket.Delete(ctx, compositionDb.PathlistURL.String)
		if err != nil {
			log.Error().Err(err).Str("key", compositionDb.PathlistURL.String).Msg("Failed to delete pathlist file")
		}
	}

	return &emptypb.Empty{}, nil
}

// Helper function to enqueue a composition for processing
func (server *Server) enqueueCompositionForProcessing(ctx context.Context, composition *models.Composition, art *models.Art) error {
	// Check if queue client is initialized
	if server.queueClient == nil {
		return fmt.Errorf("queue client not initialized")
	}

	// Create a queue message
	message := queue.NewCompositionProcessingMessage(art.ID, composition.ID)

	// Convert to JSON
	jsonData, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize composition processing message: %w", err)
	}

	// Get queue name from config
	queueName := server.config.Queue.CompositionProcessing
	if queueName == "" {
		queueName = "composition-processing" // Default queue name
	}

	// Publish to queue
	err = server.queueClient.PublishMessage(ctx, queueName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to publish composition to queue: %w", err)
	}

	log.Info().
		Str("compositionID", composition.ID).
		Str("artID", art.ID).
		Str("queue", queueName).
		Msg("Composition enqueued for processing")

	return nil
}
