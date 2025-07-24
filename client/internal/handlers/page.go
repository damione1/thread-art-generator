package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	pages "github.com/Damione1/thread-art-generator/client/internal/templates/pages"
	"github.com/Damione1/thread-art-generator/client/internal/types"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/rs/zerolog/log"
)

// PageHandler handles rendering the main application pages
type PageHandler struct {
	generatorService *services.GeneratorService
	config           *util.Config
}

// NewPageHandler creates a new page handler
func NewPageHandler(generatorService *services.GeneratorService, config *util.Config) *PageHandler {
	return &PageHandler{
		generatorService: generatorService,
		config:           config,
	}
}

// HomePage renders the home page
func (h *PageHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if authenticated
	user, _ := middleware.UserFromContext(r.Context())

	// Create page data using the new structure
	pageData := templates.NewPageData("ThreadArt - Create Beautiful Thread Art", "home").
		WithUser(user)

	// Add Firebase config for logged-in users (needed for topbar logout functionality)
	if user != nil {
		firebaseConfig := h.getFirebaseConfig()
		pageData = pageData.WithFirebaseConfig(firebaseConfig)
	}

	// Render the home page template
	err := pages.HomePage(pageData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render home page")
	}
}

// DashboardPage renders the dashboard page (protected)
func (h *PageHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	// User will be in context due to RequireAuth middleware (contains Firebase UID)
	user, _ := middleware.UserFromContext(r.Context())

	// Get internal user ID by calling GetCurrentUser API
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("firebase_uid", user.ID).Msg("Failed to get current user for DashboardPage")

		// Create error page data using middleware-provided context
		pageData := templates.NewPageDataFromRequest(r, "Dashboard - Error", "dashboard").
			WithError("Error loading user information. Please try again.")

		// Create empty dashboard data for error case
		dashboardData := &templates.DashboardPageData{
			Arts: nil,
			Sort: "create_time",
			Dir:  "desc",
		}
		pageData = pageData.WithData(dashboardData)

		err = pages.DashboardPage(pageData).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Error().Err(err).Msg("Failed to render dashboard error")
		}
		return
	}

	// Parse the user resource name to extract internal user ID
	userResource, err := resource.ParseResourceName(currentUser.ID)
	if err != nil {
		log.Error().Err(err).Str("user_resource_name", currentUser.ID).Msg("Failed to parse user resource name")

		// Create error page data using middleware-provided context
		pageData := templates.NewPageDataFromRequest(r, "Dashboard - Error", "dashboard").
			WithError("Error parsing user information. Please try again.")

		// Create empty dashboard data for error case
		dashboardData := &templates.DashboardPageData{
			Arts: nil,
			Sort: "create_time",
			Dir:  "desc",
		}
		pageData = pageData.WithData(dashboardData)

		err = pages.DashboardPage(pageData).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Error().Err(err).Msg("Failed to render dashboard error")
		}
		return
	}

	internalUserID := userResource.(*resource.User).ID

	// Read sort and dir from query params, default to create_time/desc
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "create_time"
	}
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		dir = "desc"
	}

	// Fetch user's arts with sorting using internal user ID
	arts, err := h.generatorService.ListArts(r.Context(), internalUserID, 10, "", sort, dir)
	if err != nil {
		log.Error().Err(err).Str("internal_user_id", internalUserID).Msg("Failed to fetch arts for dashboard")

		// Create error page data using middleware-provided context
		pageData := templates.NewPageDataFromRequest(r, "Dashboard - Error", "dashboard").
			WithError("Error fetching arts. Please try again.")

		err = pages.DashboardPage(pageData).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
		return
	}

	// Create dashboard-specific data
	dashboardData := &templates.DashboardPageData{
		Arts: arts.GetArts(),
		Sort: sort,
		Dir:  dir,
	}

	// Create page data using middleware-provided context
	pageData := templates.NewPageDataFromRequest(r, "Dashboard", "dashboard").
		WithData(dashboardData)

	// Render the dashboard page
	err = pages.DashboardPage(pageData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render dashboard")
	}
}

// LoginPage renders the Firebase login page
func (h *PageHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if authenticated (to redirect if already logged in)
	user, _ := middleware.UserFromContext(r.Context())

	// If user is already authenticated, redirect to dashboard
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	// Create Firebase configuration based on environment
	firebaseConfig := h.getFirebaseConfig()

	// Create page data for login page
	pageData := templates.NewPageData("Login - ThreadArt", "login").
		WithUser(user).
		WithFirebaseConfig(firebaseConfig)

	// Render the login page template
	err := pages.LoginPage(pageData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render login page")
	}
}

// SignupPage renders the Firebase signup page
func (h *PageHandler) SignupPage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if authenticated (to redirect if already logged in)
	user, _ := middleware.UserFromContext(r.Context())

	// If user is already authenticated, redirect to dashboard
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	// Create Firebase configuration based on environment
	firebaseConfig := h.getFirebaseConfig()

	// Create page data for signup page
	pageData := templates.NewPageData("Sign Up - ThreadArt", "signup").
		WithUser(user).
		WithFirebaseConfig(firebaseConfig)

	// Render the signup page template
	err := pages.SignupPage(pageData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render signup page")
	}
}

// getFirebaseConfig returns Firebase configuration based on environment
func (h *PageHandler) getFirebaseConfig() *types.FirebaseConfig {
	// Use the centralized configuration method
	coreConfig := h.config.GetFirebaseConfigForFrontend()

	// Convert from core config to client types
	return &types.FirebaseConfig{
		ProjectID:    coreConfig.ProjectID,
		APIKey:       coreConfig.APIKey,
		AuthDomain:   coreConfig.AuthDomain,
		EmulatorHost: coreConfig.EmulatorHost,
		EmulatorUI:   coreConfig.EmulatorUI,
		IsEmulator:   coreConfig.IsEmulator,
	}
}
