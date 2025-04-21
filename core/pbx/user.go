package pbx

import (
	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/volatiletech/null/v8"
)

func DbUserToProto(user *models.User) *pb.User {
	userPb := &pb.User{
		FirstName: user.FirstName,
	}

	// Handle nullable email
	if user.Email.Valid {
		userPb.Email = user.Email.String
	}

	// Avatar priority:
	// 1. Use stored AvatarID from Auth0 if available
	// 2. Fall back to Gravatar based on email
	if user.AvatarID.Valid && user.AvatarID.String != "" {
		userPb.Avatar = user.AvatarID.String
	} else if user.Email.Valid {
		// Fall back to Gravatar if we have an email
		userPb.Avatar = util.NewGravatarFromEmail(user.Email.String).GetURL()
	} else {
		// Default avatar when no email and no stored avatar
		userPb.Avatar = util.NewGravatarFromEmail("").GetURL()
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
		FirstName: user.GetFirstName(),
	}

	// Handle email conversion to null.String
	if user.GetEmail() != "" {
		userDb.Email = null.StringFrom(user.GetEmail())
	} else {
		userDb.Email = null.String{}
	}

	if user.GetName() != "" {
		userId, err := GetResourceIDByType(user.GetName(), RessourceTypeUsers)
		if err != nil {
			return nil
		}
		userDb.ID = userId
	}

	if user.GetLastName() != "" {
		userDb.LastName = null.StringFrom(user.GetLastName())
	} else {
		userDb.LastName = null.String{}
	}

	return userDb
}
