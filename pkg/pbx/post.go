package pbx

// import (
// 	"github.com/Damione1/thread-art-generator/pkg/db/models"
// 	"github.com/Damione1/thread-art-generator/pkg/pb"
// 	"github.com/volatiletech/null/v8"
// 	"google.golang.org/protobuf/types/known/timestamppb"
// )

// func DbPostToProto(post *models.Art) *pb.Art {
// 	postPb := &pb.Art{
// 		Id:         post.ID,
// 		Title:      post.Title,
// 		CreateTime: timestamppb.New(post.CreatedAt.Time),
// 		UpdateTime: timestamppb.New(post.UpdatedAt.Time),
// 	}
// 	return postPb
// }

// func ProtoPostToDb(post *pb.Art) *models.Post {
// 	postDb := &models.Post{
// 		Title: post.GetTitle(),
// 	}
// 	if post.GetId() != "" {
// 		postDb.ID = post.GetId()
// 	}
// 	if post.GetCreateTime() != nil {
// 		postDb.CreatedAt = null.TimeFrom(post.GetCreateTime().AsTime())
// 	}
// 	if post.GetUpdateTime() != nil {
// 		postDb.UpdatedAt = null.TimeFrom(post.GetUpdateTime().AsTime())
// 	}

// 	return postDb
// }
