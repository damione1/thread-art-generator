package grpcApi

import (
	"context"

	"github.com/Damione1/portfolio-playground/pkg/pbx"
	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/friendsofgo/errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (server *Server) CreatePost(ctx context.Context, req *pb.CreateArtRequest) (*pb.CreateArtResponse, error) {
	authorizeUserPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if err := validateCreateArtRequest(req); err != nil {
		return nil, err
	}

	user, err := models.FindUser(ctx, server.config.DB, authorizeUserPayload.ID.String())
	if err != nil {
		return nil, err
	}
	if user.Role != "admin" {
		return nil, errors.Wrap(err, "Unsufficient permissions to create a post")
	}

	post := pbx.ProtoPostToDb(req.GetArt())

	err = post.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &pb.CreateArtResponse{}, nil
}

func validateCreateArtRequest(req *pb.CreateArtRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Art,
			validation.Required,
			validation.By(
				func(value interface{}) error {
					return validatePost(value.(*pb.Art))
				},
			),
		),
	)
}

func (server *Server) UpdatePost(ctx context.Context, req *pb.UpdateArtRequest) (*pb.UpdateArtResponse, error) {
	if _, err := server.authorizeUser(ctx); err != nil {
		return nil, unauthenticatedError(err)
	}

	if err := validateUpdateArtRequest(req); err != nil {
		return nil, err
	}

	post := pbx.ProtoPostToDb(req.GetArt())

	_, err := post.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &pb.UpdateArtResponse{
		Art: pbx.DbPostToProto(post),
	}, nil
}

func validateUpdateArtRequest(req *pb.UpdateArtRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Art,
			validation.Required,
			validation.By(
				func(value interface{}) error {
					post := value.(*pb.Art)
					return validation.ValidateStruct(post, validation.Field(&post.Id, validation.Required))
				},
			),
			validation.By(
				func(value interface{}) error {
					return validatePost(value.(*pb.Art))
				},
			),
		),
	)
}

func (server *Server) ListPosts(ctx context.Context, req *pb.ListArtRequest) (*pb.ListArtResponse, error) {
	err := validateListPostsRequest(req)
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

	dbPosts, err := models.Posts(
		qm.OrderBy("created_at desc"),
		qm.Limit(pageSize),
		qm.Offset(offset),
	).All(ctx, server.config.DB)
	if err != nil {
		return nil, err
	}

	posts := make([]*pb.Art, 0, len(dbPosts))
	for _, dbPost := range dbPosts {
		posts = append(posts, pbx.DbPostToProto(dbPost))
	}

	nextPageToken := 0
	if len(dbPosts) == pageSize {
		nextPageToken = pageToken + 1
	}

	return &pb.ListArtResponse{
		Arts:          posts,
		NextPageToken: int32(nextPageToken),
	}, nil
}

func validateListPostsRequest(req *pb.ListArtRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.PageSize, validation.Required, validation.Min(1), validation.Max(50)),
		validation.Field(&req.PageToken, validation.Min(0)),
	)
}

func (server *Server) GetPost(ctx context.Context, req *pb.GetArtRequest) (*pb.GetArtResponse, error) {
	err := validateGetArtRequest(req)
	if err != nil {
		return nil, err
	}

	// dbPost, err := models.Posts(
	// 	models.PostWhere.ID.EQ(req.GetId()),
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

func (server *Server) DeletePost(ctx context.Context, req *pb.DeleteArtRequest) (*pb.DeleteArtResponse, error) {
	if _, err := server.authorizeUser(ctx); err != nil {
		return nil, unauthenticatedError(err)
	}

	err := validateDeleteArtRequest(req)
	if err != nil {
		return nil, err
	}

	_, err = models.Posts(
		models.PostWhere.ID.EQ(req.GetId()),
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

func validatePost(post *pb.Art) error {
	return validation.ValidateStruct(post,
		validation.Field(&post.Title, validation.Required, validation.Length(1, 255)),
	)
}
