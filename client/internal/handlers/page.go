package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	pages "github.com/Damione1/thread-art-generator/client/internal/templates/pages"
	"github.com/Damione1/thread-art-generator/client/internal/types"
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

	// Create page data using the new structure
	pageData := templates.NewPageData("ThreadArt - Create Beautiful Thread Art", "home").
		WithUser(user)

	// Render the home page template
	err := pages.HomePage(pageData).Render(r.Context(), w)
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
	arts, err := h.generatorService.ListArts(r.Context(), user.ID, 10, "", sort, dir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch arts for dashboard")
		
		// Create error page data
		pageData := templates.NewPageData("Dashboard - Error", "dashboard").
			WithUser(user).
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

	// Create page data with dashboard-specific data
	pageData := templates.NewPageData("Dashboard", "dashboard").
		WithUser(user).
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
	// Check if we're in emulator mode by checking environment variables
	emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	environment := os.Getenv("ENVIRONMENT")
	
	// Use emulator if explicitly set or in development environment
	isEmulator := emulatorHost != "" || environment == "development"
	
	if isEmulator {
		// For emulator, always use localhost for browser access
		// The browser needs to connect directly to localhost, not through Docker networking
		return &types.FirebaseConfig{
			ProjectID:    "demo-thread-art-generator",
			APIKey:       "demo-api-key", // Emulator doesn't need real API key
			AuthDomain:   "demo-thread-art-generator.firebaseapp.com",
			EmulatorHost: "localhost:9099", // Always use localhost for browser
			EmulatorUI:   "localhost:4000",
			IsEmulator:   true,
		}
	}
	
	// Production configuration with fallbacks and validation
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	webAPIKey := os.Getenv("FIREBASE_WEB_API_KEY") 
	authDomain := os.Getenv("FIREBASE_AUTH_DOMAIN")
	
	// Generate authDomain from projectID if not provided
	if authDomain == "" && projectID != "" {
		authDomain = fmt.Sprintf("%s.firebaseapp.com", projectID)
	}
	
	return &types.FirebaseConfig{
		ProjectID:  projectID,
		APIKey:     webAPIKey,
		AuthDomain: authDomain,
		IsEmulator: false,
	}
}
