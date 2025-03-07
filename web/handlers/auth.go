package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/web/client"
	"github.com/Damione1/thread-art-generator/web/middleware"
	"github.com/Damione1/thread-art-generator/web/templates"
)

// LoginHandler handles the login page
func LoginHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is already logged in
		user := middleware.GetUserFromContext(r.Context())
		if user != nil {
			// Already logged in, redirect to dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// If the request is a GET, render the login page
		if r.Method == http.MethodGet {
			component := templates.Login("")
			component.Render(r.Context(), w)
			return
		}

		// If the request is a POST, handle the login
		if r.Method == http.MethodPost {
			// Parse the form
			err := r.ParseForm()
			if err != nil {
				component := templates.Login("Error parsing form")
				component.Render(r.Context(), w)
				return
			}

			// Get the email and password
			email := r.FormValue("email")
			password := r.FormValue("password")

			// Validate the email and password
			if email == "" || password == "" {
				validationErrors := &client.ValidationErrors{
					FieldErrors: make(map[string]string),
				}

				if email == "" {
					validationErrors.FieldErrors["email"] = "cannot be blank"
				}
				if password == "" {
					validationErrors.FieldErrors["password"] = "cannot be blank"
				}

				component := templates.LoginWithValidation(validationErrors, email)
				component.Render(r.Context(), w)
				return
			}

			// Create a context with timeout
			ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			// Try to login
			session, err := grpcClient.GetClient().CreateSession(ctx, &pb.CreateSessionRequest{
				Email:    email,
				Password: password,
			})

			if err != nil {
				// Extract error details directly from the gRPC error
				errorDetails := client.ExtractErrorDetails(err)

				// Debug logging to see what's happening
				fmt.Printf("Error from API: %v\n", err)
				fmt.Printf("Extracted error details: General=%s, Fields=%v\n",
					errorDetails.GeneralError, errorDetails.FieldErrors)

				component := templates.LoginWithValidation(errorDetails, email)
				component.Render(r.Context(), w)
				return
			}

			// Set the session cookies
			client.SetSessionCookies(w, session)

			// Redirect to the dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// If the request is not a GET or POST, return 405 Method Not Allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// LogoutHandler handles the logout
func LogoutHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the refresh token
		refreshToken := client.GetRefreshToken(r)
		if refreshToken != "" {
			// Create a context with timeout
			ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			// Delete the session
			_, _ = grpcClient.GetClient().DeleteSession(ctx, &pb.DeleteSessionRequest{
				RefreshToken: refreshToken,
			})
		}

		// Clear the session cookies
		client.ClearSessionCookies(w)

		// Redirect to the home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// RegisterHandler handles the registration page
func RegisterHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is already logged in
		user := middleware.GetUserFromContext(r.Context())
		if user != nil {
			// Already logged in, redirect to dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// If the request is a GET, render the registration page
		if r.Method == http.MethodGet {
			component := templates.Register("")
			component.Render(r.Context(), w)
			return
		}

		// If the request is a POST, handle the registration
		if r.Method == http.MethodPost {
			// Parse the form
			err := r.ParseForm()
			if err != nil {
				component := templates.Register("Error parsing form")
				component.Render(r.Context(), w)
				return
			}

			// Get the form values
			firstName := r.FormValue("first_name")
			lastName := r.FormValue("last_name")
			email := r.FormValue("email")
			password := r.FormValue("password")
			confirmPassword := r.FormValue("confirm_password")

			// Validate the form values
			validationErrors := &client.ValidationErrors{
				FieldErrors: make(map[string]string),
			}

			if firstName == "" {
				validationErrors.FieldErrors["first_name"] = "cannot be blank"
			}
			if lastName == "" {
				validationErrors.FieldErrors["last_name"] = "cannot be blank"
			}
			if email == "" {
				validationErrors.FieldErrors["email"] = "cannot be blank"
			}
			if password == "" {
				validationErrors.FieldErrors["password"] = "cannot be blank"
			}
			if confirmPassword == "" {
				validationErrors.FieldErrors["confirm_password"] = "cannot be blank"
			}

			if password != confirmPassword {
				validationErrors.FieldErrors["confirm_password"] = "passwords do not match"
			}

			if len(validationErrors.FieldErrors) > 0 {
				component := templates.RegisterWithValidation(validationErrors, firstName, lastName, email)
				component.Render(r.Context(), w)
				return
			}

			// Create a context with timeout
			ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			// Try to create the user
			_, err = grpcClient.GetClient().CreateUser(ctx, &pb.CreateUserRequest{
				User: &pb.User{
					FirstName: firstName,
					LastName:  lastName,
					Email:     email,
					Password:  password,
				},
			})

			if err != nil {
				// Debug logging for raw error
				fmt.Printf("RegisterHandler - Raw error: %v\n", err)
				fmt.Printf("RegisterHandler - Error type: %T\n", err)

				// Format error as JSON for debugging
				jsonError := client.ParseGRPCError(err)
				fmt.Printf("RegisterHandler - Error as JSON: %s\n", jsonError)

				// Extract error details directly from the gRPC error
				errorDetails := client.ExtractErrorDetails(err)

				// Debug logging to see what's happening
				fmt.Printf("Error from API: %v\n", err)
				fmt.Printf("Extracted error details: General=%s, Fields=%v\n",
					errorDetails.GeneralError, errorDetails.FieldErrors)

				// Fallback if no error details were extracted
				if !errorDetails.HasErrors() {
					fmt.Printf("No error details extracted, using fallback\n")
					errorDetails.GeneralError = "An error occurred while creating your account. Please try again."
				}

				component := templates.RegisterWithValidation(errorDetails, firstName, lastName, email)
				component.Render(r.Context(), w)
				return
			}

			// Redirect to the login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If the request is not a GET or POST, return 405 Method Not Allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
