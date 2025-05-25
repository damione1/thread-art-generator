package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates/pages"
	"github.com/rs/zerolog/log"
)

// PageHandler handles rendering the main application pages
type PageHandler struct {
	generatorService *services.GeneratorService
}

// NewPageHandler creates a new page handler
func NewPageHandler(generatorService *services.GeneratorService) *PageHandler {
	return &PageHandler{
		generatorService: generatorService,
	}
}

// HomePage renders the home page
func (h *PageHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if authenticated
	user, _ := middleware.UserFromContext(r.Context())

	// Render the home page template
	err := pages.HomePage(user).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render home page")
	}
}

// DashboardPage renders the dashboard page (protected)
func (h *PageHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	// User will be in context due to RequireAuth middleware
	user, _ := middleware.UserFromContext(r.Context())

	// Read sort and dir from query params, default to create_time/desc
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "create_time"
	}
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		dir = "desc"
	}

	// Fetch user's arts with sorting
	arts, err := h.generatorService.ListArts(r.Context(), user, 10, "", sort, dir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch arts for dashboard")
		http.Error(w, "Error fetching arts", http.StatusInternalServerError)
		return
	}

	// Pass sort and dir to template for button state
	err = pages.DashboardPage(user, arts.GetArts(), sort, dir).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render dashboard")
	}
}
