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

func ArtDbToProto(ctx context.Context, dualStorage *storage.DualBucketStorage, art *models.Art) *pb.Art {
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
		Author:     resource.BuildUserResourceName(art.AuthorID),
		CreateTime: timestamppb.New(art.CreatedAt),
		UpdateTime: timestamppb.New(art.UpdatedAt),
		Status:     status,
	}
	artPb.Name = resource.BuildArtResourceName(art.AuthorID, art.ID)

	if art.ImageID.Valid && (status == pb.ArtStatus_ART_STATUS_COMPLETE) {
		imageKey := resource.BuildArtResourceName(art.AuthorID, art.ImageID.String)

		// Use public URL generator for CDN caching - art images should be publicly accessible
		publicURLGenerator := storage.NewPublicURLGenerator(dualStorage.GetPublicStorage())
		imageURL := storage.GenerateImageURL(ctx, publicURLGenerator, imageKey, storage.DefaultURLOptions())
		
		artPb.ImageUrl = imageURL
	}

	return artPb
}

func ProtoArtToDb(post *pb.Art) *models.Art {
	artDb := &models.Art{
		Title: post.GetTitle(),
	}

	if post.GetName() != "" {
		artResource, err := resource.ParseResourceName(post.GetName())
		if err != nil {
			return nil
		}

		if art, ok := artResource.(*resource.Art); ok {
			artDb.ID = art.ArtID
			artDb.AuthorID = art.UserID
		}
	}

	if post.GetCreateTime() != nil {
		artDb.CreatedAt = post.GetCreateTime().AsTime()
	}
	if post.GetUpdateTime() != nil {
		artDb.UpdatedAt = post.GetUpdateTime().AsTime()
	}
	return artDb
}

// ParseArtResourceName parses an art resource name and returns user ID and art ID
// Deprecated: Use resource.ParseResourceName instead
func ParseArtResourceName(resourceName string) (string, string, error) {
	artResource, err := resource.ParseResourceName(resourceName)
	if err != nil {
		return "", "", err
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		return "", "", fmt.Errorf("invalid art resource name")
	}

	return art.UserID, art.ArtID, nil
}
