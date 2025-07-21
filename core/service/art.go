package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/bufbuild/protovalidate-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gocloud.dev/blob"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Context keys for Firebase claims (should match interceptors/auth.go)
type contextKey string

const (
	contextKeyClaims contextKey = "firebase_claims"
)

func (server *Server) CreateArt(ctx context.Context, req *pb.CreateArtRequest) (*pb.Art, error) {
	// Get Firebase UID from context
	log.Info().Msgf("CreateArt: %s", req)

	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}
	log.Info().Msgf("CreateArt Firebase UID: %s", firebaseUID)
	
	if err := protovalidate.Validate(req); err != nil {
		log.Info().Msgf("CreateArt protovalidate: %s", err)
		return nil, pbErrors.ConvertProtoValidateError(err)
	}
	log.Info().Msgf("CreateArt protovalidate: %s", req)
	
	// JIT User Provisioning: Try to get user by Firebase UID, create if not exists
	user, err := server.getUserFromFirebaseUID(ctx, firebaseUID)
	if err != nil {
		// Check if it's a "not found" error by looking at the error message
		if err.Error() == "user not found" || errors.Is(err, sql.ErrNoRows) {
			log.Info().Msgf("CreateArt: User not found, creating new user for Firebase UID: %s", firebaseUID)
			
			// Get Firebase claims from context for user creation
			if claims, ok := ctx.Value(contextKeyClaims).(*auth.AuthClaims); ok {
				// Create new user with Firebase claims
				user, err = server.createUserFromFirebaseClaims(ctx, firebaseUID, claims.Email, claims.Name, claims.Picture)
				if err != nil {
					log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("Failed to create user from Firebase claims")
					return nil, pbErrors.InternalError("failed to create user", err)
				}
				log.Info().Str("user_id", user.ID).Str("firebase_uid", firebaseUID).Msg("Created new user from Firebase claims")
			} else {
				log.Error().Str("firebase_uid", firebaseUID).Msg("Firebase claims not found in context")
				return nil, pbErrors.InternalError("firebase claims not available", errors.New("firebase claims missing"))
			}
		} else {
			log.Info().Msgf("CreateArt failed to get user: %s", err)
			return nil, pbErrors.InternalError("failed to get user", err)
		}
	}
	log.Info().Str("user_id", user.ID).Str("firebase_uid", firebaseUID).Str("email", user.Email.String).Msg("CreateArt found/created user")
	if user.Role != models.RoleEnumUser {
		log.Info().Msgf("CreateArt user is not a user")
		return nil, pbErrors.PermissionDeniedError("only users can create art")
	}
	log.Info().Msgf("CreateArt user is a user")
	artDb := &models.Art{
		Title:    req.GetArt().GetTitle(),
		AuthorID: user.ID,
		Status:   models.ArtStatusEnumPENDING_IMAGE, // Set initial status as pending image
	}

	err = artDb.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, pbErrors.InternalError("failed to insert art", err)
	}

	return pbx.ArtDbToProto(ctx, server.bucket, artDb), nil
}

func (server *Server) UpdateArt(ctx context.Context, req *pb.UpdateArtRequest) (*pb.Art, error) {
	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	artResource, err := resource.ParseResourceName(req.GetArt().GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("art.name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("art.name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if art.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can update the art")
	}

	// Check if the art exists
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

	if req.GetArt().GetTitle() != "" {
		artDb.Title = req.GetArt().GetTitle()
	}

	_, err = artDb.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	return pbx.ArtDbToProto(ctx, server.bucket, artDb), nil
}

func (server *Server) ListArts(ctx context.Context, req *pb.ListArtsRequest) (*pb.ListArtsResponse, error) {
	// Get Firebase UID from context
	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Look up the internal user using the Firebase UID
	user, err := server.getUserFromFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	pageSize := int(req.GetPageSize())

	const (
		maxPageSize     = 1000
		defaultPageSize = 100
	)

	switch {
	case pageSize < 0:
		return nil, status.Errorf(codes.InvalidArgument, "page size is negative")
	case pageSize == 0:
		pageSize = defaultPageSize
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}

	// Parse page token to get offset
	offset := 0
	if req.GetPageToken() != "" {
		var err error
		offset, err = parseInt32PageToken(req.GetPageToken())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token: %v", err)
		}
	}

	// Determine order_by and order_direction
	orderBy := req.GetOrderBy()
	if orderBy == "" {
		orderBy = "create_time"
	}
	orderDirection := req.GetOrderDirection()
	if orderDirection == "" {
		orderDirection = "desc"
	}

	// Map proto field to DB column
	var orderColumn string
	switch orderBy {
	case "create_time":
		orderColumn = models.ArtColumns.CreatedAt
	case "update_time":
		orderColumn = models.ArtColumns.UpdatedAt
	default:
		orderColumn = models.ArtColumns.CreatedAt
	}

	// Validate direction
	dir := "DESC"
	if orderDirection == "asc" {
		dir = "ASC"
	}

	// Build query mods using internal user ID
	queryMods := []qm.QueryMod{
		models.ArtWhere.AuthorID.EQ(user.ID),
		qm.OrderBy(fmt.Sprintf("%s %s", orderColumn, dir)),
		qm.Limit(pageSize + 1),
		qm.Offset(offset),
	}

	// Query the arts with pagination and sorting
	arts, err := models.Arts(queryMods...).All(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to get arts", err)
	}

	// Check if there are more results
	hasNextPage := false
	if len(arts) > pageSize {
		hasNextPage = true
		arts = arts[:pageSize] // Trim the extra result
	}

	// Convert the arts to protobuf format
	artPbs := make([]*pb.Art, 0, len(arts))
	for _, artDb := range arts {
		artPbs = append(artPbs, pbx.ArtDbToProto(ctx, server.bucket, artDb))
	}

	// Create next page token if there are more results
	nextPageToken := ""
	if hasNextPage {
		nextPageToken = createPageToken(offset + pageSize)
	}

	return &pb.ListArtsResponse{
		Arts:          artPbs,
		NextPageToken: nextPageToken,
	}, nil
}

