package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/go-chi/chi/v5"
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
	// Get user from context
	user, _ := middleware.UserFromContext(r.Context())

	// Extract art ID from URL path
	// URL format: /dashboard/arts/{artId}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid art ID", http.StatusBadRequest)
		return
	}
	artID := pathParts[2]

	// Get the art
	art, err := h.generatorService.GetArt(r.Context(), user.ID, artID)
	if err != nil {
		http.Error(w, "Art not found", http.StatusNotFound)
		return
	}

	// Render the art page
	pageData := templates.NewPageData(fmt.Sprintf("Art: %s - ThreadArt", art.GetTitle()), "art").
		WithUser(user)
	err = templates.ArtPage(pageData, art).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render art page")
	}
}

// GetArtUploadUrl handles getting a signed upload URL for an art
func (h *ArtHandler) GetArtUploadUrl(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, _ := middleware.UserFromContext(r.Context())

	// Extract art ID from URL parameter
	artID := chi.URLParam(r, "artId")
	if artID == "" {
		http.Error(w, "Invalid art ID", http.StatusBadRequest)
		return
	}

	// Get upload URL
	uploadResponse, err := h.generatorService.GetArtUploadUrl(r.Context(), user.ID, artID)
	if err != nil {
		log.Error().Err(err).Str("art_id", artID).Msg("Failed to get upload URL")
		http.Error(w, "Failed to get upload URL", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"upload_url":      uploadResponse.GetUploadUrl(),
		"expiration_time": uploadResponse.GetExpirationTime().AsTime(),
	})
}

// ConfirmArtImageUpload handles confirming that an image has been uploaded
func (h *ArtHandler) ConfirmArtImageUpload(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, _ := middleware.UserFromContext(r.Context())

	// Extract art ID from URL parameter
	artID := chi.URLParam(r, "artId")
	if artID == "" {
		http.Error(w, "Invalid art ID", http.StatusBadRequest)
		return
	}

	// Confirm upload - construct the full resource name as expected by the service
	// Following Google AIP resource naming: users/{user_id}/arts/{art_id}
	resourceName := resource.BuildArtResourceName(user.ID, artID)

	art, err := h.generatorService.ConfirmArtImageUpload(r.Context(), resourceName)
	if err != nil {
		log.Error().Err(err).Str("art_id", artID).Str("resource_name", resourceName).Msg("Failed to confirm upload")
		http.Error(w, "Failed to confirm upload", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"status":  art.GetStatus().String(),
	})
}

// NewArtPage renders the art creation form
func (h *ArtHandler) NewArtPage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if authenticated
	user, _ := middleware.UserFromContext(r.Context())

	// Initial form data with empty values
	formData := &services.ArtFormData{
		Title:   "",
		Errors:  make(map[string][]string),
		Success: false,
	}

	// Render the art creation form
	pageData := templates.NewPageData("Create New Art - ThreadArt", "new-art").
		WithUser(user)
	err := templates.NewArtPage(pageData, formData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render new art page")
	}
}

// CreateArt handles the art creation form submission
func (h *ArtHandler) CreateArt(w http.ResponseWriter, r *http.Request) {
	// Get user from context
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

	createArtRequest := &pb.CreateArtRequest{
		Art: &pb.Art{
			Title: title,
		},
		Parent: resource.BuildUserResourceName(user.ID),
	}

	// Call service to create art with the request object for auth headers
	art, fieldErrors, err := h.generatorService.CreateArt(r.Context(), createArtRequest)
	if err != nil {
		// If there are field validation errors
		if fieldErrors != nil {
			formData.Errors = fieldErrors
			// Render form with errors
			renderErr := templates.NewArtForm(formData).Render(r.Context(), w)
			if renderErr != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				log.Error().Err(renderErr).Msg("Failed to render new art form with errors")
			}
			return
		}

		// For other errors, display a general error
		formData.Errors["_form"] = []string{"An error occurred while creating the art. Please try again."}
		renderErr := templates.NewArtForm(formData).Render(r.Context(), w)
		if renderErr != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Error().Err(renderErr).Msg("Failed to render new art form with general error")
		}
		return
	}

	// Art created successfully - parse resource name properly
	artResource, err := resource.ParseResourceName(art.GetName())
	if err != nil {
		// Fallback to dashboard if we can't parse
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}

	if parsedArt, ok := artResource.(*resource.Art); ok {
		// Redirect to the art page
		w.Header().Set("HX-Redirect", "/dashboard/arts/"+parsedArt.ArtID)
		w.WriteHeader(http.StatusOK)
	} else {
		// Fallback to dashboard if wrong type
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
	}
}
