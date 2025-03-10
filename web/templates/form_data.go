package templates

import (
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/web/client"
)

// FormData contains common form data shared across different forms
type FormData struct {
	Errors *client.ValidationErrors
}

// LoginFormData contains data for the login form
type LoginFormData struct {
	FormData
	Email string
}

// RegisterFormData contains data for the registration form
type RegisterFormData struct {
	FormData
	FirstName string
	LastName  string
	Email     string
}

// EmailValidationFormData contains data for the email validation form
type EmailValidationFormData struct {
	FormData
	Email string
}

// ProfileFormData contains data for the profile form
type ProfileFormData struct {
	FormData
	User *pb.User
}

// NewLoginFormData creates a new LoginFormData instance
func NewLoginFormData(errors *client.ValidationErrors, email string) LoginFormData {
	return LoginFormData{
		FormData: FormData{Errors: errors},
		Email:    email,
	}
}

// NewRegisterFormData creates a new RegisterFormData instance
func NewRegisterFormData(errors *client.ValidationErrors, firstName, lastName, email string) RegisterFormData {
	return RegisterFormData{
		FormData:  FormData{Errors: errors},
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
}

// NewEmailValidationFormData creates a new EmailValidationFormData instance
func NewEmailValidationFormData(errors *client.ValidationErrors, email string) EmailValidationFormData {
	return EmailValidationFormData{
		FormData: FormData{Errors: errors},
		Email:    email,
	}
}

// NewProfileFormData creates a new ProfileFormData instance
func NewProfileFormData(errors *client.ValidationErrors, user *pb.User) ProfileFormData {
	return ProfileFormData{
		FormData: FormData{Errors: errors},
		User:     user,
	}
}
