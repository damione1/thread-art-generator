package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	pages "github.com/Damione1/thread-art-generator/client/internal/templates/pages"
	coreauth "github.com/Damione1/thread-art-generator/core/auth"
	"github.com/Damione1/thread-art-generator/core/mail"
	"github.com/rs/zerolog/log"
)

type SettingsHandler struct {
	identities     coreauth.Identities
	passwords      coreauth.Passwords
	tokens         coreauth.Tokens
	emails         *mail.Emails
	sessionManager *auth.SCSSessionManager
}

func NewSettingsHandler(identities coreauth.Identities, tokens coreauth.Tokens, emails *mail.Emails, sessionManager *auth.SCSSessionManager) *SettingsHandler {
	return &SettingsHandler{
		identities:     identities,
		passwords:      coreauth.Argon2idPasswords{},
		tokens:         tokens,
		emails:         emails,
		sessionManager: sessionManager,
	}
}

func (h *SettingsHandler) Page(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	identity, _, err := h.identities.ByID(r.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("settings load user failed")
		http.Error(w, "Failed to load account", http.StatusInternalServerError)
		return
	}

	pageData := templates.NewPageDataFromRequest(r, "Settings - ThreadArt", "settings").
		WithData(&templates.SettingsPageData{
			FirstName: identity.FirstName,
			LastName:  identity.LastName,
			Email:     identity.Email,
			Saved:     r.URL.Query().Get("saved"),
		})
	if msg := settingsErrorMessage(r.URL.Query().Get("error")); msg != "" {
		pageData = pageData.WithError(msg)
	}
	if err := pages.SettingsPage(pageData).Render(r.Context(), w); err != nil {
		log.Error().Err(err).Msg("Failed to render settings page")
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *SettingsHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.renderSettings(w, r, user, "", "invalid form", nil)
		return
	}
	first := strings.TrimSpace(r.FormValue("first_name"))
	last := strings.TrimSpace(r.FormValue("last_name"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))

	fieldErrors := map[string][]string{}
	if first == "" {
		fieldErrors["first_name"] = []string{"First name is required"}
	}
	if email == "" || !strings.Contains(email, "@") {
		fieldErrors["email"] = []string{"A valid email is required"}
	}
	if len(fieldErrors) > 0 {
		h.renderSettings(w, r, user, "", "", &templates.SettingsPageData{
			FirstName:   first,
			LastName:    last,
			Email:       email,
			FieldErrors: fieldErrors,
		})
		return
	}

	identity, _, err := h.identities.ByID(r.Context(), user.ID)
	if err != nil {
		h.renderSettings(w, r, user, "", "Failed to load account", nil)
		return
	}

	if err := h.identities.UpdateProfile(r.Context(), user.ID, first, last, identity.Email); err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("update profile failed")
		h.renderSettings(w, r, user, "", "Failed to update profile", &templates.SettingsPageData{
			FirstName: first,
			LastName:  last,
			Email:     email,
		})
		return
	}

	saved := "profile"
	if email != strings.ToLower(strings.TrimSpace(identity.Email)) {
		if err := h.requestEmailChange(r, identity, email); err != nil {
			if errors.Is(err, coreauth.ErrEmailTaken) {
				h.renderSettings(w, r, user, "", "", &templates.SettingsPageData{
					FirstName: first,
					LastName:  last,
					Email:     email,
					FieldErrors: map[string][]string{
						"email": {"This email is already in use"},
					},
				})
				return
			}
			log.Error().Err(err).Str("user_id", user.ID).Msg("email change request failed")
			h.renderSettings(w, r, user, "", "Profile saved, but we could not send the confirmation email", &templates.SettingsPageData{
				FirstName: first,
				LastName:  last,
				Email:     identity.Email,
			})
			return
		}
		saved = "email"
	}

	identity, _, err = h.identities.ByID(r.Context(), user.ID)
	if err == nil {
		_ = h.sessionManager.CreateSession(w, r, identity.UserID, auth.UserInfoFromIdentity(identity), identity.SessionVersion)
	}
	http.Redirect(w, r, "/settings?saved="+saved, http.StatusSeeOther)
}

func (h *SettingsHandler) requestEmailChange(r *http.Request, identity coreauth.Identity, newEmail string) error {
	if _, _, err := h.identities.ByEmail(r.Context(), newEmail); err == nil {
		return coreauth.ErrEmailTaken
	} else if !errors.Is(err, coreauth.ErrIdentityNotFound) {
		return err
	}
	token, err := h.tokens.IssueWithPayload(r.Context(), identity.UserID, coreauth.TokenEmailChange, coreauth.EmailChangeTTL, newEmail)
	if err != nil {
		return err
	}
	return h.emails.SendEmailChange(r.Context(), mail.Address{
		Name:  strings.TrimSpace(identity.FirstName + " " + identity.LastName),
		Email: newEmail,
	}, token)
}

