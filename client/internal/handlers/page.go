package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	"github.com/rs/zerolog/log"
)

// PageHandler handles rendering the main application pages
type PageHandler struct {
	userService *services.UserService
}

// NewPageHandler creates a new page handler
func NewPageHandler(userService *services.UserService) *PageHandler {
	return &PageHandler{
		userService: userService,
	}
}

// HomePage renders the home page
func (h *PageHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if authenticated
	user, _ := middleware.UserFromContext(r.Context())

	// Render the home page template
	err := templates.HomePage(user).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render home page")
	}
}

// DashboardPage renders the dashboard page (protected)
func (h *PageHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	// User will be in context due to RequireAuth middleware
	user, _ := middleware.UserFromContext(r.Context())

	// Render dashboard with user data
	err := templates.Dashboard(user).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render dashboard")
	}
}
