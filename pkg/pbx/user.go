package pbx

import (
	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
)

func DbUserToProto(user *models.User) *pb.User {
	userPb := &pb.User{
		Id:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}
	return userPb
}

func ProtoUserToDb(user *pb.User) *models.User {
	return &models.User{
		ID:    user.GetId(),
		Email: user.GetEmail(),
		Name:  user.GetName(),
	}
}
