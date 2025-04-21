package pbx

import (
	"context"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Define the composition resource name constant
const (
	RessourceNameCompositions = "compositions"
)

// Add composition to valid resource types
func init() {
	validResourceTypesList = append(validResourceTypesList, RessourceNameCompositions)
	RessourceTypeCompositions = &ResourceType{Type: RessourceNameCompositions, Parent: RessourceTypeArts}
}

// RessourceTypeCompositions is the resource type for compositions
var RessourceTypeCompositions *ResourceType

// CompositionDbToProto converts a database composition model to a proto composition
func CompositionDbToProto(ctx context.Context, bucket *storage.BlobStorage, artDb *models.Art, composition *models.Composition) *pb.Composition {
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

	// Set the resource name
	compositionPb.Name = GetResourceName([]Resource{
		{Type: RessourceTypeUsers, ID: artDb.AuthorID},
		{Type: RessourceTypeArts, ID: artDb.ID},
		{Type: RessourceTypeCompositions, ID: composition.ID},
	})

	// Set optional result fields if they exist
	if composition.PreviewURL.Valid && bucket != nil {
		compositionPb.PreviewUrl = bucket.GetPublicURL(composition.PreviewURL.String)
	}

	if composition.GcodeURL.Valid && bucket != nil {
		compositionPb.GcodeUrl = bucket.GetPublicURL(composition.GcodeURL.String)
	}

	if composition.PathlistURL.Valid && bucket != nil {
		compositionPb.PathlistUrl = bucket.GetPublicURL(composition.PathlistURL.String)
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
		resources, err := GetResourcesFromResourceName(comp.GetName())
		if err == nil {
			if compositionID, ok := resources[RessourceNameCompositions]; ok {
				compositionDb.ID = compositionID
			}
		}
	}

	return compositionDb
}

// ParseCompositionResourceName parses a composition resource name into user ID, art ID, and composition ID
func ParseCompositionResourceName(resourceName string) (string, string, string, error) {
	resources, err := GetResourcesFromResourceName(resourceName)
	if err != nil {
		return "", "", "", err
	}

	userID, ok := resources[RessourceNameUsers]
	if !ok {
		return "", "", "", fmt.Errorf("user ID not found in resource name")
	}

	artID, ok := resources[RessourceNameArts]
	if !ok {
		return "", "", "", fmt.Errorf("art ID not found in resource name")
	}

	compositionID, ok := resources[RessourceNameCompositions]
	if !ok {
		return "", "", "", fmt.Errorf("composition ID not found in resource name")
	}

	return userID, artID, compositionID, nil
}
