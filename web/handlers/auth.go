package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
			// Check for validated=true parameter and get email from cookie
			email := client.GetEmailFromCookie(r)
			validated := r.URL.Query().Get("validated") == "true"

			if validated {
				// Create success message
				validationErrors := &client.ValidationErrors{
					SuccessMessage: "Your email has been validated. You can now log in.",
				}
				component := templates.LoginWithData(templates.NewLoginFormData(validationErrors, email))
				component.Render(r.Context(), w)
			} else {
				component := templates.LoginWithFormValues("", email)
				component.Render(r.Context(), w)
			}
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

			// Store email in cookie for future use
			if email != "" {
				client.SetEmailCookie(w, email)
			}

			// Validate the email and password
			if email == "" || password == "" {
				// Create validation errors
				validationErrors := &client.ValidationErrors{
					FieldErrors: make(map[string]string),
				}

				if email == "" {
					validationErrors.FieldErrors["email"] = "cannot be blank"
				}
				if password == "" {
					validationErrors.FieldErrors["password"] = "cannot be blank"
				}

				// Create form data and render
				formData := templates.NewLoginFormData(validationErrors, email)
				component := templates.LoginWithData(formData)
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

				// Create form data and render
				formData := templates.NewLoginFormData(errorDetails, email)
				component := templates.LoginWithData(formData)
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

			// Store email in cookie for future use
			if email != "" {
				client.SetEmailCookie(w, email)
			}

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
				// Create form data and render
				formData := templates.NewRegisterFormData(validationErrors, firstName, lastName, email)
				component := templates.RegisterWithData(formData)
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

				// Create form data and render
				formData := templates.NewRegisterFormData(errorDetails, firstName, lastName, email)
				component := templates.RegisterWithData(formData)
				component.Render(r.Context(), w)
				return
			}

			// Redirect to the email validation page with the email pre-filled
			http.Redirect(w, r, "/validate-email?email="+url.QueryEscape(email), http.StatusSeeOther)
			return
		}

		// If the request is not a GET or POST, return 405 Method Not Allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ValidateEmailHandler handles the email validation page
func ValidateEmailHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is already logged in
		user := middleware.GetUserFromContext(r.Context())
		if user != nil {
			// Already logged in, redirect to dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// If the request is a GET, render the email validation page
		if r.Method == http.MethodGet {
			// Check for email in URL parameter first
			email := r.URL.Query().Get("email")

			// If no email in URL, try to get from cookie
			if email == "" {
				email = client.GetEmailFromCookie(r)
			} else {
				// If email is in URL, save it to cookie
				client.SetEmailCookie(w, email)
			}

			component := templates.EmailValidationWithFormValues("", email)
			component.Render(r.Context(), w)
			return
		}

		// If the request is a POST, handle the validation
		if r.Method == http.MethodPost {
			// Parse the form
			err := r.ParseForm()
			if err != nil {
				component := templates.EmailValidation("Error parsing form")
				component.Render(r.Context(), w)
				return
			}

			// Get the email and validation number
			email := r.FormValue("email")
			validationNumberStr := r.FormValue("validationNumber")

			// Store email in cookie for future use
			if email != "" {
				client.SetEmailCookie(w, email)
			}

			// Validate the email and validation number
			validationErrors := &client.ValidationErrors{
				FieldErrors: make(map[string]string),
			}

			if email == "" {
				validationErrors.FieldErrors["email"] = "cannot be blank"
			}
			if validationNumberStr == "" {
				validationErrors.FieldErrors["validationNumber"] = "cannot be blank"
			}

			if len(validationErrors.FieldErrors) > 0 {
				// Create form data and render
				formData := templates.NewEmailValidationFormData(validationErrors, email)
				component := templates.EmailValidationWithData(formData)
				component.Render(r.Context(), w)
				return
			}

			// Create a context with timeout
			ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			// Convert validation number to int64
			var validationNumber int64
			_, err = fmt.Sscanf(validationNumberStr, "%d", &validationNumber)
			if err != nil {
				validationErrors.FieldErrors["validationNumber"] = "must be a valid number"
				formData := templates.NewEmailValidationFormData(validationErrors, email)
				component := templates.EmailValidationWithData(formData)
				component.Render(r.Context(), w)
				return
			}

			// Try to validate the email
			_, err = grpcClient.GetClient().ValidateEmail(ctx, &pb.ValidateEmailRequest{
				Email:            email,
				ValidationNumber: validationNumber,
			})

			if err != nil {
				// Extract error details directly from the gRPC error
				errorDetails := client.ExtractErrorDetails(err)

				// Debug logging to see what's happening
				fmt.Printf("Error from API: %v\n", err)
				fmt.Printf("Extracted error details: General=%s, Fields=%v\n",
					errorDetails.GeneralError, errorDetails.FieldErrors)

				// Create form data and render
				formData := templates.NewEmailValidationFormData(errorDetails, email)
				component := templates.EmailValidationWithData(formData)
				component.Render(r.Context(), w)
				return
			}

			// Redirect to the login page with success message
			http.Redirect(w, r, "/login?validated=true", http.StatusSeeOther)
			return
		}

		// If the request is not a GET or POST, return 405 Method Not Allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ResendValidationHandler handles the resend validation code request
func ResendValidationHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the request is not a POST, return 405 Method Not Allowed
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse the form
		err := r.ParseForm()
		if err != nil {
			http.Redirect(w, r, "/validate-email", http.StatusSeeOther)
			return
		}

		// Get the email
		email := r.FormValue("email")
		if email == "" {
			http.Redirect(w, r, "/validate-email", http.StatusSeeOther)
			return
		}

		// Store email in cookie
		client.SetEmailCookie(w, email)

		// Create a context with timeout
		ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Send validation email
		_, err = grpcClient.GetClient().SendValidationEmail(ctx, &pb.SendValidationEmailRequest{
			Email: email,
		})

		if err != nil {
			// Extract error details
			errorDetails := client.ExtractErrorDetails(err)

			// Debug logging
			fmt.Printf("Error from API: %v\n", err)
			fmt.Printf("Extracted error details: General=%s, Fields=%v\n",
				errorDetails.GeneralError, errorDetails.FieldErrors)

			// Create validation errors and add to URL as query parameter
			validationErrors := &client.ValidationErrors{
				GeneralError: "Failed to send validation email. Please try again.",
			}

			if errorDetails.HasErrors() {
				validationErrors = errorDetails
			}

			// Create form data and render
			formData := templates.NewEmailValidationFormData(validationErrors, email)
			component := templates.EmailValidationWithData(formData)
			component.Render(r.Context(), w)
			return
		}

		// Create success message
		validationErrors := &client.ValidationErrors{
			SuccessMessage: "Validation code has been sent to your email.",
		}

		// Create form data and render
		formData := templates.NewEmailValidationFormData(validationErrors, email)
		component := templates.EmailValidationWithData(formData)
		component.Render(r.Context(), w)
	}
}

// TokensForClientHandler returns tokens in JSON format for the client
func TokensForClientHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Must be authenticated
		user := middleware.GetUserFromContext(r.Context())
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Get access and refresh tokens from cookies
		accessToken := client.GetSessionToken(r)
		refreshToken := client.GetRefreshToken(r)

		// If tokens are missing, redirect to login
		if accessToken == "" || refreshToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Return tokens and expiry time in JSON format
		// We don't have exact expiry, so set to 1 hour from now
		expiryTime := time.Now().Add(1 * time.Hour).Unix()

		// Write JSON response
		fmt.Fprintf(w, `{"access_token":"%s","refresh_token":"%s","access_token_expire_time":%d}`,
			accessToken, refreshToken, expiryTime)
	}
}

// RefreshTokenAPIHandler handles token refresh for API clients
func RefreshTokenAPIHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse the request body
		var requestBody struct {
			RefreshToken string `json:"refresh_token"`
		}

		// Try to decode the request body into the struct
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil || requestBody.RefreshToken == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Create a context with timeout
		ctx, cancel := client.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Try to refresh the token
		refreshResp, err := grpcClient.RefreshToken(ctx, requestBody.RefreshToken)
		if err != nil {
			// Extract error details from the gRPC error
			errorDetails := client.ExtractErrorDetails(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(errorDetails)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Return tokens and expiry time in JSON format
		response := map[string]interface{}{
			"access_token":             refreshResp.AccessToken,
			"refresh_token":            refreshResp.RefreshToken,
			"access_token_expire_time": refreshResp.AccessTokenExpireTime.Seconds,
		}

		// Write JSON response
		json.NewEncoder(w).Encode(response)
	}
}
