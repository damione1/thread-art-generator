package service

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) CreateArt(ctx context.Context, req *pb.CreateArtRequest) (*pb.Art, error) {
	userPayload := middleware.FromAdminContext(ctx).UserPayload

	if err := validateCreateArtRequest(req); err != nil {
		return nil, err
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

func validateCreateArtRequest(req *pb.CreateArtRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Art,
			validation.Required,
			validation.By(
				func(value interface{}) error {
					return validateArt(value.(*pb.Art), true)
				},
			),
		),
	)
}

func (server *Server) UpdateArt(ctx context.Context, req *pb.UpdateArtRequest) (*pb.Art, error) {
	userPayload := middleware.FromAdminContext(ctx).UserPayload

	if err := validateUpdateArtRequest(req); err != nil {
		return nil, err
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

	updateMask := req.GetUpdateMask()
	if updateMask != nil && len(updateMask.GetPaths()) > 0 {
		for _, path := range updateMask.GetPaths() {
			switch path {
			case "title":
				if req.GetArt().GetTitle() != "" {
					artDb.Title = req.GetArt().GetTitle()
				}
			default:
				return nil, status.Errorf(codes.InvalidArgument, "Invalid field mask: %s", path)
			}
		}
	}

	_, err = artDb.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	return pbx.ArtDbToProto(ctx, server.bucket, artDb), nil
}

func validateUpdateArtRequest(req *pb.UpdateArtRequest) error {
	// Ensure the UpdateMask and Arts fields are provided
	err := validation.ValidateStruct(req,
		validation.Field(&req.UpdateMask, validation.Required),
		validation.Field(&req.Art, validation.Required),
	)
	if err != nil {
		return err
	}

	user := req.GetArt()
	updateMaskPaths := req.GetUpdateMask().GetPaths()

	// Dynamically build validation rules based on the fields present in the UpdateMask
	var rules []*validation.FieldRules

	rules = append(rules, validation.Field(&user.Name, validation.Required))

	if slices.Contains(updateMaskPaths, "title") {
		rules = append(rules, validation.Field(&user.Title, validation.Required, validation.Length(1, 255)))
	}

	// Validate the user struct based on the dynamically built rules
	return validation.ValidateStruct(user, rules...)
}

func (server *Server) ListArts(ctx context.Context, req *pb.ListArtsRequest) (*pb.ListArtsResponse, error) {
	userPayload := middleware.FromAdminContext(ctx).UserPayload

	if err := validateListArtsRequest(req); err != nil {
		return nil, err
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

func validateListArtsRequest(req *pb.ListArtsRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.PageSize, validation.Required, validation.Min(1), validation.Max(50)),
		validation.Field(&req.PageToken, validation.Min(0)),
	)
}

func (server *Server) GetArt(ctx context.Context, req *pb.GetArtRequest) (*pb.Art, error) {
	if err := validateGetArtRequest(req); err != nil {
		return nil, err
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

func validateGetArtRequest(req *pb.GetArtRequest) error {
	return validation.ValidateStruct(req) //validation.Field(&req.Id, validation.Required, is.UUIDv4),

}

func (server *Server) DeleteArt(ctx context.Context, req *pb.DeleteArtRequest) (*emptypb.Empty, error) {
	if err := validateDeleteArtRequest(req); err != nil {
		return nil, err
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

func validateDeleteArtRequest(req *pb.DeleteArtRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Name, validation.Required),
	)
}

func validateArt(art *pb.Art, isNew bool) error {
	if isNew {
		return validation.ValidateStruct(art,
			validation.Field(&art.Title, validation.Required, validation.Length(1, 255)),
		)
	} else {
		return validation.ValidateStruct(art,
			validation.Field(&art.Name, validation.Required),
			validation.Field(&art.Title, validation.Required, validation.Length(1, 255)),
		)
	}
}
