package pbx

import (
	"context"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ArtDbToProto(ctx context.Context, bucket *storage.BlobStorage, art *models.Art) *pb.Art {
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
		Author:     fmt.Sprintf("users/%s", art.AuthorID),
		CreateTime: timestamppb.New(art.CreatedAt),
		UpdateTime: timestamppb.New(art.UpdatedAt),
		Status:     status,
	}
	artPb.Name = GetResourceName([]Resource{
		{Type: RessourceTypeUsers, ID: art.AuthorID},
		{Type: RessourceTypeArts, ID: art.ID},
	})

	if art.ImageID.Valid && (status == pb.ArtStatus_ART_STATUS_COMPLETE) {
		imageKey := GetResourceName([]Resource{
			{Type: RessourceTypeUsers, ID: art.AuthorID},
			{Type: RessourceTypeArts, ID: art.ImageID.String},
		})

		// Use public URL instead of signed URL
		publicURL := bucket.GetPublicURL(imageKey)
		artPb.ImageUrl = publicURL
	}

	return artPb
}

func ProtoArtToDb(post *pb.Art) *models.Art {
	artDb := &models.Art{
		Title: post.GetTitle(),
	}

	if post.GetName() != "" {
		ressources, err := GetResourcesFromResourceName(post.GetName())
		if err != nil {
			return nil
		}

		if artId, ok := ressources[RessourceNameArts]; ok {
			artDb.ID = artId
		}

		if authorId, ok := ressources[RessourceNameUsers]; ok {
			artDb.AuthorID = authorId
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

func ParseArtResourceName(resourceName string) (string, string, error) {
	resources, err := GetResourcesFromResourceName(resourceName)
	if err != nil {
		return "", "", err
	}

	authorId, ok := resources[RessourceNameUsers]
	if !ok {
		return "", "", fmt.Errorf("author ID not found in resource name")
	}

	artId, ok := resources[RessourceNameArts]
	if !ok {
		return "", "", fmt.Errorf("art ID not found in resource name")
	}

	return authorId, artId, nil
}
