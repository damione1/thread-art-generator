package handlers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/mail"
	"github.com/rs/zerolog/log"
)

type passwordAuthRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Token           string `json:"token"`
}

// PasswordAuthHandler is cookie email/password login against Postgres.
type PasswordAuthHandler struct {
	identities     coreauth.Identities
	passwords      coreauth.Passwords
	tokens         coreauth.Tokens
	emails         *mail.Emails
	sessionManager *auth.SCSSessionManager
	limiter        *coreauth.AuthLimiter
}

func NewPasswordAuthHandler(
	identities coreauth.Identities,
	tokens coreauth.Tokens,
	emails *mail.Emails,
	sessionManager *auth.SCSSessionManager,
) *PasswordAuthHandler {
	return &PasswordAuthHandler{
		identities:     identities,
		passwords:      coreauth.Argon2idPasswords{},
		tokens:         tokens,
		emails:         emails,
		sessionManager: sessionManager,
		limiter:        coreauth.DefaultAuthLimiter(),
	}
}

func (h *PasswordAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req passwordAuthRequest
	if err := decodePasswordAuth(r, &req, true); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, false, err.Error())
		return
	}
	if !h.allowAuth(r, "login", req.Email) {
		writeAuthJSON(w, http.StatusTooManyRequests, false, "too many attempts, try again later")
		return
	}
	identity, hash, err := h.identities.ByEmail(r.Context(), req.Email)
	if err != nil {
		coreauth.Argon2idPasswords{}.CheckDummy(req.Password)
		writeAuthJSON(w, http.StatusUnauthorized, false, "incorrect email or password")
		return
	}
	if hash == "" || h.passwords.Check(req.Password, hash) != nil {
		writeAuthJSON(w, http.StatusUnauthorized, false, "incorrect email or password")
		return
	}
	if !identity.Active {
		writeAuthJSON(w, http.StatusForbidden, false, "please verify your email before signing in")
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
	if err := decodePasswordAuth(r, &req, true); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, false, err.Error())
		return
	}
	if !h.allowAuth(r, "signup", req.Email) {
		writeAuthJSON(w, http.StatusTooManyRequests, false, "too many attempts, try again later")
		return
	}
	if err := coreauth.ValidatePasswordLength(req.Password); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, false, err.Error())
		return
	}
	if strings.TrimSpace(req.FirstName) == "" {
		writeAuthJSON(w, http.StatusBadRequest, false, "first name is required")
		return
	}

	existing, _, err := h.identities.ByEmail(r.Context(), req.Email)
	if err == nil {
		if !existing.Active {
			if sendErr := h.sendVerify(r, existing); sendErr != nil {
				log.Error().Err(sendErr).Str("email", req.Email).Msg("resend verify on signup failed")
				writeAuthJSON(w, http.StatusInternalServerError, false, "failed to send verification email")
				return
			}
			writeAuthJSON(w, http.StatusOK, true, "Check your email to confirm your account")
			return
		}
		writeAuthJSON(w, http.StatusConflict, false, "email already exists")
		return
	}
	if !errors.Is(err, coreauth.ErrIdentityNotFound) {
		log.Error().Err(err).Str("email", req.Email).Msg("signup lookup failed")
		writeAuthJSON(w, http.StatusInternalServerError, false, "failed to create account")
		return
	}

	hash, err := h.passwords.Hash(req.Password)
	if err != nil {
		writeAuthJSON(w, http.StatusInternalServerError, false, "failed to hash password")
		return
	}
	identity, err := h.identities.Create(r.Context(), req.Email, hash, req.FirstName, req.LastName)
	if err != nil {
		if errors.Is(err, coreauth.ErrEmailTaken) {
			writeAuthJSON(w, http.StatusConflict, false, "email already exists")
			return
		}
		log.Warn().Err(err).Str("email", req.Email).Msg("password signup failed")
		writeAuthJSON(w, http.StatusInternalServerError, false, "failed to create account")
		return
	}
	if err := h.sendVerify(r, identity); err != nil {
		log.Error().Err(err).Str("user_id", identity.UserID).Msg("verification email failed")
		writeAuthJSON(w, http.StatusInternalServerError, false, "failed to send verification email")
		return
	}
	writeAuthJSON(w, http.StatusOK, true, "Check your email to confirm your account")
}

