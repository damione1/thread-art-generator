package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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
		passwords:      coreauth.BcryptPasswords{},
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

func (h *PasswordAuthHandler) issueSession(w http.ResponseWriter, r *http.Request, identity coreauth.Identity) error {
	info := auth.SessionUserInfo{
		ID:    identity.UserID,
		Email: identity.Email,
		Name:  identity.Email,
	}
	return h.sessionManager.CreateSession(w, r, identity.UserID, info, "", time.Now().Add(24*time.Hour))
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
