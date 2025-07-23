package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CompositionHandler handles composition-related operations
type CompositionHandler struct {
	generatorService *services.GeneratorService
}

// NewCompositionHandler creates a new composition handler
func NewCompositionHandler(generatorService *services.GeneratorService) *CompositionHandler {
	return &CompositionHandler{
		generatorService: generatorService,
	}
}

// NewCompositionForm renders the composition creation form
func (h *CompositionHandler) NewCompositionForm(w http.ResponseWriter, r *http.Request) {
	// Get user from context (contains Firebase UID)
	user, _ := middleware.UserFromContext(r.Context())

	// Extract art ID from URL
	artID := chi.URLParam(r, "artId")
	if artID == "" {
		http.Error(w, "Invalid art ID", http.StatusBadRequest)
		return
	}

	// Get internal user ID by calling GetCurrentUser API
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for NewCompositionForm")
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

	// Get the art
	art, err := h.generatorService.GetArt(r.Context(), internalUserID, artID)
	if err != nil {
		log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get art")
		http.Error(w, "Art not found", http.StatusNotFound)
		return
	}

	// Check if art is complete
	if art.GetStatus() != pb.ArtStatus_ART_STATUS_COMPLETE {
		http.Error(w, "Art must be complete to create compositions", http.StatusBadRequest)
		return
	}

	// Initialize form data with default values
	formData := &templates.CompositionFormData{
		NailsQuantity:     300,
		ImgSize:           1000,
		MaxPaths:          3000,
		StartingNail:      0,
		MinimumDifference: 15,
		BrightnessFactor:  128,
		ImageContrast:     1.2,
		PhysicalRadius:    200.0,
		Errors:            make(map[string][]string),
		Success:           false,
	}

	// Check if we need to pre-fill from an existing composition
	fromCompositionID := r.URL.Query().Get("from")
	if fromCompositionID != "" {
		// Get the source composition to copy settings from
		sourceComposition, err := h.generatorService.GetComposition(r.Context(), internalUserID, artID, fromCompositionID)
		if err != nil {
			log.Error().Err(err).
				Str("internal_user_id", internalUserID).
				Str("art_id", artID).
				Str("from_composition_id", fromCompositionID).
				Msg("Failed to get source composition for copying settings")
			// Continue with defaults if we can't load the source composition
		} else {
			// Copy settings from the source composition
			formData.NailsQuantity = sourceComposition.GetNailsQuantity()
			formData.ImgSize = sourceComposition.GetImgSize()
			formData.MaxPaths = sourceComposition.GetMaxPaths()
			formData.StartingNail = sourceComposition.GetStartingNail()
			formData.MinimumDifference = sourceComposition.GetMinimumDifference()
			formData.BrightnessFactor = sourceComposition.GetBrightnessFactor()
			formData.ImageContrast = sourceComposition.GetImageContrast()
			formData.PhysicalRadius = sourceComposition.GetPhysicalRadius()
		}
	}

	// Render the composition form
	pageData := templates.NewPageDataFromRequest(r, fmt.Sprintf("New Composition - %s - ThreadArt", art.GetTitle()), "composition")
	err = templates.NewCompositionPage(pageData, art, formData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render new composition page")
	}
}