func (h *SettingsHandler) ConfirmEmailChange(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Redirect(w, r, "/settings?error=missing_token", http.StatusSeeOther)
		return
	}
	userID, newEmail, err := h.tokens.ConsumeWithPayload(r.Context(), token, coreauth.TokenEmailChange)
	if err != nil || newEmail == "" {
		http.Redirect(w, r, "/login?error=invalid_token", http.StatusSeeOther)
		return
	}
	identity, _, err := h.identities.ByID(r.Context(), userID)
	if err != nil {
		http.Redirect(w, r, "/login?error=invalid_token", http.StatusSeeOther)
		return
	}
	if err := h.identities.UpdateProfile(r.Context(), userID, identity.FirstName, identity.LastName, newEmail); err != nil {
		if errors.Is(err, coreauth.ErrEmailTaken) {
			http.Redirect(w, r, "/settings?error=email_taken", http.StatusSeeOther)
			return
		}
		log.Error().Err(err).Str("user_id", userID).Msg("confirm email change failed")
		http.Redirect(w, r, "/settings?error=email_change", http.StatusSeeOther)
		return
	}
	ver, err := h.identities.BumpSessionVersion(r.Context(), userID)
	if err != nil {
		ver = identity.SessionVersion + 1
	}
	identity.Email = newEmail
	identity.SessionVersion = ver
	_ = h.sessionManager.CreateSession(w, r, identity.UserID, auth.UserInfoFromIdentity(identity), ver)
	http.Redirect(w, r, "/settings?saved=email-confirmed", http.StatusSeeOther)
}

func (h *SettingsHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		h.renderSettings(w, r, user, "", "invalid form", nil)
		return
	}
	current := r.FormValue("current_password")
	next := r.FormValue("new_password")
	confirm := r.FormValue("confirm_password")

	fieldErrors := map[string][]string{}
	if current == "" {
		fieldErrors["current_password"] = []string{"Current password is required"}
	}
	if err := coreauth.ValidatePasswordLength(next); err != nil {
		fieldErrors["new_password"] = []string{err.Error()}
	}
	if next != confirm {
		fieldErrors["confirm_password"] = []string{"Passwords do not match"}
	}
	identity, hash, err := h.identities.ByID(r.Context(), user.ID)
	if err != nil {
		h.renderSettings(w, r, user, "", "Failed to load account", nil)
		return
	}
	if len(fieldErrors) == 0 {
		if hash == "" || h.passwords.Check(current, hash) != nil {
			fieldErrors["current_password"] = []string{"Current password is incorrect"}
		}
	}
	if len(fieldErrors) > 0 {
		h.renderSettings(w, r, user, "", "", &templates.SettingsPageData{
			FirstName:   identity.FirstName,
			LastName:    identity.LastName,
			Email:       identity.Email,
			FieldErrors: fieldErrors,
		})
		return
	}

	newHash, err := h.passwords.Hash(next)
	if err != nil {
		h.renderSettings(w, r, user, "", "Failed to update password", nil)
		return
	}
	if err := h.identities.UpdatePassword(r.Context(), user.ID, newHash); err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("update password failed")
		h.renderSettings(w, r, user, "", "Failed to update password", nil)
		return
	}
	ver, err := h.identities.BumpSessionVersion(r.Context(), user.ID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", user.ID).Msg("password session bump failed")
		ver = identity.SessionVersion + 1
	}
	identity.SessionVersion = ver
	_ = h.sessionManager.CreateSession(w, r, identity.UserID, auth.UserInfoFromIdentity(identity), ver)
	http.Redirect(w, r, "/settings?saved=password", http.StatusSeeOther)
}

func settingsErrorMessage(code string) string {
	switch code {
	case "missing_token":
		return "That confirmation link is missing a token."
	case "email_taken":
		return "That email is already in use."
	case "email_change":
		return "Could not update your email. Try again."
	default:
		return ""
	}
}

func (h *SettingsHandler) renderSettings(w http.ResponseWriter, r *http.Request, user *auth.UserInfo, saved, errMsg string, data *templates.SettingsPageData) {
	if data == nil {
		identity, _, err := h.identities.ByID(r.Context(), user.ID)
		data = &templates.SettingsPageData{}
		if err == nil {
			data.FirstName = identity.FirstName
			data.LastName = identity.LastName
			data.Email = identity.Email
		}
	}
	data.Saved = saved
	pageData := templates.NewPageDataFromRequest(r, "Settings - ThreadArt", "settings").
		WithData(data)
	if errMsg != "" {
		pageData = pageData.WithError(errMsg)
	}
	if len(data.FieldErrors) > 0 {
		pageData = pageData.WithFieldErrors(data.FieldErrors)
	}
	if err := pages.SettingsPage(pageData).Render(r.Context(), w); err != nil {
		log.Error().Err(err).Msg("Failed to render settings page")
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
