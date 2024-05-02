package grpcApi

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
)

func (server *Server) CreateMedia(stream pb.ArtGeneratorService_CreateMediaServer) error {
	ctx := stream.Context()
	authorizeUserPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return unauthenticatedError(err)
	}

	// if err := validateCreateMediaRequest(req); err != nil {
	// 	return nil, err
	// }

	user, err := models.Users(
		models.UserWhere.ID.EQ(authorizeUserPayload.UserID),
	).One(ctx, server.config.DB)
	if err != nil {
		return errors.Wrap(err, "Failed to get user")
	}

	if user.Role != models.RoleEnumUser {
		return rolePermissionError(errors.New("Only admin can create media"))
	}

	// Initialize variables to hold the media details
	var mediaName string
	var mediaBytes []byte

	// Receive chunks from the stream
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// End of file; finish processing
			break
		}
		if err != nil {
			return errors.Wrap(err, "Failed receiving a chunk")
		}

		// Append received bytes to the media slice
		mediaBytes = append(mediaBytes, req.MediaChunk.Chunk...)
		mediaName = req.MediaChunk.Name
	}

	// Save the media to a file or a database
	mediaPath := filepath.Join("some_directory", mediaName)
	if err := os.WriteFile(mediaPath, mediaBytes, 0666); err != nil {
		return errors.Wrap(err, "Failed to save media file")
	}

	// Create a response (assuming pb.Media has fields reflecting the media storage)
	response := &pb.Media{
		Name: mediaName,
		///Path: mediaPath,
		// other fields if necessary
	}

	// Send response back to client
	if err := stream.SendAndClose(response); err != nil {
		return errors.Wrap(err, "Failed to send response")
	}

	return nil
}

// func validateCreateMediaRequest(req *pb.CreateMediaRequest) error {
// 	return validation.ValidateStruct(req,
// 		validation.Field(&req.Media,
// 			validation.Required,
// 			validation.By(
// 				func(value interface{}) error {
// 					return validateMedia(value.(*pb.Media), true)
// 				},
// 			),
// 		),
// 	)
// }

// func (server *Server) GetMedia(ctx context.Context, req *pb.GetMediaRequest) (*pb.Media, error) {
// 	authorizeUserPayload, err := server.authorizeUser(ctx)
// 	if err != nil {
// 		return nil, unauthenticatedError(err)
// 	}

// 	if err := validateGetMediaRequest(req); err != nil {
// 		return nil, err
// 	}

// 	authorId, mediaId, err := pbx.ParseMediaResourceName(req.GetName())
// 	if err != nil {
// 		return nil, notFoundError(errors.Wrap(err, "Failed to parse resource name"))
// 	}

// 	if authorId != authorizeUserPayload.UserID {
// 		return nil, rolePermissionError(errors.New("Only the author can get the media"))
// 	}

// 	// Check if the media exists
// 	mediaDb, err := models.Medias(
// 		models.MediaWhere.ID.EQ(mediaId),
// 		models.MediaWhere.AuthorID.EQ(authorId),
// 	).One(ctx, server.config.DB)
// 	if err != nil {
// 		return nil, notFoundError(errors.Wrap(err, "Failed to get media"))
// 	}

// 	return pbx.DbMediaToProto(mediaDb), nil
// }

func validateGetMediaRequest(req *pb.GetMediaRequest) error {
	return validation.ValidateStruct(req) //validation.Field(&req.Id, validation.Required, is.UUIDv4),

}

// emptu response
// func (server *Server) DeleteMedia(ctx context.Context, req *pb.DeleteMediaRequest) (*emptypb.Empty, error) {
// 	authorizeUserPayload, err := server.authorizeUser(ctx)
// 	if err != nil {
// 		return nil, unauthenticatedError(err)
// 	}

// 	if err := validateDeleteMediaRequest(req); err != nil {
// 		return nil, err
// 	}

// 	authorId, mediaId, err := pbx.ParseMediaResourceName(req.GetName())
// 	if err != nil {
// 		return nil, notFoundError(errors.Wrap(err, "Failed to parse resource name"))
// 	}

// 	if authorId != authorizeUserPayload.UserID {
// 		return nil, rolePermissionError(errors.New("Only the author can delete the media"))
// 	}

// 	// Check if the media exists
// 	mediaDb, err := models.Medias(
// 		models.MediaWhere.ID.EQ(mediaId),
// 		models.MediaWhere.AuthorID.EQ(authorId),
// 	).One(ctx, server.config.DB)
// 	if err != nil {
// 		return nil, notFoundError(errors.Wrap(err, "Failed to get media"))
// 	}

// 	_, err = mediaDb.Delete(ctx, server.config.DB)
// 	if err != nil {
// 		return nil, internalError(errors.Wrap(err, "Failed to delete media"))
// 	}

// 	return &emptypb.Empty{}, nil
// }

func validateDeleteMediaRequest(req *pb.DeleteMediaRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Name, validation.Required),
	)
}

// func validateMedia(media *pb.Media, isNew bool) error {
// 	if isNew {
// 		return validation.ValidateStruct(media,
// 			validation.Field(&media.Title, validation.Required, validation.Length(1, 255)),
// 		)
// 	} else {
// 		return validation.ValidateStruct(media,
// 			validation.Field(&media.Name, validation.Required),
// 		)
// 	}
// }
