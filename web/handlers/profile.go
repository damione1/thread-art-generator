package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/web/client"
	"github.com/Damione1/thread-art-generator/web/middleware"
	"github.com/Damione1/thread-art-generator/web/templates"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ProfileHandler handles the profile page
func ProfileHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (middleware already ensures user is logged in)
		user := middleware.GetUserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If the request is a GET, render the profile page
		if r.Method == http.MethodGet {
			templates.Profile(user, "").Render(r.Context(), w)
			return
		}

		// If the request is a POST, handle the profile update
		if r.Method == http.MethodPost {
			// Parse the form
			err := r.ParseForm()
			if err != nil {
				templates.Profile(user, "Error parsing form").Render(r.Context(), w)
				return
			}

			// Get the form values
			firstName := r.FormValue("first_name")
			lastName := r.FormValue("last_name")
			email := r.FormValue("email")
			password := r.FormValue("password")
			confirmPassword := r.FormValue("confirm_password")
			updateFields := r.FormValue("update_fields")

			// Validate the form values
			validationErrors := &client.ValidationErrors{
				FieldErrors: make(map[string]string),
			}

			// Only validate fields that are being updated
			fields := strings.Split(updateFields, ",")
			if len(fields) == 0 || updateFields == "" {
				validationErrors.GeneralError = "No fields to update"
				errJson, _ := json.Marshal(validationErrors)
				templates.Profile(user, string(errJson)).Render(r.Context(), w)
				return
			}

			// Basic validation for each field
			for _, field := range fields {
				switch field {
				case "first_name":
					if firstName == "" {
						validationErrors.FieldErrors["first_name"] = "First name cannot be blank"
					}
				case "email":
					if email == "" {
						validationErrors.FieldErrors["email"] = "Email cannot be blank"
					}
				case "password":
					if password == "" {
						validationErrors.FieldErrors["password"] = "Password cannot be blank"
					}
					if password != confirmPassword {
						validationErrors.FieldErrors["confirm_password"] = "Passwords do not match"
					}
				}
			}

			if validationErrors.HasErrors() {
				// If there are validation errors, render the form with errors
				errJson, _ := json.Marshal(validationErrors)
				templates.Profile(user, string(errJson)).Render(r.Context(), w)
				return
			}

			// Create update request with field mask
			updateRequest := &pb.UpdateUserRequest{
				User: &pb.User{
					Name:      user.GetName(),
					FirstName: firstName,
					LastName:  lastName,
					Email:     email,
					Password:  password,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: fields,
				},
			}

			// Create a context with timeout
			ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			// Call the UpdateUser API
			updatedUser, err := grpcClient.GetClient().UpdateUser(ctx, updateRequest)
			if err != nil {
				// Handle error from API
				errDetails := client.ExtractErrorDetails(err)
				errJson, _ := json.Marshal(errDetails)
				templates.Profile(user, string(errJson)).Render(r.Context(), w)
				return
			}

			// Success message
			successMsg := &client.ValidationErrors{
				SuccessMessage: "Profile updated successfully",
			}

			// Convert success message to JSON string
			successJson, _ := json.Marshal(successMsg)

			// Render the profile page with updated user and success message
			templates.Profile(updatedUser, string(successJson)).Render(r.Context(), w)
			return
		}

		// Method not allowed
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
