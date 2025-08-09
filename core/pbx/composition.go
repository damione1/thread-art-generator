package pbx

import (
	"context"
	"fmt"
	"time"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CompositionDbToProto converts a database composition model to a proto composition
func CompositionDbToProto(ctx context.Context, storageProvider storage.StorageProvider, artDb *models.Art, composition *models.Composition, authorFirebaseUID string) *pb.Composition {
	// Map status from database enum to proto enum
	var status pb.CompositionStatus
	switch composition.Status {
	case models.CompositionStatusEnumPENDING:
		status = pb.CompositionStatus_COMPOSITION_STATUS_PENDING
	case models.CompositionStatusEnumPROCESSING:
		status = pb.CompositionStatus_COMPOSITION_STATUS_PROCESSING
	case models.CompositionStatusEnumCOMPLETE:
		status = pb.CompositionStatus_COMPOSITION_STATUS_COMPLETE
	case models.CompositionStatusEnumFAILED:
		status = pb.CompositionStatus_COMPOSITION_STATUS_FAILED
	default:
		status = pb.CompositionStatus_COMPOSITION_STATUS_UNSPECIFIED
	}

	// Create proto composition with basic fields
	compositionPb := &pb.Composition{
		NailsQuantity:     int32(composition.NailsQuantity),
		ImgSize:           int32(composition.ImgSize),
		MaxPaths:          int32(composition.MaxPaths),
		StartingNail:      int32(composition.StartingNail),
		MinimumDifference: int32(composition.MinimumDifference),
		BrightnessFactor:  int32(composition.BrightnessFactor),
		ImageContrast:     float32(composition.ImageContrast),
		PhysicalRadius:    float32(composition.PhysicalRadius),
		Status:            status,
		CreateTime:        timestamppb.New(composition.CreatedAt),
		UpdateTime:        timestamppb.New(composition.UpdatedAt),
	}

	// Set the resource name using Firebase UID instead of internal ID
	compositionPb.Name = resource.BuildCompositionResourceName(authorFirebaseUID, artDb.ID, composition.ID)

	// Set optional result fields if they exist using storage provider
	if storageProvider != nil {
		// Add timeout to prevent hanging Firebase Storage operations
		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if composition.PreviewURL.Valid {
			log.Debug().
				Str("composition_id", composition.ID).
				Str("preview_path", composition.PreviewURL.String).
				Msg("CompositionDbToProto: Generating preview URL")
			
			previewURL := storageProvider.GetPublicURL(composition.PreviewURL.String)
			if timeoutCtx.Err() != nil {
				log.Error().
					Str("composition_id", composition.ID).
					Str("preview_path", composition.PreviewURL.String).
					Err(timeoutCtx.Err()).
					Msg("CompositionDbToProto: Preview URL generation timed out")
			} else {
				compositionPb.PreviewUrl = previewURL
			}
		}

		if composition.GcodeURL.Valid {
			log.Debug().
				Str("composition_id", composition.ID).
				Str("gcode_path", composition.GcodeURL.String).
				Msg("CompositionDbToProto: Generating gcode URL")
			
			gcodeURL := storageProvider.GetPublicURL(composition.GcodeURL.String)
			if timeoutCtx.Err() != nil {
				log.Error().
					Str("composition_id", composition.ID).
					Str("gcode_path", composition.GcodeURL.String).
					Err(timeoutCtx.Err()).
					Msg("CompositionDbToProto: Gcode URL generation timed out")
			} else {
				compositionPb.GcodeUrl = gcodeURL
			}
		}

		if composition.PathlistURL.Valid {
			log.Debug().
				Str("composition_id", composition.ID).
				Str("pathlist_path", composition.PathlistURL.String).
				Msg("CompositionDbToProto: Generating pathlist URL")
			
			pathlistURL := storageProvider.GetPublicURL(composition.PathlistURL.String)
			if timeoutCtx.Err() != nil {
				log.Error().
					Str("composition_id", composition.ID).
					Str("pathlist_path", composition.PathlistURL.String).
					Err(timeoutCtx.Err()).
					Msg("CompositionDbToProto: Pathlist URL generation timed out")
			} else {
				compositionPb.PathlistUrl = pathlistURL
			}
		}
	}

	if composition.ThreadLength.Valid {
		compositionPb.ThreadLength = int32(composition.ThreadLength.Int)
	}

	if composition.TotalLines.Valid {
		compositionPb.TotalLines = int32(composition.TotalLines.Int)
	}

	if composition.ErrorMessage.Valid {
		compositionPb.ErrorMessage = composition.ErrorMessage.String
	}

	return compositionPb
}

// ProtoCompositionToDb converts a proto composition to a database composition model
func ProtoCompositionToDb(comp *pb.Composition) *models.Composition {
	compositionDb := &models.Composition{
		NailsQuantity:     int(comp.GetNailsQuantity()),
		ImgSize:           int(comp.GetImgSize()),
		MaxPaths:          int(comp.GetMaxPaths()),
		StartingNail:      int(comp.GetStartingNail()),
		MinimumDifference: int(comp.GetMinimumDifference()),
		BrightnessFactor:  int(comp.GetBrightnessFactor()),
		ImageContrast:     float64(comp.GetImageContrast()),
		PhysicalRadius:    float64(comp.GetPhysicalRadius()),
	}

	// Extract resource IDs from the name if it exists
	if comp.GetName() != "" {
		compositionResource, err := resource.ParseResourceName(comp.GetName())
		if err == nil {
			if composition, ok := compositionResource.(*resource.Composition); ok {
				compositionDb.ID = composition.CompositionID
			}
		}
	}

	return compositionDb
}

// ParseCompositionResourceName parses a composition resource name into user ID, art ID, and composition ID
// Deprecated: Use resource.ParseResourceName instead
func ParseCompositionResourceName(resourceName string) (string, string, string, error) {
	compositionResource, err := resource.ParseResourceName(resourceName)
	if err != nil {
		return "", "", "", err
	}

	composition, ok := compositionResource.(*resource.Composition)
	if !ok {
		return "", "", "", fmt.Errorf("invalid composition resource name")
	}

	return composition.UserID, composition.ArtID, composition.CompositionID, nil
}