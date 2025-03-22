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
	userPayload := middleware.FromAdminContext(ctx).UserPayload

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	user, err := models.Users(
		models.UserWhere.ID.EQ(userPayload.UserID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		return nil, pbErrors.InternalError("failed to get user", err)
	}

	if user.Role != models.RoleEnumUser {
		return nil, pbErrors.PermissionDeniedError("only admin can create art")
	}

	artDb := &models.Art{
		Title:    req.GetArt().GetTitle(),
		AuthorID: user.ID,
	}

	err = artDb.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, pbErrors.InternalError("failed to insert art", err)
	}

	return pbx.ArtDbToProto(ctx, server.bucket, artDb), nil
}

func (server *Server) UpdateArt(ctx context.Context, req *pb.UpdateArtRequest) (*pb.Art, error) {
	userPayload := middleware.FromAdminContext(ctx).UserPayload

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

	if authorId != userPayload.UserID {
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
	userPayload := middleware.FromAdminContext(ctx).UserPayload

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	pageSize := int(req.GetPageSize())

	const (
		maxPageSize      = 1000
		defaultPageSize  = 100
		defaultPageToken = 0
	)

	switch {
	case pageSize < 0:
		return nil, status.Errorf(codes.InvalidArgument, "page size is negative")
	case pageSize == 0:
		pageSize = defaultPageSize
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	}

	if req.GetPageToken() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "page token is negative")
	} else if req.GetPageToken() == 0 {
		req.PageToken = defaultPageToken
	}

	// Query the arts with pagination
	arts, err := models.Arts(
		models.ArtWhere.AuthorID.EQ(userPayload.UserID),
		qm.Limit(pageSize),
		qm.Offset(int(req.PageToken)),
	).All(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to get arts", err)
	}

	// Convert the arts to protobuf format
	artPbs := make([]*pb.Art, 0, len(arts))
	for _, artDb := range arts {
		artPbs = append(artPbs, pbx.ArtDbToProto(ctx, server.bucket, artDb))
	}

	return &pb.ListArtsResponse{
		Arts:          artPbs,
		NextPageToken: req.PageToken + int32(pageSize),
	}, nil
}

func (server *Server) GetArt(ctx context.Context, req *pb.GetArtRequest) (*pb.Art, error) {
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

	if authorId != middleware.FromAdminContext(ctx).UserPayload.UserID {
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

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	if authorId != middleware.FromAdminContext(ctx).UserPayload.UserID {
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

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	userPayload := middleware.FromAdminContext(ctx).UserPayload
	if authorId != userPayload.UserID {
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
