package grpcApi

import (
	"context"
	"fmt"

	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/pbx"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) CreateArt(ctx context.Context, req *pb.CreateArtRequest) (*pb.Art, error) {
	// authorizeUserPayload, err := server.authorizeUser(ctx)
	// if err != nil {
	// 	return nil, unauthenticatedError(err)
	// }

	// if err := validateCreateArtRequest(req); err != nil {
	// 	return nil, err
	// }

	// Create a new file in the blob.bucket
	file, err := server.bucket.NewWriter(ctx, "hello.txt", nil)
	if err != nil {
		return nil, internalError(errors.Wrap(err, "Failed to create file"))
	}

	// Write content to the file
	content := []byte("Hello, world!")
	_, err = file.Write(content)

	if err != nil {
		return nil, internalError(errors.Wrap(err, "Failed to write to file"))
	}

	// Close the file
	err = file.Close()
	if err != nil {
		return nil, internalError(errors.Wrap(err, "Failed to close file"))
	}

	// Get the URL of the file
	url, err := server.bucket.SignedURL(ctx, "hello.txt", nil)
	if err != nil {
		return nil, internalError(errors.Wrap(err, "Failed to get file URL"))
	}
	// Print the URL
	fmt.Println("File URL:", url)

	return nil, nil

	// user, err := models.Users(
	// 	models.UserWhere.ID.EQ(authorizeUserPayload.UserID),
	// ).One(ctx, server.config.DB)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "Failed to get user")
	// }

	// if user.Role != models.RoleEnumUser {
	// 	return nil, rolePermissionError(errors.New("Only admin can create art"))
	// }

	// art := &models.Art{
	// 	Title:    req.GetArt().GetTitle(),
	// 	AuthorID: user.ID,
	// }

	// err = art.Insert(ctx, server.config.DB, boil.Infer())
	// if err != nil {
	// 	return nil, internalError(errors.Wrap(err, "Failed to insert art"))
	// }

	// return pbx.DbArtToProto(art), nil
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
	authorizeUserPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err := validateUpdateArtRequest(req); err != nil {
		return nil, err
	}

	authorId, artId, err := pbx.ParseArtResourceName(req.GetArt().GetName())
	if err != nil {
		return nil, notFoundError(errors.Wrap(err, "Failed to parse resource name"))
	}

	if authorId != authorizeUserPayload.UserID {
		return nil, rolePermissionError(errors.New("Only the author can update the art"))
	}

	// Check if the art exists
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
	).One(ctx, server.config.DB)
	if err != nil {
		return nil, notFoundError(errors.Wrap(err, "Failed to get art"))
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
				return nil, status.Errorf(codes.InvalidArgument, "Invalid field: %s", path)
			}
		}
	}

	// Update the art
	if req.GetArt().GetTitle() != "" {
		artDb.Title = req.GetArt().GetTitle()
	}

	_, err = artDb.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	return pbx.DbArtToProto(artDb), nil
}

func validateUpdateArtRequest(req *pb.UpdateArtRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Art,
			validation.Required,
			validation.By(
				func(value interface{}) error {
					art := value.(*pb.Art)
					return validation.ValidateStruct(art, validation.Field(&art.Name, validation.Required))
				},
			),
			validation.By(
				func(value interface{}) error {
					return validateArt(value.(*pb.Art), false)
				},
			),
		),
		validation.Field(&req.UpdateMask, validation.Required),
	)
}

func (server *Server) ListArts(ctx context.Context, req *pb.ListArtsRequest) (*pb.ListArtsResponse, error) {
	authorizeUserPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err = validateListArtsRequest(req); err != nil {
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
		models.ArtWhere.AuthorID.EQ(authorizeUserPayload.UserID),
		qm.Limit(pageSize),
		qm.Offset(int(req.PageToken)),
	).All(ctx, server.config.DB)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to query arts")
	}

	// Convert the arts to protobuf format
	pbArts := make([]*pb.Art, 0, len(arts))
	for _, dbArt := range arts {
		pbArts = append(pbArts, pbx.DbArtToProto(dbArt))
	}

	return &pb.ListArtsResponse{
		Arts:          pbArts,
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
	authorizeUserPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err := validateGetArtRequest(req); err != nil {
		return nil, err
	}

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		return nil, notFoundError(errors.Wrap(err, "Failed to parse resource name"))
	}

	if authorId != authorizeUserPayload.UserID {
		return nil, rolePermissionError(errors.New("Only the author can get the art"))
	}

	// Check if the art exists
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
	).One(ctx, server.config.DB)
	if err != nil {
		return nil, notFoundError(errors.Wrap(err, "Failed to get art"))
	}

	return pbx.DbArtToProto(artDb), nil
}

func validateGetArtRequest(req *pb.GetArtRequest) error {
	return validation.ValidateStruct(req) //validation.Field(&req.Id, validation.Required, is.UUIDv4),

}

// emptu response
func (server *Server) DeleteArt(ctx context.Context, req *pb.DeleteArtRequest) (*emptypb.Empty, error) {
	authorizeUserPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err := validateDeleteArtRequest(req); err != nil {
		return nil, err
	}

	authorId, artId, err := pbx.ParseArtResourceName(req.GetName())
	if err != nil {
		return nil, notFoundError(errors.Wrap(err, "Failed to parse resource name"))
	}

	if authorId != authorizeUserPayload.UserID {
		return nil, rolePermissionError(errors.New("Only the author can delete the art"))
	}

	// Check if the art exists
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artId),
		models.ArtWhere.AuthorID.EQ(authorId),
	).One(ctx, server.config.DB)
	if err != nil {
		return nil, notFoundError(errors.Wrap(err, "Failed to get art"))
	}

	_, err = artDb.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, internalError(errors.Wrap(err, "Failed to delete art"))
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
