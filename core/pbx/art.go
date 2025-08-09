package pbx

import (
	"context"
	"time"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ArtDbToProto(ctx context.Context, storageProvider storage.StorageProvider, art *models.Art, authorFirebaseUID string) *pb.Art {
	// Map status from database enum to proto enum
	var status pb.ArtStatus
	switch art.Status {
	case models.ArtStatusEnumPENDING_IMAGE:
		status = pb.ArtStatus_ART_STATUS_PENDING_IMAGE
	case models.ArtStatusEnumPROCESSING:
		status = pb.ArtStatus_ART_STATUS_PROCESSING
	case models.ArtStatusEnumCOMPLETE:
		status = pb.ArtStatus_ART_STATUS_COMPLETE
	case models.ArtStatusEnumFAILED:
		status = pb.ArtStatus_ART_STATUS_FAILED
	case models.ArtStatusEnumARCHIVED:
		status = pb.ArtStatus_ART_STATUS_ARCHIVED
	default:
		status = pb.ArtStatus_ART_STATUS_UNSPECIFIED
	}

	artPb := &pb.Art{
		Title:      art.Title,
		Author:     resource.BuildUserResourceName(authorFirebaseUID),
		CreateTime: timestamppb.New(art.CreatedAt),
		UpdateTime: timestamppb.New(art.UpdatedAt),
		Status:     status,
	}
	artPb.Name = resource.BuildArtResourceName(authorFirebaseUID, art.ID)

	if art.ImageID.Valid && (status == pb.ArtStatus_ART_STATUS_COMPLETE) && storageProvider != nil {
		// Construct the storage path using Firebase UID: users/{firebase_uid}/arts/{art_id}
		// This matches the upload path used by the client
		imagePath := GetResourceName([]Resource{
			{Type: RessourceTypeUsers, ID: authorFirebaseUID},
			{Type: RessourceTypeArts, ID: art.ID},
		})
		
		log.Debug().
			Str("art_id", art.ID).
			Str("author_firebase_uid", authorFirebaseUID).
			Str("image_path", imagePath).
			Msg("ArtDbToProto: About to generate Firebase Storage URL")
		
		// Add timeout to prevent hanging Firebase Storage operations
		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		
		// Use storage provider to generate public URL directly
		imageURL := storageProvider.GetPublicURL(imagePath)
		
		// Check if context was cancelled or timed out
		if timeoutCtx.Err() != nil {
			log.Error().
				Str("art_id", art.ID).
				Str("image_path", imagePath).
				Err(timeoutCtx.Err()).
				Msg("ArtDbToProto: Firebase Storage URL generation timed out")
			// Continue without image URL rather than hanging the entire request
			imageURL = ""
		} else {
			log.Debug().
				Str("art_id", art.ID).
				Str("image_url", imageURL).
				Msg("ArtDbToProto: Successfully generated Firebase Storage URL")
		}

		artPb.ImageUrl = imageURL
	}

	return artPb
}