package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/rs/zerolog/log"
)

// ArtHandler handles art-related operations
type ArtHandler struct {
	generatorService *services.GeneratorService
}

// NewArtHandler creates a new art handler
func NewArtHandler(generatorService *services.GeneratorService) *ArtHandler {
	return &ArtHandler{
		generatorService: generatorService,
	}
}

// ViewArtPage renders the art details page
func (h *ArtHandler) ViewArtPage(w http.ResponseWriter, r *http.Request) {
	// Get user from context (contains Firebase UID)
	user, _ := middleware.UserFromContext(r.Context())

	// Extract art ID from URL path
	// URL format: /dashboard/arts/{artId}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid art ID", http.StatusBadRequest)
		return
	}
	artID := pathParts[2]

	// Get internal user ID by calling GetCurrentUser API
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for ViewArtPage")
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Parse the user resource name to extract internal user ID
	userResource, err := resource.ParseResourceName(currentUser.ID)
	if err != nil {
		log.Error().Err(err).Str("user_resource_name", currentUser.ID).Msg("Failed to parse user resource name")
		http.Error(w, "Failed to parse user information", http.StatusInternalServerError)
		return
	}

	internalUserID := userResource.(*resource.User).ID

	// Get the art using internal user ID
	art, err := h.generatorService.GetArt(r.Context(), r, internalUserID, artID)
	if err != nil {
		log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get art")
		http.Error(w, "Art not found", http.StatusNotFound)
		return
	}

	// Get compositions if art is complete
	var compositions []*pb.Composition
	if art.GetStatus() == pb.ArtStatus_ART_STATUS_COMPLETE {
		compositionsResponse, err := h.generatorService.CompositionService.ListCompositionsForArt(r.Context(), r, internalUserID, artID, 50, "")
		if err != nil {
			log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get compositions for art")
			// Don't fail the page, just leave compositions empty
		} else {
			compositions = compositionsResponse.GetCompositions()
		}
	}

	// Create page data
	pageData := templates.NewPageDataFromRequest(r, fmt.Sprintf("Art: %s", art.GetTitle()), "art")

	// Render the art details page
	err = templates.ArtPage(pageData, art, compositions).Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render art details page")
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

// CreateArtForm renders the create art form
func (h *ArtHandler) CreateArtForm(w http.ResponseWriter, r *http.Request) {
	// Create page data with empty form data
	formData := &services.ArtFormData{
		Title:   "",
		Errors:  make(map[string][]string),
		Success: false,
	}

	pageData := templates.NewPageDataFromRequest(r, "Create Art", "create-art")

	// Render the create art form
	err := templates.NewArtPage(pageData, formData).Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render create art form")
		http.Error(w, "Error rendering form", http.StatusInternalServerError)
	}
}

// CreateArt handles art creation
func (h *ArtHandler) CreateArt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))

	// Get user from context (contains Firebase UID)
	user, _ := middleware.UserFromContext(r.Context())

	// Get internal user ID by calling GetCurrentUser API
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for CreateArt")
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Parse the user resource name to extract internal user ID
	userResource, err := resource.ParseResourceName(currentUser.ID)
	if err != nil {
		log.Error().Err(err).Str("user_resource_name", currentUser.ID).Msg("Failed to parse user resource name")
		http.Error(w, "Failed to parse user information", http.StatusInternalServerError)
		return
	}
	internalUserID := userResource.(*resource.User).ID

	// Call service to create art - using the new signature
	art, fieldErrors, err := h.generatorService.CreateArt(r.Context(), r, internalUserID, title)
	if err != nil {
		// If there are field validation errors
		if fieldErrors != nil {
			formData := &services.ArtFormData{
				Title:   title,
				Errors:  fieldErrors,
				Success: false,
			}

			// Check if this is an HTMX request
			if r.Header.Get("HX-Request") == "true" {
				// For HTMX requests, render only the form component
				err = templates.NewArtForm(formData).Render(r.Context(), w)
			} else {
				// For normal requests, render the full page
				pageData := templates.NewPageDataFromRequest(r, "Create Art", "create-art")
				err = templates.NewArtPage(pageData, formData).Render(r.Context(), w)
			}

			if err != nil {
				log.Error().Err(err).Msg("Failed to render create art form with errors")
				http.Error(w, "Error rendering form", http.StatusInternalServerError)
			}
			return
		}

		// Other errors
		log.Error().Err(err).Str("internal_user_id", internalUserID).Msg("Failed to create art")
		http.Error(w, "Failed to create art", http.StatusInternalServerError)
		return
	}

	// Success - redirect to art details page
	artResource, err := resource.ParseResourceName(art.GetName())
	if err != nil {
		log.Error().Err(err).Str("art_resource_name", art.GetName()).Msg("Failed to parse created art resource name")
		http.Error(w, "Failed to process created art", http.StatusInternalServerError)
		return
	}

	artID := artResource.(*resource.Art).ArtID

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// For HTMX requests, use HX-Redirect header
		w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/arts/%s", artID))
		w.WriteHeader(http.StatusOK)
	} else {
		// For normal requests, use regular redirect
		http.Redirect(w, r, fmt.Sprintf("/dashboard/arts/%s", artID), http.StatusSeeOther)
	}
}
