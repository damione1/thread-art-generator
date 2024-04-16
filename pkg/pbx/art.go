package pbx

import (
	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DbArtToProto(post *models.Art) *pb.Art {
	artPb := &pb.Art{
		Id:         post.ID,
		Title:      post.Title,
		AuthorId:   post.AuthorID,
		CreateTime: timestamppb.New(post.CreatedAt),
		UpdateTime: timestamppb.New(post.UpdatedAt),
	}
	return artPb
}

func ProtoArtToDb(post *pb.Art) *models.Art {
	artDb := &models.Art{
		Title: post.GetTitle(),
	}
	if post.GetId() != "" {
		artDb.ID = post.GetId()
	}
	if post.GetCreateTime() != nil {
		artDb.CreatedAt = post.GetCreateTime().AsTime()
	}
	if post.GetUpdateTime() != nil {
		artDb.UpdatedAt = post.GetUpdateTime().AsTime()
	}
	return artDb
}