// CreateComposition handles the composition creation form submission
func (h *CompositionHandler) CreateComposition(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, _ := middleware.UserFromContext(r.Context())

	// Extract art ID from URL
	artID := chi.URLParam(r, "artId")
	if artID == "" {
		http.Error(w, "Invalid art ID", http.StatusBadRequest)
		return
	}

	// Parse form
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		log.Error().Err(err).Msg("Failed to parse form")
		return
	}

	// Parse form values
	nailsQuantity, _ := strconv.ParseInt(r.FormValue("nails_quantity"), 10, 32)
	imgSize, _ := strconv.ParseInt(r.FormValue("img_size"), 10, 32)
	maxPaths, _ := strconv.ParseInt(r.FormValue("max_paths"), 10, 32)
	startingNail, _ := strconv.ParseInt(r.FormValue("starting_nail"), 10, 32)
	minimumDifference, _ := strconv.ParseInt(r.FormValue("minimum_difference"), 10, 32)
	brightnessFactor, _ := strconv.ParseInt(r.FormValue("brightness_factor"), 10, 32)
	imageContrast, _ := strconv.ParseFloat(r.FormValue("image_contrast"), 32)
	physicalRadius, _ := strconv.ParseFloat(r.FormValue("physical_radius"), 32)

	// Note: image_contrast comes from slider scaled by 10
	imageContrast = imageContrast / 10.0

	// Initialize form data
	formData := &templates.CompositionFormData{
		NailsQuantity:     int32(nailsQuantity),
		ImgSize:           int32(imgSize),
		MaxPaths:          int32(maxPaths),
		StartingNail:      int32(startingNail),
		MinimumDifference: int32(minimumDifference),
		BrightnessFactor:  int32(brightnessFactor),
		ImageContrast:     float32(imageContrast),
		PhysicalRadius:    float32(physicalRadius),
		Errors:            make(map[string][]string),
		Success:           false,
	}

	// Get internal user ID
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for CreateComposition")
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

	// Get the art to verify it exists and is complete
	art, err := h.generatorService.GetArt(r.Context(), internalUserID, artID)
	if err != nil {
		log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get art")
		formData.Errors["_form"] = []string{"Failed to load art. Please try again."}
		renderErr := templates.CompositionForm(art, formData).Render(r.Context(), w)
		if renderErr != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Error().Err(renderErr).Msg("Failed to render composition form with errors")
		}
		return
	}

	// Build the parent resource name
	parent := resource.BuildArtResourceName(internalUserID, artID)

	// Create composition request
	createRequest := &pb.CreateCompositionRequest{
		Parent: parent,
		Composition: &pb.Composition{
			NailsQuantity:     formData.NailsQuantity,
			ImgSize:           formData.ImgSize,
			MaxPaths:          formData.MaxPaths,
			StartingNail:      formData.StartingNail,
			MinimumDifference: formData.MinimumDifference,
			BrightnessFactor:  formData.BrightnessFactor,
			ImageContrast:     formData.ImageContrast,
			PhysicalRadius:    formData.PhysicalRadius,
		},
	}

	// Call service to create composition
	composition, fieldErrors, err := h.generatorService.CreateComposition(r.Context(), createRequest)
	if err != nil {
		// If there are field validation errors
		if fieldErrors != nil {
			formData.Errors = fieldErrors
			// Render form with errors
			renderErr := templates.CompositionForm(art, formData).Render(r.Context(), w)
			if renderErr != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				log.Error().Err(renderErr).Msg("Failed to render composition form with field errors")
			}
			return
		}

		// For other errors, display a general error
		formData.Errors["_form"] = []string{"An error occurred while creating the composition. Please try again."}
		renderErr := templates.CompositionForm(art, formData).Render(r.Context(), w)
		if renderErr != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Error().Err(renderErr).Msg("Failed to render composition form with general error")
		}
		return
	}

	// Composition created successfully - parse resource name to get composition ID
	compositionResource, err := resource.ParseResourceName(composition.GetName())
	if err != nil {
		// Fallback to art page if we can't parse
		w.Header().Set("HX-Redirect", "/dashboard/arts/"+artID)
		w.WriteHeader(http.StatusOK)
		return
	}

	if parsedComposition, ok := compositionResource.(*resource.Composition); ok {
		// Redirect to the composition detail page
		w.Header().Set("HX-Redirect", "/dashboard/arts/"+artID+"/composition/"+parsedComposition.CompositionID)
		w.WriteHeader(http.StatusOK)
	} else {
		// Fallback to art page if wrong type
		w.Header().Set("HX-Redirect", "/dashboard/arts/"+artID)
		w.WriteHeader(http.StatusOK)
	}
}

