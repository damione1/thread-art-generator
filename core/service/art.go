package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
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

func (server *Server) CreateArt(ctx context.Context, req *pb.CreateArtRequest) (*pb.Art, error) {
	// Get user ID from context using the same key used in auth interceptor
	log.Info().Msgf("CreateArt: %s", req)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}
	log.Info().Msgf("CreateArt userID: %s", userID)
	if err := protovalidate.Validate(req); err != nil {
		log.Info().Msgf("CreateArt protovalidate: %s", err)
		return nil, pbErrors.ConvertProtoValidateError(err)
	}
	log.Info().Msgf("CreateArt protovalidate: %s", req)
	user, err := models.Users(
		models.UserWhere.ID.EQ(userID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Msgf("CreateArt user not found")
			return nil, pbErrors.NotFoundError("user not found")
		}
		log.Info().Msgf("CreateArt failed to get user: %s", err)
		return nil, pbErrors.InternalError("failed to get user", err)
	}
	log.Info().Msgf("CreateArt user: %s", user)
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

	authorId, artId, err := pbx.ParseArtResourceName(req.GetArt().GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("art.name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if authorId != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can update the art")
	}

	// Check if the art exists
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
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
	// Get user ID from context using the same key used in auth interceptor
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
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

	// Query the arts with pagination
	arts, err := models.Arts(
		models.ArtWhere.AuthorID.EQ(userID),
		qm.Limit(pageSize+1), // Query one more than we need to check if there are more results
		qm.Offset(offset),
	).All(ctx, server.config.DB)
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

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if authorId != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can get the art")
	}

	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
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

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if authorId != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can delete the art")
	}

	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
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
		imageKey := pbx.GetResourceName([]pbx.Resource{
			{Type: pbx.RessourceTypeUsers, ID: artDb.AuthorID},
			{Type: pbx.RessourceTypeArts, ID: artDb.ImageID.String},
		})
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

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if authorId != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can get an upload URL for the art")
	}

	// Check if the art exists
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
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

	// Create the image key in the format "users/{userId}/arts/{imageId}"
	imageKey := pbx.GetResourceName([]pbx.Resource{
		{Type: pbx.RessourceTypeUsers, ID: artDb.AuthorID},
		{Type: pbx.RessourceTypeArts, ID: imageID},
	})

	// Generate a signed URL with 15 minutes expiration
	opts := &blob.SignedURLOptions{
		Expiry:      15 * time.Minute,
		Method:      "PUT",
		ContentType: "image/jpeg",
	}

	signedURL, err := server.bucket.SignedURL(ctx, imageKey, opts)
	if err != nil {
		return nil, pbErrors.InternalError("failed to generate signed URL", err)
	}

	// Calculate expiration time
	expirationTime := time.Now().Add(15 * time.Minute)

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

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if authorId != userID {
		return nil, pbErrors.PermissionDeniedError("only the author can confirm image upload")
	}

	// Get the art
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
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

	// Verify the image exists in the bucket
	imageKey := pbx.GetResourceName([]pbx.Resource{
		{Type: pbx.RessourceTypeUsers, ID: artDb.AuthorID},
		{Type: pbx.RessourceTypeArts, ID: artDb.ImageID.String},
	})

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
