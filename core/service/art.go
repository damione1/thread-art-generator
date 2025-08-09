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
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/bufbuild/protovalidate-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

	// Get user from database - user should already exist from auth sync
	user, err := models.Users(
		models.UserWhere.FirebaseUID.EQ(null.StringFrom(firebaseUID)),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("CreateArt failed to get user - user should have been created during auth sync")
		return nil, pbErrors.InternalError("failed to get user", err)
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

	return pbx.ArtDbToProto(ctx, server.GetStorage(), artDb, firebaseUID), nil
}

func (server *Server) UpdateArt(ctx context.Context, req *pb.UpdateArtRequest) (*pb.Art, error) {
	// Get Firebase UID from context
	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Get internal user from Firebase UID
	user, err := models.Users(
		models.UserWhere.FirebaseUID.EQ(null.StringFrom(firebaseUID)),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("UpdateArt: Failed to get user from Firebase UID")
		return nil, pbErrors.InternalError("failed to get user", err)
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

	// Compare Firebase UID with art's user ID from resource name (now using Firebase UID)
	if art.UserID != firebaseUID {
		return nil, pbErrors.PermissionDeniedError("only the author can update the art")
	}

	// Check if the art exists
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(user.ID),
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

	return pbx.ArtDbToProto(ctx, server.GetStorage(), artDb, firebaseUID), nil
}

func (server *Server) ListArts(ctx context.Context, req *pb.ListArtsRequest) (*pb.ListArtsResponse, error) {
	log.Debug().Msg("ListArts: Starting request processing")
	
	// Get Firebase UID from context
	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		log.Debug().Msg("ListArts: No Firebase UID in context")
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}
	log.Debug().Str("firebase_uid", firebaseUID).Msg("ListArts: Retrieved Firebase UID from context")

	// Get user from database - user should already exist from auth sync
	log.Debug().Str("firebase_uid", firebaseUID).Msg("ListArts: About to query user by Firebase UID")
	user, err := models.Users(
		models.UserWhere.FirebaseUID.EQ(null.StringFrom(firebaseUID)),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("ListArts failed to get user - user should have been created during auth sync")
		return nil, pbErrors.InternalError("failed to get user", err)
	}
	log.Debug().Str("user_id", user.ID).Str("firebase_uid", firebaseUID).Msg("ListArts: Successfully retrieved user from database")

	log.Debug().Msg("ListArts: About to validate request")
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}
	log.Debug().Msg("ListArts: Request validation successful")

	pageSize := int(req.GetPageSize())
	log.Debug().Int("page_size", pageSize).Msg("ListArts: Retrieved page size from request")

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
	log.Debug().
		Str("user_id", user.ID).
		Int("page_size", pageSize).
		Int("offset", offset).
		Str("order_by", orderColumn).
		Str("direction", dir).
		Msg("ListArts: About to query arts from database")

	// Query the arts with pagination and sorting
	arts, err := models.Arts(queryMods...).All(ctx, server.config.DB)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("ListArts: Database query for arts failed")
		return nil, pbErrors.InternalError("failed to get arts", err)
	}
	log.Debug().Int("arts_count", len(arts)).Str("user_id", user.ID).Msg("ListArts: Successfully retrieved arts from database")

	// Check if there are more results
	hasNextPage := false
	if len(arts) > pageSize {
		hasNextPage = true
		arts = arts[:pageSize] // Trim the extra result
	}

	log.Debug().Int("arts_to_convert", len(arts)).Msg("ListArts: About to convert arts to protobuf format")
	
	// Convert the arts to protobuf format
	artPbs := make([]*pb.Art, 0, len(arts))
	for i, artDb := range arts {
		log.Debug().Int("art_index", i).Str("art_id", artDb.ID).Msg("ListArts: Converting art to protobuf")
		artPbs = append(artPbs, pbx.ArtDbToProto(ctx, server.GetStorage(), artDb, firebaseUID))
	}
	log.Debug().Int("converted_count", len(artPbs)).Msg("ListArts: Successfully converted all arts to protobuf")

	// Create next page token if there are more results
	nextPageToken := ""
	if hasNextPage {
		nextPageToken = createPageToken(offset + pageSize)
		log.Debug().Str("next_page_token", nextPageToken).Msg("ListArts: Created next page token")
	}

	log.Debug().Int("arts_count", len(artPbs)).Bool("has_next_page", hasNextPage).Msg("ListArts: About to return response")
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

	// Get Firebase UID from context
	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Get internal user from Firebase UID
	user, err := models.Users(
		models.UserWhere.FirebaseUID.EQ(null.StringFrom(firebaseUID)),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("GetArt: Failed to get user from Firebase UID")
		return nil, pbErrors.InternalError("failed to get user", err)
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

	// Compare Firebase UID with art's user ID from resource name (now using Firebase UID)
	if art.UserID != firebaseUID {
		return nil, pbErrors.PermissionDeniedError("only the author can get the art")
	}

	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(user.ID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("art not found")
		}
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	return pbx.ArtDbToProto(ctx, server.GetStorage(), artDb, firebaseUID), nil
}

func (server *Server) DeleteArt(ctx context.Context, req *pb.DeleteArtRequest) (*emptypb.Empty, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Get Firebase UID from context
	firebaseUID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, pbErrors.PermissionDeniedError("user not authenticated")
	}

	// Get internal user from Firebase UID
	user, err := models.Users(
		models.UserWhere.FirebaseUID.EQ(null.StringFrom(firebaseUID)),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("user not found")
		}
		log.Error().Err(err).Str("firebase_uid", firebaseUID).Msg("DeleteArt: Failed to get user from Firebase UID")
		return nil, pbErrors.InternalError("failed to get user", err)
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

	// Compare Firebase UID with art's user ID from resource name (now using Firebase UID)
	if art.UserID != firebaseUID {
		return nil, pbErrors.PermissionDeniedError("only the author can delete the art")
	}

	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(user.ID),
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
		err = server.GetStorage().Delete(ctx, imageKey)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Failed to delete image %s", artDb.ImageID.String))
			return &emptypb.Empty{}, nil // Don't return a public error if the image deletion fails
		}
	}

	return &emptypb.Empty{}, nil
}