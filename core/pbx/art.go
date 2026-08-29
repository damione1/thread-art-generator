package pbx

import (
	"fmt"
	"strings"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PublicURL concatenates the public bucket base with an object key. No I/O.
func PublicURL(publicBaseURL, key string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" || key == "" {
		return ""
	}
	return base + "/" + strings.TrimLeft(key, "/")
}

func artStatus(art *models.Art) pb.ArtStatus {
	switch art.Status {
	case models.ArtStatusEnumPENDING_IMAGE:
		return pb.ArtStatus_ART_STATUS_PENDING_IMAGE
	case models.ArtStatusEnumPROCESSING:
		return pb.ArtStatus_ART_STATUS_PROCESSING
	case models.ArtStatusEnumCOMPLETE:
		return pb.ArtStatus_ART_STATUS_COMPLETE
	case models.ArtStatusEnumFAILED:
		return pb.ArtStatus_ART_STATUS_FAILED
	case models.ArtStatusEnumARCHIVED:
		return pb.ArtStatus_ART_STATUS_ARCHIVED
	default:
		return pb.ArtStatus_ART_STATUS_UNSPECIFIED
	}
}

func ArtDbToProto(art *models.Art, publicBaseURL string) *pb.Art {
	status := artStatus(art)
	artPb := &pb.Art{
		Title:      art.Title,
		Author:     resource.BuildUserResourceName(art.AuthorID),
		CreateTime: timestamppb.New(art.CreatedAt),
		UpdateTime: timestamppb.New(art.UpdatedAt),
		Status:     status,
	}
	artPb.Name = resource.BuildArtResourceName(art.AuthorID, art.ID)

	if art.ImageID.Valid && status == pb.ArtStatus_ART_STATUS_COMPLETE {
		key := resource.ArtImageObjectKey(art.AuthorID, art.ID, art.ImageID.String)
		artPb.ImageUrl = PublicURL(publicBaseURL, key)
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
