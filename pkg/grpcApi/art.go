package grpcApi

import (
	"context"
	"fmt"

	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/pbx"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (server *Server) CreateArt(ctx context.Context, req *pb.CreateArtRequest) (*pb.CreateArtResponse, error) {
	fmt.Println("🌻 start")
	authorizeUserPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}
	fmt.Println("🌻 authorized")

	if err := validateCreateArtRequest(req); err != nil {
		return nil, err
	}

	fmt.Println("🌻 authorizeUserPayload.UserID", authorizeUserPayload.UserID)
	user, err := models.Users(
		models.UserWhere.ID.EQ(authorizeUserPayload.UserID),
	).One(ctx, server.config.DB)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get user")
	}

	fmt.Println("🌻 check role", user.Role)
	if user.Role != models.RoleEnumUser {
		return nil, rolePermissionError(errors.New("Only admin can create art"))
	}

	art := &models.Art{
		Title:    req.GetArt().GetTitle(),
		AuthorID: user.ID,
	}

	err = art.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, internalError(errors.Wrap(err, "Failed to insert art"))
	}

	return &pb.CreateArtResponse{
		Art: pbx.DbArtToProto(art),
	}, nil
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

func (server *Server) UpdateArt(ctx context.Context, req *pb.UpdateArtRequest) (*pb.UpdateArtResponse, error) {
	if _, err := server.authorizeUser(ctx); err != nil {
		return nil, unauthenticatedError(err)
	}

	if err := validateUpdateArtRequest(req); err != nil {
		return nil, err
	}

	art := pbx.ProtoArtToDb(req.GetArt())

	_, err := art.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &pb.UpdateArtResponse{
		Art: pbx.DbArtToProto(art),
	}, nil
}

func validateUpdateArtRequest(req *pb.UpdateArtRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Art,
			validation.Required,
			validation.By(
				func(value interface{}) error {
					art := value.(*pb.Art)
					return validation.ValidateStruct(art, validation.Field(&art.Id, validation.Required))
				},
			),
			validation.By(
				func(value interface{}) error {
					return validateArt(value.(*pb.Art), false)
				},
			),
		),
	)
}

func (server *Server) ListArts(ctx context.Context, req *pb.ListArtRequest) (*pb.ListArtResponse, error) {
	err := validateListArtsRequest(req)
	if err != nil {
		return nil, err
	}

	pageSize := int(req.GetPageSize())
	pageToken := int(req.GetPageToken())

	// Set default page size if not provided or if it's greater than the maximum allowed
	var maxPageSize int = 50
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	offset := pageSize * pageToken

	dbArts, err := models.Arts(
		qm.OrderBy("created_at desc"),
		qm.Limit(pageSize),
		qm.Offset(offset),
	).All(ctx, server.config.DB)
	if err != nil {
		return nil, err
	}

	arts := make([]*pb.Art, 0, len(dbArts))
	for _, dbArt := range dbArts {
		arts = append(arts, pbx.DbArtToProto(dbArt))
	}

	nextPageToken := 0
	if len(dbArts) == pageSize {
		nextPageToken = pageToken + 1
	}

	return &pb.ListArtResponse{
		Arts:          arts,
		NextPageToken: int32(nextPageToken),
	}, nil
}

func validateListArtsRequest(req *pb.ListArtRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.PageSize, validation.Required, validation.Min(1), validation.Max(50)),
		validation.Field(&req.PageToken, validation.Min(0)),
	)
}

func (server *Server) GetArt(ctx context.Context, req *pb.GetArtRequest) (*pb.GetArtResponse, error) {
	err := validateGetArtRequest(req)
	if err != nil {
		return nil, err
	}

	// dbArt, err := models.Arts(
	// 	models.ArtWhere.ID.EQ(req.GetId()),
	// ).One(ctx, server.config.DB)
	// if err != nil {
	// 	return nil, err
	// }

	return &pb.GetArtResponse{}, nil
}

func validateGetArtRequest(req *pb.GetArtRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.Int),
	)
}

func (server *Server) DeleteArt(ctx context.Context, req *pb.DeleteArtRequest) (*pb.DeleteArtResponse, error) {
	if _, err := server.authorizeUser(ctx); err != nil {
		return nil, unauthenticatedError(err)
	}

	err := validateDeleteArtRequest(req)
	if err != nil {
		return nil, err
	}

	_, err = models.Arts(
		models.ArtWhere.ID.EQ(req.GetId()),
	).DeleteAll(ctx, server.config.DB)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteArtResponse{}, nil
}

func validateDeleteArtRequest(req *pb.DeleteArtRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.Int),
	)
}

func validateArt(art *pb.Art, isNew bool) error {
	if isNew {
		return validation.ValidateStruct(art,
			validation.Field(&art.Title, validation.Required, validation.Length(1, 255)),
		)
	} else {
		return validation.ValidateStruct(art,
			validation.Field(&art.Title, validation.Required, validation.Length(1, 255)),
			validation.Field(&art.Id, is.Int),
		)
	}
}
