package router

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/web/client"
	"github.com/Damione1/thread-art-generator/web/handlers"
	"github.com/Damione1/thread-art-generator/web/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates a new router
func NewRouter(grpcClient *client.GrpcClient) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Public routes
	r.Group(func(r chi.Router) {
		// Apply TryAuth middleware to all public routes
		r.Use(middleware.TryAuth(grpcClient))

		// Home page
		r.Get("/", handlers.HomeHandler(grpcClient))

		// Auth routes - these should redirect to dashboard if already logged in
		r.Get("/login", handlers.LoginHandler(grpcClient))
		r.Post("/login", handlers.LoginHandler(grpcClient))
		r.Get("/register", handlers.RegisterHandler(grpcClient))
		r.Post("/register", handlers.RegisterHandler(grpcClient))
		r.Get("/logout", handlers.LogoutHandler(grpcClient))

		// Email validation routes
		r.Get("/validate-email", handlers.ValidateEmailHandler(grpcClient))
		r.Post("/validate-email", handlers.ValidateEmailHandler(grpcClient))
		r.Post("/resend-validation", handlers.ResendValidationHandler(grpcClient))
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(grpcClient))

		// Dashboard
		r.Get("/dashboard", handlers.DashboardHandler(grpcClient))

		// Profile
		r.Get("/profile", handlers.ProfileHandler(grpcClient))
		r.Post("/profile", handlers.ProfileHandler(grpcClient))
	})

	// Static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return r
}
