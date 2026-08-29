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
	"github.com/rs/zerolog/log"
)

type SettingsHandler struct {
	identities     coreauth.Identities
	passwords      coreauth.Passwords
	sessionManager *auth.SCSSessionManager
}

func NewSettingsHandler(identities coreauth.Identities, sessionManager *auth.SCSSessionManager) *SettingsHandler {
	return &SettingsHandler{
		identities:     identities,
		passwords:      coreauth.Argon2idPasswords{},
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

	if err := h.identities.UpdateProfile(r.Context(), user.ID, first, last, email); err != nil {
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
		log.Error().Err(err).Str("user_id", user.ID).Msg("update profile failed")
		h.renderSettings(w, r, user, "", "Failed to update profile", &templates.SettingsPageData{
			FirstName: first,
			LastName:  last,
			Email:     email,
		})
		return
	}

	identity, _, err := h.identities.ByID(r.Context(), user.ID)
	if err == nil {
		_ = h.sessionManager.CreateSession(w, r, identity.UserID, auth.UserInfoFromIdentity(identity))
	}
	http.Redirect(w, r, "/settings?saved=profile", http.StatusSeeOther)
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
	if len(next) < 8 {
		fieldErrors["new_password"] = []string{"Password must be at least 8 characters"}
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
	http.Redirect(w, r, "/settings?saved=password", http.StatusSeeOther)
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