// parseInt32PageToken converts a string page token to an integer offset
func parseInt32PageToken(token string) (int, error) {
	// For simplicity, we're just converting the string to int
	// In a production system, you might want to use a more secure approach
	// such as signed or encrypted tokens
	var offset int
	_, err := fmt.Sscanf(token, "%d", &offset)
	if err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, fmt.Errorf("offset cannot be negative")
	}
	return offset, nil
}

// createPageToken creates a page token from an integer offset
func createPageToken(offset int) string {
	// For simplicity, we're just converting the int to string
	// In a production system, you might want to use a more secure approach
	return fmt.Sprintf("%d", offset)
}

func (server *Server) GetArt(ctx context.Context, req *pb.GetArtRequest) (*pb.Art, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	artResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if art.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can get the art")
	}

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

	return pbx.ArtDbToProto(ctx, server.bucket, artDb), nil
}

func (server *Server) DeleteArt(ctx context.Context, req *pb.DeleteArtRequest) (*emptypb.Empty, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	artResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if art.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can delete the art")
	}

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

	_, err = artDb.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to delete art", err)
	}

	// Delete the image from the bucket
	if artDb.ImageID.Valid {
		imageKey := resource.BuildArtResourceName(artDb.AuthorID, artDb.ImageID.String)
		err = server.bucket.Delete(ctx, imageKey)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Failed to delete image %s", artDb.ImageID.String))
			return &emptypb.Empty{}, nil // Don't return a public error if the image deletion fails
		}
	}

	return &emptypb.Empty{}, nil
}

// GetArtUploadUrl generates a signed URL for uploading an image for a specific art
func (server *Server) GetArtUploadUrl(ctx context.Context, req *pb.GetArtUploadUrlRequest) (*pb.GetArtUploadUrlResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	artResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if art.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can get an upload URL for the art")
	}

	// Check if the art exists
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

	// Generate a unique image ID if not exists
	imageID := artDb.ImageID.String
	if !artDb.ImageID.Valid {
		imageID = uuid.New().String()

		// Update the art with the new image ID
		artDb.ImageID = null.StringFrom(imageID)

		_, err = artDb.Update(ctx, server.config.DB, boil.Whitelist(models.ArtColumns.ImageID))
		if err != nil {
			return nil, pbErrors.InternalError("failed to update art with image ID", err)
		}
	}

	// Create the image key using resource builder
	imageKey := resource.BuildArtResourceName(artDb.AuthorID, imageID)

	// Generate a signed URL with 15 minutes expiration
	// Important: We're using a minimal set of options to keep the signing simple
	opts := &blob.SignedURLOptions{
		Expiry: 15 * time.Minute,
		Method: "PUT",
		// We intentionally DO NOT set ContentType to avoid signature issues
	}

	signedURL, err := server.bucket.SignedURL(ctx, imageKey, opts)
	if err != nil {
		return nil, pbErrors.InternalError("failed to generate signed URL", err)
	}

	// Calculate expiration time
	expirationTime := time.Now().Add(15 * time.Minute)

	log.Info().Msgf("Signed URL: %s", signedURL)

	return &pb.GetArtUploadUrlResponse{
		UploadUrl:      signedURL,
		ExpirationTime: timestamppb.New(expirationTime),
	}, nil
}

// ConfirmArtImageUpload marks an art as complete after successful image upload
func (server *Server) ConfirmArtImageUpload(ctx context.Context, req *pb.ConfirmArtImageUploadRequest) (*pb.Art, error) {
	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	artResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if art.UserID != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can confirm image upload")
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

	// Make sure there's an image ID
	if !artDb.ImageID.Valid {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("art has no image ID, request upload URL first")),
		})
	}

	// Verify the image exists in the bucket using resource builder
	imageKey := resource.BuildArtResourceName(artDb.AuthorID, artDb.ImageID.String)

	exists, err := server.bucket.Exists(ctx, imageKey)
	if err != nil {
		return nil, pbErrors.InternalError("failed to verify image exists", err)
	}

	if !exists {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("image not found in storage, upload the image first")),
		})
	}

	// Update status to complete
	artDb.Status = models.ArtStatusEnumCOMPLETE
	_, err = artDb.Update(ctx, server.config.DB, boil.Whitelist(models.ArtColumns.Status))
	if err != nil {
		return nil, pbErrors.InternalError("failed to update art status", err)
	}

	return pbx.ArtDbToProto(ctx, server.bucket, artDb), nil
}