func (h *PasswordAuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req passwordAuthRequest
	if err := decodePasswordAuth(r, &req, false); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, false, err.Error())
		return
	}
	if !h.allowAuth(r, "forgot", req.Email) {
		writeAuthJSON(w, http.StatusTooManyRequests, false, "too many attempts, try again later")
		return
	}
	const generic = "If that email exists, we sent a reset link"
	identity, hash, err := h.identities.ByEmail(r.Context(), req.Email)
	if err != nil || hash == "" {
		if err != nil && !errors.Is(err, coreauth.ErrIdentityNotFound) {
			log.Error().Err(err).Str("email", req.Email).Msg("forgot password lookup failed")
		} else {
			log.Info().Str("email", req.Email).Bool("found", err == nil).Bool("has_password", hash != "").Msg("forgot password skipped")
		}
		h.finishAuth(w, r, http.StatusOK, true, generic, "/forgot-password?sent=1")
		return
	}
	token, err := h.tokens.Issue(r.Context(), identity.UserID, coreauth.TokenReset, coreauth.ResetTTL)
	if err != nil {
		log.Error().Err(err).Str("user_id", identity.UserID).Msg("reset token issue failed")
		h.finishAuth(w, r, http.StatusOK, true, generic, "/forgot-password?sent=1")
		return
	}
	if err := h.emails.SendPasswordReset(r.Context(), mail.Address{
		Name:  displayName(identity),
		Email: identity.Email,
	}, token); err != nil {
		log.Error().Err(err).Str("user_id", identity.UserID).Str("email", identity.Email).Msg("reset email failed")
		h.finishAuth(w, r, http.StatusOK, true, generic, "/forgot-password?sent=1")
		return
	}
	log.Info().Str("user_id", identity.UserID).Str("email", identity.Email).Msg("reset email sent")
	h.finishAuth(w, r, http.StatusOK, true, generic, "/forgot-password?sent=1")
}

func (h *PasswordAuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	req, err := decodeResetPassword(r)
	if err != nil {
		h.resetPasswordError(w, r, req.Token, err.Error())
		return
	}
	if req.Token == "" {
		h.resetPasswordError(w, r, "", "reset token is required")
		return
	}
	if !h.allowAuth(r, "reset", req.Token) {
		if wantsAuthJSON(r) {
			writeAuthJSON(w, http.StatusTooManyRequests, false, "too many attempts, try again later")
			return
		}
		h.resetPasswordError(w, r, req.Token, "too many attempts, try again later")
		return
	}
	if err := coreauth.ValidatePasswordLength(req.Password); err != nil {
		h.resetPasswordError(w, r, req.Token, err.Error())
		return
	}
	if req.ConfirmPassword != "" && req.ConfirmPassword != req.Password {
		h.resetPasswordError(w, r, req.Token, "passwords do not match")
		return
	}
	userID, err := h.tokens.Consume(r.Context(), req.Token, coreauth.TokenReset)
	if err != nil {
		h.resetPasswordError(w, r, req.Token, "this reset link is invalid or expired")
		return
	}
	hash, err := h.passwords.Hash(req.Password)
	if err != nil {
		h.resetPasswordError(w, r, req.Token, "failed to hash password")
		return
	}
	if err := h.identities.UpdatePassword(r.Context(), userID, hash); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("reset password update failed")
		h.resetPasswordError(w, r, req.Token, "failed to update password")
		return
	}
	if _, err := h.identities.BumpSessionVersion(r.Context(), userID); err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("reset password session bump failed")
	}
	if err := h.identities.SetActive(r.Context(), userID, true); err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("reset password activate failed")
	}
	log.Info().Str("user_id", userID).Msg("password reset")
	h.finishAuth(w, r, http.StatusOK, true, "Password updated", "/login")
}

func decodeResetPassword(r *http.Request) (passwordAuthRequest, error) {
	var req passwordAuthRequest
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return req, errors.New("invalid request")
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return req, errors.New("invalid form")
		}
		req.Token = r.FormValue("token")
		req.Password = r.FormValue("password")
		req.ConfirmPassword = r.FormValue("confirm_password")
	}
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		req.Token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	return req, nil
}

func (h *PasswordAuthHandler) resetPasswordError(w http.ResponseWriter, r *http.Request, token, msg string) {
	if wantsAuthJSON(r) {
		writeAuthJSON(w, http.StatusBadRequest, false, msg)
		return
	}
	u := "/reset-password?error=" + url.QueryEscape(msg)
	if token != "" {
		u = "/reset-password?token=" + url.QueryEscape(token) + "&error=" + url.QueryEscape(msg)
	}
	http.Redirect(w, r, u, http.StatusSeeOther)
}

