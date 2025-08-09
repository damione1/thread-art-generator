package services

import (
	"github.com/Damione1/thread-art-generator/client/internal/auth"
)

// User represents the data returned from the API
type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
}

// ToAPIUser converts our User to an auth.APIUser
func (u *User) ToAPIUser() *auth.APIUser {
	return &auth.APIUser{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Avatar:    u.Avatar,
	}
}