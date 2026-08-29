package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
	"github.com/Damione1/thread-art-generator/client/internal/templates"
	pages "github.com/Damione1/thread-art-generator/client/internal/templates/pages"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/rs/zerolog/log"
)

// PageHandler handles rendering the main application pages
type PageHandler struct {
	generatorService *services.GeneratorService
}

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

	err := pages.HomePage(pageData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render home page")
	}
}

// DashboardPage renders the dashboard page (protected)
func (h *PageHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	// Get internal user ID by calling GetCurrentUser API
	currentUser, err := h.generatorService.GetCurrentUser(r.Context(), r)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("Failed to get current user for DashboardPage")

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
	arts, err := h.generatorService.ListArts(r.Context(), internalUserID, 10, "", sort+" "+dir)
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

// LoginPage renders the login page
func (h *PageHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if authenticated (to redirect if already logged in)
	user, _ := middleware.UserFromContext(r.Context())

	// If user is already authenticated, redirect to dashboard
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	pageData := templates.NewPageData("Login - ThreadArt", "login").
		WithUser(user)

	err := pages.LoginPage(pageData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render login page")
	}
}

func (h *PageHandler) SignupPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	pageData := templates.NewPageData("Sign Up - ThreadArt", "signup").
		WithUser(user)

	err := pages.SignupPage(pageData).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render signup page")
	}
}

func (h *PageHandler) ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}
	pageData := templates.NewPageData("Forgot password - ThreadArt", "forgot")
	if err := pages.ForgotPasswordPage(pageData).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render forgot password page")
	}
}

func (h *PageHandler) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	if user != nil {
		http.Redirect(w, r, "/settings", http.StatusTemporaryRedirect)
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
		return
	}
	pageData := templates.NewPageData("Reset password - ThreadArt", "reset").
		WithMeta("token", token)
	if err := pages.ResetPasswordPage(pageData).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render reset password page")
	}
}

func (h *PageHandler) CheckEmailPage(w http.ResponseWriter, r *http.Request) {
	pageData := templates.NewPageData("Check your email - ThreadArt", "check-email")
	if err := pages.CheckEmailPage(pageData).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Failed to render check email page")
	}
}