func (h *PasswordAuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Redirect(w, r, "/login?error=missing_token", http.StatusSeeOther)
		return
	}
	userID, err := h.tokens.Consume(r.Context(), token, coreauth.TokenVerify)
	if err != nil {
		http.Redirect(w, r, "/login?error=invalid_token", http.StatusSeeOther)
		return
	}
	if err := h.identities.SetActive(r.Context(), userID, true); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("activate account failed")
		http.Redirect(w, r, "/login?error=verify_failed", http.StatusSeeOther)
		return
	}
	identity, _, err := h.identities.ByID(r.Context(), userID)
	if err != nil {
		http.Redirect(w, r, "/login?verified=1", http.StatusSeeOther)
		return
	}
	if err := h.issueSession(w, r, identity); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("verify session failed")
		http.Redirect(w, r, "/login?verified=1", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *PasswordAuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req passwordAuthRequest
	if err := decodePasswordAuth(r, &req, false); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, false, err.Error())
		return
	}
	if !h.allowAuth(r, "resend", req.Email) {
		writeAuthJSON(w, http.StatusTooManyRequests, false, "too many attempts, try again later")
		return
	}
	const generic = "If that email exists, we sent a confirmation link"
	identity, _, err := h.identities.ByEmail(r.Context(), req.Email)
	if err != nil || identity.Active {
		h.finishAuth(w, r, http.StatusOK, true, generic, "/check-email?sent=1")
		return
	}
	if err := h.sendVerify(r, identity); err != nil {
		log.Error().Err(err).Str("user_id", identity.UserID).Msg("resend verification failed")
		h.finishAuth(w, r, http.StatusOK, true, generic, "/check-email?sent=1")
		return
	}
	log.Info().Str("user_id", identity.UserID).Str("email", identity.Email).Msg("verification email sent")
	h.finishAuth(w, r, http.StatusOK, true, generic, "/check-email?sent=1")
}

func (h *PasswordAuthHandler) sendVerify(r *http.Request, identity coreauth.Identity) error {
	token, err := h.tokens.Issue(r.Context(), identity.UserID, coreauth.TokenVerify, coreauth.VerifyTTL)
	if err != nil {
		return err
	}
	return h.emails.SendVerifyAccount(r.Context(), mail.Address{
		Name:  displayName(identity),
		Email: identity.Email,
	}, token)
}

func (h *PasswordAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	userID := h.sessionManager.GetUserID(r)
	if err := h.sessionManager.DestroySession(w, r); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to destroy session during logout")
		if wantsAuthJSON(r) {
			writeAuthJSON(w, http.StatusInternalServerError, false, "Logout failed due to server error")
			return
		}
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}
	if wantsAuthJSON(r) {
		writeAuthJSON(w, http.StatusOK, true, "Logout successful")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	return h.sessionManager.CreateSession(w, r, identity.UserID, auth.UserInfoFromIdentity(identity), identity.SessionVersion)
}

func (h *PasswordAuthHandler) allowAuth(r *http.Request, action, key string) bool {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || ip == "" {
		ip = r.RemoteAddr
	}
	return h.limiter.Allow(action + ":" + ip + ":" + strings.ToLower(strings.TrimSpace(key)))
}

func displayName(identity coreauth.Identity) string {
	name := strings.TrimSpace(identity.FirstName + " " + identity.LastName)
	if name != "" {
		return name
	}
	return identity.Email
}

func decodePasswordAuth(r *http.Request, req *passwordAuthRequest, requirePassword bool) error {
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
		req.Token = r.FormValue("token")
		req.ConfirmPassword = r.FormValue("confirm_password")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	if req.Email == "" {
		return errors.New("email is required")
	}
	if requirePassword && req.Password == "" {
		return errors.New("email and password are required")
	}
	return nil
}

func writeAuthJSON(w http.ResponseWriter, status int, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(AuthSyncResponse{Success: success, Message: message})
}

func wantsAuthJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	ct := r.Header.Get("Content-Type")
	return strings.Contains(accept, "application/json") || strings.Contains(ct, "application/json") || r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (h *PasswordAuthHandler) finishAuth(w http.ResponseWriter, r *http.Request, status int, success bool, message, redirect string) {
	if wantsAuthJSON(r) {
		writeAuthJSON(w, status, success, message)
		return
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
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
