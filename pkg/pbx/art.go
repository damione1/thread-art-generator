package pbx

import (
	"fmt"

	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DbArtToProto(post *models.Art) *pb.Art {

	artPb := &pb.Art{
		Title:      post.Title,
		Author:     fmt.Sprintf("users/%s", post.AuthorID),
		CreateTime: timestamppb.New(post.CreatedAt),
		UpdateTime: timestamppb.New(post.UpdatedAt),
	}
	artPb.Name = GetResourceName([]Resource{
		{Type: RessourceTypeUsers, ID: post.AuthorID},
		{Type: RessourceTypeArts, ID: post.ID},
	})

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
