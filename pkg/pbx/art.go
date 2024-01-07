package pbx

import (
	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/volatiletech/null/v8"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DbArtToProto(post *models.Art) *pb.Art {
	artPb := &pb.Art{
		Id:         post.ID,
		Title:      post.Title,
		CreateTime: timestamppb.New(post.CreatedAt.Time),
		UpdateTime: timestamppb.New(post.UpdatedAt.Time),
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
		artDb.CreatedAt = null.TimeFrom(post.GetCreateTime().AsTime())
	}
	if post.GetUpdateTime() != nil {
		artDb.UpdatedAt = null.TimeFrom(post.GetUpdateTime().AsTime())
	}

	return artDb
}
