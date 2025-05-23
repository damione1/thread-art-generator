package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	"github.com/Damione1/thread-art-generator/core/pb"
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
	err := templates.NewArtPage(user, formData).Render(r.Context(), w)
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
		Parent: user.ID,
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

	// Art created successfully
	formData.Success = true

	// Redirect to the art page or return success response
	log.Info().Str("art_id", art.GetName()).Msg("Art created successfully")

	// For HTMX, return a redirect or success message
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}
