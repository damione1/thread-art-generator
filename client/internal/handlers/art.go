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
	// Get user from context (session user)
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
		log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to get current user for ViewArtPage")
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Parse the user resource name to extract internal user ID
	userResource, err := resource.ParseResourceName(currentUser.ID)
	if err != nil {
		log.Error().Err(err).Str("user_resource_name", currentUser.ID).Msg("Failed to parse user resource name")
		http.Error(w, "Invalid user resource", http.StatusInternalServerError)
		return
	}

	internalUserID := userResource.(*resource.User).ID

	// Get the art using internal user ID
	art, err := h.generatorService.GetArt(r.Context(), internalUserID, artID)
	if err != nil {
		log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get art")
		http.Error(w, "Art not found", http.StatusNotFound)
		return
	}

	// Get compositions for this art if it's complete
	var compositions []*pb.Composition
	if art.GetStatus() == pb.ArtStatus_ART_STATUS_COMPLETE {
		compositionsResponse, err := h.generatorService.ListCompositionsForArt(r.Context(), internalUserID, artID, 50, "")
		if err != nil {
			log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get compositions for art")
			// Don't fail the page load, just log the error and continue with empty compositions
		} else {
			compositions = compositionsResponse.GetCompositions()
		}
	}

	// Render the art page using middleware-provided context
	pageData := templates.NewPageDataFromRequest(r, fmt.Sprintf("Art: %s - ThreadArt", art.GetTitle()), "art")
	err = templates.ArtPage(pageData, art, compositions).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render art page")
	}
}

// NewArtPage renders the art creation form
func (h *ArtHandler) NewArtPage(w http.ResponseWriter, r *http.Request) {
	// Initial form data with empty values
	formData := &services.ArtFormData{
		Title:   "",
		Errors:  make(map[string][]string),
		Success: false,
	}

	// Render the art creation form using middleware-provided context
	pageData := templates.NewPageDataFromRequest(r, "Create New Art - ThreadArt", "new-art")
	err := templates.NewArtPage(pageData, formData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render new art page")
	}
}

// CreateArt handles the art creation form submission
func (h *ArtHandler) CreateArt(w http.ResponseWriter, r *http.Request) {
	// Get user from context (session user)
	user, _ := middleware.UserFromContext(r.Context())

	// Parse form
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		log.Error().Err(err).Msg("Failed to parse form")
		return
	}

	// Get title from form
	title := r.FormValue("title")

	// Initialize form data
	formData := &services.ArtFormData{
		Title:   title,
		Errors:  make(map[string][]string),
		Success: false,
	}

	// Get internal user ID by calling GetCurrentUser API
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to get current user for CreateArt")
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	createArtRequest := &pb.CreateArtRequest{
		Art: &pb.Art{
			Title: title,
		},
		Parent: currentUser.ID, // currentUser.ID contains the resource name with internal user ID
	}

	// Call service to create art with the request object for auth headers
	art, fieldErrors, err := h.generatorService.CreateArt(r.Context(), createArtRequest)
	if err != nil {
		formData.Errors = fieldErrors
		if isHTMX(r) {
			renderErr := templates.NewArtForm(formData).Render(r.Context(), w)
			if renderErr != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				log.Error().Err(renderErr).Msg("Failed to render new art form with errors")
			}
			return
		}
		pageData := templates.NewPageDataFromRequest(r, "Create New Art - ThreadArt", "new-art")
		renderErr := templates.NewArtPage(pageData, formData).Render(r.Context(), w)
		if renderErr != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Error().Err(renderErr).Msg("Failed to render new art page with errors")
		}
		return
	}

	artResource, err := resource.ParseResourceName(art.GetName())
	if err != nil {
		redirect(w, r, "/dashboard")
		return
	}

	if parsedArt, ok := artResource.(*resource.Art); ok {
		redirect(w, r, "/dashboard/arts/"+parsedArt.ArtID)
		return
	}
	redirect(w, r, "/dashboard")
}
