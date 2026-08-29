package pbx

import (
	"fmt"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CompositionDbToProto(publicBaseURL string, artDb *models.Art, composition *models.Composition) *pb.Composition {
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

	compositionPb.Name = resource.BuildCompositionResourceName(artDb.AuthorID, artDb.ID, composition.ID)

	if composition.PreviewURL.Valid {
		compositionPb.PreviewUrl = PublicURL(publicBaseURL, composition.PreviewURL.String)
	}
	if composition.GcodeURL.Valid {
		compositionPb.GcodeUrl = composition.GcodeURL.String
	}
	if composition.PathlistURL.Valid {
		compositionPb.PathlistUrl = composition.PathlistURL.String
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
