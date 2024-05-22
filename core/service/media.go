package service

// import (
// 	"context"
// 	"fmt"
// 	"mime"
// 	"net/http"
// 	"time"

// 	"github.com/Damione1/thread-art-generator/core/db/models"
// 	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
// 	"github.com/Damione1/thread-art-generator/core/middleware"
// 	"github.com/Damione1/thread-art-generator/core/pb"
// 	"github.com/Damione1/thread-art-generator/core/pbx"
// 	"github.com/friendsofgo/errors"
// 	validation "github.com/go-ozzo/ozzo-validation"
// 	"github.com/volatiletech/null/v8"
// 	"github.com/volatiletech/sqlboiler/v4/boil"
// 	"gocloud.dev/blob"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// func (server *Server) GetMediaUploadUrl(ctx context.Context, req *pb.GetMediaUploadUrlRequest) (*pb.GetMediaUploadUrlResponse, error) {
// 	var ressourceType string
// 	var ressourceId string
// 	var signedUrl string

// 	userPayload := middleware.FromAdminContext(ctx).UserPayload

// 	if err := validateGetMediaUploadUrlRequest(req); err != nil {
// 		return nil, err
// 	}

// 	authorId, err := pbx.GetResourceIDByType(req.GetName(), pbx.RessourceTypeUsers)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if authorId != userPayload.UserID {
// 		return nil, pbErrors.RolePermissionError(errors.New("Only the author can upload the image"))
// 	}

// 	if ressourceId, err = pbx.GetResourceIDByType(req.GetName(), pbx.RessourceTypeArts); err == nil {
// 		ressourceType = pbx.RessourceNameArts
// 	}

// 	switch ressourceType {
// 	case pbx.RessourceNameArts:
// 		// Check if the art exists
// 		art, err := models.Arts(
// 			models.ArtWhere.ID.EQ(ressourceId),
// 			models.ArtWhere.AuthorID.EQ(authorId),
// 		).One(ctx, server.config.DB)
// 		if err != nil {
// 			return nil, pbErrors.NotFoundError(errors.Wrap(err, "Failed to get art"))
// 		}

// 		blobKey := fmt.Sprintf("users/%s/arts/%s.%s", authorId, ressourceId, req.GetExtension())
// 		// Creating a new blob signed URL
// 		signedUrl, err = server.bucket.SignedURL(ctx, blobKey, &blob.SignedURLOptions{
// 			Method:      "PUT",
// 			Expiry:      time.Duration(1) * time.Minute,
// 			ContentType: req.GetContentType(),
// 			BeforeSign: func(asFunc func(interface{}) bool) error {
// 				return nil
// 			},
// 		})
// 		if err != nil {
// 			return nil, pbErrors.InternalError(errors.Wrap(err, "Failed to create signed URL"))
// 		}

// 		// Update the art with the new image URL
// 		art.ImageID = null.NewString(ressourceId, true)
// 		if _, err = art.Update(ctx, server.config.DB, boil.Infer()); err != nil {
// 			return nil, pbErrors.InternalError(errors.Wrap(err, "Failed to update art"))
// 		}
// 	default:
// 		return nil, status.Errorf(codes.InvalidArgument, "Invalid resource type: %s", ressourceType)
// 	}

// 	// Return the upload URL
// 	return &pb.GetMediaUploadUrlResponse{
// 		UploadUrl: signedUrl,
// 	}, nil
// }

// func validateGetMediaUploadUrlRequest(req *pb.GetMediaUploadUrlRequest) error {
// 	return validation.ValidateStruct(req,
// 		validation.Field(&req.Name, validation.Required),
// 		validation.Field(&req.ContentType, validation.Required, validation.By(func(value interface{}) error {
// 			mediaType, _, err := mime.ParseMediaType(value.(string))
// 			if err != nil {
// 				return errors.New("invalid content type")
// 			}
// 			if mediaType != "image/jpeg" && mediaType != "image/png" {
// 				return errors.New("invalid content type")
// 			}
// 			return nil
// 		})),
// 		validation.Field(&req.Extension, validation.Required, validation.By(func(value interface{}) error {
// 			if value.(string) != ".jpeg" && value.(string) != ".png" {
// 				return errors.New("invalid extension")
// 			}
// 			return nil
// 		})),
// 		validation.Field(&req.Md5, validation.By(func(value interface{}) error {
// 			//to be implemented
// 			return nil
// 		})),
// 	)
// }

// func HandleBinaryFileUpload(w http.ResponseWriter, r *http.Request) {
// 	// err := r.ParseForm()
// 	// if err != nil {
// 	// 	http.Error(w, fmt.Sprintf("failed to parse form: %s", err.Error()), http.StatusBadRequest)
// 	// 	return
// 	// }

// 	// f, header, err := r.FormFile("attachment")
// 	// if err != nil {
// 	// 	http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
// 	// 	return
// 	// }
// 	// defer f.Close()

// 	// //
// 	// // Now do something with the io.Reader in `f`, i.e. read it into a buffer or stream it to a gRPC client side stream.
// 	// // Also `header` will contain the filename, size etc of the original file.
// 	// //
// }