// ViewComposition renders the composition detail page
func (h *CompositionHandler) ViewComposition(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, _ := middleware.UserFromContext(r.Context())

	// Extract IDs from URL
	artID := chi.URLParam(r, "artId")
	compositionID := chi.URLParam(r, "compositionId")
	
	if artID == "" || compositionID == "" {
		http.Error(w, "Invalid IDs", http.StatusBadRequest)
		return
	}

	// Get internal user ID
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for ViewComposition")
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

	// Get the art
	art, err := h.generatorService.GetArt(r.Context(), internalUserID, artID)
	if err != nil {
		log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get art")
		http.Error(w, "Art not found", http.StatusNotFound)
		return
	}

	// Get the composition
	composition, err := h.generatorService.GetComposition(r.Context(), internalUserID, artID, compositionID)
	if err != nil {
		log.Error().Err(err).
			Str("internal_user_id", internalUserID).
			Str("art_id", artID).
			Str("composition_id", compositionID).
			Msg("Failed to get composition")
		http.Error(w, "Composition not found", http.StatusNotFound)
		return
	}

	// Render the composition detail page
	pageData := templates.NewPageDataFromRequest(r, fmt.Sprintf("Composition - %s - ThreadArt", art.GetTitle()), "composition")
	err = templates.CompositionDetailPage(pageData, art, composition).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render composition detail page")
	}
}

// GetCompositionStatus returns the composition status for HTMX polling
func (h *CompositionHandler) GetCompositionStatus(w http.ResponseWriter, r *http.Request) {
	// Extract IDs from URL path
	// URL format: /dashboard/arts/{artId}/composition/{compositionId}/status
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	artID := pathParts[2]
	compositionID := pathParts[4]
	
	// Get user from context
	user, _ := middleware.UserFromContext(r.Context())

	// Get internal user ID
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for GetCompositionStatus")
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

	// Get the art and composition
	art, err := h.generatorService.GetArt(r.Context(), internalUserID, artID)
	if err != nil {
		log.Error().Err(err).Str("internal_user_id", internalUserID).Str("art_id", artID).Msg("Failed to get art for status")
		http.Error(w, "Art not found", http.StatusNotFound)
		return
	}

	composition, err := h.generatorService.GetComposition(r.Context(), internalUserID, artID, compositionID)
	if err != nil {
		log.Error().Err(err).
			Str("internal_user_id", internalUserID).
			Str("art_id", artID).
			Str("composition_id", compositionID).
			Msg("Failed to get composition for status")
		http.Error(w, "Composition not found", http.StatusNotFound)
		return
	}

	// Render the entire composition detail page for HTMX to swap
	pageData := templates.NewPageDataFromRequest(r, fmt.Sprintf("Composition - %s - ThreadArt", art.GetTitle()), "composition")
	err = templates.CompositionDetailPage(pageData, art, composition).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render composition status update")
	}
}

// DeleteComposition handles deleting a composition
func (h *CompositionHandler) DeleteComposition(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, _ := middleware.UserFromContext(r.Context())

	// Extract IDs from URL
	artID := chi.URLParam(r, "artId")
	compositionID := chi.URLParam(r, "compositionId")
	
	if artID == "" || compositionID == "" {
		http.Error(w, "Invalid IDs", http.StatusBadRequest)
		return
	}

	// Get internal user ID
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for DeleteComposition")
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

	// Build the composition resource name
	compositionResourceName := resource.BuildCompositionResourceName(internalUserID, artID, compositionID)

	// Delete the composition (note: we need to add this method to the service)
	err = h.generatorService.DeleteComposition(r.Context(), compositionResourceName)
	if err != nil {
		log.Error().Err(err).
			Str("internal_user_id", internalUserID).
			Str("art_id", artID).
			Str("composition_id", compositionID).
			Msg("Failed to delete composition")
		http.Error(w, "Failed to delete composition", http.StatusInternalServerError)
		return
	}

	// Return success - the HTMX target will remove the element
	w.WriteHeader(http.StatusOK)
}