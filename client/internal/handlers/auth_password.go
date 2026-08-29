package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/rs/zerolog/log"
)

type passwordAuthRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// PasswordAuthHandler is cookie email/password login against Postgres.
type PasswordAuthHandler struct {
	identities     coreauth.Identities
	passwords      coreauth.Passwords
	sessionManager *auth.SCSSessionManager
}

func NewPasswordAuthHandler(identities coreauth.Identities, sessionManager *auth.SCSSessionManager) *PasswordAuthHandler {
	return &PasswordAuthHandler{
		identities:     identities,
		passwords:      coreauth.Argon2idPasswords{},
		sessionManager: sessionManager,
	}
}

func (h *PasswordAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req passwordAuthRequest
	if err := decodePasswordAuth(r, &req); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, false, err.Error())
		return
	}
	identity, hash, err := h.identities.ByEmail(r.Context(), req.Email)
	if err != nil {
		writeAuthJSON(w, http.StatusUnauthorized, false, "incorrect email or password")
		return
	}
	if hash == "" || h.passwords.Check(req.Password, hash) != nil {
		writeAuthJSON(w, http.StatusUnauthorized, false, "incorrect email or password")
		return
	}
	if err := h.issueSession(w, r, identity); err != nil {
		log.Error().Err(err).Str("user_id", identity.UserID).Msg("password login session failed")
		writeAuthJSON(w, http.StatusInternalServerError, false, "failed to create session")
		return
	}
	writeAuthJSON(w, http.StatusOK, true, "Authentication successful")
}

func (h *PasswordAuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req passwordAuthRequest
	if err := decodePasswordAuth(r, &req); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, false, err.Error())
		return
	}
	if len(req.Password) < 8 {
		writeAuthJSON(w, http.StatusBadRequest, false, "password must be at least 8 characters")
		return
	}
	hash, err := h.passwords.Hash(req.Password)
	if err != nil {
		writeAuthJSON(w, http.StatusInternalServerError, false, "failed to hash password")
		return
	}
	identity, err := h.identities.Create(r.Context(), req.Email, hash, req.FirstName, req.LastName)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("password signup failed")
		writeAuthJSON(w, http.StatusConflict, false, "email already exists")
		return
	}
	if err := h.issueSession(w, r, identity); err != nil {
		log.Error().Err(err).Str("user_id", identity.UserID).Msg("password signup session failed")
		writeAuthJSON(w, http.StatusInternalServerError, false, "failed to create session")
		return
	}
	writeAuthJSON(w, http.StatusOK, true, "Authentication successful")
}

func (h *PasswordAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use POST or GET.", http.StatusMethodNotAllowed)
		return
	}

	userID := h.sessionManager.GetUserID(r)
	acceptsJSON := r.Header.Get("Accept") == "application/json" || r.Header.Get("Content-Type") == "application/json"
	isAjaxRequest := r.Header.Get("X-Requested-With") == "XMLHttpRequest"
	wantsJSON := acceptsJSON || isAjaxRequest || r.Method == http.MethodPost

	if err := h.sessionManager.DestroySession(w, r); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to destroy session during logout")
		if wantsJSON {
			writeAuthJSON(w, http.StatusInternalServerError, false, "Logout failed due to server error")
		} else {
			http.Redirect(w, r, "/?logout=error", http.StatusSeeOther)
		}
		return
	}

	if wantsJSON {
		writeAuthJSON(w, http.StatusOK, true, "Logout successful")
		return
	}
	http.Redirect(w, r, "/?logout=success", http.StatusSeeOther)
}

func (h *PasswordAuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID := h.sessionManager.GetUserID(r)
	if userID == "" {
		writeAuthJSON(w, http.StatusOK, false, "Not authenticated")
		return
	}
	sessionData, err := h.sessionManager.GetSession(r)
	if err != nil {
		writeAuthJSON(w, http.StatusOK, false, "Invalid session")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(AuthSyncResponse{
		Success: true,
		Message: "Authenticated",
		User: &UserProfile{
			ID:        userID,
			Name:      sessionData.UserInfo.Name,
			Email:     sessionData.UserInfo.Email,
			Picture:   sessionData.UserInfo.Picture,
			FirstName: sessionData.UserInfo.FirstName,
			LastName:  sessionData.UserInfo.LastName,
		},
	})
}

func (h *PasswordAuthHandler) issueSession(w http.ResponseWriter, r *http.Request, identity coreauth.Identity) error {
	return h.sessionManager.CreateSession(w, r, identity.UserID, auth.UserInfoFromIdentity(identity))
}

func decodePasswordAuth(r *http.Request, req *passwordAuthRequest) error {
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			return errors.New("invalid request body")
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return errors.New("invalid form")
		}
		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
		req.FirstName = r.FormValue("first_name")
		req.LastName = r.FormValue("last_name")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		return errors.New("email and password are required")
	}
	return nil
}

func writeAuthJSON(w http.ResponseWriter, status int, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(AuthSyncResponse{Success: success, Message: message})
}

type AuthSyncResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	User    *UserProfile `json:"user,omitempty"`
}

type UserProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
