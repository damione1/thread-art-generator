package pbx

import (
	"github.com/Damione1/portfolio-playground/db/models"
	"github.com/Damione1/portfolio-playground/pkg/pb"
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
