package pbx

import (
	"context"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/Damione1/thread-art-generator/core/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CompositionDbToProto converts a database composition model to a proto composition
func CompositionDbToProto(ctx context.Context, dualStorage *storage.DualBucketStorage, artDb *models.Art, composition *models.Composition) *pb.Composition {
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

	// Set the resource name using the new builder
	compositionPb.Name = resource.BuildCompositionResourceName(artDb.AuthorID, artDb.ID, composition.ID)

	// Set optional result fields if they exist using public URL generator for CDN caching
	if dualStorage != nil {
		publicURLGenerator := storage.NewPublicURLGenerator(dualStorage.GetPublicStorage())
		urlOptions := storage.DefaultURLOptions()

		if composition.PreviewURL.Valid {
			compositionPb.PreviewUrl = storage.GenerateImageURL(ctx, publicURLGenerator, composition.PreviewURL.String, urlOptions)
		}

		if composition.GcodeURL.Valid {
			compositionPb.GcodeUrl = storage.GenerateImageURL(ctx, publicURLGenerator, composition.GcodeURL.String, urlOptions)
		}

		if composition.PathlistURL.Valid {
			compositionPb.PathlistUrl = storage.GenerateImageURL(ctx, publicURLGenerator, composition.PathlistURL.String, urlOptions)
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
