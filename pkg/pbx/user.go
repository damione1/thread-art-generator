package pbx

import (
	"github.com/Damione1/thread-art-generator/pkg/db/models"
	"github.com/Damione1/thread-art-generator/pkg/pb"
	"github.com/Damione1/thread-art-generator/pkg/util"
)

func DbUserToProto(user *models.User) *pb.User {
	userPb := &pb.User{
		Email:     user.Email,
		FirstName: user.FirstName,
		Password:  "", // never return password
		Avatar:    util.NewGravatarFromEmail(user.Email).GetURL(),
	}

	if user.LastName.Valid {
		userPb.LastName = user.LastName.String
	}

	userPb.Name = GetResourceName([]Resource{
		{Type: RessourceTypeUsers, ID: user.ID},
	})

	return userPb
}

func ProtoUserToDb(user *pb.User) *models.User {
	userDb := &models.User{
		Email:     user.GetEmail(),
		FirstName: user.GetFirstName(),
	}

	if user.GetName() != "" {
		userId, err := GetResourceIDByType(user.GetName(), RessourceTypeUsers)
		if err != nil {
			return nil
		}
		userDb.ID = userId
	}

	if user.GetLastName() != "" {
		userDb.LastName.String = user.GetLastName()
		userDb.LastName.Valid = true
	}

	return userDb
}
