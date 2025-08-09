package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
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
)

// ConfirmArtImageUploadFromFunction confirms that an art image has been uploaded
// and updates the art status. This method is specifically designed for Firebase Functions
// and requires internal API key authentication (handled by the Connect adapter).
//
// Security Notes:
// - This method does NOT use Firebase authentication context
// - Firebase UID is passed as a parameter for authorization
// - Internal API key validation is enforced by the Connect adapter
// - Art ownership is validated using the provided Firebase UID
func (server *Server) ConfirmArtImageUploadFromFunction(ctx context.Context, req *pb.ConfirmArtImageUploadFromFunctionRequest) (*pb.Art, error) {
	log.Info().
		Str("art_name", req.GetName()).
		Str("firebase_uid", req.GetFirebaseUid()).
		Str("image_url", req.GetImageUrl()).
		Str("original_filename", req.GetOriginalFilename()).
		Str("content_type", req.GetContentType()).
		Int64("file_size", req.GetFileSize()).
		Msg("ConfirmArtImageUploadFromFunction: Processing Firebase Function image upload confirmation")

	// Validate the request
	if err := protovalidate.Validate(req); err != nil {
		log.Warn().
			Err(err).
			Str("art_name", req.GetName()).
			Str("firebase_uid", req.GetFirebaseUid()).
			Msg("ConfirmArtImageUploadFromFunction: Request validation failed")
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	// Parse the art resource name
	artResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		log.Warn().
			Err(err).
			Str("art_name", req.GetName()).
			Str("firebase_uid", req.GetFirebaseUid()).
			Msg("ConfirmArtImageUploadFromFunction: Invalid art resource name")
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		log.Warn().
			Str("art_name", req.GetName()).
			Str("firebase_uid", req.GetFirebaseUid()).
			Msg("ConfirmArtImageUploadFromFunction: Invalid art resource type")
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Resource name contains Firebase UID: users/{firebase_uid}/arts/{art_id}
	firebaseUidFromPath := art.UserID // This is Firebase UID from the resource path
	
	log.Info().
		Str("firebase_uid_from_path", firebaseUidFromPath).
		Str("firebase_uid_from_request", req.GetFirebaseUid()).
		Str("art_id", art.ArtID).
		Str("art_name", req.GetName()).
		Msg("ConfirmArtImageUploadFromFunction: Validating Firebase UID consistency")

	// Verify Firebase UID consistency between path and request parameter
	if firebaseUidFromPath != req.GetFirebaseUid() {
		log.Warn().
			Str("firebase_uid_from_path", firebaseUidFromPath).
			Str("firebase_uid_from_request", req.GetFirebaseUid()).
			Str("art_name", req.GetName()).
			Msg("ConfirmArtImageUploadFromFunction: Firebase UID mismatch between path and request")
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("firebase_uid", errors.New("Firebase UID in request does not match resource path")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Look up user and art in a single query using Firebase UID
	// The arts table stores author_id as internal UUID, so we need to join with users table
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		qm.InnerJoin("users ON users.id = arts.author_id"),
		qm.Where("users.firebase_uid = ?", req.GetFirebaseUid()),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().
				Str("art_id", art.ArtID).
				Str("firebase_uid", req.GetFirebaseUid()).
				Msg("ConfirmArtImageUploadFromFunction: Art not found or user not authorized")
			return nil, pbErrors.NotFoundError("art not found")
		}
		log.Error().
			Err(err).
			Str("art_id", art.ArtID).
			Str("firebase_uid", req.GetFirebaseUid()).
			Msg("ConfirmArtImageUploadFromFunction: Failed to get art from database")
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	log.Info().
		Str("art_id", artDb.ID).
		Str("current_status", string(artDb.Status)).
		Str("firebase_uid", req.GetFirebaseUid()).
		Msg("ConfirmArtImageUploadFromFunction: Found art")

	// Verify the art is in PENDING_IMAGE status
	if artDb.Status != models.ArtStatusEnumPENDING_IMAGE {
		log.Warn().
			Str("art_id", artDb.ID).
			Str("current_status", string(artDb.Status)).
			Str("firebase_uid", req.GetFirebaseUid()).
			Msg("ConfirmArtImageUploadFromFunction: Art is not in pending image status")
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", fmt.Errorf("art is not in pending image status, current status: %s", artDb.Status)),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Authorization is validated by the JOIN query:
	// The query joins arts with users table and filters by Firebase UID
	// If the query succeeds, it means the art belongs to the authenticated user

	// Upload is valid - update the art with image URL and set status to COMPLETE
	artDb.ImageID = null.StringFrom(req.GetImageUrl())
	artDb.Status = models.ArtStatusEnumCOMPLETE

	// Update the art in database
	_, err = artDb.Update(ctx, server.config.DB, boil.Infer())
	if err != nil {
		log.Error().
			Err(err).
			Str("art_id", artDb.ID).
			Str("firebase_uid", req.GetFirebaseUid()).
			Msg("ConfirmArtImageUploadFromFunction: Failed to update art")
		return nil, pbErrors.InternalError("failed to update art", err)
	}

	// Audit log: Successful image upload confirmation
	log.Info().
		Str("art_id", artDb.ID).
		Str("firebase_uid", req.GetFirebaseUid()).
		Str("author_id", artDb.AuthorID).
		Str("image_url", req.GetImageUrl()).
		Str("original_filename", req.GetOriginalFilename()).
		Str("content_type", req.GetContentType()).
		Int64("file_size", req.GetFileSize()).
		Msg("ConfirmArtImageUploadFromFunction: Successfully confirmed art image upload and updated status to COMPLETE")

	// Return the updated art
	return pbx.ArtDbToProto(ctx, server.GetStorage(), artDb, req.GetFirebaseUid()), nil
}