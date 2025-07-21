package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/rs/zerolog/log"
)

// FirebaseAuthHandler handles Firebase authentication-related routes
type FirebaseAuthHandler struct {
	firebaseAuth     coreauth.AuthService
	sessionManager   *auth.SCSSessionManager
	generatorService *services.GeneratorService
	db               *sql.DB
}

// NewFirebaseAuthHandler creates a new Firebase auth handler
func NewFirebaseAuthHandler(sessionManager *auth.SCSSessionManager) *FirebaseAuthHandler {
	return &FirebaseAuthHandler{
		sessionManager: sessionManager,
	}
}

// NewFirebaseAuthHandlerWithServices creates a new Firebase auth handler with all services
func NewFirebaseAuthHandlerWithServices(firebaseAuth coreauth.AuthService, sessionManager *auth.SCSSessionManager, generatorService *services.GeneratorService, db *sql.DB) *FirebaseAuthHandler {
	return &FirebaseAuthHandler{
		firebaseAuth:     firebaseAuth,
		sessionManager:   sessionManager,
		generatorService: generatorService,
		db:               db,
	}
}

// AuthSyncRequest represents the request body for /auth/sync
type AuthSyncRequest struct {
	IDToken string `json:"id_token"`
}

// AuthSyncResponse represents the response body for /auth/sync
type AuthSyncResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	User    *UserProfile `json:"user,omitempty"`
}

// UserProfile represents user profile data for the frontend
type UserProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AuthSync handles Firebase token validation and session creation
func (h *FirebaseAuthHandler) AuthSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req AuthSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to decode auth sync request")
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.IDToken == "" {
		h.sendErrorResponse(w, "ID token is required", http.StatusBadRequest)
		return
	}

	// Validate Firebase ID token
	userInfo, err := h.firebaseAuth.GetUserInfoFromToken(r.Context(), req.IDToken)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to validate Firebase ID token")
		h.sendErrorResponse(w, "Invalid ID token", http.StatusUnauthorized)
		return
	}

	log.Info().
		Str("firebase_uid", userInfo.ID).
		Str("email", userInfo.Email).
		Str("name", userInfo.Name).
		Msg("Firebase token validated successfully")

	// For now, we'll use the Firebase UID as the internal user ID
	// The API interceptor will handle user auto-creation when API calls are made
	internalUserID := userInfo.ID

	// Create session data
	sessionUserInfo := auth.SessionUserInfo{
		ID:        userInfo.ID,
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		Picture:   userInfo.Picture,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
	}

	// Calculate token expiry
	tokenExpiry := time.Now().Add(1 * time.Hour)

	// Create session
	err = h.sessionManager.CreateSession(w, r, internalUserID, sessionUserInfo, req.IDToken, tokenExpiry)
	if err != nil {
		log.Error().Err(err).Str("user_id", internalUserID).Msg("Failed to create session")
		h.sendErrorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Prepare response
	userProfile := &UserProfile{
		ID:        internalUserID,
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		Picture:   userInfo.Picture,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
	}

	response := AuthSyncResponse{
		Success: true,
		Message: "Authentication successful",
		User:    userProfile,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Info().
		Str("user_id", internalUserID).
		Str("firebase_uid", userInfo.ID).
		Str("email", userInfo.Email).
		Msg("User authenticated and session created")
}

// Logout handles user logout with enhanced error handling and support for multiple HTTP methods
func (h *FirebaseAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Support both POST and GET methods for logout
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		log.Warn().
			Str("method", r.Method).
			Str("endpoint", "/logout").
			Str("user_agent", r.Header.Get("User-Agent")).
			Msg("Invalid HTTP method for logout")

		http.Error(w, "Method not allowed. Use POST or GET.", http.StatusMethodNotAllowed)
		return
	}

	// Get current session info for auditing before destroying it
	userID := h.sessionManager.GetUserID(r)
	var sessionData *auth.SessionData
	if userID != "" {
		var err error
		sessionData, err = h.sessionManager.GetSession(r)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID).Msg("Failed to get session data during logout")
		}
	}

	// Check if user wants JSON response or redirect
	acceptsJSON := r.Header.Get("Accept") == "application/json" || r.Header.Get("Content-Type") == "application/json"
	isAjaxRequest := r.Header.Get("X-Requested-With") == "XMLHttpRequest"
	wantsJSON := acceptsJSON || isAjaxRequest || r.Method == http.MethodPost

	// Destroy session (always attempt to destroy, even if no active session)
	err := h.sessionManager.DestroySession(w, r)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("endpoint", "/logout").
			Str("user_agent", r.Header.Get("User-Agent")).
			Str("remote_addr", r.RemoteAddr).
			Msg("Failed to destroy session during logout")

		if wantsJSON {
			h.sendErrorResponse(w, "Logout failed due to server error", http.StatusInternalServerError)
		} else {
			// Redirect to home page even on session destroy error
			http.Redirect(w, r, "/?logout=error", http.StatusSeeOther)
		}
		return
	}

	// Comprehensive audit logging
	logEvent := log.Info().
		Str("user_id", userID).
		Str("endpoint", "/logout").
		Str("method", r.Method).
		Str("user_agent", r.Header.Get("User-Agent")).
		Str("remote_addr", r.RemoteAddr)

	if sessionData != nil {
		logEvent.
			Str("user_email", sessionData.UserInfo.Email).
			Str("user_name", sessionData.UserInfo.Name).
			Time("session_expires_at", sessionData.ExpiresAt)
	}

	logEvent.Msg("User logged out successfully")

	// Send appropriate response based on request type
	if wantsJSON {
		response := AuthSyncResponse{
			Success: true,
			Message: "Logout successful",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error().Err(err).Msg("Failed to encode logout response")
		}
	} else {
		// Redirect to home page with success indicator
		redirectURL := "/?logout=success"
		if returnURL := r.URL.Query().Get("return_url"); returnURL != "" {
			// Basic validation of return URL to prevent open redirects
			if returnURL == "/" || returnURL[0] == '/' && returnURL[1] != '/' {
				redirectURL = returnURL + "?logout=success"
			}
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

// Status returns the user's authentication status
func (h *FirebaseAuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID := h.sessionManager.GetUserID(r)
	if userID == "" {
		response := AuthSyncResponse{
			Success: false,
			Message: "Not authenticated",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get session data
	sessionData, err := h.sessionManager.GetSession(r)
	if err != nil {
		response := AuthSyncResponse{
			Success: false,
			Message: "Invalid session",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	userProfile := &UserProfile{
		ID:        userID,
		Name:      sessionData.UserInfo.Name,
		Email:     sessionData.UserInfo.Email,
		Picture:   sessionData.UserInfo.Picture,
		FirstName: sessionData.UserInfo.FirstName,
		LastName:  sessionData.UserInfo.LastName,
	}

	response := AuthSyncResponse{
		Success: true,
		Message: "Authenticated",
		User:    userProfile,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Note: User creation is now handled by the API interceptor when first API call is made

// sendErrorResponse sends a JSON error response
func (h *FirebaseAuthHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := AuthSyncResponse{
		Success: false,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
